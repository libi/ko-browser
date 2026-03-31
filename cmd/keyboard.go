package cmd

import (
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	keyboardCmd := &cobra.Command{
		Use:   "keyboard",
		Short: "Keyboard actions on the active element",
	}

	keyboardCmd.AddCommand(&cobra.Command{
		Use:   "type <text>",
		Short: "Insert text into the active element",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.KeyboardType(args[0]); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})

	keyboardCmd.AddCommand(&cobra.Command{
		Use:   "inserttext <text>",
		Short: "Insert text using Input.insertText (no key events)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.KeyboardInsertText(args[0]); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(keyboardCmd)
}
