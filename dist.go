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
	"sort"
	"time"
)

const (
	reservoirSize = 64
)

var (
	ObservedQuantiles = []float64{0, .1, .25, .5, .75, .9, .95, 1}
)

// dist is not threadsafe
type dist struct {
	low, high   time.Duration
	recent      time.Duration
	totalValues int64
	sum         time.Duration
	reservoir   [reservoirSize]float32
	lcg         lcg
	sorted      bool
}

func newDist() dist {
	return dist{lcg: newLCG()}
}

func (d *dist) Stats() (min, avg, max, recent time.Duration) {
	return d.low, d.Average(), d.high, d.recent
}

func (d *dist) Insert(val time.Duration) {
	if d.totalValues != 0 {
		if val < d.low {
			d.low = val
		}
		if val > d.high {
			d.high = val
		}
	} else {
		d.low = val
		d.high = val
	}
	d.recent = val
	d.sum += val

	index := d.totalValues
	d.totalValues += 1

	if index < reservoirSize {
		d.reservoir[index] = float32(val)
		d.sorted = false
	} else {
		// fast, but kind of biased. probably okay
		j := d.lcg.Uint64() % uint64(d.totalValues)
		if j < reservoirSize {
			d.reservoir[int(j)] = float32(val)
			d.sorted = false
		}
	}
}

func (d *dist) Average() time.Duration {
	if d.totalValues > 0 {
		return d.sum / time.Duration(d.totalValues)
	} else {
		return 0
	}
}

func (d *dist) Query(quantile float64) time.Duration {
	if quantile <= 0 {
		return d.low
	}
	if quantile >= 1 {
		return d.high
	}

	rlen := int(reservoirSize)
	if int64(rlen) > d.totalValues {
		rlen = int(d.totalValues)
	}

	idx_float := quantile * float64(rlen-1)
	idx := int(idx_float)

	reservoir := d.reservoir[:rlen]
	if !d.sorted {
		sort.Sort(float32Slice(reservoir))
		d.sorted = true
	}
	diff := idx_float - float64(idx)
	prior := float64(reservoir[idx])
	return time.Duration(prior + diff*(float64(reservoir[idx+1])-prior))
}

func (d *dist) Recent() time.Duration { return d.recent }

type float32Slice []float32

func (p float32Slice) Len() int      { return len(p) }
func (p float32Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p float32Slice) Less(i, j int) bool {
	// N.B.: usually, float comparisons should check if either value is NaN, but
	// in this package's usage, they never are here.
	return p[i] < p[j]
}
