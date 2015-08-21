package monitor

import (
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/monotime"
)

type Meter struct {
	// sync/atomic things
	lastTime   int64
	totalCount int64
	sliceCount int64
}

func newMeter() *Meter {
	return &Meter{lastTime: int64(monotime.Monotonic())}
}

func (e *Meter) Stats(cb func(name string, val float64)) {
	currentTime := int64(monotime.Monotonic())
	lastTime := atomic.SwapInt64(&e.lastTime, currentTime)
	sliceCount := atomic.SwapInt64(&e.sliceCount, 0)
	totalCount := atomic.AddInt64(&e.totalCount, sliceCount)
	cb("rate", float64(sliceCount)/time.Duration(currentTime-lastTime).Seconds())
	cb("total", float64(totalCount))
}

func (e *Meter) Mark(amount int) {
	atomic.AddInt64(&e.sliceCount, int64(amount))
}
