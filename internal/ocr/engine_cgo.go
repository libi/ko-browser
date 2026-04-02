//go:build ocr

package ocr

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/otiai10/gosseract/v2"
)

// Engine wraps a gosseract client for OCR on browser element screenshots.
// It is designed to be reused across multiple OCR calls (one per page navigation).
type Engine struct {
	client   *gosseract.Client
	langs    []string
	hasCJK   bool   // whether any CJK language is configured
	DebugDir string // if set, save preprocessed screenshots here for debugging
}

// NewEngine creates a new OCR engine with the specified languages.
// Common language codes: "eng" (English), "chi_sim" (Simplified Chinese),
// "chi_tra" (Traditional Chinese), "jpn" (Japanese), "kor" (Korean).
// If no langs are specified, defaults to "eng".
func NewEngine(langs ...string) (*Engine, error) {
	if len(langs) == 0 {
		langs = []string{"eng"}
	}

	client := gosseract.NewClient()
	if err := client.SetLanguage(langs...); err != nil {
		client.Close()
		return nil, fmt.Errorf("ocr: failed to set language: %w", err)
	}

	// Suppress debug output
	client.DisableOutput()

	// Detect CJK languages
	hasCJK := false
	for _, l := range langs {
		if strings.HasPrefix(l, "chi") || l == "jpn" || l == "kor" {
			hasCJK = true
			break
		}
	}

	// PSM_SINGLE_BLOCK (6): treat the image as a single uniform block of text.
	// This works best for logos, buttons, and mixed-language content where the
	// layout is neither a single line nor a full page.
	client.SetPageSegMode(gosseract.PSM_SINGLE_BLOCK)

	// Tell Tesseract the effective DPI (we capture at 3x browser scale → 216 DPI)
	_ = client.SetVariable("user_defined_dpi", "216")

	// For English-only, restrict char whitelist to reduce noise on small images.
	// For multi-language (CJK etc.), skip whitelist to allow full character sets.
	if !hasCJK && len(langs) == 1 && langs[0] == "eng" {
		_ = client.SetVariable("tessedit_char_whitelist",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 .-_/")
	}

	return &Engine{client: client, langs: langs, hasCJK: hasCJK}, nil
}

// Close releases the OCR engine resources. Must be called when done.
func (e *Engine) Close() {
	if e.client != nil {
		e.client.Close()
		e.client = nil
	}
}

// RecognizeElement takes a screenshot of a DOM element by its backendNodeID
// and runs OCR on it. Returns the recognized text (trimmed).
func (e *Engine) RecognizeElement(ctx context.Context, backendNodeID int64) (string, error) {
	if backendNodeID == 0 {
		return "", fmt.Errorf("ocr: invalid backendNodeID (0)")
	}

	// Step 1: Get the element's bounding box via DOM.getBoxModel
	var pngData []byte

	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Resolve the backend node to ensure it exists
		remoteObj, err := dom.ResolveNode().WithBackendNodeID(cdp.BackendNodeID(backendNodeID)).Do(ctx)
		if err != nil {
			return fmt.Errorf("resolve node: %w", err)
		}
		_ = remoteObj // needed to validate node existence

		// Get box model using the backend node ID
		boxModel, err := dom.GetBoxModel().WithBackendNodeID(cdp.BackendNodeID(backendNodeID)).Do(ctx)
		if err != nil {
			return fmt.Errorf("get box model: %w", err)
		}

		// Content quad: [x1,y1, x2,y2, x3,y3, x4,y4]
		quad := boxModel.Content
		if len(quad) < 8 {
			return fmt.Errorf("unexpected quad length: %d", len(quad))
		}

		x := quad[0]
		y := quad[1]
		w := quad[2] - quad[0]
		h := quad[5] - quad[1]

		if w <= 0 || h <= 0 {
			return fmt.Errorf("element has zero dimensions: %.0fx%.0f", w, h)
		}

		// Capture screenshot of just this region at 3x scale for better OCR
		clip := &page.Viewport{
			X:      x,
			Y:      y,
			Width:  w,
			Height: h,
			Scale:  3,
		}

		data, err := page.CaptureScreenshot().
			WithFormat(page.CaptureScreenshotFormatPng).
			WithClip(clip).
			Do(ctx)
		if err != nil {
			return fmt.Errorf("capture screenshot: %w", err)
		}

		pngData = data
		return nil
	}))
	if err != nil {
		return "", fmt.Errorf("ocr: screenshot failed: %w", err)
	}

	if len(pngData) == 0 {
		return "", fmt.Errorf("ocr: empty screenshot data")
	}

	// Keep raw data for debug before preprocessing
	rawData := pngData

	// Preprocess: decode → grayscale → binarize → add padding → re-encode
	// This dramatically improves Tesseract accuracy on colored/stylized content.
	pngData, err = preprocessForOCR(pngData, 20)
	if err != nil {
		return "", fmt.Errorf("ocr: preprocess failed: %w", err)
	}

	// Debug: save preprocessed image to disk for inspection
	if e.DebugDir != "" {
		// Save raw screenshot (before preprocessing)
		rawPath := fmt.Sprintf("%s/ocr_raw_%d.png", e.DebugDir, backendNodeID)
		_ = os.WriteFile(rawPath, rawData, 0644)
		debugPath := fmt.Sprintf("%s/ocr_debug_%d.png", e.DebugDir, backendNodeID)
		_ = os.WriteFile(debugPath, pngData, 0644)
		fmt.Fprintf(os.Stderr, "  [OCR debug] saved raw → %s, preprocessed → %s\n", rawPath, debugPath)
	}

	// Step 2: Run OCR on the preprocessed PNG bytes
	if err := e.client.SetImageFromBytes(pngData); err != nil {
		return "", fmt.Errorf("ocr: set image failed: %w", err)
	}

	text, err := e.client.Text()
	if err != nil {
		return "", fmt.Errorf("ocr: recognition failed: %w", err)
	}
	text = strings.TrimSpace(text)

	// If PSM_SINGLE_BLOCK gives a poor result, retry with the raw image
	// and PSM_AUTO (3) which does its own layout analysis.
	if len(text) < 2 || !isPlausibleText(text) {
		e.client.SetPageSegMode(gosseract.PSM_AUTO)
		if err := e.client.SetImageFromBytes(rawData); err == nil {
			if alt, err := e.client.Text(); err == nil {
				alt = strings.TrimSpace(alt)
				if len(alt) > len(text) && isPlausibleText(alt) {
					text = alt
				}
			}
		}
		// Restore default PSM
		e.client.SetPageSegMode(gosseract.PSM_SINGLE_BLOCK)
	}

	// Filter out garbage: too short, mostly non-printable, etc.
	if !isPlausibleText(text) {
		return "", nil
	}

	// Clean up OCR artifacts: normalize whitespace
	text = cleanOCRText(text)

	return text, nil
}

