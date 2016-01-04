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
	"sync"
	"time"

	"github.com/spacemonkeygo/monotime"
)

const (
	ticksToKeep = 4
	timePerTick = time.Minute
)

var (
	defaultTicker = ticker{}
)

type meterBucket struct {
	count int64
	start time.Duration
}

type Meter struct {
	mtx    sync.Mutex
	total  int64
	slices [ticksToKeep]meterBucket
}

func newMeter() *Meter {
	rv := &Meter{}
	now := monotime.Monotonic()
	for i := 0; i < ticksToKeep; i++ {
		rv.slices[i].start = now
	}
	defaultTicker.register(rv)
	return rv
}

func (e *Meter) SetTotal(total int64) {
	e.mtx.Lock()
	e.total = total
	e.mtx.Unlock()
}

func (e *Meter) Mark(amount int) {
	e.mtx.Lock()
	e.slices[ticksToKeep-1].count += int64(amount)
	e.mtx.Unlock()
}

func (e *Meter) tick(now time.Duration) {
	e.mtx.Lock()
	// only advance meter buckets if something happened. otherwise
	// rare events will always just have zero rates.
	if e.slices[ticksToKeep-1].count != 0 {
		e.total += e.slices[0].count
		copy(e.slices[:], e.slices[1:])
		e.slices[ticksToKeep-1] = meterBucket{count: 0, start: now}
	}
	e.mtx.Unlock()
}

func (e *Meter) stats(now time.Duration) (rate float64, total int64) {
	current := int64(0)
	e.mtx.Lock()
	start := e.slices[0].start
	for i := 0; i < ticksToKeep; i++ {
		current += e.slices[i].count
	}
	total = e.total
	e.mtx.Unlock()
	total += current
	rate = float64(current) / (now - start).Seconds()
	return rate, total
}

func (e *Meter) Stats(cb func(name string, val float64)) {
	rate, total := e.stats(monotime.Monotonic())
	cb("rate", rate)
	cb("total", float64(total))
}

type ticker struct {
	mtx     sync.Mutex
	started bool
	meters  []*Meter
}

func (t *ticker) register(m *Meter) {
	t.mtx.Lock()
	if !t.started {
		t.started = true
		go t.run()
	}
	t.meters = append(t.meters, m)
	t.mtx.Unlock()
}

func (t *ticker) run() {
	for {
		time.Sleep(timePerTick)
		t.mtx.Lock()
		meters := t.meters // this is safe since we only use append
		t.mtx.Unlock()
		now := monotime.Monotonic()
		for _, m := range meters {
			m.tick(now)
		}
	}
}
