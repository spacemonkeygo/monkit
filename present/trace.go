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
	"context"
	"encoding/xml"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/collect"
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
    .func { stroke: black; stroke-width: 0.5; }

    .func.hover-asParent { stroke: green; stroke-width: 1; cursor: pointer; }
    .func.hover-selected { stroke: black; stroke-width: 1; cursor: pointer; }
    .func.hover-selected .parent { stroke: green; visibility: visible; }
    .func.hover-selected .parent line { marker-end: url(#head-green); }
    .func.hover-selected text { visibility: visible; }
    .func.hover-asChild { stroke: purple; stroke-width: 1; cursor: pointer; }
    .func.hover-asChild .parent { stroke: purple; visibility: visible; }
    .func.hover-asChild .parent line { marker-end: url(#head-purple); }

    .func.click-asParent { stroke: green; stroke-width: 1; cursor: pointer; }
    .func.click-selected { stroke: black; stroke-width: 1; cursor: pointer; }
    .func.click-selected .parent { stroke: green; visibility: visible; }
    .func.click-selected .parent line { marker-end: url(#head-green); }
    .func.click-selected text { visibility: visible; }
    .func.click-asChild { stroke: purple; stroke-width: 1; cursor: pointer; }
    .func.click-asChild .parent { stroke: purple; visibility: visible; }
    .func.click-asChild .parent line { marker-end: url(#head-purple); }
  </style>
  <script>
  //<![CDATA[
    function select(el, classname) {
      if (el) { el.classList.add(classname); }
    }
    function deselect(el, classname) {
      if (el) { el.classList.remove(classname); }
    }
    function toggle(el, classname) {
      if (el) { el.classList.toggle(classname) }
    }
    function mouseApply(fn, self, parent, prefix, tooltip) {
      fn(document.getElementById("id-" + parent), prefix + "-asParent");
      fn(document.getElementById("id-" + self), prefix + "-selected");
      var children = document.getElementsByClassName("parent-" + self);
      for (var i = 0; i < children.length; i++) {
        fn(children[i], prefix + "-asChild");
      }
      if (tooltip != null) {
         document.getElementById("tooltip").textContent = tooltip;
      }
    }
    function mouseover(self, parent, tooltip) {
      mouseApply(select, self, parent, "hover", tooltip);
    }
    function mouseout(self, parent) {
      mouseApply(deselect, self, parent, "hover", "");
    }
    function mouseclick(self, parent) {
      mouseApply(toggle, self, parent, "click", null);
    }

    function moveFixed(evt) {
      var fixed = document.getElementById("fixed");
      var tfm = fixed.transform.baseVal.getItem(0);
      tfm.setTranslate(0, document.documentElement.scrollTop);
    }

    window.onscroll = moveFixed;
    window.onload = moveFixed;
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
  </defs>
 `))

	svgFooter = template.Must(template.New("footer").Parse(`
  <g id="fixed" transform="translate(0 0)">
    <rect width="100%" height="30"/>
    <text id="tooltip" y="20" x="10" font-size="20" fill="white"></text>
  </g>
</svg>
`))

	svgFunc = template.Must(template.New("func").Parse(`
  <g id="id-{{.SpanId}}" class="func parent-{{.ParentId}}" onmouseover="mouseover('{{.SpanId}}', '{{.ParentId}}', '{{.FuncName}}({{.FuncArgs}}) Duration:{{.FuncDuration}} Started:{{.FuncStartDuration}}');" onmouseout="mouseout('{{.SpanId}}', '{{.ParentId}}');" onclick="mouseclick('{{.SpanId}}', '{{.ParentId}}');">
    <clipPath id="clip-{{.SpanId}}"><rect x="{{.SpanLeft}}" y="{{.SpanTop}}" width="{{.SpanWidth}}" height="{{.SpanHeight}}"/></clipPath>
    <rect id="rect-{{.SpanId}}" x="{{.SpanLeft}}" y="{{.SpanTop}}" width="{{.SpanWidth}}" height="{{.SpanHeight}}" fill="{{.SpanColor}}"/>
    <text id="text-{{.SpanId}}" x="{{.SpanLeft}}" y="{{.TextTop}}" fill="rgb(0,0,0)" font-size="{{.FontSize}}" clip-path="url(#clip-{{.SpanId}})">{{.FuncName}}({{.FuncArgs}}) ({{.FuncDuration}})</text>
    <g class="parent"><line stroke-width="2" x1="{{.SpanLeft}}" x2="{{.ParentLeft}}" y1="{{.SpanMid}}" y2="{{.ParentMid}}" /></g>
  </g>`))
)

type spanInformation struct {
	Span        *collect.FinishedSpan
	Parent      int64
	Children    []int64
	LargestTime time.Time
	Layout      bool
	Rows        [][]int64
	Row         int
}

func computeSpanTree(spans []*collect.FinishedSpan) map[int64]*spanInformation {
	out := make(map[int64]*spanInformation)
	for _, span := range spans {
		id := span.Span.Id()
		if out[id] == nil {
			out[id] = new(spanInformation)
		}
		out[id].Span = span

		if span.Finish.After(out[id].LargestTime) {
			out[id].LargestTime = span.Finish
		}

		if parent := span.Span.Parent(); parent != nil {
			pid := parent.Id()
			out[id].Parent = pid

			if pedges := out[pid]; pedges == nil {
				out[pid] = new(spanInformation)
			}
			out[pid].Children = append(out[pid].Children, id)

			if span.Finish.After(out[pid].LargestTime) {
				out[pid].LargestTime = span.Finish
			}
		}
	}
	return out
}

func computeLayoutInformation(spans []*collect.FinishedSpan) (map[int64]*spanInformation, int) {
	spanTree := computeSpanTree(spans)
	for _, span := range spans {
		includeSpanInLayoutInformation(spanTree, span)
	}

	usedRows := 1
	for _, spanInfo := range spanTree {
		if spanInfo.Parent == 0 {
			usedRows += computeRow(spanTree, spanInfo, usedRows)
		}
	}

	return spanTree, usedRows + 1
}

func includeSpanInLayoutInformation(spanTree map[int64]*spanInformation, span *collect.FinishedSpan) {
	id := span.Span.Id()
	si := spanTree[id]
	if si.Layout {
		return
	}
	si.Layout = true

	parSpan := span.Span.Parent()
	if parSpan == nil || spanTree[parSpan.Id()] == nil {
		return
	}

	psi, ok := spanTree[parSpan.Id()]
	if !ok {
		includeSpanInLayoutInformation(spanTree, spanTree[parSpan.Id()].Span)
		psi = spanTree[parSpan.Id()]
	}

	start := span.Span.Start()
	found := false
	for i, children := range psi.Rows {
		if len(children) == 0 || start.After(spanTree[children[len(children)-1]].LargestTime) {
			psi.Rows[i] = append(psi.Rows[i], id)
			found = true
			break
		}
	}
	if !found {
		psi.Rows = append(psi.Rows, []int64{id})
	}
}

func computeRow(spanTree map[int64]*spanInformation, si *spanInformation, startingRow int) int {
	si.Row = startingRow
	usedRows := startingRow + 1
	for _, children := range si.Rows {
		maxHeight := 0
		for _, child := range children {
			childHeight := computeRow(spanTree, spanTree[child], usedRows)
			if childHeight > maxHeight {
				maxHeight = childHeight
			}
		}
		usedRows += maxHeight
	}
	return usedRows - startingRow
}

// SpansToSVG takes a list of FinishedSpans and writes them to w in SVG format.
// It draws a trace using the Spans where the Spans are ordered by start time.
func SpansToSVG(w io.Writer, spans []*collect.FinishedSpan) error {
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

	var earliestTime time.Time
	if len(spans) > 0 {
		earliestTime = spans[0].Span.Start()
	}

	lis, maxRow := computeLayoutInformation(spans)
	graphHeight := (barHeight+barSep)*maxRow + 20

	timeToX := func(t time.Time) int {
		return int(((t.UnixNano() - minStart.UnixNano()) * graphWidth) /
			(maxEnd.UnixNano() - minStart.UnixNano()))
	}

	max := func(x, y int) int {
		if x > y {
			return x
		}
		return y
	}

	err := svgHeader.Execute(w, map[string]interface{}{
		"GraphWidth":  graphWidth,
		"GraphHeight": graphHeight,
	})
	if err != nil {
		return err
	}

	for _, s := range spans {
		id := lis[s.Span.Id()].Row

		color := "rgb(128,128,255)"
		switch {
		case s.Panicked:
			color = "rgb(255,0,0)"
		case unwrapError(s.Err) == context.Canceled:
			color = "rgb(255,255,0)"
		case s.Err != nil:
			color = "rgb(255,144,0)"
		}

		templateVals := struct {
			SpanId            int64
			SpanLeft          int
			SpanTop           int
			SpanWidth         int
			SpanHeight        int
			SpanColor         string
			TextTop           int
			FontSize          int
			FuncName          string
			FuncArgs          string
			FuncDuration      string
			FuncStartDuration string
			SpanMid           int

			ParentId   int64
			ParentLeft int
			ParentMid  int
		}{
			SpanId:            s.Span.Id(),
			SpanLeft:          timeToX(s.Span.Start()),
			SpanTop:           id * (barHeight + barSep),
			SpanWidth:         max(timeToX(s.Finish)-timeToX(s.Span.Start()), 1),
			SpanHeight:        barHeight,
			SpanColor:         color,
			TextTop:           (id+1)*(barHeight+barSep) - barSep - fontOffset,
			FontSize:          fontSize,
			FuncName:          s.Span.Func().FullName(),
			FuncArgs:          strings.Join(s.Span.Args(), " "),
			FuncDuration:      s.Finish.Sub(s.Span.Start()).String(),
			FuncStartDuration: s.Span.Start().Sub(earliestTime).String(),
			SpanMid:           id*(barHeight+barSep) + barHeight/2,
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

		if parent := s.Span.Parent(); parent != nil {
			row := 0
			if pli := lis[parent.Id()]; pli != nil {
				row = pli.Row
			}
			templateVals.ParentId = parent.Id()
			templateVals.ParentLeft = timeToX(parent.Start())
			templateVals.ParentMid = barHeight/2 + row*(barHeight+barSep)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
