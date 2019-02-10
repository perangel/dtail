package monitor

// Metric is an interface for an observable metric.
type Metric interface {
	Add(other Metric)
	Multiply(other Metric)
	Less(other Metric) bool
	FloatValue() float64
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

// Float is a 64-bit observable float
type Float float64

// Add adds the value of another Float
func (f *Float) Add(other Metric) {
	o := other.(*Float)
	*f += *o
}

// Multiply multiplies self by another Metric
func (f *Float) Multiply(other Metric) {
	o := other.(*Float)
	*f *= *o
}

// Less compares self to another Metric
func (f *Float) Less(other Metric) bool {
	o := other.(*Float)
	return *f < *o
}

// Reset resets the Float to its zero-value
func (f *Float) Reset() {
	*f = 0.0
}

// Clone returns a copy of a Float
func (f *Float) Clone() Metric {
	nf := Float(*f)
	return &nf
}

// FloatValue returns the Float's values as float64
func (f *Float) FloatValue() float64 {
	return float64(*f)
}

// Counter is a simple int64 counter
type Counter int64

// NewCounter returns a new Counter
func NewCounter() *Counter {
	return new(Counter)
}

// NewCounterWithValue initializes and returns a new Counter with a specified value
func NewCounterWithValue(i int64) *Counter {
	c := Counter(i)
	return &c
}

// Value returns the current value of the Counter
func (c *Counter) Value() int64 {
	return int64(*c)
}

// Inc increments a Counter by a specified amount
func (c *Counter) Inc(i int64) {
	*c += Counter(i)
}

// Add adds the value of another Counter
func (c *Counter) Add(other Metric) {
	o := other.(*Counter)
	*c += *o
}

// Multiply multiplies self by another Metric
func (c *Counter) Multiply(other Metric) {
	o := other.(*Counter)
	*c *= *o
}

// Less compares self to another Metric
func (c *Counter) Less(other Metric) bool {
	o := other.(*Counter)
	return *c < *o
}

// Reset resets the Counter to zero
func (c *Counter) Reset() {
	*c = 0
}

// Clone returns a copy of a Counter
func (c *Counter) Clone() Metric {
	nc := Counter(*c)
	return &nc
}

// FloatValue returns the Counter' value as float64
func (c *Counter) FloatValue() float64 {
	return float64(*c)
}

// TODO: Gauge, Duration, etc
