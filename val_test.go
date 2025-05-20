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

package monkit

import (
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestRawValConcurrentSafe(t *testing.T) {
	rv := NewRawVal(NewSeriesKey("test"), Sum, Count)

	var group errgroup.Group
	defer func() { _ = group.Wait() }()
	for i := 0; i < 10; i++ {
		group.Go(func() error {
			if i%2 == 0 {
				rv.Observe(1.0)
			} else {
				rv.Stats(func(key SeriesKey, field string, val float64) {})
			}
			return nil
		})
	}
}
