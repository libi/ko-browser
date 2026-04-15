package axtree

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/chromedp"
)

// Extract fetches the full accessibility tree from the current page via CDP.
// This is a single CDP call (Accessibility.getFullAXTree) that typically
// completes in < 50ms, returning a flat list of AX nodes.
func Extract(ctx context.Context) ([]*accessibility.Node, error) {
	var nodes []*accessibility.Node
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		nodes, err = accessibility.GetFullAXTree().Do(ctx)
		return err
	}))
	return nodes, err
}

// ExtractFrame fetches the accessibility tree for a specific frame.
func ExtractFrame(ctx context.Context, frameID cdp.FrameID) ([]*accessibility.Node, error) {
	var nodes []*accessibility.Node
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		nodes, err = accessibility.GetFullAXTree().WithFrameID(frameID).Do(ctx)
		return err
	}))
	return nodes, err
}

// DumpRaw prints the raw AX node list for debugging purposes.
func DumpRaw(w io.Writer, nodes []*accessibility.Node) {
	for i, n := range nodes {
		role := valStr(n.Role)
		name := valStr(n.Name)
		if n.Ignored && name == "" && role == "" {
			continue
		}
		fmt.Fprintf(w, "[%d] id=%s role=%q name=%q ignored=%v children=%d backendID=%d\n",
			i, n.NodeID, role, name, n.Ignored, len(n.ChildIDs), n.BackendDOMNodeID)
	}
}

// --- Value extraction helpers ---
// CDP AX values use jsontext.Value ([]byte of raw JSON).
// These helpers safely unmarshal them into Go types.

// valStr extracts a Go string from an accessibility.Value.
func valStr(v *accessibility.Value) string {
	if v == nil {
		return ""
	}
	raw := []byte(v.Value)
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// Fallback: return raw bytes as string (handles unquoted values)
	return string(raw)
}

// valBool extracts a boolean from an accessibility.Value.
func valBool(v *accessibility.Value) (val bool, ok bool) {
	if v == nil {
		return false, false
	}
	raw := []byte(v.Value)
	if len(raw) == 0 {
		return false, false
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b, true
	}
	return false, false
}

// valInt extracts an integer from an accessibility.Value.
func valInt(v *accessibility.Value) (val int, ok bool) {
	if v == nil {
		return 0, false
	}
	raw := []byte(v.Value)
	if len(raw) == 0 {
		return 0, false
	}
	var i int
	if err := json.Unmarshal(raw, &i); err == nil {
		return i, true
	}
	return 0, false
}

// valTristate extracts a tristate value as "checked"/"unchecked"/"mixed".
// CDP represents checked state as tristate: "true", "false", or "mixed".
func valTristate(v *accessibility.Value) string {
	if v == nil {
		return ""
	}
	s := valStr(v)
	switch s {
	case "true":
		return "checked"
	case "false":
		return "unchecked"
	case "mixed":
		return "mixed"
	}
	// Fallback: try as boolean
	if b, ok := valBool(v); ok {
		if b {
			return "checked"
		}
		return "unchecked"
	}
	return ""
}
