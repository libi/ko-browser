package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cdpbrowser "github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

// WaitSelector waits for an element matching the CSS selector to appear in the DOM.
func (b *Browser) WaitSelector(cssSelector string, timeout ...time.Duration) error {
	ctx, cancel := b.waitContext(timeout...)
	defer cancel()

	return chromedp.Run(ctx, chromedp.WaitVisible(cssSelector, chromedp.ByQuery))
}

// WaitURL waits until the page URL contains the given substring.
func (b *Browser) WaitURL(pattern string, timeout ...time.Duration) error {
	ctx, cancel := b.waitContext(timeout...)
	defer cancel()

	return b.pollUntil(ctx, func() (bool, error) {
		url, err := b.GetURL()
		if err != nil {
			return false, err
		}
		if strings.Contains(pattern, "*") {
			return matchGlob(pattern, url), nil
		}
		return strings.Contains(url, pattern), nil
	})
}

// WaitLoad waits until the page reaches the "load" state (document.readyState === "complete").
func (b *Browser) WaitLoad(timeout ...time.Duration) error {
	ctx, cancel := b.waitContext(timeout...)
	defer cancel()

	return chromedp.Run(ctx, chromedp.WaitReady("body", chromedp.ByQuery))
}

// WaitText waits until the given text appears somewhere in the page body.
func (b *Browser) WaitText(text string, timeout ...time.Duration) error {
	ctx, cancel := b.waitContext(timeout...)
	defer cancel()

	return b.pollUntil(ctx, func() (bool, error) {
		result, err := b.evaluateString(fmt.Sprintf(
			`document.body.innerText.includes(%s) ? 'true' : 'false'`, mustJSON(text)))
		if err != nil {
			return false, err
		}
		return result == "true", nil
	})
}

// WaitFunc waits until the given JavaScript expression evaluates to a truthy value.
func (b *Browser) WaitFunc(expression string, timeout ...time.Duration) error {
	ctx, cancel := b.waitContext(timeout...)
	defer cancel()

	return b.pollUntil(ctx, func() (bool, error) {
		result, err := b.evaluateString(fmt.Sprintf(`Boolean(%s) ? 'true' : 'false'`, expression))
		if err != nil {
			return false, err
		}
		return result == "true", nil
	})
}

// waitContext creates a context with the given timeout (or default browser timeout).
func (b *Browser) waitContext(timeout ...time.Duration) (context.Context, context.CancelFunc) {
	d := b.timeout
	if len(timeout) > 0 && timeout[0] > 0 {
		d = timeout[0]
	}
	return context.WithTimeout(b.ctx, d)
}

// pollUntil polls a condition function every 100ms until it returns true or the context expires.
func (b *Browser) pollUntil(ctx context.Context, condition func() (bool, error)) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Check immediately
	if ok, err := condition(); err != nil {
		return err
	} else if ok {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait timed out: %w", ctx.Err())
		case <-ticker.C:
			ok, err := condition()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

// matchGlob performs simple glob matching with * wildcards.
func matchGlob(pattern, s string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == s
	}
	pos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(s[pos:], part)
		if idx < 0 {
			return false
		}
		if i == 0 && idx != 0 {
			return false
		}
		pos += idx + len(part)
	}
	if last := parts[len(parts)-1]; last != "" {
		return strings.HasSuffix(s, last)
	}
	return true
}

// WaitHidden waits for an element matching the CSS selector to become hidden or removed from the DOM.
func (b *Browser) WaitHidden(cssSelector string, timeout ...time.Duration) error {
	ctx, cancel := b.waitContext(timeout...)
	defer cancel()

	return b.pollUntil(ctx, func() (bool, error) {
		result, err := b.evaluateString(fmt.Sprintf(`(() => {
			const el = document.querySelector(%s);
			if (!el) return 'true';
			const style = window.getComputedStyle(el);
			if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') return 'true';
			return 'false';
		})()`, mustJSON(cssSelector)))
		if err != nil {
			return false, err
		}
		return result == "true", nil
	})
}

// WaitDownload sets up a download handler, waits for a download event, and saves the file.
// The savePath is the directory where the downloaded file will be saved.
// Returns the downloaded file path.
func (b *Browser) WaitDownload(savePath string, timeout ...time.Duration) (string, error) {
	ctx, cancel := b.waitContext(timeout...)
	defer cancel()

	// Ensure the save directory exists
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return "", fmt.Errorf("create download dir: %w", err)
	}

	// Enable download events
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return cdpbrowser.SetDownloadBehavior(cdpbrowser.SetDownloadBehaviorBehaviorAllow).
			WithDownloadPath(savePath).WithEventsEnabled(true).Do(ctx)
	})); err != nil {
		return "", fmt.Errorf("set download behavior: %w", err)
	}

	// Poll for new files in the save directory
	initialFiles := make(map[string]bool)
	entries, _ := os.ReadDir(savePath)
	for _, e := range entries {
		initialFiles[e.Name()] = true
	}

	var downloadedFile string
	err := b.pollUntil(ctx, func() (bool, error) {
		entries, err := os.ReadDir(savePath)
		if err != nil {
			return false, err
		}
		for _, e := range entries {
			name := e.Name()
			if initialFiles[name] {
				continue
			}
			// Skip partial download files (Chrome uses .crdownload)
			if strings.HasSuffix(name, ".crdownload") || strings.HasSuffix(name, ".tmp") {
				continue
			}
			downloadedFile = filepath.Join(savePath, name)
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return "", fmt.Errorf("wait download: %w", err)
	}
	return downloadedFile, nil
}
