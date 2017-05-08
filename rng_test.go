// Copyright (C) 2016 Space Monkey, Inc.

package monkit

import "testing"

var sink uint64

func BenchmarkLCG(b *testing.B) {
	l := newLCG()

	b.SetBytes(8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink += l.Uint64()
	}
}

func BenchmarkXORShift64(b *testing.B) {
	x := newXORShift64()

	b.SetBytes(8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink += x.Uint64()
	}
}

func BenchmarkXORShift1024(b *testing.B) {
	x := newXORShift1024()

	b.SetBytes(8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink += x.Uint64()
	}
}

func BenchmarkXORShift128(b *testing.B) {
	x := newXORShift128()

	b.SetBytes(8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink += x.Uint64()
	}
}
