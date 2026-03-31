package cmd

import (
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "type <id> <text>",
		Short: "Type text into an element by display ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseDisplayID(args[0])
			if err != nil {
				return err
			}
			client := session.NewClient(sessionOptions())
			if err := client.Type(id, args[1]); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})
}
