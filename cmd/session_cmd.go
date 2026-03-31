package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Display or manage browser sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default behavior: show current session name
			name := rootFlags.session
			if rootFlags.json {
				return printResult("", map[string]any{
					"session": name,
				})
			}
			fmt.Printf("Current session: %s\n", name)
			return nil
		},
	}

	// session list
	sessionCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all active browser sessions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessions, err := session.ListSessions()
			if err != nil {
				return fmt.Errorf("list sessions: %w", err)
			}

			if rootFlags.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(sessions)
			}

			if len(sessions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No active sessions")
				return nil
			}

			for _, s := range sessions {
				marker := "  "
				if s.Name == rootFlags.session {
					marker = "* "
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s\n", marker, s.Name)
			}
			return nil
		},
	})

	rootCmd.AddCommand(sessionCmd)
}
