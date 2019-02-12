package monitor

import (
	"testing"
	"time"

	"github.com/perangel/dtail/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

// simulateHighTraffic will simulate a high traffic volume by rapidly incrementing the counter over a given duration
func simulateHighTraffic(counter *metrics.Counter, cancelCh <-chan int) {
	for {
		select {
		case <-cancelCh:
			return
		default:
			// rate: 20/s
			counter.Inc(1)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// simulateLowTraffic will simulate low traffic volum by slowly incrementing the counter over a given duration
func simulateLowTraffic(counter *metrics.Counter, cancelCh <-chan int) {
	for {
		select {
		case <-cancelCh:
			return
		default:
			// rate: 1/s
			counter.Inc(1)
			time.Sleep(1 * time.Second)
		}
	}
}

// TestMonitorAlerting simulates different traffic profiles by manually manipulating the counter that the
// Monitor is watching. These tests are fairly high-level, in the sense that they depend on the execution of the Monitor
// over time to prove the correct behavior.
func TestMonitorAlerting(t *testing.T) {
	monitor := NewMonitor(&Config{
		Resolution:     1 * time.Second,
		Window:         5 * time.Second,
		Aggregator:     Mean,
		AlertThreshold: 5,
	})

	counter := metrics.NewCounter()
	monitor.Watch(counter)

	t.Run("high traffic triggers alert", func(t *testing.T) {
		// start from a non-triggered state
		monitor.isTriggered = false
		cancelCh := make(chan int, 1)
		go simulateHighTraffic(counter, cancelCh)
		for {
			select {
			case <-time.After(10 * time.Second):
				cancelCh <- 1
				t.Fail()
				return
			case evt := <-monitor.Triggered:
				cancelCh <- 1
				assert.True(t, evt.Value > monitor.threshold.Float())
				return
			case <-monitor.Resolved:
				cancelCh <- 1
				t.Fail()
				return
			}
		}
	})

	t.Run("low traffic after triggering alert resolves alert", func(t *testing.T) {
		// start from a triggered state
		monitor.isTriggered = true
		cancelCh := make(chan int, 1)
		go simulateLowTraffic(counter, cancelCh)
		for {
			select {
			case <-time.After(10 * time.Second):
				cancelCh <- 1
				t.Fail()
				return
			case <-monitor.Triggered:
				cancelCh <- 1
				t.Fail()
				return
			case evt := <-monitor.Resolved:
				cancelCh <- 1
				assert.True(t, evt.Value < monitor.threshold.Float())
				return
			}
		}
	})

	t.Run("low traffic does not trigger an alert", func(t *testing.T) {
		// start from a non-triggered state
		monitor.isTriggered = false
		cancelCh := make(chan int, 1)
		go simulateLowTraffic(counter, cancelCh)
		for {
			select {
			case <-time.After(5 * time.Second):
				cancelCh <- 1
				return
			case <-monitor.Triggered:
				cancelCh <- 1
				t.Fail()
				return
			case <-monitor.Resolved:
				cancelCh <- 1
				t.Fail()
				return
			}
		}
	})
}
