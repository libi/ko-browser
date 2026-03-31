package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/libi/ko-browser/browser"
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	// ---- cookies ----
	cookiesCmd := &cobra.Command{
		Use:   "cookies",
		Short: "Manage browser cookies",
	}

	// cookies get
	cookiesCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get all cookies for the current page",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			cookies, err := client.CookiesGet()
			if err != nil {
				return err
			}

			if rootFlags.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(cookies)
			}

			if len(cookies) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No cookies")
				return nil
			}
			for _, c := range cookies {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s (domain=%s, path=%s)\n", c.Name, c.Value, c.Domain, c.Path)
			}
			return nil
		},
	})

	// cookies set <name> <value> [--domain d] [--path p]
	setCmd := &cobra.Command{
		Use:   "set <name> <value>",
		Short: "Set a cookie",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain, _ := cmd.Flags().GetString("domain")
			path, _ := cmd.Flags().GetString("path")
			httpOnly, _ := cmd.Flags().GetBool("http-only")
			secure, _ := cmd.Flags().GetBool("secure")

			client := session.NewClient(sessionOptions())
			if err := client.CookieSet(browser.CookieInfo{
				Name:     args[0],
				Value:    args[1],
				Domain:   domain,
				Path:     path,
				HTTPOnly: httpOnly,
				Secure:   secure,
			}); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Cookie set: %s=%s\n", args[0], args[1]), map[string]any{
				"ok":   true,
				"name": args[0],
			})
		},
	}
	setCmd.Flags().String("domain", "", "Cookie domain (default: current page domain)")
	setCmd.Flags().String("path", "/", "Cookie path")
	setCmd.Flags().Bool("http-only", false, "HTTP-only cookie")
	setCmd.Flags().Bool("secure", false, "Secure cookie")
	cookiesCmd.AddCommand(setCmd)

	// cookies delete <name>
	cookiesCmd.AddCommand(&cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a cookie by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.CookieDelete(args[0]); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Cookie deleted: %s\n", args[0]), map[string]any{
				"ok":   true,
				"name": args[0],
			})
		},
	})

	// cookies clear
	cookiesCmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear all cookies",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.CookiesClear(); err != nil {
				return err
			}
			return printResult("All cookies cleared\n", map[string]any{"ok": true})
		},
	})

	rootCmd.AddCommand(cookiesCmd)

	// ---- storage ----
	storageCmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage localStorage and sessionStorage",
	}

	// storage get <key> [--type local|session]
	storageGetCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a value from storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			storageType, _ := cmd.Flags().GetString("type")
			client := session.NewClient(sessionOptions())
			value, err := client.StorageGet(storageType, args[0])
			if err != nil {
				return err
			}
			return printResult(value+"\n", map[string]any{
				"ok":    true,
				"key":   args[0],
				"value": value,
			})
		},
	}
	storageGetCmd.Flags().String("type", "local", "Storage type: local or session")
	storageCmd.AddCommand(storageGetCmd)

	// storage set <key> <value> [--type local|session]
	storageSetCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a value in storage",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			storageType, _ := cmd.Flags().GetString("type")
			client := session.NewClient(sessionOptions())
			if err := client.StorageSet(storageType, args[0], args[1]); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Storage set: %s=%s\n", args[0], args[1]), map[string]any{
				"ok":  true,
				"key": args[0],
			})
		},
	}
	storageSetCmd.Flags().String("type", "local", "Storage type: local or session")
	storageCmd.AddCommand(storageSetCmd)

	// storage delete <key> [--type local|session]
	storageDeleteCmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a key from storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			storageType, _ := cmd.Flags().GetString("type")
			client := session.NewClient(sessionOptions())
			if err := client.StorageDelete(storageType, args[0]); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Storage key deleted: %s\n", args[0]), map[string]any{
				"ok":  true,
				"key": args[0],
			})
		},
	}
	storageDeleteCmd.Flags().String("type", "local", "Storage type: local or session")
	storageCmd.AddCommand(storageDeleteCmd)

	// storage clear [--type local|session]
	storageClearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all items in storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			storageType, _ := cmd.Flags().GetString("type")
			client := session.NewClient(sessionOptions())
			if err := client.StorageClear(storageType); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("%sStorage cleared\n", storageType), map[string]any{"ok": true})
		},
	}
	storageClearCmd.Flags().String("type", "local", "Storage type: local or session")
	storageCmd.AddCommand(storageClearCmd)

	// storage list [--type local|session]
	storageListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all items in storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			storageType, _ := cmd.Flags().GetString("type")
			client := session.NewClient(sessionOptions())
			items, err := client.StorageGetAll(storageType)
			if err != nil {
				return err
			}

			if rootFlags.json {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(items)
			}

			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No items")
				return nil
			}
			for k, v := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", k, v)
			}
			return nil
		},
	}
	storageListCmd.Flags().String("type", "local", "Storage type: local or session")
	storageCmd.AddCommand(storageListCmd)

	rootCmd.AddCommand(storageCmd)
}
