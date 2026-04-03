package browser

import (
	"os"
	"runtime"
	"time"
)

type Options struct {
	Headless bool
	Timeout  time.Duration
	Logf     func(string, ...any)

	// Profile is a path to a Chrome user data directory for persistent sessions.
	// When set, cookies, IndexedDB, cache, etc. are preserved across restarts.
	Profile string

	// StatePath is a path to a JSON file containing saved browser state
	// (cookies + localStorage) to import when creating the browser.
	StatePath string

	UserAgent         string   // custom User-Agent string
	Proxy             string   // proxy server URL (e.g. "http://proxy:8080" or "socks5://proxy:1080")
	ProxyBypass       string   // comma-separated hosts to bypass proxy
	IgnoreHTTPSErrors bool     // ignore HTTPS certificate errors
	AllowFileAccess   bool     // allow file:// URL access across origins
	Extensions        []string // paths to Chrome extension directories to load
	ExtraArgs         []string // extra Chrome command-line arguments
	DownloadPath      string   // default download directory
	ScreenshotDir     string   // default screenshot output directory
	ScreenshotFormat  string   // default screenshot format: "png" or "jpeg"
}

func DefaultOptions() Options {
	return Options{
		Headless: true,
		Timeout:  30 * time.Second,
	}
}

func (o Options) normalized() Options {
	defaults := DefaultOptions()
	if o.Timeout <= 0 {
		o.Timeout = defaults.Timeout
	}
	return o
}

func shouldDisableSandbox() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if os.Getenv("KO_BROWSER_NO_SANDBOX") != "" {
		return true
	}
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return true
	}
	return os.Getenv("CI") != ""
}
