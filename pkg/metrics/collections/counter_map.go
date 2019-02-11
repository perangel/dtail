package collections

import (
	"sort"

	"github.com/perangel/dtail/pkg/metrics"
)

// CounterMap is a collection map of Counters
type CounterMap map[string]*metrics.Counter

// NewCounterMap returns a new CounterMap
func NewCounterMap() CounterMap {
	return make(CounterMap)
}

// IncKey increments the counter stored at a given key
func (c CounterMap) IncKey(key string) {
	if _, ok := c[key]; !ok {
		c[key] = metrics.NewCounterWithValue(1)
	} else {
		c[key].Inc(1)
	}
}

// TopNKeys returns the top N keys with the highest values in the map in descending order
func (c CounterMap) TopNKeys(n int) []string {
	reverseMap := c.reverseMap()

	// metrics.Metrics for sorting
	values := make(metrics.Metrics, len(c))
	i := 0
	for _, v := range c {
		values[i] = v
		i++
	}

	sort.Sort(values)
	topN := make([]string, n)
	for i, v := range values {
		topN[i] = reverseMap[v.(*metrics.Counter)]
	}

	return topN
}

// Reset clears all of the counters in the map
func (c CounterMap) Reset() {
	for k := range c {
		c[k].Reset()
	}
}

// reverseMap returns an inverted map
func (c CounterMap) reverseMap() map[*metrics.Counter]string {
	reverseMap := make(map[*metrics.Counter]string, len(c))
	for k, v := range c {
		reverseMap[v] = k
	}

	return reverseMap
}
