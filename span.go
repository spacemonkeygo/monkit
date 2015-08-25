package monitor

import (
	"fmt"
	"sort"
	"time"

	"github.com/spacemonkeygo/monotime"
	"golang.org/x/net/context"
)

type ctxKey int

const (
	spanKey ctxKey = iota
)

type Span struct {
	// sync/atomic things
	mtx spinLock

	// immutable things from construction
	Id    int64
	start time.Duration
	Func  *Func
	args  []interface{}
	context.Context

	// protected by mtx
	done     bool
	orphaned bool
	children spanBag
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
		args:    args,
		Context: ctx}

	if parent != nil {
		f.start(parent.Func)
		parent.addChild(s)
	} else {
		f.start(nil)
		f.Scope.r.traceStart(s)
	}

	return s, func(errptr *error) {
		rec := recover()
		panicked := rec != nil

		f.end(errptr, panicked, monotime.Monotonic()-s.start)

		var children []*Span
		s.mtx.Lock()
		s.done = true
		orphaned := s.orphaned
		s.children.Iterate(func(child *Span) {
			children = append(children, child)
		})
		s.mtx.Unlock()
		for _, child := range children {
			child.orphan()
		}

		if parent != nil {
			parent.removeChild(s)
			if orphaned {
				f.Scope.r.orphanEnd(s)
			}
		} else {
			f.Scope.r.traceEnd(s)
		}

		if panicked {
			panic(rec)
		}
	}
}

func (s *Span) addChild(child *Span) {
	s.mtx.Lock()
	s.children.Add(child)
	done := s.done
	s.mtx.Unlock()
	if done {
		child.orphan()
	}
}

func (s *Span) removeChild(child *Span) {
	s.mtx.Lock()
	s.children.Remove(child)
	s.mtx.Unlock()
}

func (s *Span) orphan() {
	s.mtx.Lock()
	if !s.done && !s.orphaned {
		s.orphaned = true
		s.Func.Scope.r.orphanedSpan(s)
	}
	s.mtx.Unlock()
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
	found := map[*Span]bool{}
	var sorter []*Span
	s.mtx.Lock()
	s.children.Iterate(func(s *Span) {
		if !found[s] {
			found[s] = true
			sorter = append(sorter, s)
		}
	})
	s.mtx.Unlock()
	sort.Sort(spanSorter(sorter))
	for _, s := range sorter {
		cb(s)
	}
}

func (s *Span) Args() (rv []string) {
	rv = make([]string, 0, len(s.args))
	for _, arg := range s.args {
		rv = append(rv, fmt.Sprintf("%#v", arg))
	}
	return rv
}
