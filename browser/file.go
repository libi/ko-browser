package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	bp "github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/dom"

	"github.com/chromedp/chromedp"
)

// Upload sets files on a file input element identified by display ID.
// The element must be an <input type="file"> element, or a label associated with one.
// If the element is not a file input, Upload will attempt to find an associated
// file input (e.g., via label's "for" attribute or child input).
func (b *Browser) Upload(id int, files ...string) error {
	// Resolve absolute paths
	absPaths := make([]string, len(files))
	for i, f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", f, err)
		}
		// Check file exists
		if _, err := os.Stat(abs); err != nil {
			return fmt.Errorf("file %q: %w", f, err)
		}
		absPaths[i] = abs
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	backendID, err := b.resolveBackendID(id)
	if err != nil {
		// Fallback: try JS-based approach to find file input
		return b.uploadViaJS(id, absPaths)
	}

	// Try directly setting files on the element
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return dom.SetFileInputFiles(absPaths).
			WithBackendNodeID(backendID).
			Do(ctx)
	}))
	if err == nil {
		return nil
	}

	// If it failed (e.g., element is a label, not a file input), try to find
	// the associated file input via JS and set files on that
	return b.uploadViaJS(id, absPaths)
}

// uploadViaJS finds the file input element associated with the given display ID
// (which might be a label, text node, or the input itself) and sets files on it.
func (b *Browser) uploadViaJS(id int, absPaths []string) error {
	// Use JS to find the associated file input element
	js := `function() {
		let el = this;
		// If it's a text node, go to its parent element
		if (el.nodeType === 3) el = el.parentElement;
		if (!el) return 'none';
		// If element is already a file input, return its unique selector
		if (el.tagName === 'INPUT' && el.type === 'file') {
			return 'self';
		}
		// If element is a label, find the associated input
		if (el.tagName === 'LABEL') {
			const forAttr = el.getAttribute('for');
			if (forAttr) {
				const input = document.getElementById(forAttr);
				if (input && input.tagName === 'INPUT' && input.type === 'file') {
					return 'for:' + forAttr;
				}
			}
			// Try to find a child file input
			const childInput = el.querySelector('input[type="file"]');
			if (childInput) {
				return 'child';
			}
		}
		// Try to find a nearby file input by walking up ancestors
		let ancestor = el.parentElement;
		for (let i = 0; i < 5 && ancestor; i++) {
			// Check if ancestor is a label with for attribute
			if (ancestor.tagName === 'LABEL') {
				const forAttr = ancestor.getAttribute('for');
				if (forAttr) {
					const input = document.getElementById(forAttr);
					if (input && input.tagName === 'INPUT' && input.type === 'file') {
						return 'for:' + forAttr;
					}
				}
			}
			const input = ancestor.querySelector('input[type="file"]');
			if (input) {
				if (input.id) return 'for:' + input.id;
				return 'sibling';
			}
			ancestor = ancestor.parentElement;
		}
		return 'none';
	}`

	result, err := b.evaluateOnElement(id, js)
	if err != nil {
		return fmt.Errorf("find file input for element %d: %w", id, err)
	}

	if result == "none" {
		return fmt.Errorf("element %d is not a file input and no associated file input found", id)
	}

	// Use the association info to find and upload to the correct file input
	switch {
	case result == "self":
		// The element itself is the file input - but we already failed with backendID
		// Try setting via CSS
		return b.uploadViaCSS(id, absPaths)
	case strings.HasPrefix(result, "for:"):
		inputID := result[4:]
		return b.UploadCSS("#"+inputID, absPaths...)
	case result == "child" || result == "sibling":
		return b.uploadViaCSS(id, absPaths)
	default:
		return fmt.Errorf("element %d: cannot determine associated file input", id)
	}
}

