package browser

import (
	"fmt"
	"strings"

	"github.com/libi/ko-browser/axtree"
)

const interactiveQuery = `input, textarea, select, button, a[href], [role="button"], [role="link"], [role="checkbox"], [role="radio"], [role="switch"], [tabindex]:not([tabindex="-1"])`

func (b *Browser) interactiveOrdinal(id int) (int, error) {
	if b.lastSnap == nil {
		return 0, fmt.Errorf("no snapshot yet, call Snapshot() first")
	}

	displayID := 0
	ordinal := -1
	targetOrdinal := -1

	var walk func(nodes []*axtree.Node) bool
	walk = func(nodes []*axtree.Node) bool {
		for _, node := range nodes {
			if isRootRole(node.Role) {
				if walk(node.Children) {
					return true
				}
				continue
			}

			displayID++
			interactive := isInteractiveRole(node.Role)
			if interactive {
				ordinal++
			}
			if displayID == id {
				if interactive {
					targetOrdinal = ordinal
				}
				return true
			}
			if walk(node.Children) {
				return true
			}
		}
		return false
	}

	if !walk(b.lastSnap.Nodes) {
		return 0, fmt.Errorf("element %d not found", id)
	}
	if targetOrdinal < 0 {
		return 0, fmt.Errorf("element %d is not interactive", id)
	}
	return targetOrdinal, nil
}

func isInteractiveRole(role string) bool {
	switch strings.ToLower(role) {
	case "button", "link", "textbox", "searchbox", "checkbox", "radio", "combobox", "listbox", "menuitem", "menuitemcheckbox", "menuitemradio", "option", "switch", "slider", "spinbutton", "tab", "treeitem", "scrollbar":
		return true
	default:
		return false
	}
}

func isRootRole(role string) bool {
	switch strings.ToLower(role) {
	case "rootwebarea", "webarea", "document":
		return true
	default:
		return false
	}
}

func (b *Browser) evaluateOnInteractiveElement(id int, body string) error {
	ordinal, err := b.interactiveOrdinal(id)
	if err != nil {
		return err
	}

	expression := `(() => {
		const elements = Array.from(document.querySelectorAll(` + mustJSON(interactiveQuery) + `));
		const el = elements[` + mustJSON(ordinal) + `];
		if (!el) {
			throw new Error('interactive element not found');
		}
		` + body + `
	})()`

	return b.evaluate(expression)
}
