package monitor

import (
	"fmt"
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
	s = newScope(r, name)
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
	r.Scopes(func(s *Scope) { s.Funcs(cb) })
}

func (r *Registry) Meters(cb func(*Meter)) {
	r.Scopes(func(s *Scope) { s.Meters(cb) })
}

func (r *Registry) Stats(cb func(name string, val float64)) {
	r.Scopes(func(s *Scope) {
		s.Stats(func(name string, val float64) {
			cb(fmt.Sprintf("%s.%s", s.Name, name), val)
		})
	})
}

var (
	Default      = NewRegistry()
	Package      = Default.Package
	PackageNamed = Default.PackageNamed
	LiveTraces   = Default.LiveTraces
	Scopes       = Default.Scopes
	Funcs        = Default.Funcs
	Meters       = Default.Meters
	Stats        = Default.Stats
)
