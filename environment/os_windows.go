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
	"syscall"
	"unsafe"
)

var (
	getProcessHandleCount = kernel32.MustFindProc("GetProcessHandleCount")
)

func fdCount() (count int, err error) {
	h, err := syscall.GetCurrentProcess()
	if err != nil {
		return 0, err
	}
	var procCount uint32
	r1, _, err := getProcessHandleCount.Call(
		uintptr(h), uintptr(unsafe.Pointer(&procCount)))
	if r1 == 0 {
		// if r1 == 0, then GetProcessHandleCount failed and err will be set to
		// the formatted string for whatever GetLastError() returns.
		return 0, err
	}
	return int(procCount), nil
}
