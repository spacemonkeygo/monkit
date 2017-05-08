// Copyright (C) 2016 Space Monkey, Inc.
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
//
// WARNING: THE NON-M4 VERSIONS OF THIS FILE ARE GENERATED BY GO GENERATE!
//          ONLY MAKE CHANGES TO THE M4 FILE
//

package monkit

import (
	"sort"
)

// IntDist keeps statistics about values such as
// low/high/recent/average/quantiles. Not threadsafe. Construct with
// NewIntDist(). Fields are expected to be read from but not written to.
type IntDist struct {
	// Low and High are the lowest and highest values observed since
	// construction or the last reset.
	Low, High int64

	// Recent is the last observed value.
	Recent int64

	// Count is the number of observed values since construction or the last
	// reset.
	Count int64

	// Sum is the sum of all the observed values since construction or the last
	// reset.
	Sum int64

	reservoir [ReservoirSize]float32
	lcg       lcg
	sorted    bool
}

func initIntDist(v *IntDist) {
	v.lcg = newLCG()
}

// NewIntDist creates a distribution of int64s.
func NewIntDist() (d *IntDist) {
	d = &IntDist{}
	initIntDist(d)
	return d
}

// Insert adds a value to the distribution, updating appropriate values.
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

	if index < ReservoirSize {
		d.reservoir[index] = float32(val)
		d.sorted = false
	} else {
		window := d.Count
		// careful, the capitalization of Window is important
		if Window > 0 && window > Window {
			window = Window
		}
		// fast, but kind of biased. probably okay
		j := d.lcg.Uint64() % uint64(window)
		if j < ReservoirSize {
			d.reservoir[int(j)] = float32(val)
			d.sorted = false
		}
	}
}

// FullAverage calculates and returns the average of all inserted values.
func (d *IntDist) FullAverage() int64 {
	if d.Count > 0 {
		return d.Sum / int64(d.Count)
	}
	return 0
}

// ReservoirAverage calculates the average of the current reservoir.
func (d *IntDist) ReservoirAverage() int64 {
	amount := ReservoirSize
	if d.Count < int64(amount) {
		amount = int(d.Count)
	}
	if amount <= 0 {
		return 0
	}
	var sum float32
	for i := 0; i < amount; i++ {
		sum += d.reservoir[i]
	}
	return int64(sum / float32(amount))
}

// Query will return the approximate value at the given quantile from the
// reservoir, where 0 <= quantile <= 1.
func (d *IntDist) Query(quantile float64) int64 {
	rlen := int(ReservoirSize)
	if int64(rlen) > d.Count {
		rlen = int(d.Count)
	}

	if rlen < 2 {
		return int64(d.reservoir[0])
	}

	reservoir := d.reservoir[:rlen]
	if !d.sorted {
		sort.Sort(float32Slice(reservoir))
		d.sorted = true
	}

	if quantile <= 0 {
		return int64(reservoir[0])
	}
	if quantile >= 1 {
		return int64(reservoir[rlen-1])
	}

	idx_float := quantile * float64(rlen-1)
	idx := int(idx_float)

	diff := idx_float - float64(idx)
	prior := float64(reservoir[idx])
	return int64(prior + diff*(float64(reservoir[idx+1])-prior))
}

// Copy returns a full copy of the entire distribution.
func (d *IntDist) Copy() *IntDist {
	cp := *d
	cp.lcg = newLCG()
	return &cp
}

func (d *IntDist) Reset() {
	d.Low, d.High, d.Recent, d.Count, d.Sum = 0, 0, 0, 0, 0
	// resetting count will reset the quantile reservoir
}

func (d *IntDist) Stats(cb func(name string, val float64)) {
	count := d.Count
	cb("count", float64(count))
	if count > 0 {
		cb("sum", d.toFloat64(d.Sum))
		cb("min", d.toFloat64(d.Low))
		cb("avg", d.toFloat64(d.FullAverage()))
		cb("max", d.toFloat64(d.High))
		cb("rmin", d.toFloat64(d.Query(0)))
		cb("ravg", d.toFloat64(d.ReservoirAverage()))
		cb("r50", d.toFloat64(d.Query(.5)))
		cb("r90", d.toFloat64(d.Query(.9)))
		cb("rmax", d.toFloat64(d.Query(1)))
		cb("recent", d.toFloat64(d.Recent))
	}
}
