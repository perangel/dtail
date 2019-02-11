package metrics

import "sync/atomic"

// Counter is a simple int64 counter
type Counter struct {
	i int64
}

// NewCounter returns a new Counter
func NewCounter() *Counter {
	return &Counter{0}
}

// NewCounterWithValue initializes and returns a new Counter with a specified value
func NewCounterWithValue(i int64) *Counter {
	return &Counter{i}
}

// Value returns the current value of the Counter
func (c *Counter) Value() int64 {
	return c.i
}

// Inc increments a Counter by a specified amount
func (c *Counter) Inc(i int64) {
	atomic.AddInt64(&c.i, 1)
}

// Add adds the value of another Counter
func (c *Counter) Add(other Observable) {
	o := other.(*Counter)
	atomic.AddInt64(&c.i, o.i)
}

// Multiply multiplies self by another Observable
func (c *Counter) Multiply(other Observable) {
	o := other.(*Counter)
	c.i *= o.i
}

// Less compares self to another Observable
func (c *Counter) Less(other Observable) bool {
	o := other.(*Counter)
	return c.i < o.i
}

// Reset resets the Counter to zero
func (c *Counter) Reset() {
	atomic.SwapInt64(&c.i, 0)
}

// Clone returns a copy of a Counter
func (c *Counter) Clone() Observable {
	copy := atomic.LoadInt64(&c.i)
	return &Counter{copy}
}

// Float returns the Counter' value as float64
func (c *Counter) Float() float64 {
	return float64(c.i)
}
