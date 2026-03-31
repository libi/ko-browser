package axtree

import (
	"context"
	"io"

	"github.com/chromedp/cdproto/accessibility"
	internalaxtree "github.com/libi/ko-browser/internal/axtree"
	"github.com/libi/ko-browser/ocr"
)

type Node = internalaxtree.Node
type FormatOptions = internalaxtree.FormatOptions

func Extract(ctx context.Context) ([]*accessibility.Node, error) {
	return internalaxtree.Extract(ctx)
}

func DumpRaw(w io.Writer, nodes []*accessibility.Node) {
	internalaxtree.DumpRaw(w, nodes)
}

func BuildAndFilter(rawNodes []*accessibility.Node) []*Node {
	return internalaxtree.BuildAndFilter(rawNodes)
}

func Format(nodes []*Node) string {
	return internalaxtree.Format(nodes)
}

func FormatWithOptions(nodes []*Node, opts FormatOptions) string {
	return internalaxtree.FormatWithOptions(nodes, opts)
}

func BuildIDMap(nodes []*Node) map[int]int64 {
	return internalaxtree.BuildIDMap(nodes)
}

func EnrichWithOCR(ctx context.Context, nodes []*Node, engine *ocr.Engine) {
	internalaxtree.EnrichWithOCR(ctx, nodes, engine)
}

func Count(nodes []*Node) int {
	return internalaxtree.Count(nodes)
}
