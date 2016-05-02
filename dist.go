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

// IntDist keeps statistics about ints.
type IntDist struct {
	Low, High int64
	Recent    int64
	Count     int64
	Sum       int64
	reservoir [reservoirSize]float32
	lcg       lcg
	sorted    bool
}

func NewIntDist() *IntDist {
	return &IntDist{lcg: newLCG()}
}

func (d *IntDist) Insert(val int64) {
	if d.Count != 0 {
		if val < d.Low {
			d.Low = val
		}
		if val > d.High {
			d.High = val
		}
	} else {
		d.Low = val
		d.High = val
	}
	d.Recent = val
	d.Sum += val

	index := d.Count
	d.Count += 1

	if index < reservoirSize {
		d.reservoir[index] = float32(val)
		d.sorted = false
	} else {
		// fast, but kind of biased. probably okay
		j := d.lcg.Uint64() % uint64(d.Count)
		if j < reservoirSize {
			d.reservoir[int(j)] = float32(val)
			d.sorted = false
		}
	}
}

func (d *IntDist) Average() int64 {
	if d.Count > 0 {
		return d.Sum / d.Count
	}
	return 0
}

func (d *IntDist) Query(quantile float64) int64 {
	if quantile <= 0 {
		return d.Low
	}
	if quantile >= 1 {
		return d.High
	}

	rlen := int(reservoirSize)
	if int64(rlen) > d.Count {
		rlen = int(d.Count)
	}

	if rlen < 2 {
		return int64(d.reservoir[0])
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
	return int64(prior + diff*(float64(reservoir[idx+1])-prior))
}

func (d *IntDist) Copy() *IntDist {
	cp := *d
	return &cp
}

// DurationDist keeps statistics about time.Durations.
type DurationDist struct {
	Low, High time.Duration
	Recent    time.Duration
	Count     int64
	Sum       time.Duration
	reservoir [reservoirSize]float32
	lcg       lcg
	sorted    bool
}

func NewDurationDist() *DurationDist {
	return &DurationDist{lcg: newLCG()}
}

func (d *DurationDist) Insert(val time.Duration) {
	if d.Count != 0 {
		if val < d.Low {
			d.Low = val
		}
		if val > d.High {
			d.High = val
		}
	} else {
		d.Low = val
		d.High = val
	}
	d.Recent = val
	d.Sum += val

	index := d.Count
	d.Count += 1

	if index < reservoirSize {
		d.reservoir[index] = float32(val)
		d.sorted = false
	} else {
		// fast, but kind of biased. probably okay
		j := d.lcg.Uint64() % uint64(d.Count)
		if j < reservoirSize {
			d.reservoir[int(j)] = float32(val)
			d.sorted = false
		}
	}
}

func (d *DurationDist) Average() time.Duration {
	if d.Count > 0 {
		return time.Duration(int64(d.Sum) / d.Count)
	}
	return 0
}

func (d *DurationDist) Query(quantile float64) time.Duration {
	if quantile <= 0 {
		return d.Low
	}
	if quantile >= 1 {
		return d.High
	}

	rlen := int(reservoirSize)
	if int64(rlen) > d.Count {
		rlen = int(d.Count)
	}

	if rlen < 2 {
		return time.Duration(d.reservoir[0])
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

func (d *DurationDist) Copy() *DurationDist {
	cp := *d
	return &cp
}

// FloatDist keeps statistics about floats.
type FloatDist struct {
	Low, High float64
	Recent    float64
	Count     int64
	Sum       float64
	reservoir [reservoirSize]float32
	lcg       lcg
	sorted    bool
}

func NewFloatDist() *FloatDist {
	return &FloatDist{lcg: newLCG()}
}

func (d *FloatDist) Insert(val float64) {
	if val != val {
		// NaN
		return
	}
	if d.Count != 0 {
		if val < d.Low {
			d.Low = val
		}
		if val > d.High {
			d.High = val
		}
	} else {
		d.Low = val
		d.High = val
	}
	d.Recent = val
	d.Sum += val

	index := d.Count
	d.Count += 1

	if index < reservoirSize {
		d.reservoir[index] = float32(val)
		d.sorted = false
	} else {
		// fast, but kind of biased. probably okay
		j := d.lcg.Uint64() % uint64(d.Count)
		if j < reservoirSize {
			d.reservoir[int(j)] = float32(val)
			d.sorted = false
		}
	}
}

func (d *FloatDist) Average() float64 {
	if d.Count > 0 {
		return d.Sum / float64(d.Count)
	}
	return 0
}

func (d *FloatDist) Query(quantile float64) float64 {
	if quantile <= 0 {
		return d.Low
	}
	if quantile >= 1 {
		return d.High
	}

	rlen := int(reservoirSize)
	if int64(rlen) > d.Count {
		rlen = int(d.Count)
	}

	if rlen < 2 {
		return float64(d.reservoir[0])
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
	return float64(prior + diff*(float64(reservoir[idx+1])-prior))
}

func (d *FloatDist) Copy() *FloatDist {
	cp := *d
	return &cp
}

type float32Slice []float32

func (p float32Slice) Len() int      { return len(p) }
func (p float32Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p float32Slice) Less(i, j int) bool {
	// N.B.: usually, float comparisons should check if either value is NaN, but
	// in this package's usage, they never are here.
	return p[i] < p[j]
}
