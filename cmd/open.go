package cmd

import (
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "open <url>",
		Short: "Open a URL in the browser session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.Open(args[0]); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})
}
