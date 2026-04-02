package axtree

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// buildTestTree creates a realistic tree for testing:
//
//	RootWebArea "Test Page"
//	  heading "Welcome"                  (structural, named)
//	  navigation "Main Nav"              (structural, named)
//	    list                             (structural, unnamed)
//	      listitem                       (structural, unnamed)
//	        link "Home"                  (interactive)
//	      listitem                       (structural, unnamed)
//	        link "About"                 (interactive)
//	  form "Login"                       (structural, named)
//	    textbox "Email"                  (interactive)
//	    textbox "Password"               (interactive)
//	    button "Submit"                  (interactive)
//	  paragraph "Some text"              (structural, named)
//	  group                              (structural, unnamed)
//	    checkbox "Remember me"           (interactive)
//	    link "Forgot password"           (interactive)
func buildTestTree() []*Node {
	return []*Node{
		{
			Role:      "RootWebArea",
			Name:      "Test Page",
			BackendID: 100,
			Children: []*Node{
				{
					Role: "heading", Name: "Welcome", BackendID: 101,
				},
				{
					Role: "navigation", Name: "Main Nav", BackendID: 102,
					Children: []*Node{
						{
							Role: "list", BackendID: 103,
							Children: []*Node{
								{
									Role: "listitem", BackendID: 104,
									Children: []*Node{
										{Role: "link", Name: "Home", BackendID: 105},
									},
								},
								{
									Role: "listitem", BackendID: 106,
									Children: []*Node{
										{Role: "link", Name: "About", BackendID: 107},
									},
								},
							},
						},
					},
				},
				{
					Role: "form", Name: "Login", BackendID: 108,
					Children: []*Node{
						{Role: "textbox", Name: "Email", BackendID: 109},
						{Role: "textbox", Name: "Password", BackendID: 110},
						{Role: "button", Name: "Submit", BackendID: 111},
					},
				},
				{
					Role: "paragraph", Name: "Some text", BackendID: 112,
				},
				{
					Role: "group", BackendID: 113,
					Children: []*Node{
						{Role: "checkbox", Name: "Remember me", BackendID: 114},
						{Role: "link", Name: "Forgot password", BackendID: 115},
					},
				},
			},
		},
	}
}

// extractIDsFromText parses "N: role ..." lines and returns a map[id]role.
func extractIDsFromText(text string) map[int]string {
	re := regexp.MustCompile(`(?m)^\s*(\d+):\s+(\S+)`)
	result := make(map[int]string)
	for _, match := range re.FindAllStringSubmatch(text, -1) {
		id, _ := strconv.Atoi(match[1])
		result[id] = match[2]
	}
	return result
}

// extractIDIndents parses lines and returns a map[id]indentLevel (number of 2-space indents).
func extractIDIndents(text string) map[int]int {
	re := regexp.MustCompile(`(?m)^( *)(\d+):`)
	result := make(map[int]int)
	for _, match := range re.FindAllStringSubmatch(text, -1) {
		indent := len(match[1]) / 2
		id, _ := strconv.Atoi(match[2])
		result[id] = indent
	}
	return result
}

// extractOrderedIDs returns element IDs in the order they appear in the text.
func extractOrderedIDs(text string) []int {
	re := regexp.MustCompile(`(?m)^\s*(\d+):`)
	var ids []int
	for _, match := range re.FindAllStringSubmatch(text, -1) {
		id, _ := strconv.Atoi(match[1])
		ids = append(ids, id)
	}
	return ids
}

func findNodeByBackendID(nodes []*Node, backendID int64) *Node {
	for _, n := range nodes {
		if found := findNodeByBackendIDRecursive(n, backendID); found != nil {
			return found
		}
	}
	return nil
}

