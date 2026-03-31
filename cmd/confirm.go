package cmd

import (
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "confirm <confirmation-id>",
		Short: "Approve a pending action requiring confirmation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.Confirm(args[0]); err != nil {
				return err
			}
			return printResult("Action confirmed\n", map[string]any{"ok": true, "id": args[0]})
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "deny <confirmation-id>",
		Short: "Reject a pending action requiring confirmation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.Deny(args[0]); err != nil {
				return err
			}
			return printResult("Action denied\n", map[string]any{"ok": true, "id": args[0]})
		},
	})
}
