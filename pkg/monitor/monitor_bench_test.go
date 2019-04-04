package monitor

import (
	"testing"

	"github.com/perangel/dtail/pkg/metrics"
)

func BenchmarkAppendPreAlloc(b *testing.B) {
	size := 1024
	s := make([]metrics.Observable, size)
	for i := 0; i < b.N; i++ {
		if i < size {
			s[i] = metrics.NewCounter()
		} else {
			s = append(s[1:], metrics.NewCounter())
		}
	}
}

func BenchmarkAppendNoAlloc(b *testing.B) {
	size := 1024
	s := []metrics.Observable{}
	for i := 0; i < b.N; i++ {
		if i < size {
			s = append(s, metrics.NewCounter())
		} else {
			s = append(s[1:], metrics.NewCounter())
		}
	}
}

func BenchmarkCirclularBuffer(b *testing.B) {
	size := 1024
	s := [1024]metrics.Observable{}
	for i := 0; i < b.N; i++ {
		s[i%size] = metrics.NewCounter()
	}

}
