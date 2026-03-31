package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Check and install browser dependencies",
		Long: `Verify that a compatible browser (Chrome/Chromium) is available.
If not found, provides instructions or attempts installation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			withDeps, _ := cmd.Flags().GetBool("with-deps")

			// Try to find Chrome
			chromePath := findChrome()
			if chromePath != "" {
				return printResult(fmt.Sprintf("Browser found: %s\n", chromePath), map[string]any{
					"ok":   true,
					"path": chromePath,
				})
			}

			fmt.Fprintln(os.Stderr, "No compatible browser found.")

			switch runtime.GOOS {
			case "darwin":
				fmt.Fprintln(os.Stderr, "\nTo install Chrome on macOS:")
				fmt.Fprintln(os.Stderr, "  brew install --cask google-chrome")
				fmt.Fprintln(os.Stderr, "\nOr download from: https://www.google.com/chrome/")

				if withDeps {
					fmt.Fprintln(os.Stderr, "\nAttempting installation via Homebrew...")
					c := exec.Command("brew", "install", "--cask", "google-chrome")
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					if err := c.Run(); err != nil {
						return fmt.Errorf("brew install failed: %w", err)
					}
					return printResult("Chrome installed successfully\n", map[string]any{"ok": true})
				}

			case "linux":
				fmt.Fprintln(os.Stderr, "\nTo install Chromium on Linux:")
				fmt.Fprintln(os.Stderr, "  # Debian/Ubuntu:")
				fmt.Fprintln(os.Stderr, "  sudo apt-get install -y chromium-browser")
				fmt.Fprintln(os.Stderr, "  # or:")
				fmt.Fprintln(os.Stderr, "  sudo apt-get install -y google-chrome-stable")
				fmt.Fprintln(os.Stderr, "  # Alpine:")
				fmt.Fprintln(os.Stderr, "  apk add chromium")

				if withDeps {
					fmt.Fprintln(os.Stderr, "\nAttempting installation...")
					// Try apt first
					if _, err := exec.LookPath("apt-get"); err == nil {
						c := exec.Command("apt-get", "install", "-y", "chromium-browser")
						c.Stdout = os.Stdout
						c.Stderr = os.Stderr
						if err := c.Run(); err == nil {
							return printResult("Chromium installed successfully\n", map[string]any{"ok": true})
						}
					}
					// Try apk
					if _, err := exec.LookPath("apk"); err == nil {
						c := exec.Command("apk", "add", "chromium")
						c.Stdout = os.Stdout
						c.Stderr = os.Stderr
						if err := c.Run(); err == nil {
							return printResult("Chromium installed successfully\n", map[string]any{"ok": true})
						}
					}
					return fmt.Errorf("automatic installation failed; please install manually")
				}

			case "windows":
				fmt.Fprintln(os.Stderr, "\nTo install Chrome on Windows:")
				fmt.Fprintln(os.Stderr, "  choco install googlechrome")
				fmt.Fprintln(os.Stderr, "\nOr download from: https://www.google.com/chrome/")

			default:
				fmt.Fprintf(os.Stderr, "\nDownload Chrome from: https://www.google.com/chrome/\n")
			}

			return fmt.Errorf("browser not found; install Chrome or Chromium and try again")
		},
	}

	installCmd.Flags().BoolP("with-deps", "d", false, "Attempt to install browser and system dependencies")
	rootCmd.AddCommand(installCmd)
}

// findChrome tries to locate a Chrome/Chromium executable.
func findChrome() string {
	// Common executable names
	candidates := []string{
		"google-chrome-stable",
		"google-chrome",
		"chromium-browser",
		"chromium",
		"chrome",
	}

	// On macOS, also check Application paths
	if runtime.GOOS == "darwin" {
		macPaths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		}
		for _, p := range macPaths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// On Windows, check common install paths
	if runtime.GOOS == "windows" {
		winPaths := []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		}
		for _, p := range winPaths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	return ""
}
