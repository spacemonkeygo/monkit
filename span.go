package monitor

import (
	"fmt"
	"time"

	"github.com/spacemonkeygo/monotime"
	"golang.org/x/net/context"
)

type ctxKey int

const (
	spanKey ctxKey = iota
)

type Span struct {
	// stuff with sync/atomic
	children spanBag

	// immutable things from construction
	Id     int64
	start  time.Duration
	Func   *Func
	Parent *Span
	args   []interface{}
	context.Context
}

func SpanFromCtx(ctx context.Context) *Span {
	if s, ok := ctx.(*Span); ok && s != nil {
		return s
	} else if s, ok := ctx.Value(spanKey).(*Span); ok && s != nil {
		return s
	}
	return nil
}

func newSpan(ctx context.Context, f *Func, args []interface{}) (
	s *Span, exit func(*error)) {

	var parent *Span
	if s, ok := ctx.(*Span); ok && s != nil {
		parent = s
		ctx = s.Context
	} else if s, ok := ctx.Value(spanKey).(*Span); ok && s != nil {
		parent = s
	}

	s = &Span{
		Id:      newId(),
		start:   monotime.Monotonic(),
		Func:    f,
		Parent:  parent,
		args:    args,
		Context: ctx}

	if parent != nil {
		f.start(parent.Func)
		parent.children.Add(s)
	} else {
		f.start(nil)
		f.Scope.r.traceStart(s)
	}

	return s, func(errptr *error) {
		rec := recover()
		panicked := rec != nil

		if parent != nil {
			parent.children.Remove(s)
		} else {
			f.Scope.r.traceEnd(s)
		}

		f.end(errptr, panicked, monotime.Monotonic()-s.start)

		// try to help the garbage collector
		s.Parent = nil

		if panicked {
			panic(rec)
		}
	}
}

func (s *Span) Duration() time.Duration {
	return monotime.Monotonic() - s.start
}

func (s *Span) Value(key interface{}) interface{} {
	if key == spanKey {
		return s
	}
	return s.Context.Value(key)
}

func (s *Span) String() string {
	// TODO: for working with Contexts
	return fmt.Sprintf("%v.WithSpan()", s.Context)
}

func (s *Span) Children(cb func(s *Span)) {
	s.children.Iterate(cb, true)
}

func (s *Span) Args() (rv []string) {
	rv = make([]string, 0, len(s.args))
	for _, arg := range s.args {
		rv = append(rv, fmt.Sprintf("%#v", arg))
	}
	return rv
}
