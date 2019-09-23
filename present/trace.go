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
	"encoding/xml"
	"io"
	"strings"
	"text/template"
	"time"

	"gopkg.in/spacemonkeygo/monkit.v3"
	"gopkg.in/spacemonkeygo/monkit.v3/collect"
)

const (
	graphWidth = 1200
	barHeight  = 15
	barSep     = int(barHeight / 15)
	fontSize   = int(barHeight * .6)
	fontOffset = int(barHeight * .2)
)

var (
	svgHeader = template.Must(template.New("header").Parse(
		`<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN"
  "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg version="1.1" xmlns="http://www.w3.org/2000/svg"
  xmlns:xlink="http://www.w3.org/1999/xlink"
  viewBox="0 0 {{.GraphWidth}} {{.GraphHeight}}"
  width="{{.GraphWidth}}" height="{{.GraphHeight}}">

  <style type="text/css">
    .func .parent { visibility: hidden; }
    .func.asParent { stroke: green; stroke-width: 0.5; cursor: pointer; }
    .func.selected { stroke: black; stroke-width: 0.5; cursor: pointer; }
    .func.selected .parent { stroke: green; visibility: visible; }
    .func.selected .parent line { marker-end: url(#head-green); }
    .func.asChild { stroke: purple; stroke-width: 0.5; cursor: pointer; }
    .func.asChild .parent { stroke: purple; visibility: visible; }
    .func.asChild .parent line { marker-end: url(#head-purple); }
  </style>
  <script>
  //<![CDATA[
    function select(el, classname) {
      if (el) { el.classList.add(classname); }
    }
    function deselect(el, classname) {
      if (el) { el.classList.remove(classname); }
    }
    function mouseApply(fn, self, parent) {
      fn(document.getElementById("id-" + parent), "asParent");
      fn(document.getElementById("id-" + self), "selected");
      var children = document.getElementsByClassName("parent-" + self);
      for (var i = 0; i < children.length; i++) {
        fn(children[i], "asChild");
      }
    }
    function mouseover(self, parent) {
      mouseApply(select, self, parent);
    }
    function mouseout(self, parent) {
      mouseApply(deselect, self, parent);
    }
  //]]>
  </script>
  <defs>
    <marker id="head-green" orient="auto" markerWidth="2" markerHeight="4"
        refX="0.1" refY="2" fill="green">
      <path d="M0,0 V4 L2,2 Z"/>
    </marker>
    <marker id="head-purple" orient="auto" markerWidth="2" markerHeight="4"
        refX="0.1" refY="2" fill="purple">
      <path d="M0,0 V4 L2,2 Z"/>
    </marker>
  </defs>`))

	svgFooter = template.Must(template.New("footer").Parse(`
</svg>
`))

	svgFunc = template.Must(template.New("func").Parse(`
  <g id="id-{{.SpanId}}" class="func parent-{{.ParentId}}"
    onmouseover="mouseover('{{.SpanId}}', '{{.ParentId}}');"
    onmouseout="mouseout('{{.SpanId}}', '{{.ParentId}}');">
    <rect x="{{.SpanLeft}}" y="{{.SpanTop}}"
        width="{{.SpanWidth}}" height="{{.SpanHeight}}"
        fill="{{.SpanColor}}" />
    <text x="0" y="{{.TextTop}}" fill="rgb(0,0,0)" font-size="{{.FontSize}}">
      {{.FuncName}}({{.FuncArgs}}) ({{.FuncDuration}})
    </text>
    <g class="parent">
      <line stroke-width="2"
          x1="{{.SpanLeft}}" x2="{{.ParentLeft}}"
          y1="{{.SpanMid}}" y2="{{.ParentMid}}" />
    </g>
  </g>`))
)

// SpansToSVG takes a list of FinishedSpans and writes them to w in SVG format.
// It draws a trace using the Spans where the Spans are ordered by start time.
func SpansToSVG(w io.Writer, spans []*collect.FinishedSpan) error {
	var minStart, maxEnd time.Time
	graphHeight := (barHeight + barSep) * len(spans)

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

	timeToX := func(t time.Time) int {
		return int(((t.UnixNano() - minStart.UnixNano()) * graphWidth) /
			(maxEnd.UnixNano() - minStart.UnixNano()))
	}

	err := svgHeader.Execute(w, map[string]interface{}{
		"GraphWidth":  graphWidth,
		"GraphHeight": graphHeight,
	})
	if err != nil {
		return err
	}

	positionBySpanId := map[int64]int{}

	for id, s := range spans {
		positionBySpanId[s.Span.Id()] = id

		color := "rgb(128,128,255)"
		switch {
		case s.Panicked:
			color = "rgb(255,0,0)"
		case unwrapError(s.Err) == contextCanceled:
			color = "rgb(255,255,0)"
		case s.Err != nil:
			color = "rgb(255,144,0)"
		}

		templateVals := struct {
			SpanId       int64
			SpanLeft     int
			SpanTop      int
			SpanWidth    int
			SpanHeight   int
			SpanColor    string
			TextTop      int
			FontSize     int
			FuncName     string
			FuncArgs     string
			FuncDuration string
			SpanMid      int

			ParentId   int64
			ParentLeft int
			ParentMid  int
		}{
			SpanId:       s.Span.Id(),
			SpanLeft:     timeToX(s.Span.Start()),
			SpanTop:      id * (barHeight + barSep),
			SpanWidth:    timeToX(s.Finish) - timeToX(s.Span.Start()),
			SpanHeight:   barHeight,
			SpanColor:    color,
			TextTop:      (id+1)*(barHeight+barSep) - barSep - fontOffset,
			FontSize:     fontSize,
			FuncName:     s.Span.Func().FullName(),
			FuncArgs:     strings.Join(s.Span.Args(), " "),
			FuncDuration: s.Finish.Sub(s.Span.Start()).String(),
			SpanMid:      id*(barHeight+barSep) + barHeight/2,
		}

		var buf bytes.Buffer
		err = xml.EscapeText(&buf, []byte(templateVals.FuncName))
		if err != nil {
			return err
		}
		templateVals.FuncName = buf.String()

		buf.Reset()
		err = xml.EscapeText(&buf, []byte(templateVals.FuncArgs))
		if err != nil {
			return err
		}
		templateVals.FuncArgs = buf.String()

		parent := s.Span.Parent()
		if parent != nil {
			templateVals.ParentId = parent.Id()
			templateVals.ParentLeft = timeToX(parent.Start())
			templateVals.ParentMid = barHeight/2 +
				positionBySpanId[parent.Id()]*(barHeight+barSep)
		}

		err := svgFunc.Execute(w, templateVals)
		if err != nil {
			return err
		}
	}

	return svgFooter.Execute(w, nil)
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

	spans, err := collect.WatchForSpans(ctx, reg, matcher)

	abortTimer()
	if write_err != nil {
		return nil, write_err
	}

	return spans, err
}
