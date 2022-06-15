// Copyright (C) 2022 Storj Labs, Inc.
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

package collect

import (
	"sync"

	"github.com/spacemonkeygo/monkit/v3"
)

// ObserveAllTraces will register collector with all traces present and future
// on the given monkit.Registry until cancel is called.
func ObserveAllTraces(r *monkit.Registry, collector monkit.SpanObserver) (cancel func()) {
	var mtx sync.Mutex
	var cancelers []func()
	var stopping bool
	existingTraces := map[*monkit.Trace]bool{}

	mainCanceler := r.ObserveTraces(func(t *monkit.Trace) {
		mtx.Lock()
		defer mtx.Unlock()
		if existingTraces[t] || stopping {
			return
		}
		existingTraces[t] = true
		cancelers = append(cancelers, t.ObserveSpans(collector))
	})

	// pick up live traces we can find
	r.RootSpans(func(s *monkit.Span) {
		mtx.Lock()
		defer mtx.Unlock()
		t := s.Trace()
		if existingTraces[t] || stopping {
			return
		}
		existingTraces[t] = true
		cancelers = append(cancelers, t.ObserveSpans(collector))
	})

	return func() {
		mainCanceler()
		mtx.Lock()
		defer mtx.Unlock()
		stopping = true
		for _, canceler := range cancelers {
			canceler()
		}
	}
}
