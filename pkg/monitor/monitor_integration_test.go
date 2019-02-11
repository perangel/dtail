// +build integration

package monitor

import (
	"testing"
	"time"

	"github.com/perangel/dtail/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

// simulateHighTraffic will simulate a high traffic volume by rapidly incrementing the counter over a given duration
func simulateHighTraffic(counter *metrics.Counter, duration time.Duration) {
	for {
		select {
		case <-time.After(duration * time.Second):
			return
		default:
			counter.Inc(1)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// simulateLowTraffic will simulate low traffic volum by slowly incrementing the counter over a given duration
func simulateLowTraffic(counter *metrics.Counter, duration time.Duration) {
	for {
		select {
		case <-time.After(duration * time.Second):
			return
		default:
			counter.Inc(1)
			time.Sleep(1 * time.Second)
		}
	}
}

// TestMonitorAlert simulates different traffic profiles by manually manipulating the counter that the
// Monitor is watching.
func TestMonitorAlert(t *testing.T) {
	monitor := NewMonitor(&Config{
		Resolution:     1 * time.Second,
		Window:         3 * time.Second,
		Aggregator:     Mean,
		AlertThreshold: 10,
	})

	counter := metrics.NewCounter()
	monitor.Watch(counter)

	t.Run("low traffic does not trigger an alert", func(t *testing.T) {
		// start from a non-triggered state
		monitor.isTriggered = false
		go simulateLowTraffic(counter, 5*time.Second)
		for {
			select {
			case <-time.After(10 * time.Second):
				return
			case <-monitor.Triggered:
				t.Fail()
				return
			case <-monitor.Resolved:
				t.Fail()
				return
			}
		}
	})

	t.Run("high traffic triggers alert", func(t *testing.T) {
		// start from a non-triggered state
		monitor.isTriggered = false
		go simulateHighTraffic(counter, 5*time.Second)
		for {
			select {
			case <-time.After(10 * time.Second):
				return
			case evt := <-monitor.Triggered:
				assert.True(t, evt.Value > monitor.threshold.Float())
				return
			case <-monitor.Resolved:
				t.Fail()
				return
			}
		}
	})

	t.Run("low traffic after triggering alert resolves alert", func(t *testing.T) {
		// start from a triggered state
		monitor.isTriggered = true
		go simulateLowTraffic(counter, 5*time.Second)
		for {
			select {
			case <-time.After(10 * time.Second):
				return
			case <-monitor.Triggered:
				t.Fail()
				return
			case evt := <-monitor.Resolved:
				assert.True(t, evt.Value < monitor.threshold.Float())
				return
			}
		}
	})
}
