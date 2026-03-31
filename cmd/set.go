package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/libi/ko-browser/browser"
	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Configure browser settings (viewport, device, geo, offline, headers, media)",
	}

	// set viewport <width> <height> [scale]
	setCmd.AddCommand(&cobra.Command{
		Use:   "viewport <width> <height> [scale]",
		Short: "Set browser viewport size (optional scale factor)",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			width, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid width: %w", err)
			}
			height, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid height: %w", err)
			}
			client := session.NewClient(sessionOptions())
			if len(args) == 3 {
				scale, err := strconv.ParseFloat(args[2], 64)
				if err != nil {
					return fmt.Errorf("invalid scale: %w", err)
				}
				if err := client.SetViewportWithScale(width, height, scale); err != nil {
					return err
				}
				return printResult(fmt.Sprintf("Viewport set to %dx%d (scale %.1f)\n", width, height, scale), map[string]any{
					"ok":     true,
					"width":  width,
					"height": height,
					"scale":  scale,
				})
			}
			if err := client.SetViewport(width, height); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Viewport set to %dx%d\n", width, height), map[string]any{
				"ok":     true,
				"width":  width,
				"height": height,
			})
		},
	})

	// set device <name>
	setCmd.AddCommand(&cobra.Command{
		Use:   "device <name>",
		Short: "Emulate a device (e.g., \"iPhone 12\", \"Pixel 5\")",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.Join(args, " ")
			client := session.NewClient(sessionOptions())
			if err := client.SetDevice(name); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Device set to %s\n", name), map[string]any{
				"ok":     true,
				"device": name,
			})
		},
	})

	// set geo <lat> <lon>
	setCmd.AddCommand(&cobra.Command{
		Use:   "geo <latitude> <longitude>",
		Short: "Override geolocation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			lat, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("invalid latitude: %w", err)
			}
			lon, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return fmt.Errorf("invalid longitude: %w", err)
			}
			client := session.NewClient(sessionOptions())
			if err := client.SetGeo(lat, lon); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Geolocation set to %f, %f\n", lat, lon), map[string]any{
				"ok":  true,
				"lat": lat,
				"lon": lon,
			})
		},
	})

	// set offline <true|false>
	setCmd.AddCommand(&cobra.Command{
		Use:   "offline <true|false>",
		Short: "Enable or disable offline mode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			offline, err := strconv.ParseBool(args[0])
			if err != nil {
				return fmt.Errorf("invalid boolean: %w", err)
			}
			client := session.NewClient(sessionOptions())
			if err := client.SetOffline(offline); err != nil {
				return err
			}
			status := "offline"
			if !offline {
				status = "online"
			}
			return printResult(fmt.Sprintf("Network mode: %s\n", status), map[string]any{
				"ok":      true,
				"offline": offline,
			})
		},
	})

	// set headers <key=value> [key=value...]
	setCmd.AddCommand(&cobra.Command{
		Use:   "headers <key=value>...",
		Short: "Set extra HTTP headers for all requests",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			headers := make(map[string]string)
			for _, arg := range args {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid header format %q, expected key=value", arg)
				}
				headers[parts[0]] = parts[1]
			}
			client := session.NewClient(sessionOptions())
			if err := client.SetHeaders(headers); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Set %d header(s)\n", len(headers)), map[string]any{
				"ok":      true,
				"headers": headers,
			})
		},
	})

	// set credentials <user> <pass>
	setCmd.AddCommand(&cobra.Command{
		Use:   "credentials <username> <password>",
		Short: "Set HTTP Basic Auth credentials",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := session.NewClient(sessionOptions())
			if err := client.SetCredentials(args[0], args[1]); err != nil {
				return err
			}
			return printResult("Credentials set\n", map[string]any{
				"ok":   true,
				"user": args[0],
			})
		},
	})

	// set media <feature=value> [feature=value...]
	setCmd.AddCommand(&cobra.Command{
		Use:   "media <feature=value>...",
		Short: "Set emulated CSS media features (e.g., prefers-color-scheme=dark)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var features []browser.MediaFeature
			for _, arg := range args {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid media feature format %q, expected name=value", arg)
				}
				features = append(features, browser.MediaFeature{Name: parts[0], Value: parts[1]})
			}
			client := session.NewClient(sessionOptions())
			if err := client.SetMedia(features); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Set %d media feature(s)\n", len(features)), map[string]any{
				"ok":       true,
				"features": features,
			})
		},
	})

	// set colorscheme <dark|light>
	setCmd.AddCommand(&cobra.Command{
		Use:   "colorscheme <dark|light>",
		Short: "Set emulated color scheme (shortcut for set media prefers-color-scheme=...)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scheme := args[0]
			client := session.NewClient(sessionOptions())
			if err := client.SetColorScheme(scheme); err != nil {
				return err
			}
			return printResult(fmt.Sprintf("Color scheme set to %s\n", scheme), map[string]any{
				"ok":          true,
				"colorScheme": scheme,
			})
		},
	})

	rootCmd.AddCommand(setCmd)
}
