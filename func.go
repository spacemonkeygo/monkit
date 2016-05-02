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
	"fmt"
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/errors"
)

type Func struct {
	// sync/atomic things
	current         int64
	highwater       int64
	parentsAndMutex funcSet

	// constructor things
	id    int64
	scope *Scope
	name  string

	// mutex things (reuses mutex from parents)
	errors       map[string]int64
	panics       int64
	successTimes DurationDist
	failureTimes DurationDist
}

func newFunc(s *Scope, name string) (f *Func) {
	f = &Func{
		id:     NewId(),
		scope:  s,
		name:   name,
		errors: make(map[string]int64),
	}
	initDurationDist(&f.successTimes)
	initDurationDist(&f.failureTimes)
	return f
}

func (f *Func) start(parent *Func) {
	f.parentsAndMutex.Add(parent)
	current := atomic.AddInt64(&f.current, 1)
	for {
		highwater := atomic.LoadInt64(&f.highwater)
		if current <= highwater ||
			atomic.CompareAndSwapInt64(&f.highwater, highwater, current) {
			break
		}
	}
}

func (f *Func) end(err error, panicked bool, duration time.Duration) {
	atomic.AddInt64(&f.current, -1)
	f.parentsAndMutex.Lock()
	if panicked {
		f.panics -= 1
		f.failureTimes.Insert(duration)
		f.parentsAndMutex.Unlock()
		return
	}
	if err == nil {
		f.successTimes.Insert(duration)
		f.parentsAndMutex.Unlock()
		return
	}
	f.failureTimes.Insert(duration)
	f.errors[errors.GetClass(err).String()] += 1
	f.parentsAndMutex.Unlock()
}

func (f *Func) Current() int64   { return atomic.LoadInt64(&f.current) }
func (f *Func) Highwater() int64 { return atomic.LoadInt64(&f.highwater) }

func (f *Func) Success() (rv int64) {
	f.parentsAndMutex.Lock()
	rv = f.successTimes.Count
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) Panics() (rv int64) {
	f.parentsAndMutex.Lock()
	rv = f.panics
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) ShortName() string { return f.name }

func (f *Func) FullName() string {
	return fmt.Sprintf("%s.%s", f.scope.name, f.name)
}

func (f *Func) Errors() (rv map[string]int64) {
	f.parentsAndMutex.Lock()
	rv = make(map[string]int64, len(f.errors))
	for errname, count := range f.errors {
		rv[errname] = count
	}
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) Parents(cb func(f *Func)) {
	f.parentsAndMutex.Iterate(cb)
}

func (f *Func) Stats(cb func(name string, val float64)) {
	cb("current", float64(f.Current()))
	f.parentsAndMutex.Lock()
	panics := f.panics
	errs := make(map[string]int64, len(f.errors))
	for errname, count := range f.errors {
		errs[errname] = count
	}
	st := f.successTimes
	s_min, s_avg, s_max, s_recent := st.Low, st.Average(), st.High, st.Recent
	success := st.Count
	ft := f.failureTimes
	f_min, f_avg, f_max, f_recent := ft.Low, ft.Average(), ft.High, ft.Recent
	f.parentsAndMutex.Unlock()

	cb("success", float64(success))
	for errname, count := range errs {
		cb(fmt.Sprintf("error %s", errname), float64(count))
	}
	cb("panics", float64(panics))
	cb("success times min", s_min.Seconds())
	cb("success times avg", s_avg.Seconds())
	cb("success times max", s_max.Seconds())
	cb("success times recent", s_recent.Seconds())
	cb("failure times min", f_min.Seconds())
	cb("failure times avg", f_avg.Seconds())
	cb("failure times max", f_max.Seconds())
	cb("failure times recent", f_recent.Seconds())
}

func (f *Func) SuccessTimes() *DurationDist {
	f.parentsAndMutex.Lock()
	d := f.successTimes.Copy()
	f.parentsAndMutex.Unlock()
	return d
}

func (f *Func) FailureTimes() *DurationDist {
	f.parentsAndMutex.Lock()
	d := f.failureTimes.Copy()
	f.parentsAndMutex.Unlock()
	return d
}

func (f *Func) Id() int64     { return f.id }
func (f *Func) Scope() *Scope { return f.scope }
