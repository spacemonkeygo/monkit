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
	"fmt"
	"io"

	"github.com/spacemonkeygo/monkit/v3"
)

// StatsOld is deprecated.
func StatsOld(r *monkit.Registry, w io.Writer) error { return StatsText(r, w) }

// StatsText writes all of the name/value statistics pairs the Registry knows
// to w in a text format.
func StatsText(r *monkit.Registry, w io.Writer) (err error) {
	r.Stats(func(key monkit.SeriesKey, field string, val float64) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, "%s=%f\n", key.WithField(field), val)
	})
	return err
}

// StatsJSON writes all of the name/value statistics pairs the Registry knows
// to w in a JSON format.
func StatsJSON(r *monkit.Registry, w io.Writer) (err error) {
	lw := newListWriter(w)
	r.Stats(func(key monkit.SeriesKey, field string, val float64) {
		lw.elem([]interface{}{key.Measurement, key.Tags.All(), field, val})
	})
	return lw.done()
}
