// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package http

import (
	"fmt"
	"net/http"

	"github.com/spacemonkeygo/monkit/v3"
)

// TraceHandler wraps a HTTPHandler and import trace information from header.
func TraceHandler(c http.Handler, scope *monkit.Scope, sampler func(trace *monkit.Trace)) http.Handler {
	return traceHandler{
		handler: c,
		sampler: sampler,
		scope:   scope,
	}
}

type traceHandler struct {
	handler http.Handler
	sampler func(trace *monkit.Trace)
	scope   *monkit.Scope
}

// ServeHTTP implements http.Handler with span propagation.
func (t traceHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	info := TraceInfoFromHeader(request.Header)

	traceId := monkit.NewId()
	if info.TraceId != nil {
		traceId = *info.TraceId
	}

	trace := monkit.NewTrace(traceId)
	ctx := request.Context()

	parent := int64(0)
	if info.ParentId != nil {
		parent = *info.ParentId
	}

	if info.Sampled {
		t.sampler(trace)
	}
	defer t.scope.Func().RemoteTrace(&ctx, parent, trace)(nil)

	s := monkit.SpanFromCtx(ctx)
	s.Annotate("http.uri", request.RequestURI)

	wrapped := &responseWriterObserver{w: writer}
	if info.ParentId == nil && info.Sampled {
		writer.Header().Set(traceStateHeader, fmt.Sprintf("traceid=%d,spanid=%d", s.Id(), s.Trace().Id()))
	}
	t.handler.ServeHTTP(wrapped, request.WithContext(s))

	s.Annotate("http.responsecode", fmt.Sprint(wrapped.StatusCode()))
}
