package cmd

import (
	"fmt"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	var fullPage bool
	var quality int
	var elementID int
	var annotate bool

	command := &cobra.Command{
		Use:   "screenshot <path>",
		Short: "Capture a screenshot of the current page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			client := session.NewClient(sessionOptions())
			ssArgs := session.ScreenshotArgs{
				FullPage:  fullPage,
				Quality:   quality,
				ElementID: elementID,
			}
			var err error
			if annotate {
				err = client.ScreenshotAnnotated(path, ssArgs)
			} else {
				err = client.Screenshot(path, ssArgs)
			}
			if err != nil {
				return err
			}
			if !rootFlags.json {
				fmt.Printf("Screenshot saved to %s\n", path)
			}
			return printResult("", map[string]string{"path": path})
		},
	}

	command.Flags().BoolVarP(&fullPage, "full-page", "f", false, "capture the entire scrollable page")
	command.Flags().IntVarP(&quality, "quality", "q", 0, "JPEG quality (1-100), 0 means PNG")
	command.Flags().IntVarP(&elementID, "element", "e", 0, "capture only this element (display ID)")
	command.Flags().BoolVarP(&annotate, "annotate", "a", false, "overlay element IDs on the screenshot")
	rootCmd.AddCommand(command)
}
