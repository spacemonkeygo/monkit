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
	"reflect"
	"strings"
)

// StatSource represents anything that can return named floating point values.
type StatSource interface {
	Stats(cb func(name string, val float64))
}

type StatSourceFunc func(cb func(name string, val float64))

func (f StatSourceFunc) Stats(cb func(name string, val float64)) { f(cb) }

type FilterableStatSource interface {
	StatSource
	FilteredStats(prefix string, cb func(name string, val float64))
}

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

	val = deref(val)

	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		return StatSourceFunc(func(cb func(name string, val float64)) {})
	}
	return StatSourceFunc(func(cb func(name string, val float64)) {
		for i := 0; i < typ.NumField(); i++ {
			field := val.Field(i)

			field = deref(field)

			field_type := field.Type()

			if field_type.Kind() == reflect.Struct && field.CanInterface() {
				child_source := StatSourceFromStruct(field.Interface())
				Prefix(typ.Field(i).Name+".", child_source).Stats(cb)
			} else if field_type.ConvertibleTo(f64Type) {
				cb(typ.Field(i).Name, field.Convert(f64Type).Float())
			}
		}
	})
}

// if val is a pointer, deref until it isn't
func deref(val reflect.Value) reflect.Value {
	for val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
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

// Filter takes a StatSource callback and returns a new StatSource callback
// that only passes through names that start with the given prefix. Keep in
// mind that it may be more efficient to attempt to type-assert your
// StatSource to a FilterableStatSource first.
func Filter(prefix string, cb func(string, float64)) func(string, float64) {
	return func(name string, val float64) {
		if strings.HasPrefix(name, prefix) {
			cb(name, val)
		}
	}
}