func findNodeByBackendIDRecursive(node *Node, backendID int64) *Node {
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

// --- Tests ---

func TestFormat_IDConsistencyWithBuildIDMap(t *testing.T) {
	tree := buildTestTree()
	idMap := BuildIDMap(tree)
	text := Format(tree)
	textIDs := extractIDsFromText(text)

	// Every ID in formatted text must exist in IDMap
	for id := range textIDs {
		if _, ok := idMap[id]; !ok {
			t.Errorf("Format: ID %d appears in text but not in BuildIDMap", id)
		}
	}
	// Every ID in IDMap must appear in formatted text (base Format shows all)
	for id := range idMap {
		if _, ok := textIDs[id]; !ok {
			t.Errorf("Format: ID %d is in BuildIDMap but not in formatted text", id)
		}
	}
}

func TestFormatWithOptions_IDConsistencyAcrossAllModes(t *testing.T) {
	tree := buildTestTree()
	idMap := BuildIDMap(tree)

	modes := []struct {
		name string
		opts FormatOptions
	}{
		{"default", FormatOptions{}},
		{"interactive-only", FormatOptions{InteractiveOnly: true}},
		{"compact", FormatOptions{Compact: true}},
		{"cursor", FormatOptions{Cursor: true}},
		{"max-depth-2", FormatOptions{MaxDepth: 2}},
		{"interactive+compact", FormatOptions{InteractiveOnly: true, Compact: true}},
		{"interactive+cursor", FormatOptions{InteractiveOnly: true, Cursor: true}},
		{"compact+cursor", FormatOptions{Compact: true, Cursor: true}},
		{"all-options", FormatOptions{InteractiveOnly: true, Compact: true, Cursor: true}},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			text := FormatWithOptions(tree, mode.opts)
			textIDs := extractIDsFromText(text)

			// Core invariant: every ID shown in formatted text must map to the
			// same BackendID as BuildIDMap.
			for id, role := range textIDs {
				backendID, ok := idMap[id]
				if !ok {
					t.Errorf("mode=%s: ID %d (%s) appears in text but not in BuildIDMap\ntext:\n%s",
						mode.name, id, role, text)
					continue
				}
				// Verify BackendID matches by finding the node with that BackendID
				expectedNode := findNodeByBackendID(tree, backendID)
				if expectedNode == nil {
					t.Errorf("mode=%s: ID %d maps to backendID=%d but no node found",
						mode.name, id, backendID)
					continue
				}
				// The role in the text must match the node's role
				if !strings.EqualFold(role, expectedNode.Role) {
					t.Errorf("mode=%s: ID %d shows role=%q but expected role=%q (backendID=%d)",
						mode.name, id, role, expectedNode.Role, backendID)
				}
			}
		})
	}
}

func TestFormatWithOptions_InteractivePreservesHierarchy(t *testing.T) {
	tree := buildTestTree()

	baseText := Format(tree)
	baseIndents := extractIDIndents(baseText)

	interactiveText := FormatWithOptions(tree, FormatOptions{InteractiveOnly: true})
	interactiveIndents := extractIDIndents(interactiveText)

	// Interactive mode elements should have indent >= base indent
	for id, iIndent := range interactiveIndents {
		bIndent, ok := baseIndents[id]
		if !ok {
			t.Errorf("interactive mode: ID %d appears but not in base format", id)
			continue
		}
		if iIndent < bIndent {
			t.Errorf("interactive mode: ID %d has indent=%d, less than base indent=%d",
				id, iIndent, bIndent)
		}
	}

	// IDs must be in strictly increasing order
	interactiveIDs := extractOrderedIDs(interactiveText)
	for i := 1; i < len(interactiveIDs); i++ {
		if interactiveIDs[i] <= interactiveIDs[i-1] {
			t.Errorf("interactive mode: IDs not in increasing order: ...%d, %d...",
				interactiveIDs[i-1], interactiveIDs[i])
		}
	}
}

func TestFormatWithOptions_InteractiveOnlyShowsInteractiveElements(t *testing.T) {
	tree := buildTestTree()
	text := FormatWithOptions(tree, FormatOptions{InteractiveOnly: true})
	textIDs := extractIDsFromText(text)

	// Every element shown must have an interactive role
	for id, role := range textIDs {
		roleLower := strings.ToLower(role)
		if !interactiveRoles[roleLower] {
			t.Errorf("interactive mode: ID %d has non-interactive role %q", id, role)
		}
	}

	// All interactive elements from the full tree must be present
	fullText := Format(tree)
	fullIDs := extractIDsFromText(fullText)
	for id, role := range fullIDs {
		if interactiveRoles[strings.ToLower(role)] {
			if _, ok := textIDs[id]; !ok {
				t.Errorf("interactive mode: missing interactive element ID %d (%s)", id, role)
			}
		}
	}
}

func TestFormatWithOptions_CompactPreservesIDs(t *testing.T) {
	tree := buildTestTree()
	idMap := BuildIDMap(tree)
	text := FormatWithOptions(tree, FormatOptions{Compact: true})
	textIDs := extractIDsFromText(text)

	for id, role := range textIDs {
		backendID, ok := idMap[id]
		if !ok {
			t.Errorf("compact mode: ID %d (%s) not found in BuildIDMap", id, role)
			continue
		}
		node := findNodeByBackendID(tree, backendID)
		if node == nil {
			t.Errorf("compact mode: ID %d maps to backendID=%d but no node found", id, backendID)
		}
	}
}

func TestFormatWithOptions_MaxDepthPreservesIDs(t *testing.T) {
	tree := buildTestTree()
	idMap := BuildIDMap(tree)

	for depth := 1; depth <= 5; depth++ {
		t.Run(fmt.Sprintf("depth-%d", depth), func(t *testing.T) {
			text := FormatWithOptions(tree, FormatOptions{MaxDepth: depth})
			textIDs := extractIDsFromText(text)

			for id, role := range textIDs {
				backendID, ok := idMap[id]
				if !ok {
					t.Errorf("max-depth=%d: ID %d (%s) not found in BuildIDMap", depth, id, role)
					continue
				}
				node := findNodeByBackendID(tree, backendID)
				if node == nil {
					t.Errorf("max-depth=%d: ID %d maps to backendID=%d but no node found", depth, id, backendID)
				}
			}

			// All shown elements should be within depth limit
			indents := extractIDIndents(text)
			for id, indent := range indents {
				if indent >= depth {
					t.Errorf("max-depth=%d: ID %d has indent=%d which exceeds max depth", depth, id, indent)
				}
			}
		})
	}
}

