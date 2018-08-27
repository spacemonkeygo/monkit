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

// +build cgo

package environment

import (
	"fmt"
	"os"
	"unsafe"
)

// #include <mach-o/dyld.h>
// #include <stdlib.h>
import "C"

func openProc() (*os.File, error) {
	const bufsize = 4096

	buf := (*C.char)(C.malloc(bufsize))
	defer C.free(unsafe.Pointer(buf))

	size := C.uint32_t(bufsize)
	if rc := C._NSGetExecutablePath(buf, &size); rc != 0 {
		return nil, fmt.Errorf("error in cgo call to get path: %d", rc)
	}

	return os.Open(C.GoString(buf))
}
