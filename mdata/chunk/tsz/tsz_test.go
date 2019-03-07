package tsz

import (
	"math/rand"
	"testing"
)

var T uint32
var V float64

func BenchmarkPushSeries4h(b *testing.B) {
	s := NewSeries4h(0)
	N := uint32(b.N)
	for i := uint32(1); i <= N; i++ {
		if i%10 == 0 {
			s.Push(i, 0)
		} else if i%10 == 1 {
			s.Push(i, 1)
		} else {
			s.Push(i, float64(i)+123.45)
		}
	}
	s.Finish()
	b.Logf("Series4h size: %dB", len(s.Bytes()))
}

func BenchmarkPushSeriesLong(b *testing.B) {
	s := NewSeriesLong(0)
	N := uint32(b.N)
	for i := uint32(1); i <= N; i++ {
		if i%10 == 0 {
			s.Push(i, 0)
		} else if i%10 == 1 {
			s.Push(i, 1)
		} else {
			s.Push(i, float64(i)+123.45)
		}
	}
	s.Finish()
	b.Logf("SeriesLong size: %dB", len(s.Bytes()))
}

func BenchmarkIterSeries4h(b *testing.B) {
	s := NewSeries4h(0)
	N := uint32(b.N)
	for i := uint32(1); i <= N; i++ {
		s.Push(i, 123.45)
	}
	b.ResetTimer()
	iter := s.Iter(1)
	var t uint32
	var v float64
	for iter.Next() {
		t, v = iter.Values()
	}
	err := iter.Err()
	if err != nil {
		panic(err)
	}
	T = t
	V = v
}

func BenchmarkIterSeriesLong(b *testing.B) {
	s := NewSeriesLong(0)
	N := uint32(b.N)
	for i := uint32(1); i <= N; i++ {
		s.Push(i, 123.45)
	}
	b.ResetTimer()
	iter := s.Iter()
	var t uint32
	var v float64
	for iter.Next() {
		t, v = iter.Values()
	}
	err := iter.Err()
	if err != nil {
		panic(err)
	}
	T = t
	V = v
}

func BenchmarkIterSeriesLongInterface(b *testing.B) {
	s := NewSeriesLong(0)
	N := uint32(b.N)
	for i := uint32(1); i <= N; i++ {
		s.Push(i, 123.45)
	}
	b.ResetTimer()
	var t uint32
	var v float64
	var iter Iter
	// avoid compiler optimization where it can statically assign the right type
	// and skip the overhead of the interface
	if rand.Intn(1) == 0 {
		iter = s.Iter()
	}
	for iter.Next() {
		t, v = iter.Values()
	}
	err := iter.Err()
	if err != nil {
		panic(err)
	}
	T = t
	V = v
}
