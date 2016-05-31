// Copyright (C) 2016 Space Monkey, Inc.

package monkit

import (
	"math/rand"
)

// lcg is a simple linear congruential generator based on Knuths MMIX.
type lcg uint64

// Make sure lcg is a rand.Source
var _ rand.Source = new(lcg)

func newLCG() lcg { return lcg(rand.Int63()) }

// See Knuth.
const (
	a = 6364136223846793005
	c = 1442695040888963407
	h = 0xffffffff00000000
)

// Uint64 returns a uint64.
func (l *lcg) Uint64() (ret uint64) {
	*l = a**l + c
	ret |= uint64(*l) >> 32
	*l = a**l + c
	ret |= uint64(*l) & h
	return
}

// Int63 returns a positive 63 bit integer in an int64
func (l *lcg) Int63() int64 {
	return int64(l.Uint64() >> 1)
}

// Next advances and returns an lcg. Same thing as lcg(l.Uint64())
func (l *lcg) Next() lcg {
	return lcg(l.Uint64())
}

// Seed sets the state of the lcg.
func (l *lcg) Seed(seed int64) {
	*l = lcg(seed)
}
