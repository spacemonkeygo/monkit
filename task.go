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

type taskSecretT struct{}

func (*taskSecretT) Value(key interface{}) interface{} { return nil }
func (*taskSecretT) Done() <-chan struct{}             { return nil }
func (*taskSecretT) Err() error                        { return nil }
func (*taskSecretT) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

var taskSecret context.Context = &taskSecretT{}

type Task func(ctx *context.Context, args ...interface{}) func(*error)

// Func returns the Func associated with the Task
func (f Task) Func() (out *Func) {
	// we're doing crazy things to make a function have methods that do other
	// things with internal state. basically, we have a secret argument we can
	// pass to the function that is only checked if ctx is taskSecret (
	// which it should never be) that controls what other behavior we want.
	// in this case, if arg[0] is taskGetFunc, then f will place the func in the
	// out location.
	// since someone can cast any function of this signature to a lazy task,
	// let's make sure we got roughly expected behavior and panic otherwise
	if f(&taskSecret, taskGetFunc, &out) != nil || out == nil {
		panic("Func() called on a non-Task function")
	}
	return out
}

func taskArgs(f *Func, args []interface{}) bool {
	// this function essentially does method dispatch for Tasks. returns true
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

// Task returns a new Task for use, creating an associated Func if necessary.
// It also adds a new Span to the given ctx during execution. Expected usage
// like:
//
//   var mon = monitor.Package()
//
//   func MyFunc(ctx context.Context, arg1, arg2 string) (err error) {
//     defer mon.Task()(&ctx, arg1, arg2)(&err)
//     ...
//   }
//
// or
//
//   var (
//     mon = monitor.Package()
//     funcTask = mon.Task()
//   )
//
//   func MyFunc(ctx context.Context, arg1, arg2 string) (err error) {
//     defer funcTask(&ctx, arg1, arg2)(&err)
//     ...
//   }
//
// Task uses runtime.Caller to determine the associated Func name. See
// TaskNamed if you want to supply your own name. See Func.Task if you already
// have a Func.
//
// If you want to control Trace creation, see Func.ResetTrace and
// Func.RemoteTrace
func (s *Scope) Task() Task {
	var initOnce sync.Once
	var f *Func
	init := func() {
		f = s.FuncNamed(callerFunc(3))
	}
	return Task(func(ctx *context.Context,
		args ...interface{}) func(*error) {
		if ctx == nil {
			ctx = emptyCtx()
		} else if ctx == &taskSecret && taskArgs(f, args) {
			return nil
		}
		initOnce.Do(init)
		s, exit := newSpan(*ctx, f, args, NewId(), nil)
		*ctx = s
		return exit
	})
}

// TaskNamed is like Task except you can choose the name of the associated
// Func.
func (s *Scope) TaskNamed(name string) Task {
	return s.FuncNamed(name).Task
}

// Task returns a new Task for use on this Func. It also adds a new Span to
// the given ctx during execution.
//
//   var mon = monitor.Package()
//
//   func MyFunc(ctx context.Context, arg1, arg2 string) (err error) {
//     f := mon.Func()
//     defer f.Task(&ctx, arg1, arg2)(&err)
//     ...
//   }
//
// It's more expected for you to use mon.Task directly. See RemoteTrace or
// ResetTrace if you want greater control over creating new traces.
func (f *Func) Task(ctx *context.Context, args ...interface{}) func(*error) {
	if ctx == nil {
		ctx = emptyCtx()
	} else if ctx == &taskSecret && taskArgs(f, args) {
		return nil
	}
	s, exit := newSpan(*ctx, f, args, NewId(), nil)
	*ctx = s
	return exit
}

// RemoteTrace is like Func.Task, except you can specify the trace and span id.
// Needed for things like the Zipkin plugin.
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

// ResetTrace is like Func.Task, except it always creates a new Trace.
func (f *Func) ResetTrace(ctx *context.Context,
	args ...interface{}) func(*error) {
	if ctx == nil {
		ctx = emptyCtx()
	} else if ctx == &taskSecret && taskArgs(f, args) {
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
