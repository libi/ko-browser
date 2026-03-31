package cmd

import (
	"fmt"
	"time"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for a condition or duration",
}

func init() {
	rootCmd.AddCommand(waitCmd)

	waitCmd.AddCommand(
		&cobra.Command{
			Use:   "time <duration>",
			Short: "Wait for a specified duration (e.g. 2s, 500ms)",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				d, err := time.ParseDuration(args[0])
				if err != nil {
					return fmt.Errorf("invalid duration %q: %w", args[0], err)
				}
				client := session.NewClient(sessionOptions())
				return client.Wait(d)
			},
		},
		&cobra.Command{
			Use:   "selector <css-selector>",
			Short: "Wait until an element matching the CSS selector is visible",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.WaitSelector(args[0])
			},
		},
		&cobra.Command{
			Use:   "url <pattern>",
			Short: "Wait until the URL matches the given pattern (supports * glob)",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.WaitURL(args[0])
			},
		},
		&cobra.Command{
			Use:   "load",
			Short: "Wait until the page is fully loaded",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.WaitLoad()
			},
		},
		&cobra.Command{
			Use:   "text <text>",
			Short: "Wait until the given text appears on the page",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.WaitText(args[0])
			},
		},
		&cobra.Command{
			Use:   "func <js-expression>",
			Short: "Wait until the JavaScript expression evaluates to truthy",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.WaitFunc(args[0])
			},
		},
		&cobra.Command{
			Use:   "hidden <css-selector>",
			Short: "Wait until an element is hidden or removed from the DOM",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.WaitHidden(args[0])
			},
		},
		&cobra.Command{
			Use:   "download <save-dir>",
			Short: "Wait for a download to complete and save to the specified directory",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				filePath, err := client.WaitDownload(args[0])
				if err != nil {
					return err
				}
				return printResult(filePath+"\n", map[string]any{
					"ok":       true,
					"filePath": filePath,
				})
			},
		},
	)
}
