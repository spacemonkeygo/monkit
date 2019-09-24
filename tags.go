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

type TagSet struct {
	all map[string]string
}

func (t *TagSet) Set(key, value string) *TagSet {
	return t.SetAll(map[string]string{key: value})
}

func (t *TagSet) SetAll(kvs map[string]string) *TagSet {
	all := make(map[string]string, len(t.all)+len(kvs))
	for key, value := range t.all {
		all[key] = value
	}
	for key, value := range kvs {
		all[key] = value
	}
	return &TagSet{all: all}
}

// String returns a string form of the tag set suitable for sending to influxdb.
func (t *TagSet) String() string {
	type kv struct {
		key   string
		value string
	}
	var kvs []kv

	for key, value := range t.all {
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
