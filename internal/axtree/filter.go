package axtree

import (
	"strings"

	"github.com/chromedp/cdproto/accessibility"
)

// --- Role classification maps ---

// interactiveRoles are always kept and always get an ID in the output.
var interactiveRoles = map[string]bool{
	"button": true, "link": true, "textbox": true, "searchbox": true,
	"checkbox": true, "radio": true, "combobox": true, "listbox": true,
	"menuitem": true, "menuitemcheckbox": true, "menuitemradio": true,
	"option": true, "switch": true, "slider": true, "spinbutton": true,
	"tab": true, "treeitem": true, "scrollbar": true,
}

// structuralRoles provide semantic structure — kept when they have name or children.
var structuralRoles = map[string]bool{
	"heading": true, "navigation": true, "main": true, "banner": true,
	"complementary": true, "contentinfo": true, "form": true, "search": true,
	"dialog": true, "alertdialog": true, "alert": true,
	"menu": true, "menubar": true, "tablist": true, "tabpanel": true,
	"toolbar": true, "list": true, "listitem": true,
	"table": true, "row": true, "cell": true, "columnheader": true,
	"rowheader": true, "grid": true, "gridcell": true,
	"tree": true, "treegrid": true, "region": true,
	"article": true, "figure": true, "img": true, "separator": true,
	"progressbar": true, "status": true, "tooltip": true,
	"group": true, "paragraph": true, "blockquote": true,
	"section": true,
}

// skipRoles are always transparent — their children pass through to the parent.
var skipRoles = map[string]bool{
	"none": true, "presentation": true, "inlinetextbox": true,
	"linebreak": true, "abbr": true, "ruby": true,
}

// rootRoles are document-level containers, flattened in the output.
var rootRoles = map[string]bool{
	"rootwebarea": true, "webarea": true, "document": true,
}

// BuildAndFilter converts a flat CDP AX node list into a filtered, hierarchical tree.
// This is the core of Layer 2 — deterministic local processing, zero LLM tokens.
func BuildAndFilter(rawNodes []*accessibility.Node) []*Node {
	if len(rawNodes) == 0 {
		return nil
	}

	// Build lookup index: NodeID → *accessibility.Node
	index := make(map[accessibility.NodeID]*accessibility.Node, len(rawNodes))
	for _, n := range rawNodes {
		index[n.NodeID] = n
	}

	// Find root node (the one with no parent)
	var root *accessibility.Node
	for _, n := range rawNodes {
		if n.ParentID == "" {
			root = n
			break
		}
	}
	if root == nil {
		root = rawNodes[0]
	}

	// Recursively convert and filter
	result := convertNodes(root, index)

	// Post-processing passes
	// Pass 1: remove text children that duplicate parent name
	for _, n := range result {
		removeRedundantText(n)
	}
	// Pass 2: clean up icons, merge fragments, unwrap transparent wrappers
	result = cleanTree(result)

	return result
}

// convertNodes recursively converts a CDP AX node into our simplified Node(s).
// Returns a slice because an ignored/transparent node may produce 0 or N children.
func convertNodes(raw *accessibility.Node, index map[accessibility.NodeID]*accessibility.Node) []*Node {
	if raw == nil {
		return nil
	}

	// Always process children first (needed for keep/discard decisions)
	var children []*Node
	for _, childID := range raw.ChildIDs {
		if child, ok := index[childID]; ok {
			children = append(children, convertNodes(child, index)...)
		}
	}

	// Ignored nodes are transparent — pass through their children
	if raw.Ignored {
		return children
	}

	role := valStr(raw.Role)
	name := valStr(raw.Name)
	roleLower := strings.ToLower(role)

	// Decide whether to keep this node
	keep := false
	switch {
	case rootRoles[roleLower]:
		keep = true
	case interactiveRoles[roleLower]:
		keep = true
	case structuralRoles[roleLower]:
		keep = name != "" || len(children) > 0
	case skipRoles[roleLower]:
		keep = false
	case roleLower == "generic" || roleLower == "":
		keep = name != "" // generic with a name is meaningful
	case roleLower == "statictext" || roleLower == "text":
		keep = name != ""
		if keep {
			role = "text" // normalize display role
		}
	default:
		keep = name != "" || len(children) > 0
	}

	if !keep {
		return children // transparent: pass children up
	}

	// Build our simplified node
	node := &Node{
		Role:      role,
		Name:      name,
		Value:     valStr(raw.Value),
		BackendID: int64(raw.BackendDOMNodeID),
		Children:  children,
	}
	extractProperties(node, raw.Properties)

	return []*Node{node}
}

// extractProperties reads AX properties into our Node's States and Level fields.
func extractProperties(node *Node, props []*accessibility.Property) {
	for _, prop := range props {
		switch string(prop.Name) {
		case "focused":
			if b, ok := valBool(prop.Value); ok && b {
				node.States = append(node.States, "focused")
			}
		case "checked":
			if s := valTristate(prop.Value); s != "" {
				node.States = append(node.States, s)
			}
		case "disabled":
			if b, ok := valBool(prop.Value); ok && b {
				node.States = append(node.States, "disabled")
			}
		case "expanded":
			if b, ok := valBool(prop.Value); ok {
				if b {
					node.States = append(node.States, "expanded")
				} else {
					node.States = append(node.States, "collapsed")
				}
			}
		case "selected":
			if b, ok := valBool(prop.Value); ok && b {
				node.States = append(node.States, "selected")
			}
		case "required":
			if b, ok := valBool(prop.Value); ok && b {
				node.States = append(node.States, "required")
			}
		case "readonly":
			if b, ok := valBool(prop.Value); ok && b {
				node.States = append(node.States, "readonly")
			}
		case "level":
			if l, ok := valInt(prop.Value); ok {
				node.Level = l
			}
		}
	}
}

