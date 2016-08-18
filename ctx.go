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

package monkit

//go:generate sh -c "m4 -D_STDLIB_IMPORT_='\"context\"' -D_OTHER_IMPORT_= -D_BUILD_TAG_='// +build go1.7' ctxgen.go.m4 > ctx17.go"
//go:generate sh -c "m4 -D_STDLIB_IMPORT_= -D_OTHER_IMPORT_='\"golang.org/x/net/context\"' -D_BUILD_TAG_='// +build !go1.7' ctxgen.go.m4 > xctx.go"
//go:generate gofmt -w -s ctx17.go xctx.go
