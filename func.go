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
	parentsAndMutex funcSet

	// constructor things
	id    int64
	scope *Scope
	name  string

	// mutex things (reuses mutex from parents)
	success      int64
	successTimes dist
	errors       map[string]int64
	panics       int64
	failureTimes dist
}

func newFunc(s *Scope, name string) *Func {
	return &Func{
		id:           NewId(),
		scope:        s,
		name:         name,
		errors:       make(map[string]int64),
		successTimes: newDist(),
		failureTimes: newDist(),
	}
}

func (f *Func) start(parent *Func) {
	f.parentsAndMutex.Add(parent)
	atomic.AddInt64(&f.current, 1)
}

func (f *Func) end(err error, panicked bool, duration time.Duration) {
	dur := duration.Seconds()
	atomic.AddInt64(&f.current, -1)
	f.parentsAndMutex.Lock()
	if panicked {
		f.panics -= 1
		f.failureTimes.Insert(dur)
		f.parentsAndMutex.Unlock()
		return
	}
	if err == nil {
		f.success += 1
		f.successTimes.Insert(dur)
		f.parentsAndMutex.Unlock()
		return
	}
	f.failureTimes.Insert(dur)
	f.errors[errors.GetClass(err).String()] += 1
	f.parentsAndMutex.Unlock()
}

func (f *Func) Current() int64 { return atomic.LoadInt64(&f.current) }

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
	s_min, s_med, s_max, s_recent := f.successTimes.Stats()
	f_min, f_med, f_max, f_recent := f.failureTimes.Stats()
	f.parentsAndMutex.Unlock()

	cb("success", float64(success))
	for errname, count := range errs {
		cb(fmt.Sprintf("error %s", errname), float64(count))
	}
	cb("panics", float64(panics))
	cb("success times min", s_min)
	cb("success times med", s_med)
	cb("success times max", s_max)
	cb("success times recent", s_recent)
	cb("failure times min", f_min)
	cb("failure times med", f_med)
	cb("failure times max", f_max)
	cb("failure times recent", f_recent)
}

func (f *Func) SuccessTimeQuantile(quantile float64) (rv float64) {
	f.parentsAndMutex.Lock()
	rv = f.successTimes.Query(quantile)
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) FailureTimeQuantile(quantile float64) (rv float64) {
	f.parentsAndMutex.Lock()
	rv = f.failureTimes.Query(quantile)
	f.parentsAndMutex.Unlock()
	return rv
}

func (f *Func) Id() int64     { return f.id }
func (f *Func) Scope() *Scope { return f.scope }
