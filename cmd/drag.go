package cmd

import (
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "drag <srcID> <dstID>",
		Short: "Drag an element and drop it onto another element",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcID, err := parseDisplayID(args[0])
			if err != nil {
				return err
			}
			dstID, err := parseDisplayID(args[1])
			if err != nil {
				return err
			}
			client := session.NewClient(sessionOptions())
			if err := client.Drag(srcID, dstID); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})
}
