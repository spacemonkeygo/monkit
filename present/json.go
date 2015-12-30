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
	"encoding/json"
	"fmt"
	"io"
	"time"

	"gopkg.in/spacemonkeygo/monitor.v2"
)

type finishedSpan struct {
	Span     *monitor.Span
	Err      *error
	Panicked *bool
	Finish   *time.Time
}

func formatSpan(s *monitor.Span) interface{} {
	return formatFinishedSpan(&finishedSpan{Span: s})
}

func formatFinishedSpan(s *finishedSpan) interface{} {
	js := struct {
		Id       int64  `json:"id"`
		ParentId *int64 `json:"parent_id,omitempty"`
		Func     struct {
			Package string `json:"package"`
			Name    string `json:"name"`
		} `json:"func"`
		Trace struct {
			Id int64 `json:"id"`
		} `json:"trace"`
		Start       int64      `json:"start"`
		Finish      *int64     `json:"finish,omitempty"`
		Orphaned    bool       `json:"orphaned"`
		Err         *string    `json:"err,omitempty"`
		Panicked    *bool      `json:"panicked,omitempty"`
		Args        []string   `json:"args"`
		Annotations [][]string `json:"annotations"`
	}{}
	js.Id = s.Span.Id()
	if s.Span.Parent() != nil {
		parent_id := s.Span.Parent().Id()
		js.ParentId = &parent_id
	}
	js.Func.Package = s.Span.Func().Scope().Name()
	js.Func.Name = s.Span.Func().ShortName()
	js.Trace.Id = s.Span.Trace().Id()
	js.Start = s.Span.Start().UnixNano()
	if s.Finish != nil {
		finish := s.Finish.UnixNano()
		js.Finish = &finish
	}
	js.Orphaned = s.Span.Orphaned()
	if s.Err != nil && *s.Err != nil {
		errstr := (*s.Err).Error()
		js.Err = &errstr
	}
	js.Panicked = s.Panicked
	js.Args = make([]string, 0, len(s.Span.Args()))
	for _, arg := range s.Span.Args() {
		js.Args = append(js.Args, fmt.Sprint("%#v", arg))
	}
	js.Annotations = make([][]string, 0, len(s.Span.Annotations()))
	for _, annotation := range s.Span.Annotations() {
		js.Annotations = append(js.Annotations,
			[]string{annotation.Name, annotation.Value})
	}
	return js
}

func formatFunc(f *monitor.Func) interface{} {
	js := struct {
		Id           int64            `json:"id"`
		ParentIds    []int64          `json:"parent_ids"`
		Package      string           `json:"package"`
		Name         string           `json:"name"`
		Current      int64            `json:"current"`
		Highwater    int64            `json:"highwater"`
		Success      int64            `json:"success"`
		Panics       int64            `json:"panics"`
		Entry        bool             `json:"entry"`
		Errors       map[string]int64 `json:"errors"`
		SuccessTimes struct {
			Average   time.Duration            `json:"average"`
			Recent    time.Duration            `json:"recent"`
			Quantiles map[string]time.Duration `json:"quantiles"`
		} `json:"success_times"`
		FailureTimes struct {
			Average   time.Duration            `json:"average"`
			Recent    time.Duration            `json:"recent"`
			Quantiles map[string]time.Duration `json:"quantiles"`
		} `json:"failure_times"`
	}{}

	js.Id = f.Id()
	f.Parents(func(parent *monitor.Func) {
		if parent == nil {
			js.Entry = true
		} else {
			js.ParentIds = append(js.ParentIds, parent.Id())
		}
	})
	js.Package = f.Scope().Name()
	js.Name = f.ShortName()
	js.Current = f.Current()
	js.Highwater = f.Highwater()
	js.Success = f.Success()
	js.Panics = f.Panics()
	js.Errors = f.Errors()
	js.SuccessTimes.Average = f.SuccessTimeAverage()
	js.SuccessTimes.Recent = f.SuccessTimeRecent()
	js.SuccessTimes.Quantiles = make(map[string]time.Duration,
		len(monitor.ObservedQuantiles))
	js.FailureTimes.Average = f.FailureTimeAverage()
	js.FailureTimes.Recent = f.FailureTimeRecent()
	js.FailureTimes.Quantiles = make(map[string]time.Duration,
		len(monitor.ObservedQuantiles))
	for _, quantile := range monitor.ObservedQuantiles {
		name := fmt.Sprintf("%.02f", quantile)
		js.SuccessTimes.Quantiles[name] = f.SuccessTimeQuantile(quantile)
		js.FailureTimes.Quantiles[name] = f.FailureTimeQuantile(quantile)
	}
	return js
}

type listWriter struct {
	w   io.Writer
	err error
	sep string
}

func newListWriter(w io.Writer) (rv *listWriter) {
	rv = &listWriter{
		w:   w,
		sep: "\n"}
	_, rv.err = fmt.Fprint(w, "[")
	return rv
}

func (l *listWriter) elem(elem interface{}) {
	if l.err != nil {
		return
	}
	var data []byte
	data, l.err = json.Marshal(elem)
	if l.err != nil {
		return
	}
	_, l.err = fmt.Fprintf(l.w, "%s %s", l.sep, data)
	l.sep = ",\n"
}

func (l *listWriter) done() error {
	if l.err != nil {
		return l.err
	}
	_, err := fmt.Fprint(l.w, "]\n")
	return err
}
