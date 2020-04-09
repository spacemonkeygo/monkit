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
	"sync"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/monotime"
)

var (
	startTime = monotime.Now()
)

// Process returns a StatSource including generic process data, such as
// the process uptime, and a crc of the executing binary if possible. Also
// includes a 'control' value so data collectors can accurately count how many
// unique running processes being monitored there are. Not expected to be
// called directly, as this StatSource is added by Register.
func Process() monkit.StatSource {
	return monkit.StatSourceFunc(func(cb func(key monkit.SeriesKey, field string, val float64)) {
		cb(monkit.NewSeriesKey("process"), "control", 1)
		c, err := processCRC()
		if err == nil {
			cb(monkit.NewSeriesKey("process"), "crc", float64(c))
		}
		cb(monkit.NewSeriesKey("process"), "uptime", time.Since(startTime).Seconds())
	})
}

var crcCache struct {
	once sync.Once
	crc  uint32
	err  error
}

func processCRC() (uint32, error) {
	crcCache.once.Do(func() {
		crcCache.crc, crcCache.err = getProcessCRC()
	})
	return crcCache.crc, crcCache.err
}

func getProcessCRC() (uint32, error) {
	fh, err := openProc()
	if err != nil {
		return 0, err
	}
	defer fh.Close()
	c := crc32.NewIEEE()
	_, err = io.Copy(c, fh)
	return c.Sum32(), err
}

func init() { registrations = append(registrations, Process()) }
