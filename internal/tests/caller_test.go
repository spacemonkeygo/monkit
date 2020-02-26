package monkit

import (
	"context"
	"fmt"
	"testing"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/internal/testpkg1"
)

func TestCallers(t *testing.T) {
	ctx := context.Background()
	testpkg1.TestFunc(ctx, nil)
	testpkg1.TestFunc(ctx, fmt.Errorf("new error"))
	stats := monkit.Collect(monkit.Default)

	assertEqual(t,
		stats["function,name=TestFunc,scope=github.com/spacemonkeygo/monkit/v3/internal/testpkg1 total"], 2)
	assertEqual(t,
		stats["function,name=TestFunc,scope=github.com/spacemonkeygo/monkit/v3/internal/testpkg1 successes"], 1)
	assertEqual(t,
		stats["function,name=TestFunc,scope=github.com/spacemonkeygo/monkit/v3/internal/testpkg1 errors"], 1)
	assertEqual(t,
		stats["test_event,scope=github.com/spacemonkeygo/monkit/v3/internal/testpkg1 total"], 2)
}

func assertEqual(t *testing.T, actual, expected float64) {
	t.Helper()
	if actual != expected {
		t.Fatal(fmt.Sprintf("got %v, expected %v", actual, expected))
	}
}
