package monitor

import (
	"fmt"
	"sort"
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
	c := make([]*Span, 0, len(r.liveTraces))
	for s := range r.liveTraces {
		c = append(c, s)
	}
	r.mtx.Unlock()
	sort.Sort(spanSorter(c))
	for _, s := range c {
		cb(s)
	}
}

func (r *Registry) Scopes(cb func(s *Scope)) {
	r.mtx.Lock()
	c := make([]*Scope, 0, len(r.scopes))
	for _, s := range r.scopes {
		c = append(c, s)
	}
	r.mtx.Unlock()
	sort.Sort(scopeSorter(c))
	for _, s := range c {
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
	PackageNamed = Default.PackageNamed
	LiveTraces   = Default.LiveTraces
	Scopes       = Default.Scopes
	Funcs        = Default.Funcs
	Meters       = Default.Meters
	Stats        = Default.Stats
)

func Package() *Scope {
	return PackageNamed(callerPackage(1))
}

type spanSorter []*Span

func (s spanSorter) Len() int      { return len(s) }
func (s spanSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s spanSorter) Less(i, j int) bool {
	return s[i].Func.Name() < s[j].Func.Name() && s[i].Id < s[j].Id
}

type scopeSorter []*Scope

func (s scopeSorter) Len() int           { return len(s) }
func (s scopeSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s scopeSorter) Less(i, j int) bool { return s[i].Name < s[j].Name }
