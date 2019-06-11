package monkit

import (
	"context"
	"fmt"
	"testing"

	"gopkg.in/spacemonkeygo/monkit.v2/internal/testpkg1"

	monkit "gopkg.in/spacemonkeygo/monkit.v2"
)

func TestCallers(t *testing.T) {
	ctx := context.Background()
	testpkg1.TestFunc(ctx, nil)
	testpkg1.TestFunc(ctx, fmt.Errorf("new error"))
	stats := monkit.Collect(monkit.Default)
	assertEqual(t,
		stats["gopkg.in/spacemonkeygo/monkit.v2/internal/testpkg1.TestFunc.total"], 2)
	assertEqual(t,
		stats["gopkg.in/spacemonkeygo/monkit.v2/internal/testpkg1.TestFunc.successes"], 1)
	assertEqual(t,
		stats["gopkg.in/spacemonkeygo/monkit.v2/internal/testpkg1.TestFunc.errors"], 1)
}

func assertEqual(t *testing.T, actual, expected float64) {
	t.Helper()
	if actual != expected {
		t.Fatal(fmt.Sprintf("got %v, expected %v", actual, expected))
	}
}
