package cmd

import (
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

var isCmd = &cobra.Command{
	Use:   "is",
	Short: "Query element state (visible, enabled, checked)",
}

func init() {
	rootCmd.AddCommand(isCmd)

	isCmd.AddCommand(
		&cobra.Command{
			Use:   "visible <id>",
			Short: "Check if an element is visible",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				visible, err := client.IsVisible(id)
				if err != nil {
					return err
				}
				return printResult(fmt.Sprintf("%v\n", visible), map[string]any{"visible": visible})
			},
		},
		&cobra.Command{
			Use:   "enabled <id>",
			Short: "Check if an element is enabled",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				enabled, err := client.IsEnabled(id)
				if err != nil {
					return err
				}
				return printResult(fmt.Sprintf("%v\n", enabled), map[string]any{"enabled": enabled})
			},
		},
		&cobra.Command{
			Use:   "checked <id>",
			Short: "Check if a checkbox or radio is checked",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				checked, err := client.IsChecked(id)
				if err != nil {
					return err
				}
				return printResult(fmt.Sprintf("%v\n", checked), map[string]any{"checked": checked})
			},
		},
	)
}
