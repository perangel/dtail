package collections

import (
	"testing"

	"github.com/perangel/dtail/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestCounterMap(t *testing.T) {
	cm := NewCounterMap()
	cm["jack"] = metrics.NewCounterWithValue(100)
	cm["jill"] = metrics.NewCounterWithValue(200)
	cm["mary"] = metrics.NewCounterWithValue(300)
	cm["bob"] = metrics.NewCounterWithValue(400)
	cm["sally"] = metrics.NewCounterWithValue(500)

	t.Run("top N keys returns keys in descending order (by value)", func(t *testing.T) {
		top := cm.TopNKeys(3)
		assert.Equal(t, []string{"sally", "bob", "mary"}, top)
	})

	t.Run("top N keys returns all keys if N is greater than size of map", func(t *testing.T) {
		top := cm.TopNKeys(10)
		assert.Equal(t, []string{"sally", "bob", "mary", "jill", "jack"}, top)
	})

	t.Run("reset all counters in the map", func(t *testing.T) {
		cm := NewCounterMap()
		cm["jack"] = metrics.NewCounterWithValue(100)
		cm["jill"] = metrics.NewCounterWithValue(200)
		cm["mary"] = metrics.NewCounterWithValue(300)
		cm["bob"] = metrics.NewCounterWithValue(400)
		cm["sally"] = metrics.NewCounterWithValue(500)
		cm.Reset()
		assert.Equal(t, 0, len(cm))
	})
}
