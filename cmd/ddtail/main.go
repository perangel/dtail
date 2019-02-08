package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/perangel/ddtail/pkg/monitor"
	"github.com/perangel/ddtail/pkg/parser"
	"github.com/perangel/ddtail/pkg/stats"
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
	alertThreshold int
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
	ddtailCmd.Flags().IntVarP(
		&alertThreshold,
		flagAlertThreshold, "t", 10,
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
	p := parser.NewParser()

	t, err := tail.TailFile(args[0], &tail.Config{Retry: followRetry})
	if err != nil {
		return err
	}

	mon := monitor.NewMonitor(&monitor.Config{
		Resolution:     1 * time.Second,
		AlertThreshold: 10,
		Window:         2 * time.Minute,
	})

	requestCount := stats.Counter(0)

	go func() {
		for line := range t.Lines {
			requestCount.Inc(1)
			_, err := p.ParseLine(line)
			if err != nil {
				log.Println("error:", err)
			}
		}
	}()

	mon.Watch(requestCount)

	return t.Wait()
}

func main() {
	if err := ddtailCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
