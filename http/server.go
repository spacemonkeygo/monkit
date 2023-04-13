// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package http

import (
	"fmt"
	"net/http"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/present"
)

// TraceHandler wraps a HTTPHandler and import trace information from header.
func TraceHandler(c http.Handler, scope *monkit.Scope, allowedBaggage ...string) http.Handler {
	return traceHandler{
		handler:        c,
		scope:          scope,
		allowedBaggage: allowedBaggage,
	}
}

type traceHandler struct {
	handler http.Handler
	scope   *monkit.Scope

	// allowedBaggage defines the allowed `baggage: k=v` HTTP headers which are imported as scan annotations.
	allowedBaggage []string
}

// ServeHTTP implements http.Handler with span propagation.
func (t traceHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	info := TraceInfoFromHeader(request.Header, t.allowedBaggage...)

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
		trace.Set(present.SampledKey, true)
	}
	defer t.scope.Func().RemoteTrace(&ctx, parent, trace)(nil)

	if cb, exists := trace.Get(present.SampledCBKey).(func(*monkit.Trace)); exists {
		cb(trace)
	}

	s := monkit.SpanFromCtx(ctx)
	for k, v := range info.Baggage {
		s.Annotate(k, v)
	}
	s.Annotate("http.uri", request.RequestURI)

	wrapped, statusCode := Wrap(writer)
	if info.ParentId == nil && info.Sampled {
		writer.Header().Set(traceIDHeader, fmt.Sprintf("%x", s.Trace().Id()))
		writer.Header().Set(childIDHeader, fmt.Sprintf("%x", s.Id()))
	}
	t.handler.ServeHTTP(wrapped, request.WithContext(s))

	s.Annotate("http.responsecode", fmt.Sprint(statusCode()))
}
