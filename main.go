package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/perangel/dtail/pkg/metrics"
	"github.com/perangel/dtail/pkg/metrics/collections"
	"github.com/perangel/dtail/pkg/monitor"
	"github.com/perangel/dtail/pkg/parser"
	"github.com/perangel/dtail/pkg/tail"
	"github.com/spf13/cobra"
)

var (
	// flag vars
	monitorAlertThreshold float64
	monitorAlertWindow    time.Duration
	monitorResolution     time.Duration
	retryFollow           bool
	reportInterval        time.Duration
)

const (
	defaultLogPath = "/tmp/access.log"
)

var dtailCmd = &cobra.Command{
	Use:   "dtail [FILE]",
	Short: "Tail, with more details",
	Long: `
dtail is a cli-tool for realtime monitoring of structured log files (e.g. HTTP access log).
`,
	RunE: tailFile,
}

func init() {
	dtailCmd.Flags().Float64VarP(
		&monitorAlertThreshold,
		"alert-threshold", "t", 10.0,
		"Threshold value for triggering an alert during the monitor's alert window.",
	)

	dtailCmd.Flags().DurationVarP(
		&monitorAlertWindow,
		"alert-window", "w", 2*time.Minute,
		"Time frame for evaluating a metric against the alert threshold.",
	)

	dtailCmd.Flags().DurationVarP(
		&monitorResolution,
		"monitor-resolution", "r", 1*time.Second,
		"Monitor resolution (e.g. 30s, 1m, 5h)",
	)

	dtailCmd.Flags().BoolVarP(
		&retryFollow,
		"retry-follow", "F", false,
		"Retry file after rename or deletion. Similar to `tail -F`.",
	)

	dtailCmd.Flags().DurationVarP(
		&reportInterval,
		"report-interval", "i", 10*time.Second,
		"Print a report at the given interval (e.g. 30s, 1m, 5h)",
	)
}

// TODO: Move to DSL/query package
func total4xxResponses(counterMap collections.CounterMap) int64 {
	total := int64(0)
	for k, v := range counterMap {
		// FIXME: Don't skip key on error
		i, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		if 400 <= i && i <= 499 {
			total += v.Value()
		}
	}

	return total
}

// TODO: Move to DSL/query package
func total5xxResponses(counterMap collections.CounterMap) int64 {
	total := int64(0)
	for k, v := range counterMap {
		// FIXME: Don't skip key on error
		i, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		if 500 <= i && i <= 599 {
			total += v.Value()
		}
	}

	return total

}

func tailFile(cmd *cobra.Command, args []string) error {
	var filepath string
	if len(args) < 1 {
		filepath = defaultLogPath
	} else {
		filepath = args[0]
	}

	t, err := tail.TailFile(filepath, &tail.Config{Retry: retryFollow})
	if err != nil {
		return err
	}

	fmt.Printf("\033[0;34mTailing file %s...\033[0m \n", filepath)

	// create a monitor for request rate
	requestRateMonitor := monitor.NewMonitor(&monitor.Config{
		Aggregator:     monitor.Mean, // TODO: accept aggregation on the command line
		AlertThreshold: monitorAlertThreshold,
		Resolution:     monitorResolution,
		Window:         monitorAlertWindow,
	})

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownCh
		t.Stop()
		requestRateMonitor.Stop()
	}()

	// TODO: Refactor this to pkg/dtail
	go func() {

		// NOTE: `requestCounter` will be reset at each tick of the Monitor interval,
		// so DO NOT rely on it for aggregate totals during the execution of the program.
		requestCounter := metrics.NewCounter()
		requestRateMonitor.Watch(requestCounter)

		// count total requests handled by dtail
		totalRequests := metrics.NewCounter()
		requestsByUser := collections.NewCounterMap()
		requestsByIP := collections.NewCounterMap()
		requestsBySection := collections.NewCounterMap()
		requestsByURI := collections.NewCounterMap()
		requestsByStatusCode := collections.NewCounterMap()

		parser := parser.NewParser()
		reportTick := time.NewTicker(reportInterval)
		for {
			select {
			case line := <-t.Lines:
				request, err := parser.ParseLine(line)
				if err != nil {
					log.Println("parser error: ", err)
				}

				requestCounter.Inc(1)

				requestsByUser.IncKey(request.AuthUser)
				requestsByIP.IncKey(request.RemoteHost)
				requestsBySection.IncKey(request.Section())
				requestsByURI.IncKey(request.URI)
				requestsByStatusCode.IncKey(fmt.Sprintf("%d", request.StatusCode))
				totalRequests.Inc(1)

			case evt := <-requestRateMonitor.Triggered:
				fmt.Printf("\033[0;31mHigh traffic generated an alert - hits = %.2f, triggered at %v\033[0m \n", evt.Value, evt.Time)

			case evt := <-requestRateMonitor.Resolved:
				fmt.Printf("\033[0;32mHigh traffic alert resolved - hits = %.2f, resolved at %v\033[0m \n", evt.Value, evt.Time)

			case t := <-reportTick.C:
				fmt.Println()
				fmt.Println("Traffic Report:")
				fmt.Printf("   Current time: %v\n", t)
				fmt.Printf("   Total Requests: %d\n", totalRequests.Value())
				fmt.Printf("   Top 3 IPs by # of requests: %v\n", requestsByIP.TopNKeys(3))
				fmt.Printf("   Top 3 users by # of requests: %v\n", requestsByUser.TopNKeys(3))
				fmt.Printf("   Top 3 site sections by # of requests: %v\n", requestsBySection.TopNKeys(3))
				fmt.Printf("   Top 3 URIs by # of requests: %v\n", requestsByURI.TopNKeys(3))
				fmt.Printf("   No. of 4xx responses: %v\n", total4xxResponses(requestsByStatusCode))
				fmt.Printf("   No. of 5xx responses: %v\n", total5xxResponses(requestsByStatusCode))
				fmt.Println()

				// Reset all of the counters
				totalRequests.Reset()
				requestsByIP.Reset()
				requestsByUser.Reset()
				requestsBySection.Reset()
				requestsByURI.Reset()
				requestsByStatusCode.Reset()
			}
		}
	}()

	return t.Wait()
}

func main() {
	if err := dtailCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
