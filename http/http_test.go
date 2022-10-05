// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package http

import (
	"net/http"
	"testing"
)

func TestWrapping(t *testing.T) {
	rw := &responseWriter{
		data: []byte{},
	}
	wrapped, statusCode := Wrap(rw)
	_, _ = wrapped.Write([]byte{1, 2, 3})
	wrapped.WriteHeader(123)

	_, fok := wrapped.(http.Flusher)
	if fok {
		t.Fatalf("wrapped writer is a flusher, but the original writer is not")
	}

	if len(rw.data) != 3 {
		t.Fatalf("bytes are not injected (%d size)", len(rw.data))
	}

	if statusCode() != 123 {
		t.Fatalf("Status code is not saved")
	}
}

func TestWrappingFlusher(t *testing.T) {
	rw := &responseWriterFlusher{
		data: []byte{},
	}
	wrapped, statusCode := Wrap(rw)
	_, _ = wrapped.Write([]byte{1, 2, 3})
	wrapped.WriteHeader(123)

	flusher, fok := wrapped.(http.Flusher)
	if !fok {
		t.Fatalf("wrapped writer is not a flusher")
	}
	flusher.Flush()

	if !rw.flushed {
		t.Fatalf("Not flushed")
	}

	if len(rw.data) != 3 {
		t.Fatalf("bytes are not injected (%d size)", len(rw.data))
	}

	if statusCode() != 123 {
		t.Fatalf("Status code is not saved")
	}
}

type responseWriter struct {
	data []byte
}

func (r *responseWriter) Header() http.Header {
	return http.Header{}
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	r.data = append(r.data, bytes...)
	return len(bytes), nil
}

func (r *responseWriter) WriteHeader(statusCode int) {

}

var _ http.ResponseWriter = &responseWriter{}

type responseWriterFlusher struct {
	data    []byte
	flushed bool
}

func (r *responseWriterFlusher) Flush() {
	r.flushed = true
}

func (r *responseWriterFlusher) Header() http.Header {
	return http.Header{}
}

func (r *responseWriterFlusher) Write(bytes []byte) (int, error) {
	r.data = append(r.data, bytes...)
	return len(bytes), nil
}

func (r *responseWriterFlusher) WriteHeader(statusCode int) {

}

var _ http.ResponseWriter = &responseWriterFlusher{}
var _ http.Flusher = &responseWriterFlusher{}
