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

package collect

import (
	"context"
	"fmt"

	"github.com/spacemonkeygo/monkit/v3"
)

// WatchForSpans will watch for spans that 'matcher' returns true for. As soon
// as a trace generates a matched span, all spans from that trace that finish
// from that point on are collected until the matching span completes. All
// collected spans are returned.
// To cancel this operation, simply cancel the ctx argument.
// There is a small but permanent amount of overhead added by this function to
// every trace that is started while this function is running. This only really
// affects long-running traces.
func WatchForSpans(ctx context.Context, r *monkit.Registry,
	matcher func(s *monkit.Span) bool) (spans []*FinishedSpan, err error) {
	collector := NewSpanCollector(matcher)
	defer collector.Stop()

	cancel := ObserveAllTraces(r, collector)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-collector.Done():
		return collector.Spans(), nil
	}
}

// CollectSpans is kind of like WatchForSpans, except that it uses the current
// span to figure out which trace to collect. It calls work(), then collects
// from the current trace until work() returns. CollectSpans won't work unless
// some ancestor function is also monitored and has modified the ctx.
func CollectSpans(ctx context.Context, work func(ctx context.Context)) (
	spans []*FinishedSpan) {
	s := monkit.SpanFromCtx(ctx)
	if s == nil {
		work(ctx)
		return nil
	}
	collector := NewSpanCollector(nil)
	defer collector.Stop()
	s.Trace().ObserveSpans(collector)
	f := s.Func()
	newF := f.Scope().FuncNamed(fmt.Sprintf("%s-TRACED", f.ShortName()))
	func() {
		defer newF.Task(&ctx)(nil)
		collector.ForceStart(monkit.SpanFromCtx(ctx))
		work(ctx)
	}()
	return collector.Spans()
}

// FindSpan will call matcher until matcher returns true. Due to
// the nature of span creation, matcher is likely to be concurrently
// called and therefore matcher may get more than one matching span.
func FindSpan(ctx context.Context, r *monkit.Registry,
	matcher func(s *monkit.Span) bool) {
	if matcher == nil {
		return
	}

	collector := newSpanFinder(matcher)
	defer collector.Stop()

	defer ObserveAllTraces(r, collector)()

	select {
	case <-ctx.Done():
	case <-collector.Done():
	}
}
