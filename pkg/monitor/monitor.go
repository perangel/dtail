package monitor

import "time"

// Observable is an interface for any observable metric.
// For the simplicity of the exercise an Observable is limited to an
// int64 based measturement.
type Observable interface {
	Value() int64
	Reset()
}

// Config describes the configuration for a Monitor
type Config struct {
	// The level of granularity at which the Monitor will observe a metric
	Resolution time.Duration
	// The time frame during which the thresholds are evaluated
	Window time.Duration

	// Threshold value for triggering an alert
	AlertThreshold int64
	// Threshold value for triggering a warning
	WarningThreshold int64
}

// Monitor continuously observes a Metric over time and notifies
// via a channel when the metric it is observing exceeds the configured
// threshold over a given window of time.
//
// For example, it can alert when the metric it is observing exceeds
// a given average over the last X minutes.
//
// Monitor is modeled after a DataDog monitor
type Monitor struct {
	Triggered chan time.Time
	Resolved  chan time.Time

	data       []int64
	resolution time.Duration

	// TODO: Support multiple thresholds
	threshold  int64
	evalWindow time.Duration
	ticker     *time.Ticker
	numTicks   int64

	stopCh chan bool
}

// NewMonitor initializes and returns a new Monitor.
func NewMonitor(config *Config) *Monitor {
	return &Monitor{
		Triggered:  make(chan time.Time),
		Resolved:   make(chan time.Time),
		data:       make([]int64, config.Window/config.Resolution),
		resolution: config.Resolution,
		threshold:  config.AlertThreshold,
		evalWindow: config.Window,
		stopCh:     make(chan bool, 1),
	}
}

// Watch configures the Monitor to watch an Observable
func (m *Monitor) Watch(ob Observable) {
	m.ticker = time.NewTicker(1 * m.resolution)
	for {
		select {
		default:
			// update the current index in the data frame up until it
			// reaches
		case <-m.ticker.C:
			if m.numTicks < int64(len(m.data)) {
				m.data[m.numTicks] = ob.Value()
			} else {
				m.data = append(m.data[1:], ob.Value())
			}
			m.numTicks++
			ob.Reset()
		case <-m.stopCh:
			return
		}
	}
}

// Stop stops a monitor
func (m *Monitor) Stop() {
	m.stopCh <- true
}
