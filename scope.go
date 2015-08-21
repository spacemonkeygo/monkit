package monitor

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"
)

type Scope struct {
	r       *Registry
	Name    string
	mtx     sync.Mutex
	sources map[string]StatSource
}

func newScope(r *Registry, name string) *Scope {
	return &Scope{
		r:       r,
		Name:    name,
		sources: map[string]StatSource{}}
}

func (s *Scope) Func() (
	rv func(ctx *context.Context, args ...interface{}) func(*error)) {
	var initOnce sync.Once
	var f *Func
	init := func() {
		f = s.function(callerFunc(3))
	}
	return func(ctx *context.Context, args ...interface{}) func(*error) {
		initOnce.Do(init)
		s, exit := newSpan(*ctx, f, args)
		*ctx = s
		return exit
	}
}

func (s *Scope) FuncNamed(name string) (
	rv func(ctx *context.Context, args ...interface{}) func(*error)) {
	f := s.function(name)
	return func(ctx *context.Context, args ...interface{}) func(*error) {
		s, exit := newSpan(*ctx, f, args)
		*ctx = s
		return exit
	}
}

func (s *Scope) function(name string) *Func {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	source, exists := s.sources[name]
	if !exists {
		f := newFunc(s, name)
		s.sources[name] = f
		return f
	}
	f, ok := source.(*Func)
	if !ok {
		panic(fmt.Sprintf("%s already used for another stats source: %#v",
			name, source))
	}
	return f
}

func (s *Scope) Funcs(cb func(f *Func)) {
	s.mtx.Lock()
	funcs := make(map[*Func]struct{}, len(s.sources))
	for _, source := range s.sources {
		if f, ok := source.(*Func); ok {
			funcs[f] = struct{}{}
		}
	}
	s.mtx.Unlock()
	for f := range funcs {
		cb(f)
	}
}

func (s *Scope) Meter(name string) *Meter {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	source, exists := s.sources[name]
	if !exists {
		m := newMeter()
		s.sources[name] = m
		return m
	}
	m, ok := source.(*Meter)
	if !ok {
		panic(fmt.Sprintf("%s already used for another stats source: %#v",
			name, source))
	}
	return m
}

func (s *Scope) Meters(cb func(m *Meter)) {
	s.mtx.Lock()
	meters := make(map[*Meter]struct{}, len(s.sources))
	for _, source := range s.sources {
		if m, ok := source.(*Meter); ok {
			meters[m] = struct{}{}
		}
	}
	s.mtx.Unlock()
	for m := range meters {
		cb(m)
	}
}

func (s *Scope) Stats(cb func(name string, val float64)) {
	s.mtx.Lock()
	sources := make(map[StatSource]struct{}, len(s.sources))
	for _, source := range s.sources {
		sources[source] = struct{}{}
	}
	s.mtx.Unlock()
	for source := range sources {
		source.Stats(cb)
	}
}
