package metrics

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

// Float returns the Counter' value as float64
func (c *Counter) Float() float64 {
	return float64(*c)
}
