package monitor

import (
	"sync"

	"github.com/bmizerany/perks/quantile"
)

var (
	// need to make sure to have 0, .5, and 1
	ObservedQuantiles = []float64{0, .25, .5, .75, .90, .95, .99, 1}
)

type Dist struct {
	mtx    sync.Mutex
	q      *quantile.Stream
	recent float64
}

func newDist() *Dist {
	return &Dist{q: quantile.NewTargeted(ObservedQuantiles...)}
}

func (d *Dist) Stats(cb func(name string, val float64)) {
	d.mtx.Lock()
	min, med, max, recent := d.q.Query(0), d.q.Query(.5), d.q.Query(1), d.recent
	d.mtx.Unlock()
	cb("min", min)
	cb("med", med)
	cb("max", max)
	cb("recent", recent)
}

func (d *Dist) Insert(val float64) {
	d.mtx.Lock()
	d.q.Insert(val)
	d.recent = val
	d.mtx.Unlock()
}

func (d *Dist) Query(quantile float64) (rv float64) {
	d.mtx.Lock()
	rv = d.q.Query(quantile)
	d.mtx.Unlock()
	return rv
}
