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
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestTagSet(t *testing.T) {
	assert := func(ts *TagSet, key string, value string, ok bool) {
		t.Helper()
		gotValue, gotOk := ts.all[key]
		if gotValue != value || gotOk != ok {
			t.Fatalf("exp:%q != got:%q || exp:%v != got:%v", value, gotValue, ok, gotOk)
		}
	}

	ts0 := new(TagSet)
	ts0 = ts0.Set("k0", "0")
	ts1 := ts0.SetAll(nil)
	ts1 = ts1.Set("k0", "1")
	ts2 := ts0.SetAll(nil)
	ts2 = ts2.Set("k1", "2")
	ts0 = ts0.Set("k0", "3")
	ts3 := ts0.SetAll(nil)
	ts3 = ts3.Set("k0", "4")
	ts0 = ts0.Set("k0", "5")

	assert(ts0, "k0", "5", true)
	assert(ts0, "k1", "", false)
	assert(ts1, "k0", "1", true)
	assert(ts1, "k1", "", false)
	assert(ts2, "k0", "0", true)
	assert(ts2, "k1", "2", true)
	assert(ts3, "k0", "4", true)
	assert(ts3, "k1", "", false)

	t.Log(ts0)
	t.Log(ts1)
	t.Log(ts2)
	t.Log(ts3)
}

func TestTagSetFuzz(t *testing.T) {
	ts, idx := new(TagSet), 0
	tagSets := []*TagSet{ts}
	expected := []map[string]string{{}}

	for i := 0; i < 10000; i++ {
		switch rand.Intn(10) {
		case 0, 1, 2, 3, 4, 5, 6, 7:
			key, value := fmt.Sprint(rand.Intn(10)), fmt.Sprint(rand.Intn(10))
			ts = ts.Set(key, value)
			tagSets[idx] = ts
			expected[idx][key] = value

		case 8:
			idx = rand.Intn(len(expected))
			ts = tagSets[idx]

		case 9:
			cloned := make(map[string]string, len(expected[idx]))
			for key, value := range expected[idx] {
				cloned[key] = value
			}

			ts = ts.SetAll(nil)
			tagSets = append(tagSets, ts)
			expected = append(expected, cloned)
			idx = len(tagSets) - 1
		}
	}

	for i := range tagSets {
		if got := tagSets[i].all; !reflect.DeepEqual(expected[i], got) {
			t.Fatal("mismatch: exp:", expected[i], "got:", got)
		}
	}
}
