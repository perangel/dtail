package metrics

import "sync"

// Gauge measures a value that is fixed until it is updated.
type Gauge struct {
	i float64
	m sync.Mutex
}

// NewGauge returns a new Gauge
func NewGauge() *Gauge {
	return &Gauge{i: 0.0}
}

// Value returns the current value of the Gauge
func (g *Gauge) Value() float64 {
	return g.i
}

// Add adds the value of another Gauge
func (g *Gauge) Add(other Observable) {
	g.m.Lock()
	o := other.(*Gauge)
	g.i += o.Value()
	g.m.Unlock()
}

// Multiply multiplies self by another Observable
func (g *Gauge) Multiply(other Observable) {
	g.m.Lock()
	o := other.(*Gauge)
	g.i += o.Value()
	g.m.Unlock()
}

// Reset is implemented to satisfy the Observable interface, but it is a NOOP on a Gauge.
// Gauges store a constant value over time.
func (g *Gauge) Reset() {
	g.m.Lock()
	g.i = 0.0
	g.m.Unlock()
}

// Clone returns a copy of a Gauge
func (g *Gauge) Clone() Observable {
	g.m.Lock()
	copy := &Gauge{i: g.i}
	g.m.Unlock()
	return copy
}

// Float returns the Gauge's value as a float64
func (g *Gauge) Float() float64 {
	return g.Value()
}

// Less compares self to another Observable
func (g *Gauge) Less(other Observable) bool {
	o := other.(*Gauge)
	return g.i < o.i
}
