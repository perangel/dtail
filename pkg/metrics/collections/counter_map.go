package collections

import (
	"fmt"
	"sort"

	"github.com/perangel/dtail/pkg/metrics"
)

// used for debugging
func printCounterMap(m CounterMap) {
	for k, v := range m {
		fmt.Printf("%v: %v\n", k, v.Value())
	}
}

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
	revMap := c.reverseMap()

	// metrics.Observables for sorting
	values := make(metrics.Observables, len(c))
	i := 0
	for _, v := range c {
		values[i] = v
		i++
	}

	// FIXME: Hack. First we sort then reverse in place
	sort.Sort(values)
	for i := len(values)/2 - 1; i >= 0; i-- {
		j := len(values) - 1 - i
		values[i], values[j] = values[j], values[i]
	}

	topNKeys := []string{}
	for i := 0; i < len(values); i++ {
		if i == n {
			break
		}
		topNKeys = append(topNKeys, revMap[values[i].(*metrics.Counter)])
	}

	return topNKeys
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
