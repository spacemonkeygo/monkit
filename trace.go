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

type spanWatcherRef struct {
	watcher func(s *Span, err error, panicked bool, finish time.Time)
}

type Trace struct {
	// sync/atomic things
	spanWatcher unsafe.Pointer

	// immutable things from construction
	id int64

	// protected by mtx
	mtx  sync.Mutex
	vals map[interface{}]interface{}
}

func NewTrace(id int64) *Trace {
	return &Trace{id: id}
}

func (t *Trace) observe(s *Span, err error, panicked bool, finish time.Time) {
	watcher := (*spanWatcherRef)(atomic.LoadPointer(&t.spanWatcher))
	if watcher != nil {
		watcher.watcher(s, err, panicked, finish)
	}
}

func (t *Trace) ObserveSpans(
	cb func(s *Span, err error, panicked bool, finish time.Time)) {
	if cb == nil {
		return
	}
	for {
		existing := (*spanWatcherRef)(atomic.LoadPointer(&t.spanWatcher))
		if existing == nil {
			if atomic.CompareAndSwapPointer(&t.spanWatcher, nil,
				unsafe.Pointer(&spanWatcherRef{watcher: cb})) {
				break
			}
		} else {
			other_cb := existing.watcher
			if atomic.CompareAndSwapPointer(&t.spanWatcher,
				unsafe.Pointer(existing),
				unsafe.Pointer(&spanWatcherRef{watcher: func(
					s *Span, err error, panicked bool, finish time.Time) {
					other_cb(s, err, panicked, finish)
					cb(s, err, panicked, finish)
				}})) {
				break
			}
		}
	}
}

func (t *Trace) Id() int64 { return t.id }

func (t *Trace) Get(key interface{}) (val interface{}) {
	t.mtx.Lock()
	if t.vals != nil {
		val = t.vals[key]
	}
	t.mtx.Unlock()
	return val
}

func (t *Trace) Set(key, val interface{}) {
	t.mtx.Lock()
	if t.vals == nil {
		t.vals = map[interface{}]interface{}{key: val}
	} else {
		t.vals[key] = val
	}
	t.mtx.Unlock()
}
