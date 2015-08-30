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
	"fmt"
	"sort"
	"sync"
)

type Scope struct {
	r       *Registry
	name    string
	mtx     sync.Mutex
	sources map[string]StatSource
}

func newScope(r *Registry, name string) *Scope {
	return &Scope{
		r:       r,
		name:    name,
		sources: map[string]StatSource{}}
}

func (s *Scope) Func() *Func {
	return s.FuncNamed(callerFunc(1))
}

func (s *Scope) FuncNamed(name string) *Func {
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

func (s *Scope) Stats(cb func(name string, val float64)) {
	s.mtx.Lock()
	sources := make([]namedSource, 0, len(s.sources))
	for name, source := range s.sources {
		sources = append(sources, namedSource{name: name, source: source})
	}
	s.mtx.Unlock()
	sort.Sort(namedSourceList(sources))
	for _, namedSource := range sources {
		namedSource.source.Stats(func(name string, val float64) {
			cb(fmt.Sprintf("%s.%s", namedSource.name, name), val)
		})
	}
}

func (s *Scope) Name() string { return s.name }

type namedSource struct {
	name   string
	source StatSource
}

type namedSourceList []namedSource

func (l namedSourceList) Len() int           { return len(l) }
func (l namedSourceList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l namedSourceList) Less(i, j int) bool { return l[i].name < l[j].name }
