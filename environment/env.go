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

package environment // import "github.com/spacemonkeygo/monkit/v3/environment"

import (
	"github.com/spacemonkeygo/monkit/v3"
)

var (
	registrations = []monkit.StatSource{}
)

// Register attaches all of this package's environment data to the given
// registry. It will be attached to a top-level scope called 'env'.
func Register(registry *monkit.Registry) {
	if registry == nil {
		registry = monkit.Default
	}
	pkg := registry.Package()
	for _, source := range registrations {
		pkg.Chain(source)
	}
}
