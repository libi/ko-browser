package cmd

import (
	"strconv"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		makeIDOnlyCommand("hover <id>", "Hover over an element by display ID", func(client *session.Client, id int) error {
			return client.Hover(id)
		}),
		makeIDOnlyCommand("focus <id>", "Focus an element by display ID", func(client *session.Client, id int) error {
			return client.Focus(id)
		}),
		makeIDOnlyCommand("check <id>", "Check a checkbox or radio by display ID", func(client *session.Client, id int) error {
			return client.Check(id)
		}),
		makeIDOnlyCommand("uncheck <id>", "Uncheck a checkbox by display ID", func(client *session.Client, id int) error {
			return client.Uncheck(id)
		}),
		makeIDOnlyCommand("scrollintoview <id>", "Scroll an element into view by display ID", func(client *session.Client, id int) error {
			return client.ScrollIntoView(id)
		}),
		makeIDOnlyCommand("dblclick <id>", "Double click an element by display ID", func(client *session.Client, id int) error {
			return client.DblClick(id)
		}),
	)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "select <id> <value> [value...]",
		Short: "Select one or more values in a select element",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseDisplayID(args[0])
			if err != nil {
				return err
			}
			client := session.NewClient(sessionOptions())
			if err := client.Select(id, args[1:]...); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "scroll <direction> <pixels>",
		Short: "Scroll the page in a direction by a pixel amount",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			client := session.NewClient(sessionOptions())
			if err := client.Scroll(args[0], amount); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})
}

func makeIDOnlyCommand(use string, short string, run func(client *session.Client, id int) error) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseDisplayID(args[0])
			if err != nil {
				return err
			}
			client := session.NewClient(sessionOptions())
			if err := run(client, id); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	}
}
