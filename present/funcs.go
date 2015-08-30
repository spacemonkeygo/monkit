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
	"bytes"
	"fmt"
	"io"
	"time"

	"gopkg.in/spacemonkeygo/monitor.v2"
)

func formatDist(querier func(float64) float64, indent string) (result string) {
	for _, q := range monitor.ObservedQuantiles {
		result += fmt.Sprintf("%s%.02f: %s\n", indent, q, time.Duration(
			querier(q)*float64(time.Second)))
	}
	return result
}

func FuncsDot(r *monitor.Registry, w io.Writer) (err error) {
	_, err = fmt.Fprintf(w, "digraph G {\n node [shape=box];\n")
	if err != nil {
		return err
	}
	r.Funcs(func(f *monitor.Func) {
		if err != nil {
			return
		}
		success := f.Success()
		panics := f.Panics()

		var err_out bytes.Buffer
		total_errors := int64(0)
		for errname, count := range f.Errors() {
			_, err = fmt.Fprint(&err_out, escapeDotLabel("error %s: %d\n", errname,
				count))
			if err != nil {
				return
			}
			total_errors += count
		}

		_, err = fmt.Fprintf(w, " f%d [label=\"%s", f.Id(),
			escapeDotLabel("%s\ncurrent: %d, success: %d, errors: %d, panics: %d\n",
				f.FullName(), f.Current(), success, total_errors, panics))
		if err != nil {
			return
		}

		_, err = err_out.WriteTo(w)
		if err != nil {
			return
		}

		if success > 0 {
			_, err = fmt.Fprint(w, escapeDotLabel(
				"success times:\n%s", formatDist(f.SuccessTimeQuantile, "        ")))
			if err != nil {
				return
			}
		}

		if total_errors+panics > 0 {
			_, err = fmt.Fprint(w, escapeDotLabel(
				"failure times:\n%s", formatDist(f.FailureTimeQuantile, "        ")))
			if err != nil {
				return
			}
		}

		_, err = fmt.Fprint(w, "\"];\n")
		if err != nil {
			return
		}

		f.Parents(func(parent *monitor.Func) {
			if err != nil {
				return
			}
			if parent != nil {
				_, err = fmt.Fprintf(w, " f%d -> f%d;\n", parent.Id(), f.Id())
				if err != nil {
					return
				}
			} else {
				_, err = fmt.Fprintf(w, " r%d [label=\"entry\"];\n r%d -> f%d;\n",
					f.Id(), f.Id(), f.Id())
				if err != nil {
					return
				}
			}
		})
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "}\n")
	return err
}

func FuncsText(r *monitor.Registry, w io.Writer) (err error) {
	r.Funcs(func(f *monitor.Func) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, "[%d] %s\n  parents: ", f.Id(), f.FullName())
		if err != nil {
			return
		}
		printed := false
		f.Parents(func(parent *monitor.Func) {
			if err != nil {
				return
			}
			if printed {
				_, err = fmt.Fprint(w, ", ")
				if err != nil {
					return
				}
			} else {
				printed = true
			}
			if parent != nil {
				_, err = fmt.Fprintf(w, "%d", parent.Id())
				if err != nil {
					return
				}
			} else {
				_, err = fmt.Fprintf(w, "entry")
				if err != nil {
					return
				}
			}
		})
		var err_out bytes.Buffer
		total_errors := int64(0)
		for errname, count := range f.Errors() {
			_, err = fmt.Fprintf(&err_out, "  error %s: %d\n", errname, count)
			if err != nil {
				return
			}
			total_errors += count
		}
		_, err = fmt.Fprintf(w,
			"\n  current: %d, success: %d, errors: %d, panics: %d\n",
			f.Current(), f.Success(), total_errors, f.Panics())
		if err != nil {
			return
		}
		_, err = err_out.WriteTo(w)
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, "  success times:\n%s  failure times:\n%s\n",
			formatDist(f.SuccessTimeQuantile, "    "),
			formatDist(f.FailureTimeQuantile, "    "))
	})
	return err
}

func FuncsJSON(r *monitor.Registry, w io.Writer) (err error) {
	lw := newListWriter(w)
	r.Funcs(func(f *monitor.Func) {
		lw.elem(formatFunc(f))
	})
	return lw.done()
}
