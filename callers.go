package monitor

import (
	"runtime"
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
	return f.Name()
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
	return f.Name()
}
