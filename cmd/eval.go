package cmd

import (
	"strings"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	command := &cobra.Command{
		Use:   "eval <expression>",
		Short: "Evaluate a JavaScript expression in the page context",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			expression := strings.Join(args, " ")
			client := session.NewClient(sessionOptions())
			result, err := client.Eval(expression)
			if err != nil {
				return err
			}
			return printResult(result+"\n", map[string]string{"result": result})
		},
	}

	rootCmd.AddCommand(command)
}
