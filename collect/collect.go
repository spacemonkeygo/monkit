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

package collect

import (
	"sort"
	"sync"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
)

// FinishedSpan is a Span that has completed and contains information about
// how it finished.
type FinishedSpan struct {
	Span     *monkit.Span
	Err      error
	Panicked bool
	Finish   time.Time
}

// SpanCollector implements the SpanObserver interface. It stores all Spans
// observed after it starts collecting, typically when matcher returns true.
type SpanCollector struct {
	// sync/atomic
	check *monkit.Span

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

var donePointer = new(monkit.Span)

// ForceStart starts the span collector collecting spans, stopping when endSpan
// finishes. This is typically only used if matcher was nil at construction.
func (c *SpanCollector) ForceStart(endSpan *monkit.Span) {
	compareAndSwapSpan(&c.check, nil, endSpan)
}

// Start is to implement the monkit.SpanObserver interface. Start gets called
// whenever a Span starts.
func (c *SpanCollector) Start(s *monkit.Span) {
	if loadSpan(&c.check) != nil || !c.matcher(s) {
		return
	}
	compareAndSwapSpan(&c.check, nil, s)
}

// Finish is to implement the monkit.SpanObserver interface. Finish gets
// called whenever a Span finishes.
func (c *SpanCollector) Finish(s *monkit.Span, err error, panicked bool,
	finish time.Time) {
	existing := loadSpan(&c.check)
	if existing == donePointer || existing == nil ||
		existing.Trace() != s.Trace() {
		return
	}
	fs := &FinishedSpan{Span: s, Err: err, Panicked: panicked, Finish: finish}
	c.mtx.Lock()
	if c.root != nil {
		c.mtx.Unlock()
		return
	}
	if existing == s {
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
	if swapSpan(&c.check, donePointer) != donePointer {
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

// StartTimeSorter assists with sorting a slice of FinishedSpans by start time.
type StartTimeSorter []*FinishedSpan

func (s StartTimeSorter) Len() int      { return len(s) }
func (s StartTimeSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s StartTimeSorter) Less(i, j int) bool {
	return s[i].Span.Start().UnixNano() < s[j].Span.Start().UnixNano()
}

func (s StartTimeSorter) Sort() { sort.Sort(s) }
