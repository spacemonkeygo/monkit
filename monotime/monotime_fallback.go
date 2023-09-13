//go:build !windows
// +build !windows

package monotime

import "time"

func elapsed() time.Duration { return time.Since(initTime) }
