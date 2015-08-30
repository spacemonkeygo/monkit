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
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/net/context"
	"gopkg.in/spacemonkeygo/monitor.v2"
)

// WatchForSpans will watch for traces that cross functions that 'matcher'
// returns true for. As soon as a trace generates a span for a matched
// function, all spans from that trace that finish from that point on are
// collected until the matching function span completes. Those spans are
// returned.
// To cancel this operation, simply cancel the ctx argument.
// There is a small but permanent amount of overhead added by this function to
// every trace that is started while this function is running. This only really
// affects long-running traces.
func WatchForSpans(ctx context.Context, r *monitor.Registry,
	matcher func(f *monitor.Func) bool,
	observe func(s *monitor.Span, err error, panicked bool, finish time.Time)) {
	collector := newSpanCollector(matcher, observe)
	canceler := r.ObserveTraces(func(t *monitor.Trace) {
		t.ObserveSpans(collector)
	})
	defer canceler()

	select {
	case <-ctx.Done():
		collector.Stop()
	case <-collector.Done():
	}
}

type spanCollector struct {
	// sync/atomic
	span unsafe.Pointer

	// construction
	matcher func(f *monitor.Func) bool
	observe func(s *monitor.Span, err error, panicked bool, finish time.Time)
	done    chan struct{}
}

func newSpanCollector(matcher func(f *monitor.Func) bool,
	observe func(s *monitor.Span, err error, panicked bool, finish time.Time)) (
	rv *spanCollector) {
	return &spanCollector{
		matcher: matcher,
		observe: observe,
		done:    make(chan struct{})}
}

func (c *spanCollector) Done() <-chan struct{} {
	return c.done
}

type nonce struct {
	int
}

var (
	donePointer = unsafe.Pointer(&nonce{})
)

func (c *spanCollector) Start(s *monitor.Span) {
	if atomic.LoadPointer(&c.span) != nil || !c.matcher(s.Func()) {
		return
	}
	atomic.CompareAndSwapPointer(&c.span, nil, unsafe.Pointer(s))
}

func (c *spanCollector) Finish(s *monitor.Span, err error, panicked bool,
	finish time.Time) {
	existing := atomic.LoadPointer(&c.span)
	if existing == donePointer || existing == nil ||
		((*monitor.Span)(existing)).Trace() != s.Trace() {
		return
	}
	c.observe(s, err, panicked, finish)
	if (*monitor.Span)(existing) == s {
		c.Stop()
	}
}

func (c *spanCollector) Stop() {
	atomic.StorePointer(&c.span, donePointer)
	close(c.done)
}
