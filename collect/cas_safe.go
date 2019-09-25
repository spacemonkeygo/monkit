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

// +build appengine

package collect

import (
	"sync"

	"github.com/spacemonkeygo/monkit/v3"
)

// TODO(jeff): make this mutex smaller scoped, perhaps based on the arguments
// to compare and swap?
var bigHonkinMutex sync.Mutex

func loadSpan(addr **monkit.Span) (s *monkit.Span) {
	bigHonkinMutex.Lock()
	s = *addr
	bigHonkinMutex.Unlock()
	return s
}

func swapSpan(addr **monkit.Span, new *monkit.Span) *monkit.Span {
	bigHonkinMutex.Lock()
	prev := *addr
	*addr = new
	bigHonkinMutex.Unlock()
	return prev
}

func compareAndSwapSpan(addr **monkit.Span, old, new *monkit.Span) bool {
	bigHonkinMutex.Lock()
	val := *addr
	if val == old {
		*addr = new
		bigHonkinMutex.Unlock()
		return true
	}
	bigHonkinMutex.Unlock()
	return false
}
