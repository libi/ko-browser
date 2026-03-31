package cmd

import (
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:    "_daemon",
		Short:  "Run the background browser session daemon",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return session.RunDaemon(sessionOptions())
		},
	})
}
