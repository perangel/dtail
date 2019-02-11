package monitor

import (
	"sort"

	"github.com/perangel/dtail/pkg/metrics"
)

// aggregator is a function that computes an aggregation on a collection of metrics.Observables
type aggregator func(metrics.Observables) metrics.Observable

// Sum computes the sum over a collection of metrics.Observable
var Sum aggregator = func(data metrics.Observables) metrics.Observable {
	agg := data[0].Clone()
	for _, d := range data[1:] {
		agg.Add(d)
	}
	return agg
}

// Mean computes the average over a collection of metrics.Observable
var Mean aggregator = func(data metrics.Observables) metrics.Observable {
	agg := metrics.Float(Sum(data).Float())
	ratio := metrics.Float(float64(1) / float64(len(data)))
	agg.Multiply(&ratio)
	return &agg
}

// Min returns the minimum value in a collection of metrics.Observable
var Min aggregator = func(data metrics.Observables) metrics.Observable {
	sort.Sort(data)
	return data[0]
}

// Max returns the maximum value in a collection of metrics.Observable
var Max aggregator = func(data metrics.Observables) metrics.Observable {
	sort.Sort(data)
	return data[len(data)-1]
}
