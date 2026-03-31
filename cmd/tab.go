package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	tabCmd := &cobra.Command{
		Use:   "tab",
		Short: "Manage browser tabs",
	}

	// tab list
	tabCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all open tabs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			tabs, err := client.TabList()
			if err != nil {
				return err
			}

			if rootFlags.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(tabs)
			}

			for _, t := range tabs {
				marker := " "
				if t.Active {
					marker = "*"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %d: %s (%s)\n", marker, t.Index, t.Title, t.URL)
			}
			return nil
		},
	})

	// tab new [url]
	tabCmd.AddCommand(&cobra.Command{
		Use:   "new [url]",
		Short: "Open a new tab",
		Long:  "Opens a new tab. If a URL is provided, navigates to it; otherwise opens about:blank.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := ""
			if len(args) > 0 {
				url = args[0]
			}
			client := session.NewClient(sessionOptions())
			if err := client.TabNew(url); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})

	// tab close [index]
	tabCmd.AddCommand(&cobra.Command{
		Use:   "close [index]",
		Short: "Close a tab by index (default: current tab)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			index := -1
			if len(args) > 0 {
				var err error
				index, err = strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid tab index: %s", args[0])
				}
			}
			client := session.NewClient(sessionOptions())
			if err := client.TabClose(index); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})

	// tab switch <index>
	tabCmd.AddCommand(&cobra.Command{
		Use:   "switch <index>",
		Short: "Switch to a tab by index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			index, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid tab index: %s", args[0])
			}
			client := session.NewClient(sessionOptions())
			if err := client.TabSwitch(index); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(tabCmd)
}
