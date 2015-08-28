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
	"sync/atomic"
	"unsafe"
)

type traceWatcherRef struct {
	watcher func(*Trace)
}

type Registry struct {
	// sync/atomic things
	traceWatcher unsafe.Pointer

	scopeMtx sync.Mutex
	scopes   map[string]*Scope

	spanMtx sync.Mutex
	spans   map[*Span]struct{}

	orphanMtx sync.Mutex
	orphans   map[*Span]struct{}
}

func NewRegistry() *Registry {
	return &Registry{
		scopes:  map[string]*Scope{},
		spans:   map[*Span]struct{}{},
		orphans: map[*Span]struct{}{}}
}

func (r *Registry) Package() *Scope {
	return r.PackageNamed(callerPackage(1))
}

func (r *Registry) PackageNamed(name string) *Scope {
	r.scopeMtx.Lock()
	defer r.scopeMtx.Unlock()
	s, exists := r.scopes[name]
	if exists {
		return s
	}
	s = newScope(r, name)
	r.scopes[name] = s
	return s
}

func (r *Registry) observeTrace(t *Trace) {
	watcher := (*traceWatcherRef)(atomic.LoadPointer(&r.traceWatcher))
	if watcher != nil {
		watcher.watcher(t)
	}
}

func (r *Registry) ObserveTraces(cb func(*Trace)) {
	if cb == nil {
		return
	}
	for {
		existing := (*traceWatcherRef)(atomic.LoadPointer(&r.traceWatcher))
		if existing == nil {
			if atomic.CompareAndSwapPointer(&r.traceWatcher, nil,
				unsafe.Pointer(&traceWatcherRef{watcher: cb})) {
				break
			}
		} else {
			other_cb := existing.watcher
			if atomic.CompareAndSwapPointer(&r.traceWatcher,
				unsafe.Pointer(existing),
				unsafe.Pointer(&traceWatcherRef{watcher: func(t *Trace) {
					other_cb(t)
					cb(t)
				}})) {
				break
			}
		}
	}
}

func (r *Registry) rootSpanStart(s *Span) {
	r.spanMtx.Lock()
	r.spans[s] = struct{}{}
	r.spanMtx.Unlock()
}

func (r *Registry) rootSpanEnd(s *Span) {
	r.spanMtx.Lock()
	delete(r.spans, s)
	r.spanMtx.Unlock()
}

func (r *Registry) orphanedSpan(s *Span) {
	r.orphanMtx.Lock()
	r.orphans[s] = struct{}{}
	r.orphanMtx.Unlock()
}

func (r *Registry) orphanEnd(s *Span) {
	r.orphanMtx.Lock()
	r.orphans[s] = struct{}{}
	r.orphanMtx.Unlock()
}

func (r *Registry) LiveSpans(cb func(s *Span)) {
	r.spanMtx.Lock()
	spans := make([]*Span, 0, len(r.spans))
	for s := range r.spans {
		spans = append(spans, s)
	}
	r.spanMtx.Unlock()
	r.orphanMtx.Lock()
	orphans := make([]*Span, 0, len(r.orphans))
	for s := range r.orphans {
		orphans = append(orphans, s)
	}
	r.orphanMtx.Unlock()
	spans = append(spans, orphans...)
	sort.Sort(spanSorter(spans))
	for _, s := range spans {
		cb(s)
	}
}

func (r *Registry) Scopes(cb func(s *Scope)) {
	r.scopeMtx.Lock()
	c := make([]*Scope, 0, len(r.scopes))
	for _, s := range r.scopes {
		c = append(c, s)
	}
	r.scopeMtx.Unlock()
	sort.Sort(scopeSorter(c))
	for _, s := range c {
		cb(s)
	}
}

func (r *Registry) Funcs(cb func(f *Func)) {
	r.Scopes(func(s *Scope) { s.Funcs(cb) })
}

func (r *Registry) Stats(cb func(name string, val float64)) {
	r.Scopes(func(s *Scope) {
		s.Stats(func(name string, val float64) {
			cb(fmt.Sprintf("%s.%s", s.name, name), val)
		})
	})
}

var (
	Default      = NewRegistry()
	PackageNamed = Default.PackageNamed
	LiveSpans    = Default.LiveSpans
	Scopes       = Default.Scopes
	Funcs        = Default.Funcs
	Stats        = Default.Stats
)

func Package() *Scope {
	return PackageNamed(callerPackage(1))
}

type spanSorter []*Span

func (s spanSorter) Len() int      { return len(s) }
func (s spanSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s spanSorter) Less(i, j int) bool {
	return s[i].f.FullName() < s[j].f.FullName() && s[i].id < s[j].id
}

type scopeSorter []*Scope

func (s scopeSorter) Len() int           { return len(s) }
func (s scopeSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s scopeSorter) Less(i, j int) bool { return s[i].name < s[j].name }
