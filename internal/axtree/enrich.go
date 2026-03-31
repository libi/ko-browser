package axtree

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/libi/ko-browser/internal/ocr"
)

// EnrichWithOCR traverses the filtered tree and applies OCR to image/interactive
// elements that have no accessible name. This fills in the gap where AX Tree
// provides no text info for image buttons/links.
//
// Patterns that trigger OCR (to minimize overhead):
//   - Any image/img node with no name (images are visual; missing alt = OCR)
//   - link/button nodes with no name and no text children (likely icon-only)
//
// The OCR engine should be pre-initialized and reused across calls.
func EnrichWithOCR(ctx context.Context, nodes []*Node, engine *ocr.Engine) {
	if engine == nil {
		return
	}
	for _, n := range nodes {
		enrichNode(ctx, n, engine, false)
	}
}

func enrichNode(ctx context.Context, node *Node, engine *ocr.Engine, parentIsInteractive bool) {
	if node == nil {
		return
	}

	roleLower := strings.ToLower(node.Role)
	isInteractive := interactiveRoles[roleLower]
	isImage := roleLower == "img" || roleLower == "image"

	// Case 1: Any image with no name → OCR.
	// Images are inherently visual; if the AX Tree has no alt/name, OCR is the
	// right fallback regardless of whether the parent is interactive.
	if isImage && node.Name == "" && node.BackendID > 0 {
		text, err := engine.RecognizeElement(ctx, node.BackendID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [OCR] failed for image backendID=%d: %v\n", node.BackendID, err)
		} else if text != "" {
			node.Name = text
			fmt.Fprintf(os.Stderr, "  [OCR] image backendID=%d → %q\n", node.BackendID, text)
		}
	}

	// Case 2: An interactive element (link/button) with no name and no meaningful children
	// → it's probably an icon-only button, try OCR on it
	if isInteractive && node.Name == "" && !hasNamedChildren(node) && node.BackendID > 0 {
		text, err := engine.RecognizeElement(ctx, node.BackendID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [OCR] failed for interactive backendID=%d: %v\n", node.BackendID, err)
		} else if text != "" {
			node.Name = text
			fmt.Fprintf(os.Stderr, "  [OCR] interactive %s backendID=%d → %q\n", roleLower, node.BackendID, text)
		}
	}

	// Recurse into children
	for _, child := range node.Children {
		enrichNode(ctx, child, engine, isInteractive)
	}
}

// hasNamedChildren checks if any direct child has a non-empty name.
func hasNamedChildren(node *Node) bool {
	for _, child := range node.Children {
		if child.Name != "" {
			return true
		}
	}
	return false
}
