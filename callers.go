package monitor

import (
	"runtime"
	"strings"
)

func callerPackage(frames int) string {
	pc, _, _, ok := runtime.Caller(frames + 1)
	if !ok {
		return "unknown"
	}
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown"
	}
	return strings.TrimSuffix(f.Name(), ".init")
}

func callerFunc(frames int) string {
	pc, _, _, ok := runtime.Caller(frames + 1)
	if !ok {
		return "unknown"
	}
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown"
	}
	slash_pieces := strings.Split(f.Name(), "/")
	dot_pieces := strings.SplitN(slash_pieces[len(slash_pieces)-1], ".", 2)
	return dot_pieces[len(dot_pieces)-1]
}
