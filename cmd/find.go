package cmd

import (
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	findCmd := &cobra.Command{
		Use:   "find",
		Short: "Find elements in the accessibility snapshot",
	}

	// Shared --exact flag for find subcommands
	var findExact bool

	// find role <role> [--name <name>] [--exact]
	var findRoleName string
	findRoleCmd := &cobra.Command{
		Use:   "role <role>",
		Short: "Find elements by role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			role := args[0]
			client := session.NewClient(sessionOptions())
			exact, _ := cmd.Flags().GetBool("exact")
			var result string
			var err error
			if exact {
				result, err = client.FindRoleExact(role, findRoleName, true)
			} else {
				result, err = client.FindRole(role, findRoleName)
			}
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findRoleCmd.Flags().StringVar(&findRoleName, "name", "", "filter by accessible name (substring match)")
	findRoleCmd.Flags().BoolVar(&findExact, "exact", false, "use exact matching instead of substring")
	findCmd.AddCommand(findRoleCmd)

	// find text <text> [--exact]
	findTextCmd := &cobra.Command{
		Use:   "text <text>",
		Short: "Find elements whose name contains the given text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			client := session.NewClient(sessionOptions())
			exact, _ := cmd.Flags().GetBool("exact")
			var result string
			var err error
			if exact {
				result, err = client.FindTextExact(text, true)
			} else {
				result, err = client.FindText(text)
			}
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findTextCmd.Flags().BoolVar(&findExact, "exact", false, "use exact matching instead of substring")
	findCmd.AddCommand(findTextCmd)

	// find label <label> [--exact]
	findLabelCmd := &cobra.Command{
		Use:   "label <label>",
		Short: "Find form elements by label text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]
			client := session.NewClient(sessionOptions())
			exact, _ := cmd.Flags().GetBool("exact")
			var result string
			var err error
			if exact {
				result, err = client.FindLabelExact(label, true)
			} else {
				result, err = client.FindLabel(label)
			}
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findLabelCmd.Flags().BoolVar(&findExact, "exact", false, "use exact matching instead of substring")
	findCmd.AddCommand(findLabelCmd)

	// find nth <n> <css-selector>
	findNthCmd := &cobra.Command{
		Use:   "nth <n> <css-selector>",
		Short: "Find the nth element matching a CSS selector",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var n int
			if _, err := fmt.Sscanf(args[0], "%d", &n); err != nil {
				return fmt.Errorf("invalid number: %s", args[0])
			}
			cssSelector := args[1]
			client := session.NewClient(sessionOptions())
			result, err := client.FindNth(cssSelector, n)
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findCmd.AddCommand(findNthCmd)

	// find first <css-selector>
	findFirstCmd := &cobra.Command{
		Use:   "first <css-selector>",
		Short: "Find the first element matching a CSS selector",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cssSelector := args[0]
			client := session.NewClient(sessionOptions())
			result, err := client.FindNth(cssSelector, 1)
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findCmd.AddCommand(findFirstCmd)

	// find last <css-selector>
	findLastCmd := &cobra.Command{
		Use:   "last <css-selector>",
		Short: "Find the last element matching a CSS selector",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cssSelector := args[0]
			client := session.NewClient(sessionOptions())
			result, err := client.FindLast(cssSelector)
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findCmd.AddCommand(findLastCmd)

	// find placeholder <text> [--exact]
	findPlaceholderCmd := &cobra.Command{
		Use:   "placeholder <text>",
		Short: "Find elements by placeholder attribute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			exact, _ := cmd.Flags().GetBool("exact")
			result, err := client.FindPlaceholder(args[0], exact)
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findPlaceholderCmd.Flags().BoolVar(&findExact, "exact", false, "use exact matching instead of substring")
	findCmd.AddCommand(findPlaceholderCmd)

	// find alt <text> [--exact]
	findAltCmd := &cobra.Command{
		Use:   "alt <text>",
		Short: "Find image elements by alt text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			exact, _ := cmd.Flags().GetBool("exact")
			result, err := client.FindAlt(args[0], exact)
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findAltCmd.Flags().BoolVar(&findExact, "exact", false, "use exact matching instead of substring")
	findCmd.AddCommand(findAltCmd)

	// find title <text> [--exact]
	findTitleCmd := &cobra.Command{
		Use:   "title <text>",
		Short: "Find elements by title attribute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			exact, _ := cmd.Flags().GetBool("exact")
			result, err := client.FindTitle(args[0], exact)
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findTitleCmd.Flags().BoolVar(&findExact, "exact", false, "use exact matching instead of substring")
	findCmd.AddCommand(findTitleCmd)

	// find testid <testid>
	findTestIDCmd := &cobra.Command{
		Use:   "testid <testid>",
		Short: "Find elements by data-testid attribute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			result, err := client.FindTestID(args[0])
			if err != nil {
				return err
			}
			return printResult(result, map[string]any{"result": result})
		},
	}
	findCmd.AddCommand(findTestIDCmd)

	rootCmd.AddCommand(findCmd)
}
