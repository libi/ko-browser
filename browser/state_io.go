package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// BrowserState represents the serializable state of a browser session,
// including cookies and localStorage data.
type BrowserState struct {
	Cookies      []CookieInfo      `json:"cookies,omitempty"`
	LocalStorage map[string]string `json:"localStorage,omitempty"`
	// Origin is the URL origin used for localStorage operations.
	Origin string `json:"origin,omitempty"`
}

// ExportState exports the current browser state (cookies + localStorage) to a JSON file.
func (b *Browser) ExportState(outputPath string) error {
	state := &BrowserState{}

	ctx, cancel := b.operationContext()
	defer cancel()

	// Export cookies
	cookies, err := b.CookiesGet()
	if err != nil {
		return fmt.Errorf("get cookies: %w", err)
	}
	state.Cookies = cookies

	// Export localStorage (try to get current origin)
	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Evaluate("location.href", &currentURL)); err == nil && currentURL != "" && currentURL != "about:blank" {
		localStorage, err := b.StorageGetAll("local")
		if err == nil {
			state.LocalStorage = localStorage
			state.Origin = currentURL
		}
	}

	// Marshal and write
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// ImportState imports browser state (cookies + localStorage) from a JSON file.
func (b *Browser) ImportState(inputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read state file: %w", err)
	}

	var state BrowserState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("parse state file: %w", err)
	}

	return b.ApplyState(&state)
}

// ApplyState applies a BrowserState to the current browser session.
func (b *Browser) ApplyState(state *BrowserState) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	// Import cookies
	if len(state.Cookies) > 0 {
		for _, cookie := range state.Cookies {
			if err := b.importCookie(ctx, cookie); err != nil {
				// Log but don't fail on individual cookie errors
				continue
			}
		}
	}

	// Import localStorage if we have an origin to navigate to
	if state.Origin != "" && len(state.LocalStorage) > 0 {
		// Navigate to the origin first to set localStorage
		if err := chromedp.Run(ctx, chromedp.Navigate(state.Origin)); err != nil {
			return fmt.Errorf("navigate to origin for localStorage: %w", err)
		}

		for key, value := range state.LocalStorage {
			if err := b.StorageSet("local", key, value); err != nil {
				// Log but don't fail on individual storage errors
				continue
			}
		}
	}

	return nil
}

// importCookie sets a single cookie via CDP.
func (b *Browser) importCookie(ctx context.Context, cookie CookieInfo) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		params := network.SetCookie(cookie.Name, cookie.Value)
		if cookie.Domain != "" {
			params = params.WithDomain(cookie.Domain)
		}
		if cookie.Path != "" {
			params = params.WithPath(cookie.Path)
		}
		if cookie.URL != "" {
			params = params.WithURL(cookie.URL)
		}
		if cookie.HTTPOnly {
			params = params.WithHTTPOnly(true)
		}
		if cookie.Secure {
			params = params.WithSecure(true)
		}
		if cookie.Expires > 0 {
			t := cdp.TimeSinceEpoch(time.Unix(int64(cookie.Expires), 0))
			params = params.WithExpires(&t)
		}
		if cookie.SameSite != "" {
			ss := network.CookieSameSite(cookie.SameSite)
			params = params.WithSameSite(ss)
		}
		return params.Do(ctx)
	}))
}
