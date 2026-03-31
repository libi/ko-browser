package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	// ---- console ----
	consoleCmd := &cobra.Command{
		Use:   "console",
		Short: "Console message management",
	}

	// console start
	consoleCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start collecting console messages",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.ConsoleStart(); err != nil {
				return err
			}
			return printResult("Console listening started\n", map[string]any{"ok": true})
		},
	})

	// console messages [--level log|warning|error|info|debug]
	msgCmd := &cobra.Command{
		Use:   "messages",
		Short: "Get collected console messages",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			level, _ := cmd.Flags().GetString("level")
			client := session.NewClient(sessionOptions())
			msgs, err := client.ConsoleMessages(level)
			if err != nil {
				return err
			}

			if rootFlags.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(msgs)
			}

			if len(msgs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No console messages")
				return nil
			}
			for _, m := range msgs {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", m.Level, m.Text)
			}
			return nil
		},
	}
	msgCmd.Flags().String("level", "", "Filter by level (log, warning, error, info, debug)")
	consoleCmd.AddCommand(msgCmd)

	// console clear
	consoleCmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear collected console messages",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.ConsoleClear(); err != nil {
				return err
			}
			return printResult("Console messages cleared\n", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(consoleCmd)

	// ---- errors ----
	errorsCmd := &cobra.Command{
		Use:   "errors",
		Short: "Get page JavaScript errors",
	}

	// errors list
	errorsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List collected page errors",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			errs, err := client.PageErrors()
			if err != nil {
				return err
			}

			if rootFlags.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(errs)
			}

			if len(errs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No page errors")
				return nil
			}
			for _, e := range errs {
				if e.URL != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "%s (at %s:%d:%d)\n", e.Message, e.URL, e.Line, e.Column)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", e.Message)
				}
			}
			return nil
		},
	})

	// errors clear
	errorsCmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear collected page errors",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.PageErrorsClear(); err != nil {
				return err
			}
			return printResult("Page errors cleared\n", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(errorsCmd)

	// ---- highlight ----
	rootCmd.AddCommand(makeIDOnlyCommand("highlight <id>", "Highlight an element with a red border", func(client *session.Client, id int) error {
		return client.Highlight(id)
	}))

	// ---- inspect (devtools) ----
	rootCmd.AddCommand(&cobra.Command{
		Use:   "inspect",
		Short: "Open DevTools (non-headless mode only)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.OpenDevTools(); err != nil {
				return err
			}
			return printResult("DevTools opened\n", map[string]any{"ok": true})
		},
	})

	// ---- clipboard ----
	clipboardCmd := &cobra.Command{
		Use:   "clipboard",
		Short: "Clipboard read/write operations",
	}

	// clipboard read
	clipboardCmd.AddCommand(&cobra.Command{
		Use:   "read",
		Short: "Read text from clipboard",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			text, err := client.ClipboardRead()
			if err != nil {
				return err
			}
			return printResult(text+"\n", map[string]any{
				"ok":   true,
				"text": text,
			})
		},
	})

	// clipboard write <text>
	clipboardCmd.AddCommand(&cobra.Command{
		Use:   "write <text>",
		Short: "Write text to clipboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.ClipboardWrite(args[0]); err != nil {
				return err
			}
			return printResult("Text written to clipboard\n", map[string]any{
				"ok": true,
			})
		},
	})

	// clipboard copy
	clipboardCmd.AddCommand(&cobra.Command{
		Use:   "copy",
		Short: "Copy the current selection to clipboard",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.ClipboardCopy(); err != nil {
				return err
			}
			return printResult("Selection copied to clipboard\n", map[string]any{"ok": true})
		},
	})

	// clipboard paste
	clipboardCmd.AddCommand(&cobra.Command{
		Use:   "paste",
		Short: "Paste clipboard content into the focused element",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.ClipboardPaste(); err != nil {
				return err
			}
			return printResult("Clipboard content pasted\n", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(clipboardCmd)
}
