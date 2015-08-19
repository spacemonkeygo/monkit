package monitor

import (
	"sync/atomic"
)

var (
	idCounter int64
)

func newId() int64 {
	return atomic.AddInt64(&idCounter, 1)
}
