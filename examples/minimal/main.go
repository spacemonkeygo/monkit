package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/environment"
	"github.com/spacemonkeygo/monkit/v3/present"
)

var mon = monkit.Package()

func main() {
	environment.Register(monkit.Default)

	go http.ListenAndServe("127.0.0.1:9000", present.HTTP(monkit.Default))

	for {
		time.Sleep(100 * time.Millisecond)
		if err := DoStuff(context.Background()); err != nil {
			fmt.Println("error", err)
		}
	}
}

func DoStuff(ctx context.Context) (err error) {
	defer mon.Task()(&ctx, "query", []interface{}{[]byte{1, 2, 3}, "args", time.Now()})(&err)

	result, err := ComputeThing(ctx, 1, 2)
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}

func ComputeThing(ctx context.Context, arg1, arg2 int) (res int, err error) {
	defer mon.Task()(&ctx)(&err)

	timer := mon.Timer("subcomputation").Start()
	res = arg1 + arg2
	timer.Stop()

	if res == 3 {
		mon.Event("hit 3")
	}

	mon.BoolVal("was-4").Observe(res == 4)
	mon.IntVal("res").Observe(int64(res))
	mon.DurationVal("took").Observe(time.Second + time.Duration(rand.Intn(int(10*time.Second))))
	mon.Counter("calls").Inc(1)
	mon.Gauge("arg1", func() float64 { return float64(arg1) })
	mon.Meter("arg2").Mark(arg2)

	time.Sleep(time.Second)

	return arg1 + arg2, nil
}
