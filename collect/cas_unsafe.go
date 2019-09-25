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

// +build !appengine

package collect

import (
	"sync/atomic"
	"unsafe"

	"github.com/spacemonkeygo/monkit/v3"
)

func loadSpan(addr **monkit.Span) (s *monkit.Span) {
	return (*monkit.Span)(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(addr))))
}

func swapSpan(addr **monkit.Span, new *monkit.Span) *monkit.Span {
	return (*monkit.Span)(atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(addr)),
		unsafe.Pointer(new)))
}

func compareAndSwapSpan(addr **monkit.Span, old, new *monkit.Span) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(addr)),
		unsafe.Pointer(old),
		unsafe.Pointer(new))
}
