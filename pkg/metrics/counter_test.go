package metrics

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	t.Run("initialize new Counter", func(t *testing.T) {
		c := NewCounter()
		assert.Equal(t, int64(0), c.Value(), "initial value should be 0")
	})

	t.Run("initialize new Counter with value", func(t *testing.T) {
		c := NewCounterWithValue(99)
		assert.Equal(t, int64(99), c.Value(), "initial value should be 99")
	})

	t.Run("get counter value", func(t *testing.T) {
		c := NewCounterWithValue(10)
		assert.Equal(t, int64(10), c.Value(), "value should be 10")
	})

	t.Run("get counter value as a float64", func(t *testing.T) {
		c := NewCounterWithValue(10)
		assert.Equal(t, float64(10), c.Float(), "float value for counter shoudl be 10.0")
	})

	t.Run("increment counter", func(t *testing.T) {
		c := NewCounter()
		c.Inc(1)
		assert.Equal(t, int64(1), c.Value(), "value should have been incremented by 1")
	})
}

func TestCounterImplementsMetricIface(t *testing.T) {
	t.Run("add two counters", func(t *testing.T) {
		c1 := NewCounterWithValue(10)
		c2 := NewCounterWithValue(5)
		c1.Add(c2)
		assert.Equal(t, int64(15), c1.Value(), "new value should be 15")
	})

	t.Run("multiply two counters", func(t *testing.T) {
		c1 := NewCounterWithValue(10)
		c2 := NewCounterWithValue(10)
		c1.Multiply(c2)
		assert.Equal(t, int64(100), c1.Value(), "new value should be 100")
	})

	t.Run("compare two counters", func(t *testing.T) {
		c1 := NewCounterWithValue(1)
		c2 := NewCounterWithValue(10)
		assert.True(t, c1.Less(c2))
		assert.False(t, c2.Less(c1))
	})

	t.Run("reset a counter", func(t *testing.T) {
		c := NewCounterWithValue(10)
		c.Reset()
		assert.Equal(t, int64(0), c.Value(), "value should be reset to 0")
	})

	t.Run("clone a counter", func(t *testing.T) {
		c1 := NewCounterWithValue(10)
		c2 := c1.Clone()
		assert.Equal(t, c1, c2)
		c1.Inc(10)
		assert.NotEqual(t, c1, c2, "incrementing source counter should not increment copy")
	})
}

func TestMetricsImplementsSortIface(t *testing.T) {
	metrics := make(Observables, 10)
	// 1..10
	values := []int64{3, 10, 2, 4, 1, 8, 6, 9, 7, 5}
	for i, v := range values {
		metrics[i] = NewCounterWithValue(v)
	}

	t.Run("implements Len()", func(t *testing.T) {
		assert.Equal(t, 10, len(metrics))
	})

	t.Run("implements Less()", func(t *testing.T) {
		assert.True(t, metrics.Less(0, 9))
	})

	t.Run("implements Swap()", func(t *testing.T) {
		// values (3, 5)
		metrics.Swap(0, len(metrics)-1)
		assert.Equal(t, int64(5), metrics[0].(*Counter).Value())
		assert.Equal(t, int64(3), metrics[len(metrics)-1].(*Counter).Value())
	})

	t.Run("can be sorted", func(t *testing.T) {
		sort.Sort(metrics)
		for i, m := range metrics {
			c := m.(*Counter)
			assert.Equal(t, int64(i+1), c.Value())
		}
	})
}
