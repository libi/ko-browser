package browser

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// NetworkRequest records information about a completed network request.
type NetworkRequest struct {
	URL        string `json:"url"`
	Method     string `json:"method"`
	Status     int64  `json:"status"`
	StatusText string `json:"statusText,omitempty"`
	Type       string `json:"type,omitempty"`
}

// networkState holds internal state for network interception.
type networkState struct {
	mu       sync.Mutex
	routes   map[string]*routeEntry
	requests []NetworkRequest
	fetching bool // whether fetch domain is enabled
	logging  bool // whether request logging is enabled
}

// RouteAction defines what to do with a matched request.
type RouteAction int

const (
	RouteBlock    RouteAction = iota // Block the request (fail it)
	RouteContinue                    // Let it continue (optionally modified)
)

// routeEntry stores a registered route pattern with its action.
type routeEntry struct {
	pattern string
	action  RouteAction
	body    string // optional: custom response body for fulfill
}

// initNetworkState initializes the network state if not already done.
func (b *Browser) initNetworkState() {
	if b.networkState == nil {
		b.networkState = &networkState{
			routes: make(map[string]*routeEntry),
		}
	}
}

// NetworkRoute registers a route pattern to intercept requests.
// Matching requests will be handled based on the action:
//   - RouteBlock: the request is failed (blocked)
//   - RouteContinue: the request is allowed through
func (b *Browser) NetworkRoute(pattern string, action RouteAction) error {
	b.initNetworkState()
	b.networkState.mu.Lock()
	b.networkState.routes[pattern] = &routeEntry{pattern: pattern, action: action}
	b.networkState.mu.Unlock()

	// Enable fetch domain if not already enabled
	if !b.networkState.fetching {
		ctx, cancel := b.operationContext()
		defer cancel()

		err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			// Enable fetch interception with a wildcard pattern
			return fetch.Enable().WithPatterns([]*fetch.RequestPattern{
				{URLPattern: "*"},
			}).Do(ctx)
		}))
		if err != nil {
			return err
		}
		b.networkState.fetching = true

		// Set up event listener for fetch.RequestPaused
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			if paused, ok := ev.(*fetch.EventRequestPaused); ok {
				go b.handleFetchPaused(paused)
			}
		})
	}
	return nil
}

// NetworkUnroute removes a previously registered route pattern.
func (b *Browser) NetworkUnroute(pattern string) error {
	b.initNetworkState()
	b.networkState.mu.Lock()
	delete(b.networkState.routes, pattern)
	remaining := len(b.networkState.routes)
	b.networkState.mu.Unlock()

	// If no more routes, disable fetch domain
	if remaining == 0 && b.networkState.fetching {
		ctx, cancel := b.operationContext()
		defer cancel()
		_ = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return fetch.Disable().Do(ctx)
		}))
		b.networkState.fetching = false
	}
	return nil
}

// NetworkRequests returns a list of recorded network requests.
// Call NetworkStartLogging() first to begin recording.
func (b *Browser) NetworkRequests() ([]NetworkRequest, error) {
	b.initNetworkState()
	b.networkState.mu.Lock()
	defer b.networkState.mu.Unlock()

	result := make([]NetworkRequest, len(b.networkState.requests))
	copy(result, b.networkState.requests)
	return result, nil
}

// NetworkStartLogging enables recording of network requests.
func (b *Browser) NetworkStartLogging() error {
	b.initNetworkState()
	if b.networkState.logging {
		return nil
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return network.Enable().Do(ctx)
	}))
	if err != nil {
		return err
	}
	b.networkState.logging = true

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if resp, ok := ev.(*network.EventResponseReceived); ok {
			b.networkState.mu.Lock()
			b.networkState.requests = append(b.networkState.requests, NetworkRequest{
				URL:        resp.Response.URL,
				Method:     "", // Will be enriched from requestWillBeSent if available
				Status:     resp.Response.Status,
				StatusText: resp.Response.StatusText,
				Type:       string(resp.Type),
			})
			b.networkState.mu.Unlock()
		}
	})

	return nil
}

// NetworkClearRequests clears all recorded network requests.
func (b *Browser) NetworkClearRequests() {
	b.initNetworkState()
	b.networkState.mu.Lock()
	b.networkState.requests = nil
	b.networkState.mu.Unlock()
}

// handleFetchPaused handles a paused fetch request.
func (b *Browser) handleFetchPaused(ev *fetch.EventRequestPaused) {
	ctx, cancel := b.operationContext()
	defer cancel()

	b.networkState.mu.Lock()
	matched := false
	var action RouteAction
	for pattern, entry := range b.networkState.routes {
		if matchURLPattern(ev.Request.URL, pattern) {
			matched = true
			action = entry.action
			break
		}
	}
	b.networkState.mu.Unlock()

	_ = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		if matched && action == RouteBlock {
			return fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
		}
		// Continue the request
		return fetch.ContinueRequest(ev.RequestID).Do(ctx)
	}))
}

// matchURLPattern checks if a URL matches a simple glob-like pattern.
// Supports * as wildcard and ** for path matching.
func matchURLPattern(url, pattern string) bool {
	// Exact match
	if url == pattern {
		return true
	}
	// Simple contains check for patterns like "*google*"
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		inner := pattern[1 : len(pattern)-1]
		return strings.Contains(url, inner)
	}
	// Suffix match: "*.js"
	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(url, suffix)
	}
	// Prefix match: "https://api.*"
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(url, prefix)
	}
	// Path matching with **
	if strings.Contains(pattern, "**") {
		parts := strings.SplitN(pattern, "**", 2)
		return strings.HasPrefix(url, parts[0]) && strings.HasSuffix(url, parts[1])
	}
	return false
}

// FormatNetworkRequests formats network request list as human-readable text.
func FormatNetworkRequests(reqs []NetworkRequest) string {
	if len(reqs) == 0 {
		return "No requests recorded\n"
	}
	var out string
	for i, r := range reqs {
		method := r.Method
		if method == "" {
			method = "GET"
		}
		out += fmt.Sprintf("%d: %s %s → %d %s\n", i+1, method, r.URL, r.Status, r.StatusText)
	}
	return out
}
