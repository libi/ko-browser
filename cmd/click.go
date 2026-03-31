package cmd

import (
	"strconv"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	var newTab bool

	clickCmd := &cobra.Command{
		Use:   "click <id>",
		Short: "Click an element by display ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			client := session.NewClient(sessionOptions())
			if newTab {
				if err := client.ClickNewTab(id); err != nil {
					return err
				}
			} else {
				if err := client.Click(id); err != nil {
					return err
				}
			}
			return printResult("", map[string]any{"ok": true})
		},
	}
	clickCmd.Flags().BoolVar(&newTab, "new-tab", false, "Open the link in a new tab")
	rootCmd.AddCommand(clickCmd)
}
