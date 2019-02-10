package monitor

import (
	"time"
)

// Config describes the configuration for a Monitor
type Config struct {
	// The level of granularity at which the Monitor will observe a metric
	Resolution time.Duration
	// The time frame during which the thresholds are evaluated
	Window time.Duration

	// The aggregation function
	Aggregtor aggregator

	// Threshold value for triggering an alert
	AlertThreshold float64
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

	isTriggered bool

	data       []Metric
	resolution time.Duration

	threshold  *Float
	evalWindow time.Duration
	aggrF      aggregator
	ticker     *time.Ticker
	ticks      *Counter

	stopCh chan bool
}

// NewMonitor initializes and returns a new Monitor.
func NewMonitor(config *Config) *Monitor {
	threshold := Float(config.AlertThreshold)
	return &Monitor{
		Triggered:  make(chan time.Time),
		Resolved:   make(chan time.Time),
		data:       make([]Metric, config.Window/config.Resolution),
		ticks:      NewCounter(),
		resolution: config.Resolution,
		threshold:  &threshold,
		evalWindow: config.Window,
		aggrF:      config.Aggregtor,
		stopCh:     make(chan bool, 1),
	}
}

// checkTrigger runs the aggregator function over the monitor's collected data.
// If the result below the configured threshold then the Monitor notifies the time at which the
// alert was triggered via the Triggered channel. If the Monitor was previously triggered
// and the value is now below the threshold then the Montior notifies via the Resolved channel.
func (m *Monitor) checkTrigger() {
	agg := m.aggrF(m.data)

	if !m.isTriggered && !agg.Less(m.threshold) {
		// Alert: if we are not in a triggered state and we've hit the threshold
		m.Triggered <- time.Now().UTC()
		m.isTriggered = true

	} else if m.isTriggered && agg.Less(m.threshold) {
		// Recover: if we are in a triggered state and we are below the threshold
		m.Resolved <- time.Now().UTC()
		m.isTriggered = false
	}
}

// Watch configures the Monitor to watch an Metric
func (m *Monitor) Watch(metric Metric) {
	go func() {
		m.ticker = time.NewTicker(1 * m.resolution)
		for {
			select {
			case <-m.ticker.C:
				ticks := m.ticks.Value()
				if ticks < int64(len(m.data)) {
					m.data[ticks] = metric.Clone()
				} else {
					m.data = append(m.data[1:], metric)
					m.checkTrigger()
				}
				m.ticks.Inc(1)
				metric.Reset()
			case <-m.stopCh:
				return
			}
		}
	}()
}

// Stop stops a monitor
func (m *Monitor) Stop() {
	m.stopCh <- true
}
