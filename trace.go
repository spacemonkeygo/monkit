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

package monitor

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// SpanObserver is the interface plugins must implement if they want to observe
// all spans on a given trace as they happen.
type SpanObserver interface {
	// Start is called when a Span starts
	Start(s *Span)

	// Finish is called when a Span finishes, along with an error if any, whether
	// or not it panicked, and what time it finished.
	Finish(s *Span, err error, panicked bool, finish time.Time)
}

type spanObserverTuple struct{ first, second SpanObserver }

func (l spanObserverTuple) Start(s *Span) {
	l.first.Start(s)
	l.second.Start(s)
}

func (l spanObserverTuple) Finish(s *Span, err error, panicked bool,
	finish time.Time) {
	l.first.Finish(s, err, panicked, finish)
	l.second.Finish(s, err, panicked, finish)
}

type spanObserverRef struct {
	observer SpanObserver
}

// Trace represents a 'trace' of execution. A 'trace' is the collection of all
// of the 'spans' kicked off from the same root execution context. A trace is
// a concurrency-supporting analog of a stack trace, where a span is somewhat
// like a stack frame.
type Trace struct {
	// sync/atomic things
	spanObserver unsafe.Pointer

	// immutable things from construction
	id int64

	// protected by mtx
	mtx  sync.Mutex
	vals map[interface{}]interface{}
}

// NewTrace creates a new Trace.
func NewTrace(id int64) *Trace {
	return &Trace{id: id}
}

func (t *Trace) getObserver() SpanObserver {
	observer := (*spanObserverRef)(atomic.LoadPointer(&t.spanObserver))
	if observer == nil {
		return nil
	}
	return observer.observer
}

// ObserveSpans lets you register a SpanObserver for all future Spans on the
// Trace.
func (t *Trace) ObserveSpans(observer SpanObserver) {
	for {
		existing := (*spanObserverRef)(atomic.LoadPointer(&t.spanObserver))
		if existing == nil {
			if atomic.CompareAndSwapPointer(&t.spanObserver, nil,
				unsafe.Pointer(&spanObserverRef{observer: observer})) {
				break
			}
		} else {
			otherObserver := existing.observer
			if atomic.CompareAndSwapPointer(&t.spanObserver,
				unsafe.Pointer(existing),
				unsafe.Pointer(&spanObserverRef{observer: spanObserverTuple{
					first: otherObserver, second: observer}})) {
				break
			}
		}
	}
}

// Id returns the id of the Trace
func (t *Trace) Id() int64 { return t.id }

// Get returns a value associated with a key on a trace. See Set.
func (t *Trace) Get(key interface{}) (val interface{}) {
	t.mtx.Lock()
	if t.vals != nil {
		val = t.vals[key]
	}
	t.mtx.Unlock()
	return val
}

// Set sets a value associated with a key on a trace. See Get.
func (t *Trace) Set(key, val interface{}) {
	t.mtx.Lock()
	if t.vals == nil {
		t.vals = map[interface{}]interface{}{key: val}
	} else {
		t.vals[key] = val
	}
	t.mtx.Unlock()
}
