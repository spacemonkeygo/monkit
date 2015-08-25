package monitor

import (
	"runtime"
	"sync/atomic"
)

type spinLock uint32

func (s *spinLock) Lock() {
	for {
		if atomic.CompareAndSwapUint32((*uint32)(s), 0, 1) {
			return
		}
		runtime.Gosched()
	}
}

func (s *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(s), 0)
}
