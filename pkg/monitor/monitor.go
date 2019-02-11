package monitor

import (
	"time"

	"github.com/perangel/dtail/pkg/metrics"
)

// Config describes the configuration for a Monitor
type Config struct {
	// The level of granularity at which the Monitor will observe a metric
	Resolution time.Duration
	// The time frame during which the thresholds are evaluated
	Window time.Duration

	// An aggregation function (e.g. Mean, Min, Max, Sum, etc)
	// For available aggregator functions see aggregator.go
	Aggregator aggregator

	// Threshold value for triggering an alert
	AlertThreshold float64
}

// monitorEventType is the type of event emitted by the monitor
type monitorEventType string

const (
	// EventTypeTriggered is the event type for a triggered alert
	EventTypeTriggered monitorEventType = "triggered"
	// EventTypeResolved is the event type for a resolved alert
	EventTypeResolved monitorEventType = "resolved"
)

// Event represents a monitor event (e.g. Triggered, Resovled)
type Event struct {
	Type  monitorEventType
	Value float64
	Time  time.Time
}

// Monitor watches a Observable over time and notifies via channel
// when the metric it is observing exceeds the configured
// threshold over a given window of time.
//
// Monitors are configured with an Aggregator function which computes some
// descriptive statistic over the data points that it has collected over
// the evaluation window.
//
// For example, it can alert when the metric it is observing exceeds
// a given average over the last X minutes.
//
// Monitor is modeled after a DataDog monitor
type Monitor struct {
	Triggered chan *Event
	Resolved  chan *Event

	isTriggered bool

	// data is a fixed buffer of datapoints, which is sized to evalWindow/resolution
	// e.g.  2 minutes at a 1-second resolution == [120]metrics.Observable
	data       []metrics.Observable
	resolution time.Duration

	threshold  *metrics.Float
	evalWindow time.Duration
	aggrF      aggregator
	ticker     *time.Ticker
	ticks      *metrics.Counter

	stopCh chan bool
}

// NewMonitor initializes and returns a new Monitor.
func NewMonitor(config *Config) *Monitor {
	threshold := metrics.Float(config.AlertThreshold)
	return &Monitor{
		Triggered:  make(chan *Event),
		Resolved:   make(chan *Event),
		data:       make([]metrics.Observable, config.Window/config.Resolution),
		ticks:      metrics.NewCounter(),
		resolution: config.Resolution,
		threshold:  &threshold,
		evalWindow: config.Window,
		aggrF:      config.Aggregator,
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
		m.Triggered <- &Event{
			Type:  EventTypeTriggered,
			Value: agg.Float(),
			Time:  time.Now().UTC(),
		}
		m.isTriggered = true

	} else if m.isTriggered && agg.Less(m.threshold) {
		// Recover: if we are in a triggered state and we are below the threshold
		m.Resolved <- &Event{
			Type:  EventTypeResolved,
			Value: agg.Float(),
			Time:  time.Now().UTC(),
		}
		m.isTriggered = false
	}
}

// Watch configures the Monitor to watch an Observable
func (m *Monitor) Watch(metric metrics.Observable) {
	go func() {
		m.ticker = time.NewTicker(1 * m.resolution)
		for {
			select {
			case <-m.ticker.C:
				ticks := m.ticks.Value()
				// for the first N ticks, where N == size of m.data, simply set via index
				if ticks < int64(len(m.data)) {
					m.data[ticks] = metric.Clone()
				} else {
					// drop the oldest value and append the latest one to the end
					m.data = append(m.data[1:], metric.Clone())
					// only check the trigger after we have enough datapoints
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
