package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage saved authentication profiles",
	}

	// auth save <name>
	saveCmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Save an authentication profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url, _ := cmd.Flags().GetString("url")
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")
			passwordStdin, _ := cmd.Flags().GetBool("password-stdin")
			usernameSelector, _ := cmd.Flags().GetString("username-selector")
			passwordSelector, _ := cmd.Flags().GetString("password-selector")
			submitSelector, _ := cmd.Flags().GetString("submit-selector")

			if url == "" {
				return fmt.Errorf("--url is required")
			}
			if username == "" {
				return fmt.Errorf("--username is required")
			}
			if password == "" && !passwordStdin {
				return fmt.Errorf("--password or --password-stdin is required")
			}
			if passwordStdin {
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					password = strings.TrimSpace(scanner.Text())
				}
				if password == "" {
					return fmt.Errorf("no password provided on stdin")
				}
			}

			vault, err := session.NewAuthVault()
			if err != nil {
				return err
			}
			if err := vault.Save(session.AuthProfile{
				Name:             name,
				URL:              url,
				Username:         username,
				Password:         password,
				UsernameSelector: usernameSelector,
				PasswordSelector: passwordSelector,
				SubmitSelector:   submitSelector,
			}); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Auth profile %q saved\n", name), map[string]any{
				"ok":   true,
				"name": name,
			})
		},
	}
	saveCmd.Flags().String("url", "", "Login page URL (required)")
	saveCmd.Flags().String("username", "", "Username (required)")
	saveCmd.Flags().String("password", "", "Password")
	saveCmd.Flags().Bool("password-stdin", false, "Read password from stdin")
	saveCmd.Flags().String("username-selector", "", "Custom CSS selector for username field")
	saveCmd.Flags().String("password-selector", "", "Custom CSS selector for password field")
	saveCmd.Flags().String("submit-selector", "", "Custom CSS selector for submit button")
	authCmd.AddCommand(saveCmd)

	// auth login <name>
	authCmd.AddCommand(&cobra.Command{
		Use:   "login <name>",
		Short: "Login using a saved auth profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			vault, err := session.NewAuthVault()
			if err != nil {
				return err
			}
			profile, ok := vault.Get(name)
			if !ok {
				return fmt.Errorf("auth profile %q not found", name)
			}

			client := session.NewClient(sessionOptions())

			// Navigate to login page
			if err := client.Open(profile.URL); err != nil {
				return fmt.Errorf("navigate to %s: %w", profile.URL, err)
			}

			// Wait for page to load
			_ = client.WaitText("")
			time.Sleep(500 * time.Millisecond)

			// Fill username
			if profile.UsernameSelector != "" {
				if _, err := client.Eval(fmt.Sprintf(
					`document.querySelector(%q).value = %q; document.querySelector(%q).dispatchEvent(new Event('input', {bubbles:true}))`,
					profile.UsernameSelector, profile.Username, profile.UsernameSelector,
				)); err != nil {
					return fmt.Errorf("fill username: %w", err)
				}
			} else {
				// Try common selectors
				if _, err := client.Eval(fmt.Sprintf(
					`(function(){var e=document.querySelector('input[type="email"],input[type="text"],input[name="username"],input[name="login"],input[name="email"],input[id="username"],input[id="email"],input[id="login"]');if(e){e.value=%q;e.dispatchEvent(new Event('input',{bubbles:true}));return 'ok'}return 'not found'})()`,
					profile.Username,
				)); err != nil {
					return fmt.Errorf("fill username: %w", err)
				}
			}

			// Fill password
			if profile.PasswordSelector != "" {
				if _, err := client.Eval(fmt.Sprintf(
					`document.querySelector(%q).value = %q; document.querySelector(%q).dispatchEvent(new Event('input', {bubbles:true}))`,
					profile.PasswordSelector, profile.Password, profile.PasswordSelector,
				)); err != nil {
					return fmt.Errorf("fill password: %w", err)
				}
			} else {
				if _, err := client.Eval(fmt.Sprintf(
					`(function(){var e=document.querySelector('input[type="password"]');if(e){e.value=%q;e.dispatchEvent(new Event('input',{bubbles:true}));return 'ok'}return 'not found'})()`,
					profile.Password,
				)); err != nil {
					return fmt.Errorf("fill password: %w", err)
				}
			}

			// Click submit
			if profile.SubmitSelector != "" {
				if _, err := client.Eval(fmt.Sprintf(
					`document.querySelector(%q).click()`, profile.SubmitSelector,
				)); err != nil {
					return fmt.Errorf("click submit: %w", err)
				}
			} else {
				if _, err := client.Eval(
					`(function(){var e=document.querySelector('button[type="submit"],input[type="submit"]');if(e){e.click();return 'ok'}return 'not found'})()`,
				); err != nil {
					return fmt.Errorf("click submit: %w", err)
				}
			}

			return printResult(fmt.Sprintf("Logged in with profile %q\n", name), map[string]any{
				"ok":   true,
				"name": name,
				"url":  profile.URL,
			})
		},
	})

	// auth list
	authCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List saved auth profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			vault, err := session.NewAuthVault()
			if err != nil {
				return err
			}
			profiles := vault.List()

			if rootFlags.json {
				// Redact passwords
				type safeProfile struct {
					Name string `json:"name"`
					URL  string `json:"url"`
					User string `json:"username"`
				}
				var safe []safeProfile
				for _, p := range profiles {
					safe = append(safe, safeProfile{Name: p.Name, URL: p.URL, User: p.Username})
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(safe)
			}

			if len(profiles) == 0 {
				fmt.Println("No saved auth profiles")
				return nil
			}
			for _, p := range profiles {
				fmt.Printf("  %s  %s  (%s)\n", p.Name, p.URL, p.Username)
			}
			return nil
		},
	})

	// auth show <name>
	authCmd.AddCommand(&cobra.Command{
		Use:   "show <name>",
		Short: "Show auth profile details (password redacted)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vault, err := session.NewAuthVault()
			if err != nil {
				return err
			}
			profile, ok := vault.Get(args[0])
			if !ok {
				return fmt.Errorf("auth profile %q not found", args[0])
			}

			if rootFlags.json {
				safe := map[string]any{
					"name":     profile.Name,
					"url":      profile.URL,
					"username": profile.Username,
					"password": "********",
				}
				if profile.UsernameSelector != "" {
					safe["usernameSelector"] = profile.UsernameSelector
				}
				if profile.PasswordSelector != "" {
					safe["passwordSelector"] = profile.PasswordSelector
				}
				if profile.SubmitSelector != "" {
					safe["submitSelector"] = profile.SubmitSelector
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(safe)
			}

			fmt.Printf("Name:     %s\n", profile.Name)
			fmt.Printf("URL:      %s\n", profile.URL)
			fmt.Printf("Username: %s\n", profile.Username)
			fmt.Printf("Password: ********\n")
			if profile.UsernameSelector != "" {
				fmt.Printf("Username selector: %s\n", profile.UsernameSelector)
			}
			if profile.PasswordSelector != "" {
				fmt.Printf("Password selector: %s\n", profile.PasswordSelector)
			}
			if profile.SubmitSelector != "" {
				fmt.Printf("Submit selector:   %s\n", profile.SubmitSelector)
			}
			return nil
		},
	})

	// auth delete <name>
	authCmd.AddCommand(&cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a saved auth profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vault, err := session.NewAuthVault()
			if err != nil {
				return err
			}
			if err := vault.Delete(args[0]); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Auth profile %q deleted\n", args[0]), map[string]any{
				"ok":   true,
				"name": args[0],
			})
		},
	})

	rootCmd.AddCommand(authCmd)
}
