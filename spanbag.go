package monitor

import (
	"sort"
	"sync"
	"sync/atomic"
	"unsafe"
)

// spanBag is a bag data structure (can add 0 or more references to a span,
// where every add needs to be matched with an equivalent remove). spanBag has
// a fast path for dealing with cases where the bag only has one element (the
// common case).
type spanBag struct {
	// sync/atomic things
	first unsafe.Pointer

	// protected by mtx
	mtx  sync.Mutex
	rest map[*Span]int
}

func (b *spanBag) Add(s *Span) {
	if atomic.CompareAndSwapPointer(&b.first, nil, unsafe.Pointer(s)) {
		return
	}
	// we had some kind of contention. let's just put it into b.rest
	b.mtx.Lock()
	if b.rest == nil {
		b.rest = map[*Span]int{}
	}
	b.rest[s] += 1
	b.mtx.Unlock()
}

func (b *spanBag) Remove(s *Span) {
	if atomic.CompareAndSwapPointer(&b.first, unsafe.Pointer(s), nil) {
		return
	}
	// okay it must be in b.rest
	b.mtx.Lock()
	if b.rest == nil {
		return
	}
	count := b.rest[s]
	if count <= 1 {
		delete(b.rest, s)
	} else {
		b.rest[s] = count - 1
	}
	b.mtx.Unlock()
}

// Iterate loops over all elements of the bag after removing duplicates.
func (b *spanBag) Iterate(cb func(s *Span)) {
	b.mtx.Lock()
	uniq := make(map[*Span]struct{}, len(b.rest)+1)
	for s := range b.rest {
		uniq[s] = struct{}{}
	}
	b.mtx.Unlock()
	s := (*Span)(atomic.LoadPointer(&b.first))
	if s != nil {
		uniq[s] = struct{}{}
	}
	s_sorted := make([]*Span, 0, len(uniq))
	for s := range uniq {
		s_sorted = append(s_sorted, s)
	}
	sort.Sort(spanSorter(s_sorted))
	for _, s := range s_sorted {
		cb(s)
	}
}
