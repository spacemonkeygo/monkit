package main

import (
	"context"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/environment"
)

var mon = monkit.Package()

func main() {
	ctx := context.Background()
	environment.Register(monkit.Default)

	var err error
	defer mon.Task()(&ctx)(&err)
}
