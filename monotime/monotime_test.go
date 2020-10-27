package monotime

import (
	"testing"
	"time"
)

func TestMonotime(t *testing.T) {
	const sleep = time.Second
	start := Now()
	time.Sleep(sleep)
	finish := Now()

	delta := finish.Sub(start) - sleep
	if delta < 0 {
		delta = -delta
	}
	if delta > time.Second/4 {
		t.Errorf("delta too large %v", delta)
	}
}
