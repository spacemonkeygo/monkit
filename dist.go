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
	"github.com/bmizerany/perks/quantile"
)

var (
	// need to make sure to have 0, .5, and 1
	ObservedQuantiles = []float64{0, .25, .5, .75, .90, .95, .99, 1}
)

// dist is not threadsafe
type dist struct {
	q      *quantile.Stream
	recent float64
}

func newDist() dist {
	return dist{q: quantile.NewTargeted(ObservedQuantiles...)}
}

func (d *dist) Stats() (min, med, max, recent float64) {
	return d.q.Query(0), d.q.Query(.5), d.q.Query(1), d.recent
}

func (d *dist) Insert(val float64) {
	d.q.Insert(val)
	d.recent = val
}

func (d *dist) Query(quantile float64) (rv float64) {
	return d.q.Query(quantile)
}

func (d *dist) Recent() float64 { return d.recent }
