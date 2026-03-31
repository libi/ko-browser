package main

import (
	"os"

	"github.com/libi/ko-browser/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
