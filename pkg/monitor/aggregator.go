package monitor

import (
	"sort"

	"github.com/perangel/ddtail/pkg/metrics"
)

// aggregator is a function that computes an aggregation on a collection of metrics.Metrics
type aggregator func(metrics.Metrics) metrics.Metric

// Sum computes the sum over a collection of metrics.Metric
var Sum aggregator = func(data metrics.Metrics) metrics.Metric {
	agg := data[0].Clone()
	for _, d := range data[1:] {
		agg.Add(d)
	}
	return agg
}

// Mean computes the average over a collection of metrics.Metric
var Mean aggregator = func(data metrics.Metrics) metrics.Metric {
	agg := metrics.Float(Sum(data).Float())
	ratio := metrics.Float(float64(1) / float64(len(data)))
	agg.Multiply(&ratio)
	return &agg
}

// Min returns the minimum value in a collection of metrics.Metric
var Min aggregator = func(data metrics.Metrics) metrics.Metric {
	sort.Sort(data)
	return data[0]
}

// Max returns the maximum value in a collection of metrics.Metric
var Max aggregator = func(data metrics.Metrics) metrics.Metric {
	sort.Sort(data)
	return data[len(data)-1]
}
