package cmd

import (
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show daemon and browser process status",
		Long:  "Show the status of the daemon and browser process for the current session.\nThis works independently of whether the daemon is running.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := session.GetStatus(rootFlags.session)

			if rootFlags.json {
				return printResult("", info)
			}

			return printResult(session.FormatStatus(info), nil)
		},
	})
}
