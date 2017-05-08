// Copyright (C) 2016 Space Monkey, Inc.
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

package monkit

import (
	"sort"
	"sync"
	"time"

	"github.com/spacemonkeygo/monotime"
)

// Timer is a threadsafe convenience wrapper around a DurationDist. You should
// construct with NewTimer(), though the expected usage is from a Scope like
// so:
//
//   var mon = monkit.Package()
//
//   func MyFunc() {
//     ...
//     timer := mon.Timer("event")
//     // perform event
//     timer.Stop()
//     ...
//   }
//
// Timers implement StatSource.
type Timer struct {
	mtx    sync.Mutex
	times  *DurationDist
	splits map[string]*DurationDist
}

// NewTimer constructs a new Timer.
func NewTimer() *Timer {
	return &Timer{times: NewDurationDist()}
}

// Start constructs a RunningTimer
func (t *Timer) Start() *RunningTimer {
	return &RunningTimer{
		start: monotime.Monotonic(),
		t:     t}
}

// RunningTimer should be constructed from a Timer.
type RunningTimer struct {
	start   time.Duration
	t       *Timer
	stopped bool
}

// Split constructs new named child DurationDists and adds the current elapsed
// time to them.
func (r *RunningTimer) Split(name string) time.Duration {
	elapsed := r.Elapsed()
	r.t.mtx.Lock()
	if !r.stopped {
		if r.t.splits == nil {
			r.t.splits = map[string]*DurationDist{}
		}
		if r.t.splits[name] == nil {
			r.t.splits[name] = NewDurationDist()
		}
		r.t.splits[name].Insert(elapsed)
	}
	r.t.mtx.Unlock()
	return elapsed
}

// Elapsed just returns the amount of time since the timer started
func (r *RunningTimer) Elapsed() time.Duration {
	return monotime.Monotonic() - r.start
}

// Stop stops the timer, adds the duration to the statistics information, and
// returns the elapsed time.
func (r *RunningTimer) Stop() time.Duration {
	elapsed := r.Elapsed()
	r.t.mtx.Lock()
	if !r.stopped {
		r.t.times.Insert(elapsed)
		r.stopped = true
	}
	r.t.mtx.Unlock()
	return elapsed
}

// Values returns the main timer values
func (t *Timer) Values() *DurationDist {
	t.mtx.Lock()
	rv := t.times.Copy()
	t.mtx.Unlock()
	return rv
}

// SplitValues returns the timer values for the named split
func (t *Timer) SplitValues(name string) (rv *DurationDist) {
	t.mtx.Lock()
	if t.splits != nil {
		found := t.splits[name]
		if found != nil {
			rv = found.Copy()
		}
	}
	t.mtx.Unlock()
	if rv == nil {
		rv = NewDurationDist()
	}
	return rv
}

// Stats implements the StatSource interface
func (t *Timer) Stats(cb func(name string, val float64)) {
	t.mtx.Lock()
	times := t.times.Copy()
	splits := make(map[string]*DurationDist, len(t.splits))
	for name, dist := range t.splits {
		splits[name] = dist.Copy()
	}
	t.mtx.Unlock()

	call := func(prefix string, times *DurationDist) {
		times.Stats(func(name string, val float64) {
			cb(prefix+name, val)
		})
	}

	call("", times)

	splitNames := make([]string, 0, len(splits))
	for name := range splits {
		splitNames = append(splitNames, name)
	}
	sort.Strings(splitNames)

	for _, name := range splitNames {
		call(name+" - ", splits[name])
	}
}
