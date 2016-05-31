// Copyright (C) 2015 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package present

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/net/context"
	"gopkg.in/spacemonkeygo/monkit.v2"
)

// FinishedSpan is a Span that has completed and contains information about
// how it finished.
type FinishedSpan struct {
	Span     *monkit.Span
	Err      error
	Panicked bool
	Finish   time.Time
}

// WatchForSpans will watch for spans that 'matcher' returns true for. As soon
// as a trace generates a matched span, all spans from that trace that finish
// from that point on are collected until the matching span completes. All
// collected spans are returned.
// To cancel this operation, simply cancel the ctx argument.
// There is a small but permanent amount of overhead added by this function to
// every trace that is started while this function is running. This only really
// affects long-running traces.
func WatchForSpans(ctx context.Context, r *monkit.Registry,
	matcher func(s *monkit.Span) bool) (spans []*FinishedSpan, err error) {
	collector := NewSpanCollector(matcher)
	defer collector.Stop()
	canceler := r.ObserveTraces(func(t *monkit.Trace) {
		t.ObserveSpans(collector)
	})
	defer canceler()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-collector.Done():
		return collector.Spans(), nil
	}
}

// CollectSpans is kind of like WatchForSpans, except that it uses the current
// span to figure out which trace to collect. It calls work(), then collects
// from the current trace until work() returns. CollectSpans won't work unless
// some ancestor function is also monitored and has modified the ctx.
func CollectSpans(ctx context.Context, work func(ctx context.Context)) (
	spans []*FinishedSpan) {
	s := monkit.SpanFromCtx(ctx)
	if s == nil {
		work(ctx)
		return nil
	}
	collector := NewSpanCollector(nil)
	defer collector.Stop()
	s.Trace().ObserveSpans(collector)
	f := s.Func()
	newF := f.Scope().FuncNamed(fmt.Sprintf("%s-TRACED", f.ShortName()))
	func() {
		defer newF.Task(&ctx)(nil)
		collector.ForceStart(monkit.SpanFromCtx(ctx))
		work(ctx)
	}()
	return collector.Spans()
}

// SpanCollector implements the SpanObserver interface. It stores all Spans
// observed after it starts collecting, typically when matcher returns true.
type SpanCollector struct {
	// sync/atomic
	check unsafe.Pointer

	// construction
	matcher func(s *monkit.Span) bool
	done    chan struct{}

	// mtx protected
	mtx           sync.Mutex
	root          *FinishedSpan
	spansByParent map[*monkit.Span][]*FinishedSpan
}

// NewSpanCollector takes a matcher that will return true when a span is found
// that should start collection. matcher can be nil if you intend to use
// ForceStart instead.
func NewSpanCollector(matcher func(s *monkit.Span) bool) (
	rv *SpanCollector) {
	if matcher == nil {
		matcher = func(*monkit.Span) bool { return false }
	}
	return &SpanCollector{
		matcher:       matcher,
		done:          make(chan struct{}),
		spansByParent: map[*monkit.Span][]*FinishedSpan{},
	}
}

// Done returns a channel that's closed when the SpanCollector has collected
// everything it cares about.
func (c *SpanCollector) Done() <-chan struct{} {
	return c.done
}

type nonce struct {
	int
}

var (
	donePointer = unsafe.Pointer(&nonce{})
)

// ForceStart starts the span collector collecting spans, stopping when endSpan
// finishes. This is typically only used if matcher was nil at construction.
func (c *SpanCollector) ForceStart(endSpan *monkit.Span) {
	atomic.CompareAndSwapPointer(&c.check, nil, unsafe.Pointer(endSpan))
}

// Start is to implement the monkit.SpanObserver interface. Start gets called
// whenever a Span starts.
func (c *SpanCollector) Start(s *monkit.Span) {
	if atomic.LoadPointer(&c.check) != nil || !c.matcher(s) {
		return
	}
	atomic.CompareAndSwapPointer(&c.check, nil, unsafe.Pointer(s))
}

// Finish is to implement the monkit.SpanObserver interface. Finish gets
// called whenever a Span finishes.
func (c *SpanCollector) Finish(s *monkit.Span, err error, panicked bool,
	finish time.Time) {
	existing := atomic.LoadPointer(&c.check)
	if existing == donePointer || existing == nil ||
		((*monkit.Span)(existing)).Trace() != s.Trace() {
		return
	}
	fs := &FinishedSpan{Span: s, Err: err, Panicked: panicked, Finish: finish}
	c.mtx.Lock()
	if c.root != nil {
		c.mtx.Unlock()
		return
	}
	if (*monkit.Span)(existing) == s {
		c.root = fs
		c.mtx.Unlock()
		c.Stop()
	} else {
		c.spansByParent[s.Parent()] = append(c.spansByParent[s.Parent()], fs)
		c.mtx.Unlock()
	}
}

// Stop stops the SpanCollector from collecting.
func (c *SpanCollector) Stop() {
	if atomic.SwapPointer(&c.check, donePointer) != donePointer {
		close(c.done)
	}
}

// Spans returns all spans found, rooted from the Span the collector started
// on.
func (c *SpanCollector) Spans() (spans []*FinishedSpan) {
	var walkSpans func(s *FinishedSpan)
	walkSpans = func(s *FinishedSpan) {
		spans = append(spans, s)
		for _, child := range c.spansByParent[s.Span] {
			walkSpans(child)
		}
	}
	c.mtx.Lock()
	if c.root != nil {
		walkSpans(c.root)
	}
	c.mtx.Unlock()
	return spans
}
