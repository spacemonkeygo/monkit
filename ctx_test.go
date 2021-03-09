package monkit

import (
	"context"
	"testing"
	"time"
)

// TestLateObserver checks that if you add an observer to a trace after it has
// begun, the original span that started the trace will also be observed
func TestLateObserver(t *testing.T) {
	ctx := context.Background()
	mon := Package()
	mock := &mockSpanObserver{}
	func() {
		// Start the first span, this is the trace
		defer mon.Task()(&ctx)(nil)

		// Start a second span, this is a span on the trace, but still occurs
		// before we register a span observer
		defer mon.Task()(&ctx)(nil)

		// Now start observing all new traces
		mon.r.ObserveTraces(func(t *Trace) {
			t.ObserveSpans(mock)
		})

		// Now go through all root spans, and observe the traces associated
		// with those spans
		mon.r.RootSpans(func(s *Span) {
			s.Trace().ObserveSpans(mock)
		})

		// One more after-the-fact
		defer mon.Task()(&ctx)(nil)

		// Here is a new trace, to prove that is working
		ctx2 := context.Background()
		defer mon.Task()(&ctx2)(nil)
	}()
	expStarts := 2
	expFinishes := 4
	if mock.starts != expStarts {
		t.Errorf("Expected %d, got %d", expStarts, mock.starts)
	}
	if mock.finishes != expFinishes {
		t.Errorf("Expected %d, got %d", expFinishes, mock.finishes)
	}
}

type mockSpanObserver struct {
	starts, finishes int
}

func (m *mockSpanObserver) Start(s *Span) {
	m.starts++
}

func (m *mockSpanObserver) Finish(s *Span, err error, panicked bool, finish time.Time) {
	m.finishes++
}

func BenchmarkTask(b *testing.B) {
	mon := Package()
	pctx := context.Background()
	for i := 0; i < b.N; i++ {
		var err error
		func() {
			ctx := pctx
			defer mon.Task()(&ctx)(&err)
		}()
	}
}
func BenchmarkTaskNested(b *testing.B) {
	mon := Package()
	pctx := context.Background()
	var errout error
	defer mon.Task()(&pctx)(&errout)
	for i := 0; i < b.N; i++ {
		var err error
		func() {
			ctx := pctx
			defer mon.Task()(&ctx)(&err)
		}()
	}
}
