package metrics

// Metric is an interface for an observable metric.
type Metric interface {
	Add(other Metric)
	Multiply(other Metric)
	Less(other Metric) bool
	Float() float64
	Reset()
	Clone() Metric
}

// Metrics is a collection of Metric that implements sort.Interface
type Metrics []Metric

func (m Metrics) Len() int {
	return len(m)
}

func (m Metrics) Less(i, j int) bool {
	return m[i].Less(m[j])
}

func (m Metrics) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
