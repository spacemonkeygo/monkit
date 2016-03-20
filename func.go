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
	success      int64
	successTimes intDist
	errors       map[string]int64
	panics       int64
	failureTimes intDist
}

func newFunc(s *Scope, name string) *Func {
	return &Func{
		id:           NewId(),
		scope:        s,
		name:         name,
		errors:       make(map[string]int64),
		successTimes: newIntDist(),
		failureTimes: newIntDist(),
	}
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
		f.failureTimes.Insert(int64(duration))
		f.parentsAndMutex.Unlock()
		return
	}
	if err == nil {
		f.success += 1
		f.successTimes.Insert(int64(duration))
		f.parentsAndMutex.Unlock()
		return
	}
	f.failureTimes.Insert(int64(duration))
	f.errors[errors.GetClass(err).String()] += 1
	f.parentsAndMutex.Unlock()
}

func (f *Func) Current() int64   { return atomic.LoadInt64(&f.current) }
func (f *Func) Highwater() int64 { return atomic.LoadInt64(&f.highwater) }

func (f *Func) Success() (rv int64) {
	f.parentsAndMutex.Lock()
	rv = f.success
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
	success, panics := f.success, f.panics
	errs := make(map[string]int64, len(f.errors))
	for errname, count := range f.errors {
		errs[errname] = count
	}
	s_min, s_avg, s_max, s_recent := f.successTimes.Stats()
	f_min, f_avg, f_max, f_recent := f.failureTimes.Stats()
	f.parentsAndMutex.Unlock()

	cb("success", float64(success))
	for errname, count := range errs {
		cb(fmt.Sprintf("error %s", errname), float64(count))
	}
	cb("panics", float64(panics))
	cb("success times min", time.Duration(s_min).Seconds())
	cb("success times avg", time.Duration(s_avg).Seconds())
	cb("success times max", time.Duration(s_max).Seconds())
	cb("success times recent", time.Duration(s_recent).Seconds())
	cb("failure times min", time.Duration(f_min).Seconds())
	cb("failure times avg", time.Duration(f_avg).Seconds())
	cb("failure times max", time.Duration(f_max).Seconds())
	cb("failure times recent", time.Duration(f_recent).Seconds())
}

func (f *Func) SuccessTimeAverage() (rv time.Duration) {
	f.parentsAndMutex.Lock()
	rv = time.Duration(f.successTimes.Average())
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) SuccessTimeRecent() (rv time.Duration) {
	f.parentsAndMutex.Lock()
	rv = time.Duration(f.successTimes.Recent())
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) SuccessTimeQuantile(quantile float64) (rv time.Duration) {
	f.parentsAndMutex.Lock()
	rv = time.Duration(f.successTimes.Query(quantile))
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) FailureTimeAverage() (rv time.Duration) {
	f.parentsAndMutex.Lock()
	rv = time.Duration(f.failureTimes.Average())
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) FailureTimeRecent() (rv time.Duration) {
	f.parentsAndMutex.Lock()
	rv = time.Duration(f.failureTimes.Recent())
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) FailureTimeQuantile(quantile float64) (rv time.Duration) {
	f.parentsAndMutex.Lock()
	rv = time.Duration(f.failureTimes.Query(quantile))
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) Id() int64     { return f.id }
func (f *Func) Scope() *Scope { return f.scope }
