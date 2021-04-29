package monkit

import "testing"

func TestExtractFuncName(t *testing.T) {
	for _, test := range []struct {
		in string
		fn string
		ok bool
	}{
		{"", "", false},
		{"a/", "", false},
		{"a/v.", "", false},
		{"a/v.x", "x", true},
		{"github.com/spacemonkeygo/monkit/v3.BenchmarkTask.func1", "BenchmarkTask.func1", true},
		{"main.DoThings.func1", "DoThings.func1", true},
		{"main.DoThings", "DoThings", true},
	} {
		fn, ok := extractFuncName(test.in)
		if fn != test.fn || ok != test.ok {
			t.Errorf("failed %q, got %q %v, expected %q %v", test.in, fn, ok, test.fn, test.ok)
		}
	}
}