// removeRedundantText removes text children that simply duplicate their parent's name.
// e.g. button "Submit" → StaticText "Submit" is redundant.
func removeRedundantText(node *Node) {
	if node == nil {
		return
	}
	var filtered []*Node
	for _, child := range node.Children {
		roleLower := strings.ToLower(child.Role)
		// Skip text nodes that just repeat parent name
		if (roleLower == "text" || roleLower == "statictext") &&
			child.Name == node.Name && child.Name != "" && len(child.Children) == 0 {
			continue
		}
		removeRedundantText(child)
		filtered = append(filtered, child)
	}
	node.Children = filtered
}

// cleanTree applies multiple simplification passes to the tree.
func cleanTree(nodes []*Node) []*Node {
	var result []*Node
	for _, n := range nodes {
		n = cleanNode(n)
		if n != nil {
			result = append(result, n)
		}
	}
	return result
}

func cleanNode(node *Node) *Node {
	if node == nil {
		return nil
	}

	// Recursively clean children first
	node.Children = cleanTree(node.Children)

	roleLower := strings.ToLower(node.Role)

	// Remove pure-icon text nodes (Unicode private use area: E000-F8FF)
	if (roleLower == "text" || roleLower == "statictext") && isPureIcon(node.Name) {
		return nil
	}

	// Remove image/img nodes with no name (decorative)
	if (roleLower == "img" || roleLower == "image") && node.Name == "" && len(node.Children) == 0 {
		return nil
	}

	// Unwrap single-child structural containers that add no info
	// e.g. paragraph > link "About" → just link "About"
	if isTransparentWrapper(node) {
		return node.Children[0]
	}

	// Merge fragmented text children under interactive elements
	// e.g. link [text "5", text "some news"] → link "5 some news"
	if interactiveRoles[roleLower] && allChildrenAreText(node.Children) && len(node.Children) > 0 {
		if node.Name == "" {
			var parts []string
			for _, c := range node.Children {
				if c.Name != "" && !isPureIcon(c.Name) {
					parts = append(parts, c.Name)
				}
			}
			node.Name = strings.Join(parts, " ")
		}
		// Remove text children that are now merged into name, keep non-text children
		var keep []*Node
		for _, c := range node.Children {
			cr := strings.ToLower(c.Role)
			if cr != "text" && cr != "statictext" {
				keep = append(keep, c)
			}
		}
		node.Children = keep
	}

	// Remove generic wrapper with same name as single child
	if roleLower == "generic" && len(node.Children) == 1 && node.Name == node.Children[0].Name {
		return node.Children[0]
	}

	// Strip embedded private-use-area chars from names (icon fonts mixed with text)
	node.Name = stripIconChars(node.Name)

	// After stripping, if a non-interactive node has no name and no children, discard
	if !interactiveRoles[roleLower] && node.Name == "" && len(node.Children) == 0 && node.Value == "" {
		// Keep structural roles that might still be useful (e.g. separator)
		if roleLower != "separator" && roleLower != "img" {
			return nil
		}
	}

	return node
}

// isPureIcon checks if a string consists entirely of Unicode private use area chars or icon-like chars.
func isPureIcon(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		// Private Use Area: U+E000 to U+F8FF
		// Supplementary PUA: U+F0000 to U+FFFFD, U+100000 to U+10FFFD
		if !((r >= 0xE000 && r <= 0xF8FF) ||
			(r >= 0xF0000 && r <= 0xFFFFD) ||
			(r >= 0x100000 && r <= 0x10FFFD)) {
			return false
		}
	}
	return true
}

// isPrivateUseChar checks if a rune is in a Unicode Private Use Area.
func isPrivateUseChar(r rune) bool {
	return (r >= 0xE000 && r <= 0xF8FF) ||
		(r >= 0xF0000 && r <= 0xFFFFD) ||
		(r >= 0x100000 && r <= 0x10FFFD)
}

// stripIconChars removes private-use-area characters from a string and trims whitespace.
func stripIconChars(s string) string {
	var buf strings.Builder
	for _, r := range s {
		if !isPrivateUseChar(r) {
			buf.WriteRune(r)
		}
	}
	return strings.TrimSpace(buf.String())
}

// isTransparentWrapper checks if a node is a meaningless wrapper that can be unwrapped.
func isTransparentWrapper(node *Node) bool {
	if len(node.Children) != 1 || len(node.States) > 0 {
		return false
	}
	roleLower := strings.ToLower(node.Role)
	// paragraph, group, section with no name wrapping a single meaningful child
	if (roleLower == "paragraph" || roleLower == "group" || roleLower == "section") && node.Name == "" {
		return true
	}
	return false
}

// allChildrenAreText checks if all children are text/statictext leaf nodes.
func allChildrenAreText(children []*Node) bool {
	for _, c := range children {
		roleLower := strings.ToLower(c.Role)
		if roleLower != "text" && roleLower != "statictext" {
			return false
		}
		if len(c.Children) > 0 {
			return false
		}
	}
	return true
}

// Count returns the total number of displayable nodes in the tree.
// Root document nodes are excluded from the count (they are flattened).
func Count(nodes []*Node) int {
	count := 0
	for _, n := range nodes {
		roleLower := strings.ToLower(n.Role)
		if rootRoles[roleLower] {
			count += Count(n.Children)
		} else {
			count++
			count += Count(n.Children)
		}
	}
	return count
}
