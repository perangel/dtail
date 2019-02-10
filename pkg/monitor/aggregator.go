package monitor

import "sort"

// aggregator is a function that computes an aggregation on a collection of Metrics
type aggregator func(Metrics) Metric

// Sum computes the sum over a collection of Metric
var Sum aggregator = func(data Metrics) Metric {
	agg := data[0].Clone()
	for _, d := range data[1:] {
		agg.Add(d)
	}
	return agg
}

// Mean computes the average over a collection of Metric
var Mean aggregator = func(data Metrics) Metric {
	agg := Float(Sum(data).FloatValue())
	ratio := Float(1 / len(data))
	agg.Multiply(&ratio)
	return &agg
}

// Min returns the minimum value in a collection of Metric
var Min aggregator = func(data Metrics) Metric {
	sort.Sort(data)
	return data[0]
}

// Max returns the maximum value in a collection of Metric
var Max aggregator = func(data Metrics) Metric {
	sort.Sort(data)
	return data[len(data)-1]
}
