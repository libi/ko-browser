package browser

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-json-experiment/json"
)

// Connect connects to an already-running Chrome instance via CDP.
// The target can be:
//   - A port number string like "9222" → connects to http://localhost:9222
//   - A WebSocket URL like "ws://localhost:9222/devtools/browser/..."
//   - A remote WebSocket URL like "wss://remote.example.com/cdp?token=..."
func Connect(target string, opts Options) (*Browser, error) {
	opts = opts.normalized()

	wsURL, err := resolveWSURL(target, opts.Timeout)
	if err != nil {
		return nil, fmt.Errorf("resolve CDP endpoint: %w", err)
	}

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), wsURL)

	ctxOpts := []chromedp.ContextOption{}
	if opts.Logf != nil {
		ctxOpts = append(ctxOpts, chromedp.WithLogf(opts.Logf))
	}

	ctx, cancel := chromedp.NewContext(allocCtx, ctxOpts...)

	// Verify connectivity by running a trivial action
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		allocCancel()
		return nil, fmt.Errorf("connect to browser: %w", err)
	}

	return &Browser{
		ctx:       ctx,
		cancel:    cancel,
		allocCtx:  allocCtx,
		allocCanc: allocCancel,
		timeout:   opts.Timeout,
	}, nil
}

// resolveWSURL converts a target string to a full WebSocket URL.
func resolveWSURL(target string, timeout time.Duration) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("empty target")
	}

	// Already a WebSocket URL
	if strings.HasPrefix(target, "ws://") || strings.HasPrefix(target, "wss://") {
		return target, nil
	}

	// Port number → discover from /json/version
	port := target
	if !isNumeric(port) {
		return "", fmt.Errorf("invalid target %q: expected port number or ws:// URL", target)
	}

	return discoverWSURL(fmt.Sprintf("http://localhost:%s", port), timeout)
}

// discoverWSURL queries a Chrome DevTools HTTP endpoint to find the WebSocket debugger URL.
func discoverWSURL(baseURL string, timeout time.Duration) (string, error) {
	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(baseURL + "/json/version")
	if err != nil {
		return "", fmt.Errorf("query %s/json/version: %w", baseURL, err)
	}
	defer resp.Body.Close()

	var info struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.UnmarshalRead(resp.Body, &info); err != nil {
		return "", fmt.Errorf("parse /json/version: %w", err)
	}
	if info.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("no webSocketDebuggerUrl in /json/version response")
	}

	return info.WebSocketDebuggerURL, nil
}

// GetCDPURL returns the CDP WebSocket URL for the current browser session.
// This queries the browser target's URL from the active context.
func (b *Browser) GetCDPURL() (string, error) {
	targets, err := chromedp.Targets(b.ctx)
	if err != nil {
		return "", fmt.Errorf("get targets: %w", err)
	}
	if len(targets) == 0 {
		return "", fmt.Errorf("no browser targets available")
	}
	// Return the URL of the first page target
	return targets[0].URL, nil
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
