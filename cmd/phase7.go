package cmd

import (
	"fmt"

	"github.com/libi/ko-browser/browser"
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	// ---- diff ----
	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare snapshots or screenshots",
	}

	// diff snapshot
	diffSnapCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Compare current snapshot with previous or baseline",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			baseline, _ := cmd.Flags().GetString("baseline")
			client := session.NewClient(sessionOptions())
			resp, err := client.DiffSnapshot(baseline, browser.SnapshotOptions{})
			if err != nil {
				return err
			}

			if rootFlags.json {
				return printResult("", resp.DiffResult)
			}

			if resp.DiffResult != nil && !resp.DiffResult.Changed {
				return printResult("No differences found\n", resp.DiffResult)
			}

			return printResult(resp.Text, resp.DiffResult)
		},
	}
	diffSnapCmd.Flags().StringP("baseline", "b", "", "baseline snapshot file to compare against")
	diffCmd.AddCommand(diffSnapCmd)

	// diff screenshot
	diffScreenCmd := &cobra.Command{
		Use:   "screenshot",
		Short: "Pixel-level screenshot comparison with baseline",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			baseline, _ := cmd.Flags().GetString("baseline")
			output, _ := cmd.Flags().GetString("output")
			threshold, _ := cmd.Flags().GetFloat64("threshold")
			fullPage, _ := cmd.Flags().GetBool("full")

			if baseline == "" {
				return fmt.Errorf("--baseline is required for screenshot diff")
			}

			client := session.NewClient(sessionOptions())
			resp, err := client.DiffScreenshot(baseline, output, threshold, fullPage)
			if err != nil {
				return err
			}

			if rootFlags.json {
				return printResult("", resp.ScreenshotDiff)
			}

			if resp.ScreenshotDiff != nil {
				result := resp.ScreenshotDiff
				text := fmt.Sprintf("Diff: %.2f%% pixels changed (%d / %d)\n",
					result.DiffPercent, result.DiffCount, result.TotalPixels)
				if output != "" && result.Changed {
					text += fmt.Sprintf("Diff image saved to: %s\n", output)
				}
				return printResult(text, result)
			}
			return nil
		},
	}
	diffScreenCmd.Flags().StringP("baseline", "b", "", "baseline PNG image (required)")
	diffScreenCmd.Flags().StringP("output", "o", "", "output path for diff image")
	diffScreenCmd.Flags().Float64P("threshold", "t", 0.1, "color distance threshold (0-1)")
	diffScreenCmd.Flags().Bool("full", false, "full page screenshot")
	diffCmd.AddCommand(diffScreenCmd)

	// diff url
	diffURLCmd := &cobra.Command{
		Use:   "url <url1> <url2>",
		Short: "Compare two URLs by navigating to each and diffing",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			screenshot, _ := cmd.Flags().GetBool("screenshot")
			fullPage, _ := cmd.Flags().GetBool("full")
			threshold, _ := cmd.Flags().GetFloat64("threshold")

			client := session.NewClient(sessionOptions())
			resp, err := client.DiffURL(args[0], args[1], screenshot, fullPage, threshold, browser.SnapshotOptions{})
			if err != nil {
				return err
			}

			if rootFlags.json {
				return printResult("", resp.DiffURLResult)
			}

			return printResult(resp.Text, resp.DiffURLResult)
		},
	}
	diffURLCmd.Flags().Bool("screenshot", false, "also compare screenshots")
	diffURLCmd.Flags().Bool("full", false, "full page screenshot")
	diffURLCmd.Flags().Float64("threshold", 0.1, "screenshot diff threshold (0-1)")
	diffCmd.AddCommand(diffURLCmd)

	rootCmd.AddCommand(diffCmd)

	// ---- trace ----
	traceCmd := &cobra.Command{
		Use:   "trace",
		Short: "Record Chrome trace",
	}

	traceStartCmd := &cobra.Command{
		Use:   "start [path]",
		Short: "Start recording a trace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			categories, _ := cmd.Flags().GetString("categories")
			client := session.NewClient(sessionOptions())
			if err := client.TraceStart(categories); err != nil {
				return err
			}
			return printResult("Trace recording started\n", map[string]any{"ok": true})
		},
	}
	traceStartCmd.Flags().String("categories", "", "comma-separated tracing categories")
	traceCmd.AddCommand(traceStartCmd)

	traceCmd.AddCommand(&cobra.Command{
		Use:   "stop [path]",
		Short: "Stop recording and save trace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputPath := ""
			if len(args) > 0 {
				outputPath = args[0]
			}
			if outputPath == "" {
				outputPath = "trace.json"
			}
			client := session.NewClient(sessionOptions())
			if err := client.TraceStop(outputPath); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Trace saved to %s\n", outputPath), map[string]any{
				"ok":   true,
				"path": outputPath,
			})
		},
	})
	rootCmd.AddCommand(traceCmd)

	// ---- profiler ----
	profilerCmd := &cobra.Command{
		Use:   "profiler",
		Short: "CPU profiling",
	}

	profilerCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start CPU profiling",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.ProfilerStart(); err != nil {
				return err
			}
			return printResult("Profiler started\n", map[string]any{"ok": true})
		},
	})

	profilerCmd.AddCommand(&cobra.Command{
		Use:   "stop [path]",
		Short: "Stop profiling and save result",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputPath := ""
			if len(args) > 0 {
				outputPath = args[0]
			}
			if outputPath == "" {
				outputPath = "profile.json"
			}
			client := session.NewClient(sessionOptions())
			if err := client.ProfilerStop(outputPath); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Profile saved to %s\n", outputPath), map[string]any{
				"ok":   true,
				"path": outputPath,
			})
		},
	})
	rootCmd.AddCommand(profilerCmd)

	// ---- record ----
	recordCmd := &cobra.Command{
		Use:   "record",
		Short: "Record browser session as screenshots",
	}

	recordCmd.AddCommand(&cobra.Command{
		Use:   "start <path>",
		Short: "Start recording",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.RecordStart(args[0]); err != nil {
				return err
			}
			return printResult("Recording started\n", map[string]any{"ok": true})
		},
	})

	recordCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop recording and save",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			frames, err := client.RecordStop()
			if err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Recording stopped: %d frames captured\n", frames), map[string]any{
				"ok":     true,
				"frames": frames,
			})
		},
	})
	rootCmd.AddCommand(recordCmd)

	// ---- state ----
	stateCmd := &cobra.Command{
		Use:   "state",
		Short: "Export or import browser state (cookies + localStorage)",
	}

	stateCmd.AddCommand(&cobra.Command{
		Use:   "export <path>",
		Short: "Export browser state to JSON file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.ExportState(args[0]); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("State exported to %s\n", args[0]), map[string]any{
				"ok":   true,
				"path": args[0],
			})
		},
	})

	stateCmd.AddCommand(&cobra.Command{
		Use:   "import <path>",
		Short: "Import browser state from JSON file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.ImportState(args[0]); err != nil {
				return err
			}
			return printResult("State imported successfully\n", map[string]any{"ok": true})
		},
	})
	rootCmd.AddCommand(stateCmd)
}
