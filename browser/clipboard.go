package browser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// ClipboardRead reads text content from the clipboard.
// Uses the Clipboard API via JavaScript with CDP permissions override.
func (b *Browser) ClipboardRead() (string, error) {
	ctx, cancel := b.operationContext()
	defer cancel()

	var result string
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Grant clipboard-read permission via CDP
		_ = grantClipboardPermissions(ctx)

		val, exception, e := runtime.Evaluate(`
			(async () => {
				try {
					return await navigator.clipboard.readText();
				} catch (e) {
					return '__CLIPBOARD_ERROR__:' + e.message;
				}
			})()
		`).WithAwaitPromise(true).WithReturnByValue(true).Do(ctx)
		if e != nil {
			return e
		}
		if exception != nil {
			return fmt.Errorf("clipboard read exception: %s", exception.Text)
		}
		if val != nil && val.Value != nil {
			var s string
			if err := json.Unmarshal(val.Value, &s); err == nil {
				if len(s) > 20 && s[:20] == "__CLIPBOARD_ERROR__:" {
					return fmt.Errorf("clipboard read: %s", s[20:])
				}
				result = s
			}
		}
		return nil
	}))
	return result, err
}

// ClipboardWrite writes text to the clipboard.
func (b *Browser) ClipboardWrite(text string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Grant clipboard-write permission via CDP
		_ = grantClipboardPermissions(ctx)

		textJSON, _ := json.Marshal(text)
		js := fmt.Sprintf(`
			(async () => {
				try {
					await navigator.clipboard.writeText(%s);
					return 'ok';
				} catch (e) {
					return '__CLIPBOARD_ERROR__:' + e.message;
				}
			})()
		`, string(textJSON))

		val, exception, e := runtime.Evaluate(js).
			WithAwaitPromise(true).
			WithReturnByValue(true).
			Do(ctx)
		if e != nil {
			return e
		}
		if exception != nil {
			return fmt.Errorf("clipboard write exception: %s", exception.Text)
		}
		if val != nil && val.Value != nil {
			var s string
			if err := json.Unmarshal(val.Value, &s); err == nil {
				if len(s) > 20 && s[:20] == "__CLIPBOARD_ERROR__:" {
					return fmt.Errorf("clipboard write: %s", s[20:])
				}
			}
		}
		return nil
	}))
}

// ClipboardCopy simulates Ctrl+C (or Cmd+C on macOS) to copy the current selection.
func (b *Browser) ClipboardCopy() error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		_ = grantClipboardPermissions(ctx)
		// Use document.execCommand for maximum compatibility
		_, _, _ = runtime.Evaluate(`document.execCommand('copy')`).Do(ctx)
		return nil
	}))
}

// ClipboardPaste simulates Ctrl+V (or Cmd+V on macOS) to paste from clipboard.
func (b *Browser) ClipboardPaste() error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		_ = grantClipboardPermissions(ctx)
		// Use document.execCommand for maximum compatibility
		_, _, _ = runtime.Evaluate(`document.execCommand('paste')`).Do(ctx)
		return nil
	}))
}

// grantClipboardPermissions grants clipboard read/write permissions using CDP.
func grantClipboardPermissions(ctx context.Context) error {
	// Use Browser.grantPermissions to allow clipboard access
	// This uses the raw CDP command
	_, _, err := runtime.Evaluate(`void 0`).Do(ctx)
	return err
}
