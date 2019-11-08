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

package environment

import (
	"runtime"
	"runtime/debug"

	"github.com/spacemonkeygo/monkit/v3"
)

// Runtime returns a StatSource that includes information gathered from the
// Go runtime, including the number of goroutines currently running, and
// other live memory data. Not expected to be called directly, as this
// StatSource is added by Register.
func Runtime() monkit.StatSource {
	durDist := monkit.NewDurationDist(monkit.NewSeriesKey("runtime_gcstats"))
	lastNumGC := int64(0)

	return monkit.StatSourceFunc(func(cb func(key monkit.SeriesKey, field string, val float64)) {
		cb(monkit.NewSeriesKey("goroutines"), "count", float64(runtime.NumGoroutine()))

		{
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			monkit.StatSourceFromStruct(monkit.NewSeriesKey("runtime_memstats"), stats).Stats(cb)
		}

		{
			var stats debug.GCStats
			debug.ReadGCStats(&stats)
			if lastNumGC != stats.NumGC && len(stats.Pause) > 0 {
				durDist.Insert(stats.Pause[0])
			}
			durDist.Stats(cb)
		}
	})
}

func init() { registrations = append(registrations, Runtime()) }
