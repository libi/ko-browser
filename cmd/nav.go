package cmd

import (
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "back",
			Short: "Navigate back in browser history",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.Back()
			},
		},
		&cobra.Command{
			Use:   "forward",
			Short: "Navigate forward in browser history",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.Forward()
			},
		},
		&cobra.Command{
			Use:   "reload",
			Short: "Reload the current page",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				return client.Reload()
			},
		},
	)
}
