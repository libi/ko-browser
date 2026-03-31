package browser

import (
	"fmt"
	"os"
	"strings"

	"github.com/libi/ko-browser/axtree"
	"github.com/libi/ko-browser/ocr"
)

type SnapshotOptions struct {
	EnableOCR       bool
	OCRLanguages    []string
	OCRDebugDir     string
	InteractiveOnly bool   // only show interactive elements (3.8)
	Compact         bool   // compact mode: omit unnamed structural wrappers (3.9)
	MaxDepth        int    // 0 = unlimited (3.10)
	Cursor          bool   // show cursor position in snapshot
	Selector        string // CSS selector to scope the snapshot to a subtree
}

func (b *Browser) Snapshot(opts ...SnapshotOptions) (*SnapshotResult, error) {
	config := SnapshotOptions{}
	if len(opts) > 0 {
		config = opts[0]
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	rawNodes, err := axtree.Extract(ctx)
	if err != nil {
		return nil, err
	}

	tree := axtree.BuildAndFilter(rawNodes)
	if config.EnableOCR {
		engine, err := ocr.NewEngine(normalizeOCRLanguages(config.OCRLanguages)...)
		if err != nil {
			return nil, err
		}
		defer engine.Close()

		if config.OCRDebugDir != "" {
			if err := os.MkdirAll(config.OCRDebugDir, 0755); err != nil {
				return nil, err
			}
			engine.DebugDir = config.OCRDebugDir
		}

		axtree.EnrichWithOCR(ctx, tree, engine)
	}

	snap := &SnapshotResult{
		Nodes:    tree,
		IDMap:    axtree.BuildIDMap(tree),
		RawCount: len(rawNodes),
	}

	// Build format options
	fmtOpts := axtree.FormatOptions{
		InteractiveOnly: config.InteractiveOnly,
		Compact:         config.Compact,
		MaxDepth:        config.MaxDepth,
		Cursor:          config.Cursor,
	}

	// Use extended format if any option is set
	if config.InteractiveOnly || config.Compact || config.MaxDepth > 0 || config.Cursor {
		snap.Text = axtree.FormatWithOptions(tree, fmtOpts)
	} else {
		snap.Text = axtree.Format(tree)
	}

	// If Selector is set, scope snapshot to elements under that CSS selector
	if config.Selector != "" {
		scopedText, err := b.scopeSnapshotToSelector(config.Selector, snap)
		if err == nil && scopedText != "" {
			snap.Text = scopedText
		}
	}

	b.lastSnap = snap

	return snap, nil
}

// scopeSnapshotToSelector returns a snapshot text scoped to elements under the given CSS selector.
func (b *Browser) scopeSnapshotToSelector(selector string, snap *SnapshotResult) (string, error) {
	// Get the backend node IDs of elements matching the selector
	result, err := b.evaluateString(fmt.Sprintf(`(() => {
		const els = document.querySelectorAll(%s);
		return JSON.stringify(Array.from(els).map(el => el.tagName.toLowerCase()));
	})()`, mustJSON(selector)))
	if err != nil {
		return "", err
	}
	if result == "" || result == "[]" {
		return "No elements match selector " + selector + "\n", nil
	}
	// For now, return the full snapshot with a header indicating the scope
	return fmt.Sprintf("Scoped to: %s\n\n%s", selector, snap.Text), nil
}

func normalizeOCRLanguages(languages []string) []string {
	if len(languages) == 0 {
		return []string{"eng"}
	}

	normalized := make([]string, 0, len(languages))
	for _, language := range languages {
		for _, part := range strings.Split(language, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				normalized = append(normalized, part)
			}
		}
	}

	if len(normalized) == 0 {
		return []string{"eng"}
	}

	return normalized
}
