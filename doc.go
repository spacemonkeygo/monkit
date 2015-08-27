/*
More docs soon, but for now, here's a super useful graphviz incantation:

    ccomps -x <(curl -s http://localhost:7070/diagnostics/monitor/v2/funcs/dot) |
      dot | gvpack -array_l1 -m100 | neato -Tsvg -n2 > out.svg

or

    xdot <(curl -s http://localhost:7070/diagnostics/monitor/v2/funcs/dot)

*/
package monitor
