// Copyright (C) 2017 Space Monkey, Inc.
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
	"net/http"
)

type errorKind string

const (
	errBadRequest errorKind = "Bad Request"
	errNotFound   errorKind = "Not Found"
)

var statusCodes = map[errorKind]int{
	errBadRequest: http.StatusBadRequest,
	errNotFound:   http.StatusNotFound,
}

func getStatusCode(err error, def int) int {
	if err, ok := err.(errorT); ok {
		if code, ok := statusCodes[err.kind]; ok {
			return code
		}
	}
	return def
}

func (e errorKind) New(format string, args ...interface{}) error {
	return errorT{
		kind:    e,
		message: fmt.Sprintf(format, args...),
	}
}

type errorT struct {
	kind    errorKind
	message string
}

func (e errorT) Error() string {
	return fmt.Sprintf("%s: %s", string(e.kind), e.message)
}
