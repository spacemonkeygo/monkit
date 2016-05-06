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
	"reflect"
)

// StatSource represents anything that can return named floating point values.
type StatSource interface {
	Stats(cb func(name string, val float64))
}

type StatSourceFunc func(cb func(name string, val float64))

func (f StatSourceFunc) Stats(cb func(name string, val float64)) { f(cb) }

// Collect takes something that implements the StatSource interface and returns
// a key/value map.
func Collect(mon StatSource) map[string]float64 {
	rv := make(map[string]float64)
	mon.Stats(func(name string, val float64) {
		rv[name] = val
	})
	return rv
}

var f64Type = reflect.TypeOf(float64(0))

// StatSourceFromStruct uses the reflect package to implement the Stats call
// across all float64-castable fields of the struct.
func StatSourceFromStruct(structData interface{}) StatSource {
	val := reflect.ValueOf(structData)
	for val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		return StatSourceFunc(func(cb func(name string, val float64)) {})
	}
	return StatSourceFunc(func(cb func(name string, val float64)) {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.Type.ConvertibleTo(f64Type) {
				cb(field.Name, val.Field(i).Convert(f64Type).Float())
			}
		}
	})
}

// Prefix takes a StatSource and returns a new StatSource where all names have
// the given prefix.
func Prefix(prefix string, source StatSource) StatSource {
	return StatSourceFunc(func(cb func(string, float64)) {
		source.Stats(func(name string, val float64) {
			cb(prefix+name, val)
		})
	})
}
