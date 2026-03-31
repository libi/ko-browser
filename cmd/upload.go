package cmd

import (
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "upload <id> <file> [file...]",
		Short: "Upload one or more files to a file input element",
		Long:  `Set files on an <input type="file"> element identified by display ID.`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseDisplayID(args[0])
			if err != nil {
				return err
			}
			files := args[1:]
			client := session.NewClient(sessionOptions())
			if err := client.Upload(id, files...); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Uploaded %d file(s)", len(files)), map[string]any{
				"ok":    true,
				"files": files,
			})
		},
	})
}
