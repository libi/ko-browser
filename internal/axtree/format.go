package axtree

import (
	"fmt"
	"strings"
)

// Format renders the filtered tree as structured text suitable for LLM consumption.
// Output format:
//
//	1: heading "Page Title"
//	2: button "Submit" disabled
//	3: navigation "Main Nav"
//	  4: link "Home"
//	  5: link "About"
//	6: textbox "Search" focused
func Format(nodes []*Node) string {
	var buf strings.Builder
	counter := 0
	for _, node := range nodes {
		formatNode(&buf, node, 0, &counter)
	}
	return buf.String()
}

func formatNode(buf *strings.Builder, node *Node, depth int, counter *int) {
	if node == nil {
		return
	}

	roleLower := strings.ToLower(node.Role)

	// Root document nodes: print title line, then flatten children at same depth
	if rootRoles[roleLower] {
		if node.Name != "" {
			buf.WriteString(fmt.Sprintf("Page: %q\n\n", node.Name))
		}
		for _, child := range node.Children {
			formatNode(buf, child, depth, counter)
		}
		return
	}

	*counter++

	// Indentation (2 spaces per level)
	indent := strings.Repeat("  ", depth)
	buf.WriteString(indent)

	// id: role
	buf.WriteString(fmt.Sprintf("%d: %s", *counter, node.Role))

	// "name" (quoted)
	if node.Name != "" {
		name := truncate(node.Name, 80)
		buf.WriteString(fmt.Sprintf(" %q", name))
	}

	// value="..." (only if different from name, useful for form fields)
	if node.Value != "" && node.Value != node.Name {
		val := truncate(node.Value, 50)
		buf.WriteString(fmt.Sprintf(" value=%q", val))
	}

	// States: focused, checked, disabled, etc.
	for _, s := range node.States {
		buf.WriteString(" " + s)
	}

	buf.WriteByte('\n')

	// Recurse into children
	for _, child := range node.Children {
		formatNode(buf, child, depth+1, counter)
	}
}

// FormatOptions controls the formatting behavior.
type FormatOptions struct {
	InteractiveOnly bool // only show interactive elements
	Compact         bool // compact mode: omit structural wrappers without names
	MaxDepth        int  // 0 = unlimited
	Cursor          bool // annotate the focused element with [cursor]
}

// FormatWithOptions renders the filtered tree with additional formatting controls.
func FormatWithOptions(nodes []*Node, opts FormatOptions) string {
	var buf strings.Builder
	counter := 0
	for _, node := range nodes {
		formatNodeWithOptions(&buf, node, 0, &counter, &opts)
	}
	return buf.String()
}

func formatNodeWithOptions(buf *strings.Builder, node *Node, depth int, counter *int, opts *FormatOptions) {
	if node == nil {
		return
	}

	roleLower := strings.ToLower(node.Role)

	// Root document nodes: print title line, then flatten children at same depth
	if rootRoles[roleLower] {
		if node.Name != "" {
			buf.WriteString(fmt.Sprintf("Page: %q\n\n", node.Name))
		}
		for _, child := range node.Children {
			formatNodeWithOptions(buf, child, depth, counter, opts)
		}
		return
	}

	isInteractive := interactiveRoles[roleLower]

	// InteractiveOnly: skip non-interactive nodes but still recurse their children.
	// We still increment the counter to keep IDs consistent with BuildIDMap,
	// and use depth+1 to preserve the tree hierarchy.
	if opts.InteractiveOnly && !isInteractive {
		*counter++
		for _, child := range node.Children {
			formatNodeWithOptions(buf, child, depth+1, counter, opts)
		}
		return
	}

	// Compact: skip structural wrappers without names that have children.
	// We still increment the counter to keep IDs consistent with BuildIDMap.
	if opts.Compact && !isInteractive && node.Name == "" && len(node.Children) > 0 {
		*counter++
		for _, child := range node.Children {
			formatNodeWithOptions(buf, child, depth, counter, opts)
		}
		return
	}

	*counter++

	// MaxDepth: if we've exceeded the limit, still count children but don't print.
	// We must recurse to keep counter consistent with BuildIDMap.
	if opts.MaxDepth > 0 && depth >= opts.MaxDepth {
		for _, child := range node.Children {
			countNodeOnly(child, counter)
		}
		return
	}

	// Indentation (2 spaces per level)
	indent := strings.Repeat("  ", depth)
	buf.WriteString(indent)

	// id: role
	buf.WriteString(fmt.Sprintf("%d: %s", *counter, node.Role))

	// "name" (quoted)
	if node.Name != "" {
		name := truncate(node.Name, 80)
		buf.WriteString(fmt.Sprintf(" %q", name))
	}

	// value="..." (only if different from name, useful for form fields)
	if node.Value != "" && node.Value != node.Name {
		val := truncate(node.Value, 50)
		buf.WriteString(fmt.Sprintf(" value=%q", val))
	}

	// States: focused, checked, disabled, etc.
	for _, s := range node.States {
		buf.WriteString(" " + s)
	}

	// Cursor annotation: mark focused element
	if opts.Cursor {
		for _, s := range node.States {
			if s == "focused" {
				buf.WriteString(" ←[cursor]")
				break
			}
		}
	}

	buf.WriteByte('\n')

	// Recurse into children
	for _, child := range node.Children {
		formatNodeWithOptions(buf, child, depth+1, counter, opts)
	}
}

// countNodeOnly increments the counter for a node and all its descendants
// without producing any output. Used to keep IDs consistent when nodes are
// hidden by MaxDepth or other filters.
func countNodeOnly(node *Node, counter *int) {
	if node == nil {
		return
	}
	roleLower := strings.ToLower(node.Role)
	if rootRoles[roleLower] {
		for _, child := range node.Children {
			countNodeOnly(child, counter)
		}
		return
	}
	*counter++
	for _, child := range node.Children {
		countNodeOnly(child, counter)
	}
}

// truncate shortens a string to maxLen runes, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
