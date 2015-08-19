/*
More docs soon, but for now, here's a super useful graphviz incantation:

    ccomps -x <(curl -s http://localhost:7070/diagnostics/monitor2/funcs/dot) |
      dot | gvpack -array_l1 -m100 | neato -Tsvg -n2 > out.svg
*/
package monitor
