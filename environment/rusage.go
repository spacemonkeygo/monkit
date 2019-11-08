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

// +build !windows

package environment

import (
	"syscall"

	"github.com/spacemonkeygo/monkit/v3"
)

// Rusage returns a StatSource that provides as many statistics as possible
// gathered from the Rusage syscall. Not expected to be called directly, as
// this StatSource is added by Register.
func Rusage() monkit.StatSource {
	return monkit.StatSourceFunc(func(cb func(key monkit.SeriesKey, field string, val float64)) {
		var rusage syscall.Rusage
		err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage)
		if err == nil {
			monkit.StatSourceFromStruct(monkit.NewSeriesKey("rusage"), &rusage).Stats(cb)
		}
	})
}

func init() { registrations = append(registrations, Rusage()) }
