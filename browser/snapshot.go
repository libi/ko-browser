package browser

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
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
	totalRawCount := len(rawNodes)
	if err := b.attachIframeContents(ctx, tree, &totalRawCount); err != nil {
		return nil, err
	}
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
		RawCount: totalRawCount,
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

func (b *Browser) attachIframeContents(ctx context.Context, tree []*axtree.Node, totalRawCount *int) error {
	var frameTree *page.FrameTree
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		frameTree, err = page.GetFrameTree().Do(ctx)
		return err
	})); err != nil {
		return err
	}
	if frameTree == nil {
		return nil
	}
	return b.attachChildFrameContents(ctx, tree, frameTree, totalRawCount)
}

func (b *Browser) attachChildFrameContents(ctx context.Context, tree []*axtree.Node, frameTree *page.FrameTree, totalRawCount *int) error {
	if frameTree == nil {
		return nil
	}

	for _, childFrame := range frameTree.ChildFrames {
		if childFrame == nil || childFrame.Frame == nil {
			continue
		}

		var ownerBackendID int64
		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			backendNodeID, _, err := dom.GetFrameOwner(childFrame.Frame.ID).Do(ctx)
			ownerBackendID = int64(backendNodeID)
			return err
		})); err != nil {
			return err
		}

		iframeNode := findNodeByBackendID(tree, ownerBackendID)
		if iframeNode == nil {
			continue
		}

		rawNodes, err := axtree.ExtractFrame(ctx, childFrame.Frame.ID)
		if err != nil {
			return err
		}
		*totalRawCount += len(rawNodes)

		childTree := axtree.BuildAndFilter(rawNodes)
		if err := b.attachChildFrameContents(ctx, childTree, childFrame, totalRawCount); err != nil {
			return err
		}

		iframeNode.Children = append(iframeNode.Children, flattenDocumentRoots(childTree)...)
	}

	return nil
}

func findNodeByBackendID(nodes []*axtree.Node, backendID int64) *axtree.Node {
	for _, node := range nodes {
		if found := findNodeByBackendIDRecursive(node, backendID); found != nil {
			return found
		}
	}
	return nil
}

func findNodeByBackendIDRecursive(node *axtree.Node, backendID int64) *axtree.Node {
	if node == nil {
		return nil
	}
	if node.BackendID == backendID {
		return node
	}
	for _, child := range node.Children {
		if found := findNodeByBackendIDRecursive(child, backendID); found != nil {
			return found
		}
	}
	return nil
}

func flattenDocumentRoots(nodes []*axtree.Node) []*axtree.Node {
	var flattened []*axtree.Node
	for _, node := range nodes {
		if node == nil {
			continue
		}
		switch strings.ToLower(node.Role) {
		case "rootwebarea", "webarea", "document":
			flattened = append(flattened, flattenDocumentRoots(node.Children)...)
		default:
			flattened = append(flattened, node)
		}
	}
	return flattened
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
