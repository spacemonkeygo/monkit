package monitor

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"
)

type PresentationFunc func(*Registry, io.Writer) error

func escapeDotLabel(format string, args ...interface{}) string {
	val := fmt.Sprintf(format, args...)
	var rv []byte
	for _, b := range []byte(val) {
		switch {
		case 'A' <= b && b <= 'Z', 'a' <= b && b <= 'z', '0' <= b && b <= '9',
			128 <= b, ' ' == b:
			rv = append(rv, b)
		case b == '\n':
			rv = append(rv, []byte(`\l`)...)
		default:
			rv = append(rv, []byte(fmt.Sprintf("&#%d;", int(b)))...)
		}
	}
	return string(rv)
}

func outputDotSpan(w io.Writer, s *Span) error {
	_, err := fmt.Fprintf(w,
		" f%d [label=\"%s\"];\n",
		s.Id, escapeDotLabel("%s(%s)\nelapsed: %s\n",
			s.Func.Name(), strings.Join(s.Args(), ", "), s.Duration()))
	if err != nil {
		return err
	}
	s.Children(func(child *Span) {
		if err != nil {
			return
		}
		err = outputDotSpan(w, child)
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, " f%d -> f%d;\n", s.Id, child.Id)
		if err != nil {
			return
		}
	})
	return err
}

func PresentSpansDot(r *Registry, w io.Writer) error {
	_, err := fmt.Fprintf(w, "digraph G {\n node [shape=box];\n")
	if err != nil {
		return err
	}
	r.LiveTraces(func(s *Span) {
		if err != nil {
			return
		}
		err = outputDotSpan(w, s)
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "}\n")
	return err
}

func outputTextSpan(w io.Writer, s *Span, indent string) (err error) {
	_, err = fmt.Fprintf(w, "%s[%d] %s(%s) (elapsed: %s)\n",
		indent, s.Id, s.Func.Name(), strings.Join(s.Args(), ", "), s.Duration())
	if err != nil {
		return err
	}
	s.Children(func(s *Span) {
		if err != nil {
			return
		}
		err = outputTextSpan(w, s, indent+" ")
	})
	return err
}

func PresentSpansText(r *Registry, w io.Writer) (err error) {
	r.LiveTraces(func(s *Span) {
		if err != nil {
			return
		}
		err = outputTextSpan(w, s, "")
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, "\n")
	})
	return err
}

func formatDist(d *Dist, indent string) (result string) {
	for _, q := range ObservedQuantiles {
		result += fmt.Sprintf("%s%.02f: %s\n", indent, q, time.Duration(
			d.Query(q)*float64(time.Second)))
	}
	return result
}

func PresentFuncsDot(r *Registry, w io.Writer) (err error) {
	_, err = fmt.Fprintf(w, "digraph G {\n node [shape=box];\n")
	if err != nil {
		return err
	}
	r.Funcs(func(f *Func) {
		if err != nil {
			return
		}
		success := f.Success()
		panics := f.Panics()

		var err_out bytes.Buffer
		total_errors := int64(0)
		for errname, count := range f.Errors() {
			_, err = fmt.Fprint(&err_out, escapeDotLabel("error %s: %d\n", errname,
				count))
			if err != nil {
				return
			}
			total_errors += count
		}

		_, err = fmt.Fprintf(w, " f%d [label=\"%s", f.Id,
			escapeDotLabel("%s\ncurrent: %d, success: %d, errors: %d, panics: %d\n",
				f.Name(), f.Current(), success, total_errors, panics))
		if err != nil {
			return
		}

		_, err = err_out.WriteTo(w)
		if err != nil {
			return
		}

		if success > 0 {
			_, err = fmt.Fprint(w, escapeDotLabel(
				"success times:\n%s", formatDist(f.SuccessTimes, "        ")))
			if err != nil {
				return
			}
		}

		if total_errors+panics > 0 {
			_, err = fmt.Fprint(w, escapeDotLabel(
				"failure times:\n%s", formatDist(f.FailureTimes, "        ")))
			if err != nil {
				return
			}
		}

		_, err = fmt.Fprint(w, "\"];\n")
		if err != nil {
			return
		}

		f.Parents(func(parent *Func) {
			if err != nil {
				return
			}
			if parent != nil {
				_, err = fmt.Fprintf(w, " f%d -> f%d;\n", parent.Id, f.Id)
				if err != nil {
					return
				}
			} else {
				_, err = fmt.Fprintf(w, " r%d [label=\"entry\"];\n r%d -> f%d;\n",
					f.Id, f.Id, f.Id)
				if err != nil {
					return
				}
			}
		})
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "}\n")
	return err
}

func PresentFuncsText(r *Registry, w io.Writer) (err error) {
	r.Funcs(func(f *Func) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, "[%d] %s\n  parents: ", f.Id, f.Name())
		if err != nil {
			return
		}
		printed := false
		f.Parents(func(parent *Func) {
			if err != nil {
				return
			}
			if printed {
				_, err = fmt.Fprint(w, ", ")
				if err != nil {
					return
				}
			} else {
				printed = true
			}
			if parent != nil {
				_, err = fmt.Fprintf(w, "%d", parent.Id)
				if err != nil {
					return
				}
			} else {
				_, err = fmt.Fprintf(w, "entry")
				if err != nil {
					return
				}
			}
		})
		var err_out bytes.Buffer
		total_errors := int64(0)
		for errname, count := range f.Errors() {
			_, err = fmt.Fprintf(&err_out, "  error %s: %d\n", errname, count)
			if err != nil {
				return
			}
			total_errors += count
		}
		_, err = fmt.Fprintf(w,
			"\n  current: %d, success: %d, errors: %d, panics: %d\n",
			f.Current(), f.Success(), total_errors, f.Panics())
		if err != nil {
			return
		}
		_, err = err_out.WriteTo(w)
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, "  success times:\n%s  failure times:\n%s\n",
			formatDist(f.SuccessTimes, "    "), formatDist(f.FailureTimes, "    "))
	})
	return err
}

func PresentationFromPath(path string) (
	f PresentationFunc, contentType string, found bool) {
	first, rest := shift(path)
	switch first {
	case "ps":
		second, _ := shift(rest)
		switch second {
		case "":
			return PresentSpansText, "text/plain; charset=utf-8", true
		case "dot":
			return PresentSpansDot, "text/plain; charset=utf-8", true
		}

	case "funcs":
		second, _ := shift(rest)
		switch second {
		case "":
			return PresentFuncsText, "text/plain; charset=utf-8", true
		case "dot":
			return PresentFuncsDot, "text/plain; charset=utf-8", true
		}

	case "stats":
		return PresentStatsText, "text/plain; charset=utf-8", true

	}
	return nil, "", false
}

func shift(path string) (dir, left string) {
	path = strings.TrimLeft(path, "/")
	split := strings.Index(path, "/")
	if split == -1 {
		return path, ""
	}
	return path[:split], path[split:]
}

func PresentStatsText(r *Registry, w io.Writer) (err error) {
	r.Stats(func(name string, val float64) {
		if err != nil {
			return
		}
		_, err = fmt.Fprintf(w, "%s\t%f\n", name, val)
	})
	return err
}
