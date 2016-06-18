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

/*
Package monkit is a flexible code instrumenting and data collection library.

With this package, it's easy to monitor and watch all sorts of data.
A motivating example:

	package main

	import (
		"net/http"

		"golang.org/x/net/context"
		"gopkg.in/spacemonkeygo/monkit.v2"
		"gopkg.in/spacemonkeygo/monkit.v2/present"
	)

	var (
		mon = monkit.Package()
	)

	func FixSerenity(ctx context.Context) (err error) {
		defer mon.Task()(&ctx)(&err)

		if SerenityBroken(ctx) {
			err := CallKaylee(ctx)
			mon.Event("kaylee called")
			if err != nil {
				return err
			}
		}

		stowaway_count := StowawaysNeedingHiding(ctx)
		mon.IntVal("stowaway count").Observe(stowaway_count)
		err = HideStowaways(ctx, stowaway_count)
		if err != nil {
			return err
		}

		return nil
	}

	func Monitor() {
		go http.ListenAndServe(":8080", present.HTTP(monkit.Default))
	}

In this example, calling FixSerenity will cause the endpoint at
http://localhost:8080/ to return all sorts of data, such as:

 * How many times we've needed to fix the Serenity
   (the Task monitor infers the statistic name from the callstack)
 * How many times we've succeeded
 * How many times we've failed
 * How long it's taken each time (min/max/avg/recent)
 * How many times we needed to call Kaylee
 * How many errors we've received (per error type!)
 * Statistics on how many stowaways we usually have (min/max/avg/recent/etc)
 * Call graphs of currently running functions
 * Call graphs of all known functions
 * Trace diagrams of ongoing traces

For example, here is how to generate a callgraph:

  ccomps -x <(curl -s http://localhost:8080/funcs/dot) |
    dot | gvpack -array_l1 -m100 | neato -Tsvg -n2 > out.svg

or

  xdot <(curl -s http://localhost:8080/funcs/dot)

*/

package monkit // import "gopkg.in/spacemonkeygo/monkit.v2"
