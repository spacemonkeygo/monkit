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
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
	"gopkg.in/spacemonkeygo/monitor.v2"
)

var (
	BadRequest = errors.NewClass("Bad Request", errhttp.SetStatusCode(400))
	NotFound   = errors.NewClass("Not Found", errhttp.SetStatusCode(404))
)

type Func func(io.Writer) error

func curry(reg *monitor.Registry,
	f func(*monitor.Registry, io.Writer) error) func(io.Writer) error {
	return func(w io.Writer) error {
		return f(reg, w)
	}
}

func FromRequest(reg *monitor.Registry, path string, query url.Values) (
	f Func, contentType string, err error) {
	first, rest := shift(path)
	second, _ := shift(rest)
	switch first {
	case "ps":
		switch second {
		case "", "text":
			return curry(reg, SpansText), "text/plain; charset=utf-8", nil
		case "dot":
			return curry(reg, SpansDot), "text/plain; charset=utf-8", nil
		case "json":
			return curry(reg, SpansJSON), "application/json; charset=utf-8", nil
		}

	case "funcs":
		switch second {
		case "", "text":
			return curry(reg, FuncsText), "text/plain; charset=utf-8", nil
		case "dot":
			return curry(reg, FuncsDot), "text/plain; charset=utf-8", nil
		case "json":
			return curry(reg, FuncsJSON), "application/json; charset=utf-8", nil
		}

	case "stats":
		switch second {
		case "", "text":
			return curry(reg, StatsText), "text/plain; charset=utf-8", nil
		case "json":
			return curry(reg, StatsJSON), "application/json; charset=utf-8", nil
		}
	case "trace":
		re, err := regexp.Compile(query.Get("regex"))
		if err != nil {
			return nil, "", BadRequest.New("invalid regex %#v: %v",
				query.Get("regex"), err)
		}
		preselect := true
		if query.Get("preselect") != "" {
			preselect, err = strconv.ParseBool(query.Get("preselect"))
			if err != nil {
				return nil, "", BadRequest.New("invalid preselect %#v: %v",
					query.Get("preselect"), err)
			}
		}
		matcher := func(f *monitor.Func) bool {
			return re.MatchString(f.FullName())
		}
		if preselect {
			funcs := map[*monitor.Func]bool{}
			reg.Funcs(func(f *monitor.Func) {
				if matcher(f) {
					funcs[f] = true
				}
			})
			if len(funcs) <= 0 {
				return nil, "", BadRequest.New("regex preselect matches 0 functions")
			}
			matcher = func(f *monitor.Func) bool { return funcs[f] }
		}
		spanMatcher := func(s *monitor.Span) bool { return matcher(s.Func()) }
		switch second {
		case "svg":
			return func(w io.Writer) error {
				return TraceQuerySVG(reg, w, spanMatcher)
			}, "image/svg+xml; charset=utf-8", nil
		case "json":
			return func(w io.Writer) error {
				return TraceQueryJSON(reg, w, spanMatcher)
			}, "application/json; charset=utf-8", nil
		}
	}
	return nil, "", NotFound.New("path not found: %s", path)
}

func shift(path string) (dir, left string) {
	path = strings.TrimLeft(path, "/")
	split := strings.Index(path, "/")
	if split == -1 {
		return path, ""
	}
	return path[:split], path[split:]
}
