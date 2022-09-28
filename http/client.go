// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spacemonkeygo/monkit/v3"
)

// TraceRequest will perform an HTTP request, creating a new Span for the HTTP
// request and sending the Span in the HTTP request headers.
// Compare to http.Client.Do.
func TraceRequest(ctx context.Context, scope *monkit.Scope, cl Client, req *http.Request) (
	resp *http.Response, err error) {
	defer scope.TaskNamed(req.Method)(&ctx)(&err)

	s := monkit.SpanFromCtx(ctx)
	s.Annotate("http.uri", req.URL.String())
	TraceInfoFromSpan(s).SetHeader(req.Header)
	resp, err = cl.Do(req)
	if err != nil {
		return resp, err
	}
	s.Annotate("http.responsecode", fmt.Sprint(resp.StatusCode))
	return resp, nil
}
