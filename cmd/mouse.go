package cmd

import (
	"strconv"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	mouseCmd := &cobra.Command{
		Use:   "mouse",
		Short: "Low-level mouse operations",
	}

	mouseCmd.AddCommand(&cobra.Command{
		Use:   "move <x> <y>",
		Short: "Move the mouse to the given page coordinates",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			x, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return err
			}
			y, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return err
			}
			client := session.NewClient(sessionOptions())
			if err := client.MouseMove(x, y); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	})

	downCmd := &cobra.Command{
		Use:   "down <x> <y>",
		Short: "Press a mouse button at the given coordinates",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			x, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return err
			}
			y, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return err
			}
			button, _ := cmd.Flags().GetString("button")
			client := session.NewClient(sessionOptions())
			if err := client.MouseDown(x, y, button); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	}
	downCmd.Flags().String("button", "left", "Mouse button: left, right, middle")

	upCmd := &cobra.Command{
		Use:   "up <x> <y>",
		Short: "Release a mouse button at the given coordinates",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			x, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return err
			}
			y, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return err
			}
			button, _ := cmd.Flags().GetString("button")
			client := session.NewClient(sessionOptions())
			if err := client.MouseUp(x, y, button); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	}
	upCmd.Flags().String("button", "left", "Mouse button: left, right, middle")

	wheelCmd := &cobra.Command{
		Use:   "wheel <x> <y> <deltaY>",
		Short: "Dispatch a mouse wheel event",
		Long:  "Dispatches a mouse wheel event at the given coordinates. Positive deltaY scrolls down.",
		Args:  cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			x, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return err
			}
			y, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return err
			}
			deltaY, err := strconv.ParseFloat(args[2], 64)
			if err != nil {
				return err
			}
			var deltaX float64
			if len(args) == 4 {
				deltaX, err = strconv.ParseFloat(args[3], 64)
				if err != nil {
					return err
				}
			}
			client := session.NewClient(sessionOptions())
			if err := client.MouseWheel(x, y, deltaX, deltaY); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	}

	clickCmd := &cobra.Command{
		Use:   "click <x> <y>",
		Short: "Perform a mouse click at the given coordinates",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			x, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return err
			}
			y, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return err
			}
			button, _ := cmd.Flags().GetString("button")
			client := session.NewClient(sessionOptions())
			if err := client.MouseClick(x, y, button); err != nil {
				return err
			}
			return printResult("", map[string]any{"ok": true})
		},
	}
	clickCmd.Flags().String("button", "left", "Mouse button: left, right, middle")

	mouseCmd.AddCommand(downCmd, upCmd, wheelCmd, clickCmd)
	rootCmd.AddCommand(mouseCmd)
}