// cleanOCRText normalizes OCR output by collapsing whitespace and
// removing line breaks (element text should be a single label).
func cleanOCRText(s string) string {
	// Replace newlines with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	// Collapse multiple spaces
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

// isPlausibleText checks if OCR output looks like real text rather than noise.
func isPlausibleText(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Count meaningful runes (letters, digits, CJK)
	meaningful := 0
	total := 0
	for _, r := range s {
		total++
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			(r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
			(r >= 0x3000 && r <= 0x303F) || // CJK Symbols
			(r >= 0x3040 && r <= 0x30FF) || // Hiragana + Katakana
			(r >= 0xAC00 && r <= 0xD7AF) { // Hangul
			meaningful++
		}
	}
	// At least 50% of chars should be meaningful
	return meaningful > 0 && float64(meaningful)/float64(total) >= 0.5
}

// preprocessForOCR converts a PNG screenshot into an optimized image for Tesseract:
//  1. Decode PNG
//  2. "Non-white is foreground" binarization — any pixel that is NOT close to
//     white becomes black (text). This handles red, blue, black, or any colored
//     text on a white/light background — which covers virtually all web UI elements.
//  3. Add white padding around the image
//  4. Re-encode as PNG
func preprocessForOCR(pngBytes []byte, pad int) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Binarize using "distance from white" approach:
	// If a pixel's max channel deviation from 255 exceeds threshold → it's foreground (black)
	// This captures colored text (red, blue, green, etc.) as well as dark text.
	const whiteThreshold = 60 // pixels with any channel < (255-60)=195 are foreground
	bw := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			// Max deviation from pure white
			dr := 255 - r8
			dg := 255 - g8
			db := 255 - b8
			maxDev := dr
			if dg > maxDev {
				maxDev = dg
			}
			if db > maxDev {
				maxDev = db
			}

			if maxDev > whiteThreshold {
				bw.SetGray(x, y, color.Gray{Y: 0}) // black (foreground/text)
			} else {
				bw.SetGray(x, y, color.Gray{Y: 255}) // white (background)
			}
		}
	}

	// Add white padding
	newW := w + pad*2
	newH := h + pad*2
	dst := image.NewGray(image.Rect(0, 0, newW, newH))
	// Fill with white
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			dst.SetGray(x, y, color.Gray{Y: 255})
		}
	}
	// Copy binarized image to center
	draw.Draw(dst, image.Rect(pad, pad, pad+w, pad+h), bw, image.Point{}, draw.Src)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
