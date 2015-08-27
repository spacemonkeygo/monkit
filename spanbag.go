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
	count int32

	// protected by mtx
	mtx  sync.Mutex
	rest map[*Span]int32
}

func (b *spanBag) Add(s *Span) {
	atomic.AddInt32(&b.count, 1)
	if atomic.CompareAndSwapPointer(&b.first, nil, unsafe.Pointer(s)) {
		return
	}
	// we had some kind of contention. let's just put it into b.rest
	b.mtx.Lock()
	if b.rest == nil {
		b.rest = map[*Span]int32{}
	}
	b.rest[s] += 1
	b.mtx.Unlock()
}

func (b *spanBag) Remove(s *Span) {
	if atomic.CompareAndSwapPointer(&b.first, unsafe.Pointer(s), nil) {
		atomic.AddInt32(&b.count, -1)
		return
	}
	// okay it must be in b.rest
	b.mtx.Lock()
	count := b.rest[s]
	if count <= 1 {
		delete(b.rest, s)
	} else {
		b.rest[s] = count - 1
	}
	b.mtx.Unlock()
	atomic.AddInt32(&b.count, -1)
}

// Iterate loops over all elements of the bag after removing duplicates.
func (b *spanBag) Iterate(cb func(s *Span), shouldSort bool) {
	if atomic.LoadInt32(&b.count) == 0 {
		return
	}
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
	if !shouldSort {
		for s := range uniq {
			cb(s)
		}
		return
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
