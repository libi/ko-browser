package axtree

// Node is a simplified, filtered accessibility tree node.
// It carries only the information needed for LLM consumption.
type Node struct {
	Role      string   // semantic role: button, link, heading, textbox, etc.
	Name      string   // accessible name (label)
	Value     string   // current value (e.g. text in an input field)
	States    []string // active states: focused, checked, unchecked, disabled, etc.
	Level     int      // heading level (1-6), 0 if not applicable
	BackendID int64    // backendDOMNodeId — for future CDP click/type operations
	Children  []*Node
}
