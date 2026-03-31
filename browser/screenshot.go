package browser

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// ScreenshotOptions controls the screenshot behavior.
type ScreenshotOptions struct {
	FullPage  bool // capture the entire scrollable page
	Quality   int  // JPEG quality (1-100), 0 means PNG format
	ElementID int  // if > 0, capture only this element
}

// Screenshot captures a screenshot and saves it to the given path.
// Supported formats: .png (default) and .jpg/.jpeg.
func (b *Browser) Screenshot(path string, opts ...ScreenshotOptions) error {
	config := ScreenshotOptions{}
	if len(opts) > 0 {
		config = opts[0]
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	var data []byte
	var err error

	if config.ElementID > 0 {
		data, err = b.screenshotElement(ctx, config)
	} else if config.FullPage {
		data, err = b.screenshotFullPage(ctx, config, path)
	} else {
		data, err = b.screenshotViewport(ctx, config, path)
	}
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	return os.WriteFile(path, data, 0644)
}

// ScreenshotToBytes captures a screenshot and returns the raw bytes.
func (b *Browser) ScreenshotToBytes(opts ...ScreenshotOptions) ([]byte, error) {
	config := ScreenshotOptions{}
	if len(opts) > 0 {
		config = opts[0]
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	if config.ElementID > 0 {
		return b.screenshotElement(ctx, config)
	}
	if config.FullPage {
		return b.screenshotFullPage(ctx, config, "out.png")
	}
	return b.screenshotViewport(ctx, config, "out.png")
}

func (b *Browser) screenshotViewport(ctx context.Context, config ScreenshotOptions, path string) ([]byte, error) {
	format := page.CaptureScreenshotFormatPng
	if isJPEG(path) || config.Quality > 0 {
		format = page.CaptureScreenshotFormatJpeg
	}

	var buf []byte
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		params := page.CaptureScreenshot().WithFormat(format)
		if format == page.CaptureScreenshotFormatJpeg && config.Quality > 0 {
			params = params.WithQuality(int64(config.Quality))
		}
		data, e := params.Do(ctx)
		if e != nil {
			return e
		}
		buf = data
		return nil
	}))
	return buf, err
}

func (b *Browser) screenshotFullPage(ctx context.Context, config ScreenshotOptions, path string) ([]byte, error) {
	format := page.CaptureScreenshotFormatPng
	if isJPEG(path) || config.Quality > 0 {
		format = page.CaptureScreenshotFormatJpeg
	}

	var buf []byte
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Get page metrics for full-page capture
		_, _, _, _, _, contentSize, e := page.GetLayoutMetrics().Do(ctx)
		if e != nil {
			return fmt.Errorf("get layout metrics: %w", e)
		}

		params := page.CaptureScreenshot().
			WithFormat(format).
			WithCaptureBeyondViewport(true).
			WithClip(&page.Viewport{
				X:      0,
				Y:      0,
				Width:  contentSize.Width,
				Height: contentSize.Height,
				Scale:  1,
			})
		if format == page.CaptureScreenshotFormatJpeg && config.Quality > 0 {
			params = params.WithQuality(int64(config.Quality))
		}

		data, e := params.Do(ctx)
		if e != nil {
			return e
		}
		buf = data
		return nil
	}))
	return buf, err
}

func (b *Browser) screenshotElement(ctx context.Context, config ScreenshotOptions) ([]byte, error) {
	box, err := b.GetBox(config.ElementID)
	if err != nil {
		return nil, fmt.Errorf("get element box: %w", err)
	}

	var buf []byte
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		params := page.CaptureScreenshot().
			WithFormat(page.CaptureScreenshotFormatPng).
			WithClip(&page.Viewport{
				X:      box.X,
				Y:      box.Y,
				Width:  box.Width,
				Height: box.Height,
				Scale:  1,
			})

		data, e := params.Do(ctx)
		if e != nil {
			return e
		}
		buf = data
		return nil
	}))
	return buf, err
}

// PDFOptions controls PDF generation.
type PDFOptions struct {
	Landscape   bool
	PrintBG     bool    // print background graphics
	PaperWidth  float64 // inches, default 8.5
	PaperHeight float64 // inches, default 11
}

// PDF generates a PDF of the current page and saves it to path.
func (b *Browser) PDF(path string, opts ...PDFOptions) error {
	config := PDFOptions{}
	if len(opts) > 0 {
		config = opts[0]
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	var data []byte
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		params := page.PrintToPDF().
			WithPrintBackground(config.PrintBG).
			WithLandscape(config.Landscape)
		if config.PaperWidth > 0 {
			params = params.WithPaperWidth(config.PaperWidth)
		}
		if config.PaperHeight > 0 {
			params = params.WithPaperHeight(config.PaperHeight)
		}

		result, _, e := params.Do(ctx)
		if e != nil {
			return e
		}

		// result is base64-encoded
		decoded, e := base64.StdEncoding.DecodeString(string(result))
		if e != nil {
			// maybe already raw bytes
			data = result
			return nil
		}
		data = decoded
		return nil
	}))
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	return os.WriteFile(path, data, 0644)
}

// ScreenshotAnnotated captures a screenshot with snapshot element IDs overlaid
// as numbered badges near each interactive element. This is useful for AI agents
// that need to correlate visual positions with snapshot IDs.
func (b *Browser) ScreenshotAnnotated(path string, opts ...ScreenshotOptions) error {
	// Ensure we have a fresh snapshot for ID positions
	if b.lastSnap == nil {
		if _, err := b.Snapshot(); err != nil {
			return fmt.Errorf("take snapshot for annotation: %w", err)
		}
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	// Inject annotation overlays via JS
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		js := `(() => {
			// Remove any previous annotations
			document.querySelectorAll('[data-ko-annotate]').forEach(el => el.remove());

			const interactive = 'a,button,input,select,textarea,[role="button"],[role="link"],[role="textbox"],[role="checkbox"],[role="radio"],[role="combobox"],[role="tab"],[role="menuitem"],[tabindex]';
			const elements = document.querySelectorAll(interactive);
			let id = 0;
			elements.forEach((el, idx) => {
				const rect = el.getBoundingClientRect();
				if (rect.width === 0 && rect.height === 0) return;
				if (window.getComputedStyle(el).display === 'none') return;
				if (window.getComputedStyle(el).visibility === 'hidden') return;
				id++;

				const badge = document.createElement('div');
				badge.setAttribute('data-ko-annotate', 'true');
				badge.style.cssText = 'position:fixed;z-index:2147483647;pointer-events:none;' +
					'background:#e53e3e;color:white;font:bold 11px/1 monospace;' +
					'padding:1px 3px;border-radius:3px;white-space:nowrap;' +
					'left:' + Math.max(0, rect.left) + 'px;' +
					'top:' + Math.max(0, rect.top - 14) + 'px;';
				badge.textContent = id.toString();
				document.body.appendChild(badge);
			});
			return id;
		})()`
		val, exc, e := runtime.Evaluate(js).Do(ctx)
		if e != nil {
			return e
		}
		if exc != nil {
			return fmt.Errorf("annotation JS error: %s", exc.Text)
		}
		_ = val
		return nil
	}))
	if err != nil {
		return fmt.Errorf("inject annotations: %w", err)
	}

	// Take the screenshot
	screenshotErr := b.Screenshot(path, opts...)

	// Clean up annotations
	_ = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, _ = runtime.Evaluate(`document.querySelectorAll('[data-ko-annotate]').forEach(el => el.remove())`).Do(ctx)
		return nil
	}))

	return screenshotErr
}

func isJPEG(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg"
}
