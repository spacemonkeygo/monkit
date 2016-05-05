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
More docs soon, but for now, here's a super useful graphviz incantation:

    ccomps -x <(curl -s http://localhost:7070/diagnostics/monitor/v2/funcs/dot) |
      dot | gvpack -array_l1 -m100 | neato -Tsvg -n2 > out.svg

or

    xdot <(curl -s http://localhost:7070/diagnostics/monitor/v2/funcs/dot)

*/
package monitor // import "gopkg.in/spacemonkeygo/monitor.v2"
