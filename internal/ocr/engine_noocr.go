//go:build !ocr

package ocr

import (
	"context"
	"fmt"
)

// Engine is a stub OCR engine when built without the "ocr" build tag.
// All methods return an error indicating OCR is not available.
type Engine struct {
	DebugDir string
}

// NewEngine returns an error because OCR support is not compiled in.
// Build with -tags=ocr to enable OCR (requires Tesseract).
func NewEngine(langs ...string) (*Engine, error) {
	return nil, fmt.Errorf("OCR is not available: binary was built without OCR support.\n" +
		"To enable OCR, rebuild with: go build -tags=ocr\n" +
		"This requires Tesseract to be installed on your system.")
}

// Close is a no-op for the stub engine.
func (e *Engine) Close() {}

// RecognizeElement always returns an error for the stub engine.
func (e *Engine) RecognizeElement(ctx context.Context, backendNodeID int64) (string, error) {
	return "", fmt.Errorf("OCR is not available: binary was built without OCR support")
}
