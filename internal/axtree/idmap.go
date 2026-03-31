package axtree

import "strings"

// BuildIDMap traverses the tree using the same counter logic as Format(),
// returning a map from display ID (1, 2, ...) to BackendDOMNodeID.
// This is essential for translating user-facing element numbers into
// CDP-level identifiers for actions like click/type.
func BuildIDMap(nodes []*Node) map[int]int64 {
	m := make(map[int]int64)
	counter := 0
	for _, node := range nodes {
		buildIDMapNode(node, &counter, m)
	}
	return m
}

func buildIDMapNode(node *Node, counter *int, m map[int]int64) {
	if node == nil {
		return
	}

	roleLower := strings.ToLower(node.Role)

	// Root document nodes: skip counter (same logic as formatNode)
	if rootRoles[roleLower] {
		for _, child := range node.Children {
			buildIDMapNode(child, counter, m)
		}
		return
	}

	*counter++
	m[*counter] = node.BackendID

	for _, child := range node.Children {
		buildIDMapNode(child, counter, m)
	}
}
