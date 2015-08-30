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
	"sort"
	"sync"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/spacemonkeygo/monitor.v2"
)

const (
	graphWidth = 1200
	barHeight  = 15
	barSep     = int(barHeight / 15)
	fontSize   = int(barHeight * .6)
	fontOffset = int(barHeight * .2)
)

func TraceQuerySVG(reg *monitor.Registry, w io.Writer,
	matcher func(*monitor.Func) bool) error {
	_, err := fmt.Fprint(w, `<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN"
  "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg version="1.1" xmlns="http://www.w3.org/2000/svg"
  xmlns:xlink="http://www.w3.org/1999/xlink"`)
	if err != nil {
		return err
	}

	spans, err := traceQuery(reg, w, matcher, []byte(" "))
	if err != nil {
		return err
	}

	var minStart, maxEnd time.Time
	for _, s := range spans {
		start := s.Span.Start()
		finish := *s.Finish
		if minStart.IsZero() || start.Before(minStart) {
			minStart = start
		}
		if maxEnd.IsZero() || finish.After(maxEnd) {
			maxEnd = finish
		}
	}
	sort.Sort(startTimeSorter(spans))

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
		_, err := fmt.Fprintf(w, `
  <g class="func">
    <rect x="%d" y="%d" width="%d" height="%d" fill="rgb(128,128,255)" />
    <text x="0" y="%d" fill="rgb(0,0,0)" font-size="%d">%s (%s)</text>
  </g>`, timeToX(s.Span.Start()), id*(barHeight+barSep),
			timeToX(*s.Finish)-timeToX(s.Span.Start()), barHeight,
			(id+1)*(barHeight+barSep)-barSep-fontOffset, fontSize,
			s.Span.Func().FullName(), s.Finish.Sub(s.Span.Start()))
		if err != nil {
			return err
		}
	}

	_, err = w.Write([]byte("\n</svg>\n"))
	return err
}

func TraceQueryJSON(reg *monitor.Registry, w io.Writer,
	matcher func(*monitor.Func) bool) (write_err error) {

	spans, err := traceQuery(reg, w, matcher, []byte(" "))
	if err != nil {
		return err
	}

	lw := newListWriter(w)
	for _, s := range spans {
		lw.elem(formatFinishedSpan(s))
	}
	return lw.done()
}

func traceQuery(reg *monitor.Registry, w io.Writer,
	matcher func(f *monitor.Func) bool,
	keepalive []byte) (spans []*finishedSpan, write_err error) {
	ctx, cancel := context.WithCancel(context.Background())

	abort_ch := make(chan struct{})
	var abort_ch_closed bool
	abort := func() {
		if !abort_ch_closed {
			abort_ch_closed = true
			close(abort_ch)
		}
	}
	defer abort()
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
			case <-abort_ch:
				return
			}
		}
	}()

	var mtx sync.Mutex

	WatchForSpans(ctx, reg, matcher,
		func(s *monitor.Span, err error, panicked bool, finish time.Time) {
			mtx.Lock()
			spans = append(spans,
				&finishedSpan{
					Span:     s,
					Err:      &err,
					Panicked: &panicked,
					Finish:   &finish})
			mtx.Unlock()
		})

	abort()
	if write_err != nil {
		return nil, write_err
	}

	return spans, nil
}

type startTimeSorter []*finishedSpan

func (s startTimeSorter) Len() int      { return len(s) }
func (s startTimeSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s startTimeSorter) Less(i, j int) bool {
	return s[i].Span.Start().UnixNano() < s[j].Span.Start().UnixNano()
}
