package stats

// Counter is a simple int64 counter
type Counter int64

// CounterWithValue initializes and returns a new Counter with a specified value
func CounterWithValue(i int64) Counter {
	return Counter(i)
}

// Inc increments a Counter by a specified amount
func (c Counter) Inc(i int64) {
	c += Counter(i)
}

// Value returns the current value of the Counter
func (c Counter) Value() int64 {
	return int64(c)
}

// Reset resets the Counter to zero
func (c Counter) Reset() {
	c = 0
}
