package monkit

import (
	"context"
	"fmt"
	"testing"

	"github.com/spacemonkeygo/monkit/v3/internal/testpkg1"

	monkit "github.com/spacemonkeygo/monkit/v3"
)

func TestCallers(t *testing.T) {
	ctx := context.Background()
	testpkg1.TestFunc(ctx, nil)
	testpkg1.TestFunc(ctx, fmt.Errorf("new error"))
	stats := monkit.Collect(monkit.Default)
	assertEqual(t,
		stats["github.com/spacemonkeygo/monkit/v3/internal/testpkg1.TestFunc.total"], 2)
	assertEqual(t,
		stats["github.com/spacemonkeygo/monkit/v3/internal/testpkg1.TestFunc.successes"], 1)
	assertEqual(t,
		stats["github.com/spacemonkeygo/monkit/v3/internal/testpkg1.TestFunc.errors"], 1)
	assertEqual(t,
		stats["github.com/spacemonkeygo/monkit/v3/internal/testpkg1.test_event.total"], 2)
}

func assertEqual(t *testing.T, actual, expected float64) {
	t.Helper()
	if actual != expected {
		t.Fatal(fmt.Sprintf("got %v, expected %v", actual, expected))
	}
}