// uploadViaCSS finds and uploads to the file input associated with the given display ID.
func (b *Browser) uploadViaCSS(id int, absPaths []string) error {
	ordinal, err := b.interactiveOrdinal(id)
	if err != nil {
		// If the element isn't interactive (e.g., a label), try to find a nearby input
		// Use JS to get the parent and find the file input
		jsExpr := `(() => {
			const labels = document.querySelectorAll('label');
			for (const label of labels) {
				const forAttr = label.getAttribute('for');
				if (forAttr) {
					const input = document.getElementById(forAttr);
					if (input && input.tagName === 'INPUT' && input.type === 'file') {
						return '#' + forAttr;
					}
				}
			}
			const inputs = document.querySelectorAll('input[type="file"]');
			if (inputs.length > 0) {
				if (inputs[0].id) return '#' + inputs[0].id;
				return 'input[type="file"]';
			}
			return '';
		})()`
		selector, evalErr := b.evaluateString(jsExpr)
		if evalErr != nil || selector == "" {
			return fmt.Errorf("cannot find file input for element %d: %w", id, err)
		}
		return b.UploadCSS(selector, absPaths...)
	}
	_ = ordinal

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.SetUploadFiles("input[type='file']", absPaths, chromedp.ByQuery),
	)
}

// DownloadOptions controls download behavior.
type DownloadOptions struct {
	Timeout time.Duration // timeout waiting for download to complete (default: 30s)
}

// Download clicks a download link/button and saves the file to the specified directory.
// Returns the path to the downloaded file.
func (b *Browser) Download(id int, saveDir string, opts ...DownloadOptions) (string, error) {
	config := DownloadOptions{Timeout: b.timeout}
	if len(opts) > 0 {
		config = opts[0]
		if config.Timeout <= 0 {
			config.Timeout = b.timeout
		}
	}

	// Ensure save directory exists
	absDir, err := filepath.Abs(saveDir)
	if err != nil {
		return "", fmt.Errorf("resolve save dir: %w", err)
	}
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return "", fmt.Errorf("create save dir: %w", err)
	}

	ctx, cancel := context.WithTimeout(b.ctx, config.Timeout)
	defer cancel()

	// Enable download handling
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return bp.SetDownloadBehavior(bp.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(absDir).
			WithEventsEnabled(true).
			Do(ctx)
	}))
	if err != nil {
		return "", fmt.Errorf("set download behavior: %w", err)
	}

	// Set up channels to track download progress
	downloadGUID := ""
	downloadDone := make(chan string, 1)

	// Listen for download events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *bp.EventDownloadWillBegin:
			downloadGUID = e.GUID
		case *bp.EventDownloadProgress:
			if e.GUID == downloadGUID && e.State == bp.DownloadProgressStateCompleted {
				downloadDone <- e.GUID
			} else if e.GUID == downloadGUID && e.State == bp.DownloadProgressStateCanceled {
				downloadDone <- ""
			}
		}
	})

	// Click the download element
	if err := b.Click(id); err != nil {
		return "", fmt.Errorf("click download element: %w", err)
	}

	// Wait for download to complete
	select {
	case guid := <-downloadDone:
		if guid == "" {
			return "", fmt.Errorf("download was cancelled")
		}
		// Find the downloaded file in the save directory
		return b.findDownloadedFile(absDir, guid)
	case <-ctx.Done():
		return "", fmt.Errorf("download timed out after %v", config.Timeout)
	}
}

// findDownloadedFile locates the downloaded file in the save directory.
func (b *Browser) findDownloadedFile(dir string, guid string) (string, error) {
	// The file might be saved with the GUID or with the original filename
	// Try to find any new file in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read save dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip hidden files and temp files
		if strings.HasPrefix(name, ".") {
			continue
		}
		fullPath := filepath.Join(dir, name)
		return fullPath, nil
	}

	return "", fmt.Errorf("no downloaded file found in %s", dir)
}

// UploadCSS sets files on a file input element identified by CSS selector.
func (b *Browser) UploadCSS(cssSelector string, files ...string) error {
	absPaths := make([]string, len(files))
	for i, f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", f, err)
		}
		if _, err := os.Stat(abs); err != nil {
			return fmt.Errorf("file %q: %w", f, err)
		}
		absPaths[i] = abs
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.SetUploadFiles(cssSelector, absPaths, chromedp.ByQuery),
	)
}
