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
	"io"

	"gopkg.in/spacemonkeygo/monitor.v2"
)

func StatsText(r *monitor.Registry, w io.Writer) (err error) {
	r.Stats(func(name string, val float64) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, "%s\t%f\n", name, val)
	})
	return err
}

func StatsJSON(r *monitor.Registry, w io.Writer) (err error) {
	lw := newListWriter(w)
	r.Stats(func(name string, val float64) {
		lw.elem([]interface{}{name, val})
	})
	return lw.done()
}