func TestFormatWithOptions_CursorAnnotation(t *testing.T) {
	tree := []*Node{
		{
			Role:      "RootWebArea",
			Name:      "Test",
			BackendID: 1,
			Children: []*Node{
				{Role: "textbox", Name: "Search", BackendID: 2, States: []string{"focused"}},
				{Role: "button", Name: "Go", BackendID: 3},
			},
		},
	}

	text := FormatWithOptions(tree, FormatOptions{Cursor: true})
	if !strings.Contains(text, "\u2190[cursor]") {
		t.Errorf("cursor mode: expected cursor annotation, got:\n%s", text)
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "\u2190[cursor]") && !strings.Contains(line, "Search") {
			t.Errorf("cursor should be on 'Search' textbox, got: %s", line)
		}
		if strings.Contains(line, "\"Go\"") && strings.Contains(line, "\u2190[cursor]") {
			t.Errorf("cursor should NOT be on 'Go' button, got: %s", line)
		}
	}

	// IDs still consistent
	idMap := BuildIDMap(tree)
	textIDs := extractIDsFromText(text)
	for id := range textIDs {
		if _, ok := idMap[id]; !ok {
			t.Errorf("cursor mode: ID %d not found in BuildIDMap", id)
		}
	}
}

func TestFormat_BaselineOutput(t *testing.T) {
	tree := buildTestTree()
	text := Format(tree)

	if !strings.Contains(text, `Page: "Test Page"`) {
		t.Errorf("expected page title, got:\n%s", text)
	}

	// Verify all elements present with correct IDs
	expected := map[int]string{
		1: "heading", 2: "navigation", 3: "list", 4: "listitem",
		5: "link", 6: "listitem", 7: "link", 8: "form",
		9: "textbox", 10: "textbox", 11: "button", 12: "paragraph",
		13: "group", 14: "checkbox", 15: "link",
	}

	textIDs := extractIDsFromText(text)
	for id, expectedRole := range expected {
		role, ok := textIDs[id]
		if !ok {
			t.Errorf("missing ID %d (expected %s) in output:\n%s", id, expectedRole, text)
			continue
		}
		if role != expectedRole {
			t.Errorf("ID %d: expected role=%s, got=%s", id, expectedRole, role)
		}
	}
	if len(textIDs) != len(expected) {
		t.Errorf("expected %d elements, got %d\ntext:\n%s", len(expected), len(textIDs), text)
	}
}

func TestFormatWithOptions_InteractiveBaselineOutput(t *testing.T) {
	tree := buildTestTree()
	text := FormatWithOptions(tree, FormatOptions{InteractiveOnly: true})

	// Interactive mode: only links, buttons, textboxes, checkboxes
	// but they must keep their original IDs from the full tree
	expectedIDs := map[int]string{
		5: "link", 7: "link", 9: "textbox", 10: "textbox",
		11: "button", 14: "checkbox", 15: "link",
	}

	textIDs := extractIDsFromText(text)
	for id, expectedRole := range expectedIDs {
		role, ok := textIDs[id]
		if !ok {
			t.Errorf("interactive mode: missing ID %d (expected %s)\ntext:\n%s", id, expectedRole, text)
			continue
		}
		if role != expectedRole {
			t.Errorf("interactive mode: ID %d expected role=%s, got=%s", id, expectedRole, role)
		}
	}
	for id, role := range textIDs {
		if _, ok := expectedIDs[id]; !ok {
			t.Errorf("interactive mode: unexpected ID %d (%s)", id, role)
		}
	}
}

func TestFormatWithOptions_InteractiveHierarchyOutput(t *testing.T) {
	tree := buildTestTree()
	text := FormatWithOptions(tree, FormatOptions{InteractiveOnly: true})
	indents := extractIDIndents(text)

	// Links under navigation>list>listitem should have deeper indent
	// than textbox/button under form
	linkHomeIndent := indents[5]  // link "Home": nav > list > listitem > link
	linkAboutIndent := indents[7] // link "About": nav > list > listitem > link
	emailIndent := indents[9]     // textbox "Email": form > textbox

	if linkHomeIndent <= 0 {
		t.Errorf("link 'Home' (ID=5) should have indent > 0, got %d\ntext:\n%s", linkHomeIndent, text)
	}
	if linkAboutIndent <= 0 {
		t.Errorf("link 'About' (ID=7) should have indent > 0, got %d", linkAboutIndent)
	}
	if linkHomeIndent != linkAboutIndent {
		t.Errorf("sibling links should have same indent: Home=%d, About=%d", linkHomeIndent, linkAboutIndent)
	}
	// Links nested under nav>list>listitem should be deeper than textbox under form
	if linkHomeIndent <= emailIndent {
		t.Errorf("link 'Home' (indent=%d) should be deeper than textbox 'Email' (indent=%d)",
			linkHomeIndent, emailIndent)
	}
}
