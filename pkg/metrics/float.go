package metrics

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

// Float returns the Float's values as float64
func (f *Float) Float() float64 {
	return float64(*f)
}
