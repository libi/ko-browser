package cmd

import (
	"time"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:     "stop",
		Aliases: []string{"close"},
		Short:   "Stop the browser session",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.Close(); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "restart",
		Short: "Restart the browser session (close then re-launch)",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := sessionOptions()
			client := session.NewClient(opts)
			// Best-effort close; ignore errors if no session is running
			_ = client.Close()
			// Brief pause to let the daemon fully exit
			time.Sleep(200 * time.Millisecond)
			// Re-create client to start a fresh daemon
			client = session.NewClient(opts)
			if err := client.Open("about:blank"); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true, "restarted": true})
		},
	})
}
