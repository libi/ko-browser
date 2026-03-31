package cmd

import (
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	var landscape bool
	var printBG bool

	command := &cobra.Command{
		Use:   "pdf <path>",
		Short: "Generate a PDF of the current page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			client := session.NewClient(sessionOptions())
			err := client.PDF(path, session.PDFArgs{
				Landscape: landscape,
				PrintBG:   printBG,
			})
			if err != nil {
				return err
			}
			if !rootFlags.json {
				fmt.Printf("PDF saved to %s\n", path)
			}
			return printResult("", map[string]string{"path": path})
		},
	}

	command.Flags().BoolVarP(&landscape, "landscape", "l", false, "landscape orientation")
	command.Flags().BoolVar(&printBG, "print-background", false, "print background graphics")
	rootCmd.AddCommand(command)
}
