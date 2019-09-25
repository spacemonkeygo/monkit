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
	"net/http"

	"github.com/spacemonkeygo/monkit/v3"
)

type handler struct {
	Registry *monkit.Registry
}

// HTTP makes an http.Handler out of a Registry. It serves paths using this
// package's FromRequest request router. Usually HTTP is called with the
// Default registry.
func HTTP(r *monkit.Registry) http.Handler {
	return handler{Registry: r}
}

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p, contentType, err := FromRequest(h.Registry, req.URL.Path, req.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), getStatusCode(err, 500))
		return
	}
	w.Header().Set("Content-Type", contentType)
	p(w)
}
