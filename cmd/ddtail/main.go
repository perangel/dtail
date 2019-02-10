package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/perangel/ddtail/pkg/monitor"
	"github.com/perangel/ddtail/pkg/parser"
	"github.com/perangel/ddtail/pkg/tail"
	"github.com/spf13/cobra"
)

const (
	// flag names
	flagAlertThreshold = "alert-threshold"
	flagAlertWindow    = "alert-window"
	flagFollowRetry    = "follow-retry"
)

var (
	// flag vars
	alertThreshold float64
	alertWindow    int
	followRetry    bool
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
		&alertWindow, flagAlertWindow, "w", 120,
		"Time frame (in seconds) for evaluating the alert threshold.",
	)

	ddtailCmd.Flags().BoolVarP(
		&followRetry, flagFollowRetry, "F", false,
		"Keep trying to open a file after it is rename or removed. Useful for logrotate.",
	)
}

func tailFile(cmd *cobra.Command, args []string) error {
	t, err := tail.TailFile(args[0], &tail.Config{Retry: followRetry})
	if err != nil {
		return err
	}

	// create a new monitor
	requestRateMonitor := monitor.NewMonitor(&monitor.Config{
		Resolution:     1 * time.Second,
		AlertThreshold: alertThreshold,
		Aggregtor:      monitor.Mean,
		Window:         2 * time.Minute,
	})

	// create a counter for tracking requests
	requestCount := new(monitor.Counter)
	requestRateMonitor.Watch(requestCount)

	// parse the lines as they are emitted by the Tail, incrementing the
	// counter for each request line
	p := parser.NewParser()
	go func() {
		for line := range t.Lines {
			_, err := p.ParseLine(line)
			if err != nil {
				log.Println("error:", err)
			}
			requestCount.Inc(1)
		}
	}()

	go func() {
		for {
			select {
			case t := <-requestRateMonitor.Triggered:
				fmt.Println("triggered at: ", t)
			case t := <-requestRateMonitor.Resolved:
				fmt.Println("resolved at: ", t)
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
