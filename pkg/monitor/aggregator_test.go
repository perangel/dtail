package monitor

import (
	"testing"

	"github.com/perangel/dtail/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestAggregatorOverCounters(t *testing.T) {
	data := make(metrics.Observables, 5)

	for i, v := range []int64{1, 2, 3, 4, 5} {
		data[i] = metrics.NewCounterWithValue(v)
	}

	t.Run("calculate sum over a collection of counters", func(t *testing.T) {
		sum := Sum(data)
		assert.Equal(t, 15.0, sum.Float())
	})

	t.Run("calculate mean over a collection of counters", func(t *testing.T) {
		m := Mean(data)
		assert.Equal(t, 3.0, m.Float())
	})

	t.Run("calculate min over a collection of counters", func(t *testing.T) {
		min := Min(data)
		assert.Equal(t, 1.0, min.Float())
	})

	t.Run("calculate max over a collection of counters", func(t *testing.T) {
		max := Max(data)
		assert.Equal(t, 5.0, max.Float())
	})
}
