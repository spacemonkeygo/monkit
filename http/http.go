// Copyright (C) 2014 Space Monkey, Inc.
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

package http

import (
	"net/http"
)

// Client is an interface that matches an http.Client
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type responseWriterObserver struct {
	w  http.ResponseWriter
	sc int
}

func (w *responseWriterObserver) WriteHeader(statusCode int) {
	w.sc = statusCode
	w.w.WriteHeader(statusCode)
}

func (w *responseWriterObserver) Write(p []byte) (n int, err error) {
	if w.sc == 0 {
		w.sc = 200
	}
	return w.w.Write(p)
}

func (w *responseWriterObserver) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriterObserver) StatusCode() int {
	if w.sc == 0 {
		return 200
	}
	return w.sc
}
