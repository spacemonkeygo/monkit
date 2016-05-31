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

package monkit

import (
	"sync"
	"sync/atomic"
)

// IntVal is a convenience wrapper around an IntDist. Constructed using
// NewIntVal, though it's expected usage is like:
//
//   var mon = monkit.Package()
//
//   func MyFunc() {
//     ...
//     mon.IntVal("size").Observe(val)
//     ...
//   }
//
type IntVal struct {
	mtx  sync.Mutex
	dist IntDist
}

// NewIntVal creates an IntVal
func NewIntVal() (v *IntVal) {
	v = &IntVal{}
	initIntDist(&v.dist)
	return v
}

// Observe observes an integer value
func (v *IntVal) Observe(val int64) {
	v.mtx.Lock()
	v.dist.Insert(val)
	v.mtx.Unlock()
}

// Stats implements the StatSource interface.
func (v *IntVal) Stats(cb func(name string, val float64)) {
	v.mtx.Lock()
	vd := v.dist
	min, max, recent, sum, count := vd.Low, vd.High, vd.Recent, vd.Sum, vd.Count
	v.mtx.Unlock()
	if count > 0 {
		cb("avg", float64(sum/count))
	}
	cb("count", float64(count))
	cb("max", float64(max))
	cb("min", float64(min))
	cb("recent", float64(recent))
	cb("sum", float64(sum))
}

// Quantile returns an estimate of the requested quantile of observed values.
// 0 <= quantile <= 1
func (v *IntVal) Quantile(quantile float64) (rv int64) {
	v.mtx.Lock()
	rv = v.dist.Query(quantile)
	v.mtx.Unlock()
	return rv
}

// FloatVal is a convenience wrapper around an FloatDist. Constructed using
// NewFloatVal, though it's expected usage is like:
//
//   var mon = monkit.Package()
//
//   func MyFunc() {
//     ...
//     mon.FloatVal("size").Observe(val)
//     ...
//   }
//
type FloatVal struct {
	mtx  sync.Mutex
	dist FloatDist
}

// NewFloatVal creates a FloatVal
func NewFloatVal() (v *FloatVal) {
	v = &FloatVal{}
	initFloatDist(&v.dist)
	return v
}

// Observe observes an floating point value
func (v *FloatVal) Observe(val float64) {
	v.mtx.Lock()
	v.dist.Insert(val)
	v.mtx.Unlock()
}

// Stats implements the StatSource interface.
func (v *FloatVal) Stats(cb func(name string, val float64)) {
	v.mtx.Lock()
	vd := v.dist
	min, max, recent, sum, count := vd.Low, vd.High, vd.Recent, vd.Sum, vd.Count
	v.mtx.Unlock()
	if count > 0 {
		cb("avg", sum/float64(count))
	}
	cb("count", float64(count))
	cb("max", max)
	cb("min", min)
	cb("recent", recent)
	cb("sum", sum)
}

// Quantile returns an estimate of the requested quantile of observed values.
// 0 <= quantile <= 1
func (v *FloatVal) Quantile(quantile float64) (rv float64) {
	v.mtx.Lock()
	rv = v.dist.Query(quantile)
	v.mtx.Unlock()
	return rv
}

// BoolVal keeps statistics about boolean values. It keeps the number of trues,
// number of falses, and the disposition (number of trues minus number of
// falses). Constructed using NewBoolVal, though it's expected usage is like:
//
//   var mon = monkit.Package()
//
//   func MyFunc() {
//     ...
//     mon.BoolVal("flipped").Observe(bool)
//     ...
//   }
//
type BoolVal struct {
	trues  int64
	falses int64
}

// NewBoolVal creates a BoolVal
func NewBoolVal() *BoolVal {
	return &BoolVal{}
}

// Observe observes a boolean value
func (v *BoolVal) Observe(val bool) {
	if val {
		atomic.AddInt64(&v.trues, 1)
	} else {
		atomic.AddInt64(&v.falses, 1)
	}
}

// Stats implements the StatSource interface.
func (v *BoolVal) Stats(cb func(name string, val float64)) {
	trues := atomic.LoadInt64(&v.trues)
	falses := atomic.LoadInt64(&v.falses)
	cb("disposition", float64(trues-falses))
	cb("false", float64(falses))
	cb("true", float64(trues))
}
