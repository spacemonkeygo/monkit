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
	"hash/crc32"
	"io"

	"github.com/spacemonkeygo/monotime"
	"gopkg.in/spacemonkeygo/monitor.v2"
)

var (
	startTime = monotime.Monotonic()
)

func Process() monitor.StatSource {
	return monitor.StatSourceFunc(func(cb func(name string, val float64)) {
		cb("control", 1)
		c, err := processCRC()
		if err == nil {
			cb("crc", float64(c))
		}
		cb("uptime", (monotime.Monotonic() - startTime).Seconds())
	})
}

func processCRC() (uint32, error) {
	fh, err := openProc()
	if err != nil {
		return 0, err
	}
	defer fh.Close()
	c := crc32.NewIEEE()
	_, err = io.Copy(c, fh)
	return c.Sum32(), err
}

func init() {
	registrations["process"] = Process()
}
