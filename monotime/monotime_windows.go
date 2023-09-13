package monotime

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	modkernel32                   = syscall.NewLazyDLL("kernel32.dll")
	queryPerformanceFrequencyProc = modkernel32.NewProc("QueryPerformanceFrequency")
	queryPerformanceCounterProc   = modkernel32.NewProc("QueryPerformanceCounter")

	qpcFrequency = queryPerformanceFrequency()
	qpcBase      = queryPerformanceCounter()
)

func elapsed() time.Duration {
	elapsed := queryPerformanceCounter() - qpcBase
	return time.Duration(elapsed) * time.Second / (time.Duration(qpcFrequency) * time.Nanosecond)
}

func queryPerformanceCounter() int64 {
	var count int64
	syscall.SyscallN(queryPerformanceCounterProc.Addr(), uintptr(unsafe.Pointer(&count)))
	return count
}

func queryPerformanceFrequency() int64 {
	var freq int64
	syscall.SyscallN(queryPerformanceFrequencyProc.Addr(), uintptr(unsafe.Pointer(&freq)))
	return freq
}
