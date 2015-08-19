package monitor

import (
	"sync"
)

type Registry struct {
	mtx        sync.Mutex
	scopes     map[string]*Scope
	liveTraces map[*Span]struct{}
}

func NewRegistry() *Registry {
	return &Registry{
		scopes:     map[string]*Scope{},
		liveTraces: map[*Span]struct{}{}}
}

func (r *Registry) Package() *Scope {
	return r.PackageNamed(callerPackage(1))
}

func (r *Registry) PackageNamed(name string) *Scope {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	s, exists := r.scopes[name]
	if exists {
		return s
	}
	s = newScope(r)
	r.scopes[name] = s
	return s
}

func (r *Registry) traceStart(s *Span) {
	r.mtx.Lock()
	r.liveTraces[s] = struct{}{}
	r.mtx.Unlock()
}

func (r *Registry) traceEnd(s *Span) {
	r.mtx.Lock()
	delete(r.liveTraces, s)
	r.mtx.Unlock()
}

func (r *Registry) LiveTraces(cb func(s *Span)) {
	r.mtx.Lock()
	c := make(map[*Span]struct{}, len(r.liveTraces))
	for s := range r.liveTraces {
		c[s] = struct{}{}
	}
	r.mtx.Unlock()
	for s := range c {
		cb(s)
	}
}

func (r *Registry) Scopes(cb func(s *Scope)) {
	r.mtx.Lock()
	c := make(map[*Scope]struct{}, len(r.scopes))
	for _, s := range r.scopes {
		c[s] = struct{}{}
	}
	r.mtx.Unlock()
	for s := range c {
		cb(s)
	}
}

func (r *Registry) Funcs(cb func(f *Func)) {
	r.Scopes(func(s *Scope) {
		s.Funcs(cb)
	})
}

var (
	Default      = NewRegistry()
	Package      = Default.Package
	PackageNamed = Default.PackageNamed
)
