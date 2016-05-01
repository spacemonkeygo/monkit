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
	"sync/atomic"
)

type IntVal struct {
	mtx  sync.Mutex
	dist intDist
}

func newIntVal() StatSource {
	return &IntVal{dist: newIntDist()}
}

func (v *IntVal) Observe(val int64) {
	v.mtx.Lock()
	v.dist.Insert(val)
	v.mtx.Unlock()
}

func (v *IntVal) Stats(cb func(name string, val float64)) {
	v.mtx.Lock()
	min, avg, max, recent, sum := v.dist.Stats()
	v.mtx.Unlock()
	cb("avg", float64(avg))
	cb("max", float64(max))
	cb("min", float64(min))
	cb("recent", float64(recent))
	cb("sum", float64(sum))
}

func (v *IntVal) Quantile(quantile float64) (rv int64) {
	v.mtx.Lock()
	rv = v.dist.Query(quantile)
	v.mtx.Unlock()
	return rv
}

type FloatVal struct {
	mtx  sync.Mutex
	dist floatDist
}

func newFloatVal() StatSource {
	return &FloatVal{dist: newFloatDist()}
}

func (v *FloatVal) Observe(val float64) {
	v.mtx.Lock()
	v.dist.Insert(val)
	v.mtx.Unlock()
}

func (v *FloatVal) Stats(cb func(name string, val float64)) {
	v.mtx.Lock()
	min, avg, max, recent, sum := v.dist.Stats()
	v.mtx.Unlock()
	cb("avg", avg)
	cb("max", max)
	cb("min", min)
	cb("recent", recent)
	cb("sum", sum)
}

func (v *FloatVal) Quantile(quantile float64) (rv float64) {
	v.mtx.Lock()
	rv = v.dist.Query(quantile)
	v.mtx.Unlock()
	return rv
}

type BoolVal struct {
	trues  int64
	falses int64
}

func newBoolVal() StatSource {
	return &BoolVal{}
}

func (v *BoolVal) Observe(val bool) {
	if val {
		atomic.AddInt64(&v.trues, 1)
	} else {
		atomic.AddInt64(&v.falses, 1)
	}
}

func (v *BoolVal) Stats(cb func(name string, val float64)) {
	trues := atomic.LoadInt64(&v.trues)
	falses := atomic.LoadInt64(&v.falses)
	cb("disposition", float64(trues-falses))
	cb("false", float64(falses))
	cb("true", float64(trues))
}
