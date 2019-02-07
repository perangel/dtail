package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/perangel/ddtail/pkg/parser"
	"github.com/perangel/ddtail/pkg/tail"
	"github.com/spf13/cobra"
)

const (
	flagAlertThreshold   = "alert-threshold"
	flagAlertWindow      = "alert-window"
	flagAutoResumeFollow = "auto-resume"
	flagFileMustExist    = "file-must-exist"
)

var (
	alertThreshold   int
	alertWindow      int
	autoResumeFollow bool
	fileMustExist    bool
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
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.NewParser()

		t, err := tail.TailFile(args[0], &tail.Config{
			FileMustExist:  fileMustExist,
			ResumeWatching: true,
		})
		if err != nil {
			return err
		}

		for line := range t.Lines {
			req, err := p.ParseLine(line)
			if err != nil {
				log.Error(err)
			}
			fmt.Printf("%+v\n", req)
		}

		return nil
	},
}

func init() {
	ddtailCmd.Flags().IntVarP(
		&alertThreshold,
		flagAlertThreshold, "t", 10,
		"Number of requests per second that when exceeded on average over the alert window triggers an alert.",
	)

	ddtailCmd.Flags().IntVarP(
		&alertWindow, flagAlertWindow, "w", 120,
		"Time window (in seconds) during which alerts are evaluated.",
	)

	ddtailCmd.Flags().BoolVarP(
		&fileMustExist, flagFileMustExist, "F", false,
		"If true, the program will exist if the file being tailed does not exist, or is removed.",
	)
}

func main() {
	if err := ddtailCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
