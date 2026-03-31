package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get information from the page or elements",
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.AddCommand(
		&cobra.Command{
			Use:   "title",
			Short: "Get the page title",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				title, err := client.GetTitle()
				if err != nil {
					return err
				}
				return printResult(title+"\n", map[string]any{"title": title})
			},
		},
		&cobra.Command{
			Use:   "url",
			Short: "Get the current page URL",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				url, err := client.GetURL()
				if err != nil {
					return err
				}
				return printResult(url+"\n", map[string]any{"url": url})
			},
		},
		&cobra.Command{
			Use:   "text <id>",
			Short: "Get the inner text of an element",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				text, err := client.GetText(id)
				if err != nil {
					return err
				}
				return printResult(text+"\n", map[string]any{"text": text})
			},
		},
		&cobra.Command{
			Use:   "html <id>",
			Short: "Get the inner HTML of an element",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				html, err := client.GetHTML(id)
				if err != nil {
					return err
				}
				return printResult(html+"\n", map[string]any{"html": html})
			},
		},
		&cobra.Command{
			Use:   "value <id>",
			Short: "Get the value of a form element",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				value, err := client.GetValue(id)
				if err != nil {
					return err
				}
				return printResult(value+"\n", map[string]any{"value": value})
			},
		},
		&cobra.Command{
			Use:   "attr <id> <name>",
			Short: "Get an attribute value of an element",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				value, err := client.GetAttr(id, args[1])
				if err != nil {
					return err
				}
				return printResult(value+"\n", map[string]any{"attr": args[1], "value": value})
			},
		},
		&cobra.Command{
			Use:   "count <css-selector>",
			Short: "Count elements matching a CSS selector",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client := session.NewClient(sessionOptions())
				count, err := client.GetCount(args[0])
				if err != nil {
					return err
				}
				return printResult(fmt.Sprintf("%d\n", count), map[string]any{"count": count})
			},
		},
		&cobra.Command{
			Use:   "box <id>",
			Short: "Get the bounding box of an element",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				box, err := client.GetBox(id)
				if err != nil {
					return err
				}
				text := fmt.Sprintf("x=%.0f y=%.0f width=%.0f height=%.0f\n",
					box.X, box.Y, box.Width, box.Height)
				return printResult(text, box)
			},
		},
		&cobra.Command{
			Use:   "styles <id>",
			Short: "Get computed styles of an element",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id, err := parseDisplayID(args[0])
				if err != nil {
					return err
				}
				client := session.NewClient(sessionOptions())
				styles, err := client.GetStyles(id)
				if err != nil {
					return err
				}
				if rootFlags.json {
					// pretty-print the JSON
					var parsed map[string]any
					if json.Unmarshal([]byte(styles), &parsed) == nil {
						return printResult("", parsed)
					}
				}
				return printResult(styles+"\n", map[string]any{"styles": styles})
			},
		},
	)

	// get cdp-url
	getCmd.AddCommand(&cobra.Command{
		Use:   "cdp-url",
		Short: "Get the Chrome DevTools Protocol WebSocket URL",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			url, err := client.GetCDPURL()
			if err != nil {
				return err
			}
			return printResult(url+"\n", map[string]any{"cdpUrl": url})
		},
	})
}
