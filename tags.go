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
	"sort"
	"strings"
)

// cloneKVs clones the input key value map and returns it. If the input
// is nil, then the returned value is also nil.
func cloneKVs(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

// TagSet holds an unordered collection of string key value pairs. It has
// methods to cheaply clone and inspect.
type TagSet struct {
	parent *TagSet           // the parent of this tag set
	cow    bool              // if true, copy kvs before modifying it
	kvs    map[string]string // key, value pairs in this layer
	all    map[string]string // cached value
	str    string            // cached value
}

// copy returns a shallow copy of the tag set.
func (t *TagSet) copy() *TagSet {
	tc := *t
	return &tc
}

// Apply returns a copy of the tag set with all the values in ts set on
// the receiver tag set. It is O(1) in the number of keys in each tag set.
// Passing in nil will return a copy of the receiver.
func (t *TagSet) Apply(ts *TagSet) *TagSet {
	t.cow = true
	if ts != nil {
		ts.cow = true
	}
	return &TagSet{
		parent: t.copy(),
		cow:    true,
		kvs:    ts.getAll(),
	}
}

// Set associates the tag key with the given tag value.
func (t *TagSet) Set(key, value string) {
	if t.cow {
		t.kvs = cloneKVs(t.kvs)
		t.all = cloneKVs(t.all)
		t.cow = false
	}
	if t.kvs == nil {
		t.kvs = make(map[string]string)
	}

	t.str = ""
	t.kvs[key] = value
	if t.all != nil {
		t.all[key] = value
	}
}

func (t *TagSet) getAll() map[string]string {
	if t == nil {
		return nil
	}
	t.cacheAll()
	return t.all
}

func (t *TagSet) cacheAll() {
	if t == nil || t.all != nil {
		return
	}
	t.all = make(map[string]string)
	for key, value := range t.parent.getAll() {
		t.all[key] = value
	}
	for key, value := range t.kvs {
		t.all[key] = value
	}
}

// String returns a string form of the tag set suitable for sending to influxdb.
func (t *TagSet) String() string {
	if t.str == "" {
		t.str = t.cacheString()
	}
	return t.str
}

// cacheString caches a string representation of the tag set.
func (t *TagSet) cacheString() string {
	type kv struct {
		key   string
		value string
	}
	var kvs []kv

	for key, value := range t.parent.getAll() {
		if _, ok := t.kvs[key]; !ok {
			kvs = append(kvs, kv{key, value})
		}
	}
	for key, value := range t.kvs {
		kvs = append(kvs, kv{key, value})
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].key < kvs[j].key
	})

	var builder strings.Builder
	for i, kv := range kvs {
		if i > 0 {
			builder.WriteByte(',')
		}
		writeTag(&builder, kv.key)
		builder.WriteByte('=')
		writeTag(&builder, kv.value)
	}

	return builder.String()
}

// writeTag writes a tag key or value to the builder.
func writeTag(builder *strings.Builder, tag string) {
	if strings.IndexByte(tag, ',') == -1 &&
		strings.IndexByte(tag, '=') == -1 &&
		strings.IndexByte(tag, ' ') == -1 {

		builder.WriteString(tag)
		return
	}

	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' ||
			tag[i] == '=' ||
			tag[i] == ' ' {
			builder.WriteByte('\\')
		}
		builder.WriteByte(tag[i])
	}
}
