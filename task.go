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

	"golang.org/x/net/context"
)

type taskKey int

const (
	taskGetFunc taskKey = 0
)

type LazyTask func(ctx *context.Context, args ...interface{}) func(*error)

func (f LazyTask) Func() (out *Func) {
	// we're doing crazy things to make a function have methods that do other
	// things with internal state. basically, we have a secret argument we can
	// pass to the function that is only checked if ctx is nil (which it should
	// never be) that controls what other behavior we want.
	// in this case, if arg[0] is taskGetFunc, then f will place the func in the
	// out location.
	f(nil, taskGetFunc, &out)
	return out
}

func taskArgs(f *Func, args []interface{}) bool {
	// this function essentially does method dispatch for LazyTasks. returns true
	// if a method got dispatched and normal behavior should be aborted
	if len(args) != 2 {
		return false
	}
	val, ok := args[0].(taskKey)
	if !ok {
		return false
	}
	switch val {
	case taskGetFunc:
		*(args[1].(**Func)) = f
		return true
	}
	return false
}

func (s *Scope) Task() LazyTask {
	var initOnce sync.Once
	var f *Func
	init := func() {
		f = s.FuncNamed(callerFunc(3))
	}
	return LazyTask(func(ctx *context.Context,
		args ...interface{}) func(*error) {
		initOnce.Do(init)
		if ctx == nil && taskArgs(f, args) {
			return nil
		}
		s, exit := newSpan(*ctx, f, args, NewId(), nil)
		*ctx = s
		return exit
	})
}

func (s *Scope) TaskNamed(name string) LazyTask {
	return s.FuncNamed(name).Task
}

func (f *Func) Task(ctx *context.Context, args ...interface{}) func(*error) {
	if ctx == nil && taskArgs(f, args) {
		return nil
	}
	s, exit := newSpan(*ctx, f, args, NewId(), nil)
	*ctx = s
	return exit
}

func (f *Func) RemoteTrace(ctx *context.Context, spanId int64, trace *Trace,
	args ...interface{}) func(*error) {
	if trace != nil {
		f.scope.r.observeTrace(trace)
	}
	s, exit := newSpan(*ctx, f, args, spanId, trace)
	*ctx = s
	return exit
}
