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
	"time"
)

var (
	ObservedQuantiles = []float64{0, 1}
)

// dist is not threadsafe
type dist struct {
	low, high   time.Duration
	recent      time.Duration
	totalValues int64
	sum         time.Duration
}

func newDist() dist {
	return dist{}
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
	d.totalValues += 1
	d.sum += val
}

func (d *dist) Average() time.Duration {
	if d.totalValues > 0 {
		return d.sum / time.Duration(d.totalValues)
	} else {
		return 0
	}
}
func (d *dist) Query(quantile float64) (rv time.Duration) {
	if quantile < .5 {
		return d.low
	} else {
		return d.high
	}
}

func (d *dist) Recent() time.Duration { return d.recent }
