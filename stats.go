package monitor

type StatSource interface {
	Stats(cb func(name string, val float64))
}
