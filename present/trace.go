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
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"gopkg.in/spacemonkeygo/monkit.v2"
	"gopkg.in/spacemonkeygo/monkit.v2/collect"
)

const (
	graphWidth = 1200
	barHeight  = 15
	barSep     = int(barHeight / 15)
	fontSize   = int(barHeight * .6)
	fontOffset = int(barHeight * .2)
)

// SpansToSVG takes a list of FinishedSpans and writes them to w in SVG format.
// It draws a trace using the Spans where the Spans are ordered by start time.
func SpansToSVG(w io.Writer, spans []*collect.FinishedSpan) error {
	_, err := fmt.Fprint(w, `<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN"
  "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg version="1.1" xmlns="http://www.w3.org/2000/svg"
  xmlns:xlink="http://www.w3.org/1999/xlink"`)
	if err != nil {
		return err
	}

	var minStart, maxEnd time.Time
	for _, s := range spans {
		start := s.Span.Start()
		finish := s.Finish
		if minStart.IsZero() || start.Before(minStart) {
			minStart = start
		}
		if maxEnd.IsZero() || finish.After(maxEnd) {
			maxEnd = finish
		}
	}
	collect.StartTimeSorter(spans).Sort()

	timeToX := func(t time.Time) int64 {
		return ((t.UnixNano() - minStart.UnixNano()) * graphWidth) /
			(maxEnd.UnixNano() - minStart.UnixNano())
	}

	graphHeight := (barHeight + barSep) * len(spans)
	_, err = fmt.Fprintf(w, ` viewBox="0 0 %d %d" width="%d" height="%d">

  <style type="text/css">
    .func:hover { stroke: black; stroke-width: 0.5; cursor: pointer; }
  </style>`, graphWidth, graphHeight, graphWidth, graphHeight)
	if err != nil {
		return err
	}

	for id, s := range spans {
		color := "rgb(128,128,255)"
		switch {
		case s.Panicked:
			color = "rgb(255,0,0)"
		case unwrapError(s.Err) == context.Canceled:
			color = "rgb(255,255,0)"
		case s.Err != nil:
			color = "rgb(255,144,0)"
		}
		args := strings.Join(s.Span.Args(), " ")
		_, err := fmt.Fprintf(w, `
  <g class="func">
    <rect x="%d" y="%d" width="%d" height="%d" fill="%s" />
    <text x="0" y="%d" fill="rgb(0,0,0)" font-size="%d">%s(%s) (%s)</text>
  </g>`, timeToX(s.Span.Start()), id*(barHeight+barSep),
			timeToX(s.Finish)-timeToX(s.Span.Start()), barHeight, color,
			(id+1)*(barHeight+barSep)-barSep-fontOffset, fontSize,
			s.Span.Func().FullName(), args, s.Finish.Sub(s.Span.Start()))
		if err != nil {
			return err
		}
	}

	_, err = w.Write([]byte("\n</svg>\n"))
	return err
}

type wrappedError interface {
	WrappedErr() error
}

func unwrapError(err error) error {
	for {
		wrapper, ok := err.(wrappedError)
		if !ok {
			return err
		}
		wrapped_error := wrapper.WrappedErr()
		if wrapped_error == nil {
			return err
		}
		err = wrapped_error
	}
}

// TraceQuerySVG uses WatchForSpans to write all Spans from 'reg' matching
// 'matcher' to 'w' in SVG format.
func TraceQuerySVG(reg *monkit.Registry, w io.Writer,
	matcher func(*monkit.Span) bool) error {
	spans, err := watchForSpansWithKeepalive(
		reg, w, matcher, []byte("\n"))
	if err != nil {
		return err
	}

	return SpansToSVG(w, spans)

}

// TraceQueryJSON uses WatchForSpans to write all Spans from 'reg' matching
// 'matcher' to 'w' in JSON format.
func TraceQueryJSON(reg *monkit.Registry, w io.Writer,
	matcher func(*monkit.Span) bool) (write_err error) {

	spans, err := watchForSpansWithKeepalive(
		reg, w, matcher, []byte("\n"))
	if err != nil {
		return err
	}

	return SpansToJSON(w, spans)
}

// SpansToJSON turns a list of FinishedSpans into JSON format.
func SpansToJSON(w io.Writer, spans []*collect.FinishedSpan) error {
	lw := newListWriter(w)
	for _, s := range spans {
		lw.elem(formatFinishedSpan(s))
	}
	return lw.done()
}

func watchForSpansWithKeepalive(reg *monkit.Registry, w io.Writer,
	matcher func(s *monkit.Span) bool, keepalive []byte) (
	spans []*collect.FinishedSpan, write_err error) {
	ctx, cancel := contextWithCancel()

	abortTimerCh := make(chan struct{})
	var abortTimerChClosed bool
	abortTimer := func() {
		if !abortTimerChClosed {
			abortTimerChClosed = true
			close(abortTimerCh)
		}
	}
	defer abortTimer()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, write_err = w.Write(keepalive)
				if write_err != nil {
					cancel()
				}
			case <-abortTimerCh:
				return
			}
		}
	}()

	spans, err := WatchForSpans(ctx, reg, matcher)

	abortTimer()
	if write_err != nil {
		return nil, write_err
	}

	return spans, err
}
