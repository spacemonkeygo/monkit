package monitor

import (
	"sync"

	"golang.org/x/net/context"
)

type Scope struct {
	r     *Registry
	mtx   sync.Mutex
	funcs map[string]*Func
}

func newScope(r *Registry) *Scope {
	return &Scope{
		r:     r,
		funcs: map[string]*Func{}}
}

func (s *Scope) Func() (
	rv func(ctx *context.Context, args ...interface{}) func(*error)) {
	var initOnce sync.Once
	var state *Func
	init := func() {
		state = s.function(callerFunc(3))
	}
	return func(ctx *context.Context, args ...interface{}) func(*error) {
		initOnce.Do(init)
		s, exit := newSpan(*ctx, state, args)
		*ctx = s
		return exit
	}
}

func (s *Scope) FuncNamed(name string) (
	rv func(ctx *context.Context, args ...interface{}) func(*error)) {
	state := s.function(name)
	return func(ctx *context.Context, args ...interface{}) func(*error) {
		s, exit := newSpan(*ctx, state, args)
		*ctx = s
		return exit
	}
}

func (s *Scope) function(name string) *Func {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	state, exists := s.funcs[name]
	if !exists {
		state = newFunc(s, name)
		s.funcs[name] = state
	}
	return state
}

func (s *Scope) Funcs(cb func(f *Func)) {
	s.mtx.Lock()
	funcs := make(map[*Func]struct{}, len(s.funcs))
	for _, f := range s.funcs {
		funcs[f] = struct{}{}
	}
	s.mtx.Unlock()
	for f := range funcs {
		cb(f)
	}
}
