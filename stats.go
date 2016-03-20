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

type StatSource interface {
	Stats(cb func(name string, val float64))
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
