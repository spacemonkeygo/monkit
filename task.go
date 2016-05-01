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
	"time"

	"golang.org/x/net/context"
)

type taskKey int

const taskGetFunc taskKey = 0

type lazyTaskSecretT struct{}

func (*lazyTaskSecretT) Value(key interface{}) interface{} { return nil }
func (*lazyTaskSecretT) Done() <-chan struct{}             { return nil }
func (*lazyTaskSecretT) Err() error                        { return nil }
func (*lazyTaskSecretT) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

var lazyTaskSecret context.Context = &lazyTaskSecretT{}

type LazyTask func(ctx *context.Context, args ...interface{}) func(*error)

func (f LazyTask) Func() (out *Func) {
	// we're doing crazy things to make a function have methods that do other
	// things with internal state. basically, we have a secret argument we can
	// pass to the function that is only checked if ctx is lazyTaskSecret (
	// which it should never be) that controls what other behavior we want.
	// in this case, if arg[0] is taskGetFunc, then f will place the func in the
	// out location.
	// since someone can cast any function of this signature to a lazy task,
	// let's make sure we got roughly expected behavior and panic otherwise
	if f(&lazyTaskSecret, taskGetFunc, &out) != nil || out == nil {
		panic("Func() called on a non-LazyTask function")
	}
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
		if ctx == nil {
			ctx = emptyCtx()
		} else if ctx == &lazyTaskSecret && taskArgs(f, args) {
			return nil
		}
		initOnce.Do(init)
		s, exit := newSpan(*ctx, f, args, NewId(), nil)
		*ctx = s
		return exit
	})
}

func (s *Scope) TaskNamed(name string) LazyTask {
	return s.FuncNamed(name).Task
}

func (f *Func) Task(ctx *context.Context, args ...interface{}) func(*error) {
	if ctx == nil {
		ctx = emptyCtx()
	} else if ctx == &lazyTaskSecret && taskArgs(f, args) {
		return nil
	}
	s, exit := newSpan(*ctx, f, args, NewId(), nil)
	*ctx = s
	return exit
}

func (f *Func) RemoteTrace(ctx *context.Context, spanId int64, trace *Trace,
	args ...interface{}) func(*error) {
	if ctx == nil {
		ctx = emptyCtx()
	}
	if trace != nil {
		f.scope.r.observeTrace(trace)
	}
	s, exit := newSpan(*ctx, f, args, spanId, trace)
	*ctx = s
	return exit
}

func (f *Func) ResetTrace(ctx *context.Context,
	args ...interface{}) func(*error) {
	if ctx == nil {
		ctx = emptyCtx()
	} else if ctx == &lazyTaskSecret && taskArgs(f, args) {
		return nil
	}
	trace := NewTrace(NewId())
	f.scope.r.observeTrace(trace)
	s, exit := newSpan(*ctx, f, args, trace.Id(), trace)
	*ctx = s
	return exit
}

func emptyCtx() *context.Context {
	// TODO: maybe we should generate some special parent for these unparented
	// spans
	ctx := context.Background()
	return &ctx
}
