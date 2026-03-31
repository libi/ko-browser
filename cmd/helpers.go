package cmd

import (
	"fmt"

	"github.com/libi/ko-browser/selector"
)

func parseDisplayID(arg string) (int, error) {
	sel, err := selector.Parse(arg)
	if err != nil {
		return 0, err
	}
	if !sel.IsDisplayID() {
		return 0, fmt.Errorf("only display ID selectors are supported for this command")
	}
	return sel.DisplayID, nil
}
