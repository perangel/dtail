package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/perangel/ddtail/pkg/metrics"
	"github.com/perangel/ddtail/pkg/metrics/collections"
	"github.com/perangel/ddtail/pkg/monitor"
	"github.com/perangel/ddtail/pkg/parser"
	"github.com/perangel/ddtail/pkg/tail"
	"github.com/spf13/cobra"
)

const (
	// flag names
	flagAlertThreshold        = "alert-threshold"
	flagAlertWindow           = "alert-window"
	flagFollowRetry           = "follow-retry"
	flagReportIntervalSeconds = "report-interval"
)

var (
	// flag vars
	alertThreshold        float64
	alertWindow           int
	followRetry           bool
	reportIntervalSeconds int
)

var ddtailCmd = &cobra.Command{
	Use:   "ddtail [FILE]",
	Short: "Analyze a logfile in realtime",
	Long: `
ddtail is a command-line utility for real-time analysis of a live log file (e.g. HTTP access log).
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("missing required argument [FILE]")
		}
		return nil
	},
	RunE: tailFile,
}

func init() {
	ddtailCmd.Flags().Float64VarP(
		&alertThreshold,
		flagAlertThreshold, "t", 10.0,
		"Average request/sec that will trigger an alert within a given alert window.",
	)

	ddtailCmd.Flags().IntVarP(
		&alertWindow, flagAlertWindow, "w", 2,
		"Time frame (in minutes) for evaluating the alert threshold.",
	)

	ddtailCmd.Flags().BoolVarP(
		&followRetry, flagFollowRetry, "F", false,
		"Keep trying to open a file after it is rename or removed. Useful for logrotate.",
	)

	ddtailCmd.Flags().IntVarP(
		&reportIntervalSeconds, flagReportIntervalSeconds, "I", 10,
		"Generate a report of traffic statistics every N seconds",
	)
}

func total4xxResponses(counterMap collections.CounterMap) int64 {
	total := int64(0)
	for k, v := range counterMap {
		// TODO: err on strconv will skip the key. This should be handled better
		i, err := strconv.Atoi(k)
		fmt.Println(i)
		if err != nil {
			continue
		}
		if 400 <= i && i <= 499 {
			total += v.Value()
		}
	}

	return total
}

func total5xxResponses(counterMap collections.CounterMap) int64 {
	total := int64(0)
	for k, v := range counterMap {
		// TODO: err on strconv will skip the key. This should be handled better
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
	filepath := args[0]
	t, err := tail.TailFile(filepath, &tail.Config{Retry: followRetry})
	if err != nil {
		return err
	}

	fmt.Printf("\033[0;34mTailing file %s...\033[0m \n", filepath)

	// create a new monitor
	requestRateMonitor := monitor.NewMonitor(&monitor.Config{
		Resolution:     1 * time.Second,
		AlertThreshold: alertThreshold,
		Aggregator:     monitor.Mean,
		//Window:         time.Duration(alertWindow) * time.Minute,
		Window: 5 * time.Second,
	})

	go func() {
		// create a counter for tracking requests
		// NOTE: requestCounter will be reset at each tick of the Monitor interval,
		// so DO NOT rely on it for aggregate totals during the execution of the program.
		requestCounter := metrics.NewCounter()
		requestRateMonitor.Watch(requestCounter)

		// count total requests handled by ddtail
		totalRequests := metrics.NewCounter()
		requestsByUser := collections.NewCounterMap()
		requestsBySection := collections.NewCounterMap()
		requestsByURI := collections.NewCounterMap()
		requestsByStatusCode := collections.NewCounterMap()

		parser := parser.NewParser()
		tick := time.NewTicker(time.Duration(reportIntervalSeconds) * time.Second)
		for {
			select {
			case line := <-t.Lines:
				request, err := parser.ParseLine(line)
				if err != nil {
					log.Println("parser error: ", err)
				}

				requestCounter.Inc(1)

				requestsByUser.IncKey(request.AuthUser)
				requestsBySection.IncKey(request.Section())
				requestsByURI.IncKey(request.URI)
				requestsByStatusCode.IncKey(fmt.Sprintf("%d", request.StatusCode))
				totalRequests.Inc(1)

			case evt := <-requestRateMonitor.Triggered:
				fmt.Printf("\033[0;31mHigh traffic generated an alert - hits = %.2f, triggered at %v\033[0m \n", evt.Value, evt.Time)

			case evt := <-requestRateMonitor.Resolved:
				fmt.Printf("\033[0;32mHigh traffic alert resolved - hits = %.2f, resolved at %v\033[0m \n", evt.Value, evt.Time)

			case t := <-tick.C:
				fmt.Println()
				fmt.Println("Traffic Report:")
				fmt.Printf("   Current time: %v\n", t)
				fmt.Printf("   Total Requests: %d\n", totalRequests.Value())
				fmt.Printf("   Top 3 users by # of requests: %v\n", requestsByUser.TopNKeys(3))
				fmt.Printf("   Top 3 site sections by # of requests: %v\n", requestsBySection.TopNKeys(3))
				fmt.Printf("   Top 3 URIs by # of requests: %v\n", requestsByURI.TopNKeys(3))
				fmt.Printf("   No. of 4xx responses: %v\n", total4xxResponses(requestsByStatusCode))
				fmt.Printf("   No. of 5xx responses: %v\n", total5xxResponses(requestsByStatusCode))
				fmt.Println()
			}
		}
	}()

	return t.Wait()
}

func main() {
	if err := ddtailCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
