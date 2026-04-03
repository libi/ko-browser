package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/libi/ko-browser/internal/session"
	"github.com/spf13/cobra"
)

var rootFlags struct {
	session           string
	headed            bool
	json              bool
	debug             bool
	timeout           time.Duration
	contentBoundaries bool
	profile           string
	statePath         string
	config            string

	userAgent         string
	proxy             string
	proxyBypass       string
	ignoreHTTPSErrors bool
	allowFileAccess   bool
	extensions        []string
	extraArgs         []string
	downloadPath      string
	screenshotDir     string
	screenshotFormat  string
	confirmActions    []string
}

var rootCmd = &cobra.Command{
	Use:           "ko-browser",
	Short:         "A fast, token-efficient browser for AI agents",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootFlags.session, "session", "default", "isolated session name")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.headed, "headed", false, "show browser window")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.json, "json", false, "JSON output")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.debug, "debug", false, "debug output")
	rootCmd.PersistentFlags().DurationVar(&rootFlags.timeout, "timeout", 30*time.Second, "operation timeout")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.contentBoundaries, "content-boundaries", false, "wrap output with <ko-browser-content> boundaries")
	rootCmd.PersistentFlags().StringVar(&rootFlags.profile, "profile", "", "Chrome user data directory for persistent sessions")
	rootCmd.PersistentFlags().StringVar(&rootFlags.statePath, "state", "", "JSON file to load saved browser state (cookies + localStorage)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.config, "config", "", "path to config file (default: auto-detect)")

	rootCmd.PersistentFlags().StringVar(&rootFlags.userAgent, "user-agent", "", "custom User-Agent string")
	rootCmd.PersistentFlags().StringVar(&rootFlags.proxy, "proxy", "", "proxy server URL (e.g. http://proxy:8080)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.proxyBypass, "proxy-bypass", "", "comma-separated hosts to bypass proxy")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.ignoreHTTPSErrors, "ignore-https-errors", false, "ignore HTTPS certificate errors")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.allowFileAccess, "allow-file-access", false, "allow file:// URL access across origins")
	rootCmd.PersistentFlags().StringSliceVar(&rootFlags.extensions, "extension", nil, "Chrome extension directory to load (can be specified multiple times)")
	rootCmd.PersistentFlags().StringSliceVar(&rootFlags.extraArgs, "args", nil, "extra Chrome command-line arguments")
	rootCmd.PersistentFlags().StringVar(&rootFlags.downloadPath, "download-path", "", "default download directory")
	rootCmd.PersistentFlags().StringVar(&rootFlags.screenshotDir, "screenshot-dir", "", "default screenshot output directory")
	rootCmd.PersistentFlags().StringVar(&rootFlags.screenshotFormat, "screenshot-format", "", "default screenshot format (png or jpeg)")
	rootCmd.PersistentFlags().StringSliceVar(&rootFlags.confirmActions, "confirm-actions", nil, "action categories that require confirmation")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return loadConfigFile()
	}
}

func sessionOptions() session.Options {
	return session.Options{
		Name:              rootFlags.session,
		Headed:            rootFlags.headed,
		Timeout:           rootFlags.timeout,
		Debug:             rootFlags.debug,
		Profile:           rootFlags.profile,
		StatePath:         rootFlags.statePath,
		UserAgent:         rootFlags.userAgent,
		Proxy:             rootFlags.proxy,
		ProxyBypass:       rootFlags.proxyBypass,
		IgnoreHTTPSErrors: rootFlags.ignoreHTTPSErrors,
		AllowFileAccess:   rootFlags.allowFileAccess,
		Extensions:        rootFlags.extensions,
		ExtraArgs:         rootFlags.extraArgs,
		DownloadPath:      rootFlags.downloadPath,
		ScreenshotDir:     rootFlags.screenshotDir,
		ScreenshotFormat:  rootFlags.screenshotFormat,
		ConfirmActions:    rootFlags.confirmActions,
	}
}

func printResult(text string, payload any) error {
	if rootFlags.json {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}
	if text != "" {
		if rootFlags.contentBoundaries {
			fmt.Println("<ko-browser-content>")
			fmt.Print(text)
			fmt.Println("</ko-browser-content>")
		} else {
			fmt.Print(text)
		}
	}
	return nil
}

// configData represents the JSON config file structure.
type configData struct {
	Headed            *bool    `json:"headed,omitempty"`
	Session           string   `json:"session,omitempty"`
	Profile           string   `json:"profile,omitempty"`
	State             string   `json:"state,omitempty"`
	Timeout           string   `json:"timeout,omitempty"`
	Debug             *bool    `json:"debug,omitempty"`
	JSON              *bool    `json:"json,omitempty"`
	ContentBoundaries *bool    `json:"contentBoundaries,omitempty"`
	UserAgent         string   `json:"userAgent,omitempty"`
	Proxy             string   `json:"proxy,omitempty"`
	ProxyBypass       string   `json:"proxyBypass,omitempty"`
	IgnoreHTTPSErrors *bool    `json:"ignoreHttpsErrors,omitempty"`
	AllowFileAccess   *bool    `json:"allowFileAccess,omitempty"`
	Extensions        []string `json:"extensions,omitempty"`
	ExtraArgs         []string `json:"args,omitempty"`
	DownloadPath      string   `json:"downloadPath,omitempty"`
	ScreenshotDir     string   `json:"screenshotDir,omitempty"`
	ScreenshotFormat  string   `json:"screenshotFormat,omitempty"`
}

