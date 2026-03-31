package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Network interception and monitoring",
	}

	// network route <pattern> --action block|continue
	routeCmd := &cobra.Command{
		Use:   "route <pattern>",
		Short: "Intercept requests matching a URL pattern",
		Long: `Register a URL pattern to intercept network requests.
Pattern supports simple glob matching:
  *         - matches any characters
  *.js      - matches URLs ending in .js
  *api*     - matches URLs containing "api"
  https://** - matches URLs starting with https://`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := args[0]
			action, _ := cmd.Flags().GetString("action")
			client := session.NewClient(sessionOptions())
			if err := client.NetworkRoute(pattern, action); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Route registered: %s → %s\n", pattern, action), map[string]any{
				"ok":      true,
				"pattern": pattern,
				"action":  action,
			})
		},
	}
	routeCmd.Flags().String("action", "block", "Action for matched requests: block or continue")
	networkCmd.AddCommand(routeCmd)

	// network unroute <pattern>
	networkCmd.AddCommand(&cobra.Command{
		Use:   "unroute <pattern>",
		Short: "Remove a previously registered route pattern",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := args[0]
			client := session.NewClient(sessionOptions())
			if err := client.NetworkUnroute(pattern); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Route removed: %s\n", pattern), map[string]any{
				"ok":      true,
				"pattern": pattern,
			})
		},
	})

	// network requests
	networkCmd.AddCommand(&cobra.Command{
		Use:   "requests",
		Short: "List recorded network requests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			reqs, err := client.NetworkRequests()
			if err != nil {
				return err
			}

			if rootFlags.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(reqs)
			}

			if len(reqs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No requests recorded")
				return nil
			}
			for i, r := range reqs {
				method := r.Method
				if method == "" {
					method = "GET"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%d: %s %s → %d %s\n", i+1, method, r.URL, r.Status, r.StatusText)
			}
			return nil
		},
	})

	// network start-logging
	networkCmd.AddCommand(&cobra.Command{
		Use:   "start-logging",
		Short: "Start recording network requests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.NetworkStartLogging(); err != nil {
				return err
			}
			return printResult("Network logging started\n", map[string]any{"ok": true})
		},
	})

	// network clear
	networkCmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear recorded network requests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.NetworkClearRequests(); err != nil {
				return err
			}
			return printResult("Network requests cleared\n", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(networkCmd)
}
