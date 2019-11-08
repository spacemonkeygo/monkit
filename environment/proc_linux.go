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
	"fmt"
	"io/ioutil"

	"github.com/spacemonkeygo/monkit/v3"
)

func proc(cb func(key monkit.SeriesKey, field string, val float64)) {
	var stat procSelfStat
	err := readProcSelfStat(&stat)
	if err == nil {
		monkit.StatSourceFromStruct(monkit.NewSeriesKey("proc_stat"), &stat).Stats(cb)
	}

	var statm procSelfStatm
	err = readProcSelfStatm(&statm)
	if err == nil {
		monkit.StatSourceFromStruct(monkit.NewSeriesKey("proc_statm"), &statm).Stats(cb)
	}
}

type procSelfStat struct {
	Pid                 int64
	Comm                string
	State               byte
	Ppid                int64
	Pgrp                int64
	Session             int64
	TtyNr               int64
	Tpgid               int64
	Flags               uint64
	Minflt              uint64
	Cminflt             uint64
	Majflt              uint64
	Cmajflt             uint64
	Utime               uint64
	Stime               uint64
	Cutime              int64
	Cstime              int64
	Priority            int64
	Nice                int64
	NumThreads          int64
	Itrealvalue         int64
	Starttime           uint64
	Vsize               uint64
	Rss                 int64
	Rsslim              uint64
	Startcode           uint64
	Endcode             uint64
	Startstack          uint64
	Kstkesp             uint64
	Kstkeip             uint64
	Signal              uint64
	Blocked             uint64
	Sigignore           uint64
	Sigcatch            uint64
	Wchan               uint64
	Nswap               uint64
	Cnswap              uint64
	ExitSignal          int64
	Processor           int64
	RtPriority          uint64
	Policy              uint64
	DelayAcctBlkioTicks uint64
	GuestTime           uint64
	CguestTime          int64
}

func readProcSelfStat(s *procSelfStat) error {
	data, err := ioutil.ReadFile("/proc/self/stat")
	if err != nil {
		return err
	}
	_, err = fmt.Sscanf(string(data), "%d %s %c %d %d %d %d %d %d %d %d "+
		"%d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d "+
		"%d %d %d %d %d %d %d %d %d %d", &s.Pid, &s.Comm, &s.State, &s.Ppid,
		&s.Pgrp, &s.Session, &s.TtyNr, &s.Tpgid, &s.Flags, &s.Minflt, &s.Cminflt,
		&s.Majflt, &s.Cmajflt, &s.Utime, &s.Stime, &s.Cutime, &s.Cstime,
		&s.Priority, &s.Nice, &s.NumThreads, &s.Itrealvalue, &s.Starttime,
		&s.Vsize, &s.Rss, &s.Rsslim, &s.Startcode, &s.Endcode, &s.Startstack,
		&s.Kstkesp, &s.Kstkeip, &s.Signal, &s.Blocked, &s.Sigignore, &s.Sigcatch,
		&s.Wchan, &s.Nswap, &s.Cnswap, &s.ExitSignal, &s.Processor, &s.RtPriority,
		&s.Policy, &s.DelayAcctBlkioTicks, &s.GuestTime, &s.CguestTime)
	return err
}

type procSelfStatm struct {
	Size     int
	Resident int
	Share    int
	Text     int
	Lib      int
	Data     int
	Dt       int
}

func readProcSelfStatm(s *procSelfStatm) error {
	data, err := ioutil.ReadFile("/proc/self/statm")
	if err != nil {
		return err
	}
	_, err = fmt.Sscanf(string(data), "%d %d %d %d %d %d %d", &s.Size,
		&s.Resident, &s.Share, &s.Text, &s.Lib, &s.Data, &s.Dt)
	return err
}
