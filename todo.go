package monitor

//import (
//	"golang.org/x/net/context"
//)

//func (s *Scope) Gauge(name string, cb func() float64) {
//}

//func (s *Scope) Chain(
//	prefix string, data func(cb func(key string, val float64))) {
//}

//type Counter struct{}

//func (s *Scope) Counter(name string) *Counter {
//	return &Counter{}
//}

//func (c *Counter) Inc(delta int) {}
//func (c *Counter) Dec(delta int) { c.Inc(-delta) }

//type Timer struct{}

//func (s *Scope) Timer(name string) *Timer { return &Timer{} }

//type TimerCtx struct{}

//func (t *Timer) Start() *TimerCtx { return &TimerCtx{} }
//func (t *TimerCtx) Stop()         {}

//func (s *Scope) Crit(ctx context.Context)   {}
//func (s *Scope) Debug(ctx context.Context)  {}
//func (s *Scope) Info(ctx context.Context)   {}
//func (s *Scope) Error(ctx context.Context)  {}
//func (s *Scope) Notice(ctx context.Context) {}
//func (s *Scope) Warn(ctx context.Context)   {}

//// Gauge - register a function
//// Counter - inc/dec
//// Timer - distribution/meter

//// Chain - register a subgroup
//// Data - observe a tuple
//// Event/Meter
//// Val
//// Task

//// Log