// loadConfigFile loads configuration from JSON files with priority:
// 1. ~/.ko-browser/config.json (user-level)
// 2. ./ko-browser.json (project-level, overrides user-level)
// 3. --config flag (overrides everything)
// CLI flags always take final precedence over config file values.
func loadConfigFile() error {
	var configPaths []string

	// 1. User-level config
	if home, err := os.UserHomeDir(); err == nil {
		userConfig := filepath.Join(home, ".ko-browser", "config.json")
		if _, err := os.Stat(userConfig); err == nil {
			configPaths = append(configPaths, userConfig)
		}
	}

	// 2. Project-level config
	projectConfig := "ko-browser.json"
	if _, err := os.Stat(projectConfig); err == nil {
		configPaths = append(configPaths, projectConfig)
	}

	// 3. Explicit --config flag (highest priority)
	if rootFlags.config != "" {
		configPaths = append(configPaths, rootFlags.config)
	}

	// Also check KO_BROWSER_CONFIG env var
	if envConfig := os.Getenv("KO_BROWSER_CONFIG"); envConfig != "" {
		configPaths = append(configPaths, envConfig)
	}

	for _, path := range configPaths {
		if err := applyConfigFile(path); err != nil {
			// Only error on explicitly-specified config files
			if path == rootFlags.config {
				return fmt.Errorf("load config %s: %w", path, err)
			}
			// Silently ignore errors from auto-detected configs
			continue
		}
	}

	return nil
}

// applyConfigFile reads and applies a single config file.
// Config file values only apply when the corresponding CLI flag was not explicitly set.
func applyConfigFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg configData
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	// Only apply config values when CLI flag was not explicitly set
	if cfg.Headed != nil && !rootCmd.PersistentFlags().Lookup("headed").Changed {
		rootFlags.headed = *cfg.Headed
	}
	if cfg.Session != "" && !rootCmd.PersistentFlags().Lookup("session").Changed {
		rootFlags.session = cfg.Session
	}
	if cfg.Profile != "" && !rootCmd.PersistentFlags().Lookup("profile").Changed {
		rootFlags.profile = cfg.Profile
	}
	if cfg.State != "" && !rootCmd.PersistentFlags().Lookup("state").Changed {
		rootFlags.statePath = cfg.State
	}
	if cfg.Timeout != "" && !rootCmd.PersistentFlags().Lookup("timeout").Changed {
		d, err := time.ParseDuration(cfg.Timeout)
		if err == nil {
			rootFlags.timeout = d
		}
	}
	if cfg.Debug != nil && !rootCmd.PersistentFlags().Lookup("debug").Changed {
		rootFlags.debug = *cfg.Debug
	}
	if cfg.JSON != nil && !rootCmd.PersistentFlags().Lookup("json").Changed {
		rootFlags.json = *cfg.JSON
	}
	if cfg.ContentBoundaries != nil && !rootCmd.PersistentFlags().Lookup("content-boundaries").Changed {
		rootFlags.contentBoundaries = *cfg.ContentBoundaries
	}

	if cfg.UserAgent != "" && !rootCmd.PersistentFlags().Lookup("user-agent").Changed {
		rootFlags.userAgent = cfg.UserAgent
	}
	if cfg.Proxy != "" && !rootCmd.PersistentFlags().Lookup("proxy").Changed {
		rootFlags.proxy = cfg.Proxy
	}
	if cfg.ProxyBypass != "" && !rootCmd.PersistentFlags().Lookup("proxy-bypass").Changed {
		rootFlags.proxyBypass = cfg.ProxyBypass
	}
	if cfg.IgnoreHTTPSErrors != nil && !rootCmd.PersistentFlags().Lookup("ignore-https-errors").Changed {
		rootFlags.ignoreHTTPSErrors = *cfg.IgnoreHTTPSErrors
	}
	if cfg.AllowFileAccess != nil && !rootCmd.PersistentFlags().Lookup("allow-file-access").Changed {
		rootFlags.allowFileAccess = *cfg.AllowFileAccess
	}
	if len(cfg.Extensions) > 0 && !rootCmd.PersistentFlags().Lookup("extension").Changed {
		rootFlags.extensions = cfg.Extensions
	}
	if len(cfg.ExtraArgs) > 0 && !rootCmd.PersistentFlags().Lookup("args").Changed {
		rootFlags.extraArgs = cfg.ExtraArgs
	}
	if cfg.DownloadPath != "" && !rootCmd.PersistentFlags().Lookup("download-path").Changed {
		rootFlags.downloadPath = cfg.DownloadPath
	}
	if cfg.ScreenshotDir != "" && !rootCmd.PersistentFlags().Lookup("screenshot-dir").Changed {
		rootFlags.screenshotDir = cfg.ScreenshotDir
	}
	if cfg.ScreenshotFormat != "" && !rootCmd.PersistentFlags().Lookup("screenshot-format").Changed {
		rootFlags.screenshotFormat = cfg.ScreenshotFormat
	}

	return nil
}
