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

package present

import (
	"fmt"
)

func escapeDotLabel(format string, args ...interface{}) string {
	val := fmt.Sprintf(format, args...)
	var rv []byte
	for _, b := range []byte(val) {
		switch {
		case 'A' <= b && b <= 'Z', 'a' <= b && b <= 'z', '0' <= b && b <= '9',
			128 <= b, ' ' == b:
			rv = append(rv, b)
		case b == '\n':
			rv = append(rv, []byte(`\l`)...)
		default:
			rv = append(rv, []byte(fmt.Sprintf("&#%d;", int(b)))...)
		}
	}
	return string(rv)
}
