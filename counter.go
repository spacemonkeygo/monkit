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
	"math"
	"sync"
)

type Counter struct {
	mtx       sync.Mutex
	val       int64
	low, high *int64
}

func newCounter() StatSource {
	return &Counter{}
}

func (c *Counter) set(val int64) {
	c.val = val
	if c.low == nil || val < *c.low {
		c.low = &val
	}
	if c.high == nil || *c.high < val {
		c.high = &val
	}
}

func (c *Counter) Set(val int64) {
	c.mtx.Lock()
	c.set(val)
	c.mtx.Unlock()
}

func (c *Counter) Inc(delta int64) {
	c.mtx.Lock()
	c.set(c.val + delta)
	c.mtx.Unlock()
}

func (c *Counter) Dec(delta int64) { c.Inc(-delta) }

func (c *Counter) Stats(cb func(name string, val float64)) {
	c.mtx.Lock()
	val, low, high := c.val, c.low, c.high
	c.mtx.Unlock()
	if high != nil {
		cb("high", float64(*high))
	} else {
		cb("high", math.NaN())
	}
	if low != nil {
		cb("low", float64(*low))
	} else {
		cb("low", math.NaN())
	}
	cb("val", float64(val))
}
