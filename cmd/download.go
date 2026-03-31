package cmd

import (
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "download <id> <saveDir>",
		Short: "Click a download link/button and save the file",
		Long:  `Triggers a download by clicking the element and saves the downloaded file to the specified directory.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseDisplayID(args[0])
			if err != nil {
				return err
			}
			saveDir := args[1]
			client := session.NewClient(sessionOptions())
			path, err := client.Download(id, saveDir)
			if err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Downloaded: %s", path), map[string]any{
				"ok":   true,
				"path": path,
			})
		},
	})
}
