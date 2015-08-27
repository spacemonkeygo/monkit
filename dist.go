package monitor

import (
	"github.com/bmizerany/perks/quantile"
)

var (
	// need to make sure to have 0, .5, and 1
	ObservedQuantiles = []float64{0, .25, .5, .75, .90, .95, .99, 1}
)

// dist is not threadsafe
type dist struct {
	q      *quantile.Stream
	recent float64
}

func newDist() dist {
	return dist{q: quantile.NewTargeted(ObservedQuantiles...)}
}

func (d *dist) Stats() (min, med, max, recent float64) {
	return d.q.Query(0), d.q.Query(.5), d.q.Query(1), d.recent
}

func (d *dist) Insert(val float64) {
	d.q.Insert(val)
	d.recent = val
}

func (d *dist) Query(quantile float64) (rv float64) {
	return d.q.Query(quantile)
}
