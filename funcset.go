package monitor

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// funcSet is a set data structure (keeps track of unique functions). funcSet
// has a fast path for dealing with cases where the set only has one element.
//
// to reduce memory usage for functions, funcSet exposes its mutex for use in
// other contexts
type funcSet struct {
	// sync/atomic things
	first unsafe.Pointer

	// protected by mtx
	sync.Mutex
	rest map[*Func]struct{}
}

var (
	// used to signify that we've specifically added a nil function, since nil is
	// used internally to specify an empty set.
	nilFunc = &Func{name: "nil function"}
)

func (s *funcSet) Add(f *Func) {
	if f == nil {
		f = nilFunc
	}
	if atomic.LoadPointer(&s.first) == unsafe.Pointer(f) {
		return
	}
	if atomic.CompareAndSwapPointer(&s.first, nil, unsafe.Pointer(f)) {
		return
	}
	s.Mutex.Lock()
	if s.rest == nil {
		s.rest = map[*Func]struct{}{}
	}
	s.rest[f] = struct{}{}
	s.Mutex.Unlock()
}

// Iterate loops over all unique elements of the set.
func (s *funcSet) Iterate(cb func(f *Func)) {
	s.Mutex.Lock()
	uniq := make(map[*Func]struct{}, len(s.rest)+1)
	for f := range s.rest {
		uniq[f] = struct{}{}
	}
	s.Mutex.Unlock()
	f := (*Func)(atomic.LoadPointer(&s.first))
	if f != nil {
		uniq[f] = struct{}{}
	}
	for f := range uniq {
		if f == nilFunc {
			cb(nil)
		} else {
			cb(f)
		}
	}
}
