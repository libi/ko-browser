package cmd

import (
	"github.com/libi/ko-browser/browser"
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	var enableOCR bool
	var ocrLangs []string
	var ocrDebugDir string
	var interactiveOnly bool
	var compact bool
	var maxDepth int
	var cursor bool
	var selector string

	command := &cobra.Command{
		Use:   "snapshot",
		Short: "Capture the current accessibility snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			resp, err := client.Snapshot(browser.SnapshotOptions{
				EnableOCR:       enableOCR,
				OCRLanguages:    ocrLangs,
				OCRDebugDir:     ocrDebugDir,
				InteractiveOnly: interactiveOnly,
				Compact:         compact,
				MaxDepth:        maxDepth,
				Cursor:          cursor,
				Selector:        selector,
			})
			if err != nil {
				return err
			}
			return printResult(resp.Text, resp)
		},
	}

	command.Flags().BoolVar(&enableOCR, "ocr", false, "enable OCR for nameless interactive images")
	command.Flags().StringSliceVar(&ocrLangs, "lang", []string{"eng"}, "OCR language(s)")
	command.Flags().StringVar(&ocrDebugDir, "ocr-debug", "", "directory for OCR debug images")
	command.Flags().BoolVarP(&interactiveOnly, "interactive", "i", false, "only show interactive elements")
	command.Flags().BoolVarP(&compact, "compact", "c", false, "compact mode: omit unnamed structural wrappers")
	command.Flags().IntVarP(&maxDepth, "depth", "d", 0, "maximum tree depth (0 = unlimited)")
	command.Flags().BoolVarP(&cursor, "cursor", "C", false, "annotate the focused element with [cursor]")
	command.Flags().StringVarP(&selector, "selector", "s", "", "CSS selector to scope the snapshot to a subtree")
	rootCmd.AddCommand(command)
}
