package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cdp "github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
)

// CookieInfo represents a browser cookie.
type CookieInfo struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires,omitempty"`
	HTTPOnly bool    `json:"httpOnly,omitempty"`
	Secure   bool    `json:"secure,omitempty"`
	SameSite string  `json:"sameSite,omitempty"`
	URL      string  `json:"url,omitempty"`
}

// CookiesGet returns all cookies for the current page.
func (b *Browser) CookiesGet() ([]CookieInfo, error) {
	ctx, cancel := b.operationContext()
	defer cancel()

	var cookies []CookieInfo
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		rawCookies, e := storage.GetCookies().Do(ctx)
		if e != nil {
			return e
		}
		for _, c := range rawCookies {
			cookies = append(cookies, CookieInfo{
				Name:     c.Name,
				Value:    c.Value,
				Domain:   c.Domain,
				Path:     c.Path,
				Expires:  c.Expires,
				HTTPOnly: c.HTTPOnly,
				Secure:   c.Secure,
				SameSite: string(c.SameSite),
			})
		}
		return nil
	}))
	return cookies, err
}

// CookieSet sets a cookie. At minimum, name and value are required.
// Domain defaults to the current page's domain if empty.
func (b *Browser) CookieSet(cookie CookieInfo) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		domain := cookie.Domain
		if domain == "" {
			// Get domain from current page URL
			var loc string
			if err := chromedp.Evaluate(`location.hostname`, &loc).Do(ctx); err == nil && loc != "" {
				domain = loc
			}
		}

		params := network.SetCookie(cookie.Name, cookie.Value).
			WithDomain(domain).
			WithPath(cookie.Path)

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

		return params.Do(ctx)
	}))
}

// CookieDelete deletes a cookie by name.
func (b *Browser) CookieDelete(name string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// First get all cookies to find the domain
		cookies, err := storage.GetCookies().Do(ctx)
		if err != nil {
			return err
		}
		for _, c := range cookies {
			if c.Name == name {
				return network.DeleteCookies(name).
					WithDomain(c.Domain).
					WithPath(c.Path).
					Do(ctx)
			}
		}
		// If cookie not found by name, try deleting anyway with current domain
		var hostname string
		_ = chromedp.Evaluate(`location.hostname`, &hostname).Do(ctx)
		return network.DeleteCookies(name).WithDomain(hostname).Do(ctx)
	}))
}

// CookiesClear clears all cookies.
func (b *Browser) CookiesClear() error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return network.ClearBrowserCookies().Do(ctx)
	}))
}

// StorageGet gets a value from localStorage or sessionStorage.
// storageType should be "local" or "session".
func (b *Browser) StorageGet(storageType, key string) (string, error) {
	ctx, cancel := b.operationContext()
	defer cancel()

	jsStorage := "localStorage"
	if storageType == "session" {
		jsStorage = "sessionStorage"
	}

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(
		fmt.Sprintf(`%s.getItem(%q)`, jsStorage, key),
		&result,
	))
	return result, err
}

// StorageSet sets a value in localStorage or sessionStorage.
// storageType should be "local" or "session".
func (b *Browser) StorageSet(storageType, key, value string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	jsStorage := "localStorage"
	if storageType == "session" {
		jsStorage = "sessionStorage"
	}

	return chromedp.Run(ctx, chromedp.Evaluate(
		fmt.Sprintf(`%s.setItem(%q, %q)`, jsStorage, key, value),
		nil,
	))
}

// StorageDelete removes a key from localStorage or sessionStorage.
func (b *Browser) StorageDelete(storageType, key string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	jsStorage := "localStorage"
	if storageType == "session" {
		jsStorage = "sessionStorage"
	}

	return chromedp.Run(ctx, chromedp.Evaluate(
		fmt.Sprintf(`%s.removeItem(%q)`, jsStorage, key),
		nil,
	))
}

// StorageClear clears all items in localStorage or sessionStorage.
func (b *Browser) StorageClear(storageType string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	jsStorage := "localStorage"
	if storageType == "session" {
		jsStorage = "sessionStorage"
	}

	return chromedp.Run(ctx, chromedp.Evaluate(
		fmt.Sprintf(`%s.clear()`, jsStorage),
		nil,
	))
}

// StorageGetAll returns all key-value pairs from localStorage or sessionStorage.
func (b *Browser) StorageGetAll(storageType string) (map[string]string, error) {
	ctx, cancel := b.operationContext()
	defer cancel()

	jsStorage := "localStorage"
	if storageType == "session" {
		jsStorage = "sessionStorage"
	}

	var jsonStr string
	err := chromedp.Run(ctx, chromedp.Evaluate(
		fmt.Sprintf(`JSON.stringify(Object.fromEntries(Object.entries(%s)))`, jsStorage),
		&jsonStr,
	))
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// FormatCookies formats cookies as human-readable text.
func FormatCookies(cookies []CookieInfo) string {
	if len(cookies) == 0 {
		return "No cookies\n"
	}
	var out string
	for _, c := range cookies {
		flags := []string{}
		if c.HTTPOnly {
			flags = append(flags, "httpOnly")
		}
		if c.Secure {
			flags = append(flags, "secure")
		}
		flagStr := ""
		if len(flags) > 0 {
			flagStr = " [" + strings.Join(flags, ", ") + "]"
		}
		out += fmt.Sprintf("%s=%s (domain=%s, path=%s)%s\n", c.Name, c.Value, c.Domain, c.Path, flagStr)
	}
	return out
}

// FormatStorage formats a storage map as human-readable text.
func FormatStorage(items map[string]string) string {
	if len(items) == 0 {
		return "No items\n"
	}
	var out string
	for k, v := range items {
		out += fmt.Sprintf("%s=%s\n", k, v)
	}
	return out
}
