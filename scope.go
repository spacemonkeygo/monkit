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

func (s *Scope) newSource(name string, constructor func() StatSource) (
	rv StatSource) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if source, exists := s.sources[name]; exists {
		return source
	}
	ss := constructor()
	s.sources[name] = ss
	return ss
}

func (s *Scope) FuncNamed(name string) *Func {
	source := s.newSource(name, func() StatSource { return newFunc(s, name) })
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
	source := s.newSource(name, newMeter)
	m, ok := source.(*Meter)
	if !ok {
		panic(fmt.Sprintf("%s already used for another stats source: %#v",
			name, source))
	}
	return m
}

func (s *Scope) Event(name string) {
	s.Meter(name).Mark(1)
}

func (s *Scope) DiffMeter(name string, m1, m2 *Meter) {
	source := s.newSource(name, func() StatSource {
		return newDiffMeter(m1, m2)
	})
	if _, ok := source.(*DiffMeter); !ok {
		panic(fmt.Sprintf("%s already used for another stats source: %#v",
			name, source))
	}
}

func (s *Scope) IntVal(name string) *IntVal {
	source := s.newSource(name, newIntVal)
	m, ok := source.(*IntVal)
	if !ok {
		panic(fmt.Sprintf("%s already used for another stats source: %#v",
			name, source))
	}
	return m
}

func (s *Scope) FloatVal(name string) *FloatVal {
	source := s.newSource(name, newFloatVal)
	m, ok := source.(*FloatVal)
	if !ok {
		panic(fmt.Sprintf("%s already used for another stats source: %#v",
			name, source))
	}
	return m
}

func (s *Scope) BoolVal(name string) *BoolVal {
	source := s.newSource(name, newBoolVal)
	m, ok := source.(*BoolVal)
	if !ok {
		panic(fmt.Sprintf("%s already used for another stats source: %#v",
			name, source))
	}
	return m
}

func (s *Scope) Counter(name string) *Counter {
	source := s.newSource(name, newCounter)
	m, ok := source.(*Counter)
	if !ok {
		panic(fmt.Sprintf("%s already used for another stats source: %#v",
			name, source))
	}
	return m
}

func (s *Scope) Gauge(name string, cb func() float64) {
	// gauges allow overwriting
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if source, exists := s.sources[name]; exists {
		if _, ok := source.(gauge); !ok {
			panic(fmt.Sprintf("%s already used for another stats source: %#v",
				name, source))
		}
	}
	s.sources[name] = gauge{cb: cb}
}

func (s *Scope) Chain(name string, source StatSource) {
	// chains allow overwriting
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if source, exists := s.sources[name]; exists {
		if _, ok := source.(chain); !ok {
			panic(fmt.Sprintf("%s already used for another stats source: %#v",
				name, source))
		}
	}
	s.sources[name] = chain{source: source}
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
