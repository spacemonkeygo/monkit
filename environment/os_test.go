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

//go:build (unix || windows) && !darwin
// +build unix windows
// +build !darwin

package environment

import "testing"

func TestFDCount(t *testing.T) {
	count, err := fdCount()
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("fd count should not be zero")
	}
}
