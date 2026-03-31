package browser

import (
	"fmt"
	"strings"

	"github.com/libi/ko-browser/axtree"
)

// FindResult represents a found element in the accessibility tree.
type FindResult struct {
	ID   int    // display ID from the snapshot
	Role string // semantic role
	Name string // accessible name
}

// FindResults holds the results of a find operation along with a formatted text view.
type FindResults struct {
	Items []FindResult
	Text  string // formatted text for CLI output
}

// FindRole searches the latest snapshot for elements with the given role.
// If name is non-empty, only elements whose name contains name (case-insensitive) are returned.
// If exact is true, the name must match exactly.
func (b *Browser) FindRole(role string, name string, opts ...FindOption) (*FindResults, error) {
	if b.lastSnap == nil {
		return nil, fmt.Errorf("no snapshot yet, call Snapshot() first")
	}

	cfg := parseFindOptions(opts)
	roleLower := strings.ToLower(role)
	nameLower := strings.ToLower(name)

	var items []FindResult
	walkTree(b.lastSnap.Nodes, func(displayID int, node *axtree.Node) {
		if strings.ToLower(node.Role) != roleLower {
			return
		}
		if name != "" {
			if cfg.Exact {
				if node.Name != name {
					return
				}
			} else if !strings.Contains(strings.ToLower(node.Name), nameLower) {
				return
			}
		}
		items = append(items, FindResult{ID: displayID, Role: node.Role, Name: node.Name})
	})

	return formatFindResults(items), nil
}

// FindText searches the latest snapshot for elements whose name contains the given text.
func (b *Browser) FindText(text string, opts ...FindOption) (*FindResults, error) {
	if b.lastSnap == nil {
		return nil, fmt.Errorf("no snapshot yet, call Snapshot() first")
	}

	cfg := parseFindOptions(opts)
	textLower := strings.ToLower(text)

	var items []FindResult
	walkTree(b.lastSnap.Nodes, func(displayID int, node *axtree.Node) {
		if cfg.Exact {
			if node.Name != text {
				return
			}
		} else if !strings.Contains(strings.ToLower(node.Name), textLower) {
			return
		}
		items = append(items, FindResult{ID: displayID, Role: node.Role, Name: node.Name})
	})

	return formatFindResults(items), nil
}

// FindLabel searches for form elements associated with the given label text.
func (b *Browser) FindLabel(label string, opts ...FindOption) (*FindResults, error) {
	if b.lastSnap == nil {
		return nil, fmt.Errorf("no snapshot yet, call Snapshot() first")
	}

	cfg := parseFindOptions(opts)
	labelLower := strings.ToLower(label)

	var items []FindResult
	walkTree(b.lastSnap.Nodes, func(displayID int, node *axtree.Node) {
		if !isFormRole(node.Role) {
			return
		}
		if cfg.Exact {
			if node.Name != label {
				return
			}
		} else if !strings.Contains(strings.ToLower(node.Name), labelLower) {
			return
		}
		items = append(items, FindResult{ID: displayID, Role: node.Role, Name: node.Name})
	})

	return formatFindResults(items), nil
}

// FindPlaceholder searches for elements by placeholder attribute via DOM query.
func (b *Browser) FindPlaceholder(text string, opts ...FindOption) (*FindResults, error) {
	cfg := parseFindOptions(opts)

	jsCode := fmt.Sprintf(`(() => {
		const searchText = %s;
		const exact = %t;
		const els = document.querySelectorAll('[placeholder]');
		const results = [];
		for (let i = 0; i < els.length; i++) {
			const el = els[i];
			const ph = el.getAttribute('placeholder') || '';
			const match = exact
				? ph === searchText
				: ph.toLowerCase().includes(searchText.toLowerCase());
			if (match) {
				results.push({
					tag: el.tagName.toLowerCase(),
					placeholder: ph,
					name: el.name || el.id || '',
					type: el.type || ''
				});
			}
		}
		return JSON.stringify(results);
	})()`, mustJSON(text), cfg.Exact)

	result, err := b.evaluateString(jsCode)
	if err != nil {
		return nil, fmt.Errorf("find placeholder: %w", err)
	}

	if result == "" || result == "[]" {
		return &FindResults{Text: "No elements found\n"}, nil
	}

	return &FindResults{
		Items: []FindResult{{Role: "placeholder", Name: result}},
		Text:  result + "\n",
	}, nil
}

// FindAlt searches for elements (typically images) whose name matches the alt text.
func (b *Browser) FindAlt(text string, opts ...FindOption) (*FindResults, error) {
	if b.lastSnap == nil {
		return nil, fmt.Errorf("no snapshot yet, call Snapshot() first")
	}

	cfg := parseFindOptions(opts)
	textLower := strings.ToLower(text)

	var items []FindResult
	walkTree(b.lastSnap.Nodes, func(displayID int, node *axtree.Node) {
		role := strings.ToLower(node.Role)
		if role != "image" && role != "img" && role != "figure" {
			return
		}
		if cfg.Exact {
			if node.Name != text {
				return
			}
		} else if !strings.Contains(strings.ToLower(node.Name), textLower) {
			return
		}
		items = append(items, FindResult{ID: displayID, Role: node.Role, Name: node.Name})
	})

	return formatFindResults(items), nil
}

// FindTitle searches for elements by title attribute via DOM query.
func (b *Browser) FindTitle(text string, opts ...FindOption) (*FindResults, error) {
	cfg := parseFindOptions(opts)

	jsCode := fmt.Sprintf(`(() => {
		const searchText = %s;
		const exact = %t;
		const els = document.querySelectorAll('[title]');
		const results = [];
		for (let i = 0; i < els.length; i++) {
			const el = els[i];
			const title = el.getAttribute('title') || '';
			const match = exact
				? title === searchText
				: title.toLowerCase().includes(searchText.toLowerCase());
			if (match) {
				results.push({
					tag: el.tagName.toLowerCase(),
					title: title,
					text: (el.innerText || el.textContent || '').substring(0, 80)
				});
			}
		}
		return JSON.stringify(results);
	})()`, mustJSON(text), cfg.Exact)

	result, err := b.evaluateString(jsCode)
	if err != nil {
		return nil, fmt.Errorf("find title: %w", err)
	}

	if result == "" || result == "[]" {
		return &FindResults{Text: "No elements found\n"}, nil
	}

	return &FindResults{
		Items: []FindResult{{Role: "title", Name: result}},
		Text:  result + "\n",
	}, nil
}

// FindTestID searches for elements by data-testid attribute.
// This operates on the DOM, not the snapshot.
func (b *Browser) FindTestID(testID string) (*FindResults, error) {
	result, err := b.evaluateString(fmt.Sprintf(`(() => {
		const els = document.querySelectorAll('[data-testid=%s]');
		if (els.length === 0) return '';
		const results = [];
		for (let i = 0; i < els.length; i++) {
			const el = els[i];
			results.push({
				tag: el.tagName.toLowerCase(),
				text: (el.innerText || el.textContent || '').substring(0, 80),
				id: el.id || '',
				testid: el.getAttribute('data-testid')
			});
		}
		return JSON.stringify(results);
	})()`, mustJSON(testID)))
	if err != nil {
		return nil, fmt.Errorf("find testid: %w", err)
	}

	if result == "" {
		return &FindResults{Text: "No elements found\n"}, nil
	}

	items := []FindResult{
		{Role: "testid", Name: result},
	}
	return &FindResults{
		Items: items,
		Text:  result + "\n",
	}, nil
}

// FindOption configures find behavior.
type FindOption func(*findConfig)

type findConfig struct {
	Exact bool
}

// WithExact enables exact matching instead of substring matching.
func WithExact() FindOption {
	return func(c *findConfig) {
		c.Exact = true
	}
}

func parseFindOptions(opts []FindOption) findConfig {
	var cfg findConfig
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

// FindNth returns the nth match (1-based) from a CSS selector query.
// This operates on the DOM, not the snapshot.
func (b *Browser) FindNth(cssSelector string, n int) (*FindResults, error) {
	if n < 1 {
		return nil, fmt.Errorf("n must be >= 1, got %d", n)
	}

	// Use JS to find the nth element matching the selector
	result, err := b.evaluateString(fmt.Sprintf(`(() => {
		const els = document.querySelectorAll(%s);
		if (%d > els.length) {
			return JSON.stringify({error: 'only ' + els.length + ' elements match', count: els.length});
		}
		const el = els[%d - 1];
		return JSON.stringify({
			tag: el.tagName.toLowerCase(),
			text: (el.innerText || el.textContent || '').substring(0, 80),
			id: el.id || '',
			className: el.className || '',
			count: els.length
		});
	})()`, mustJSON(cssSelector), n, n))
	if err != nil {
		return nil, fmt.Errorf("find nth: %w", err)
	}

	items := []FindResult{
		{ID: n, Role: "match", Name: result},
	}
	return &FindResults{
		Items: items,
		Text:  result,
	}, nil
}

// FindFirst returns the first element matching the given CSS selector.
func (b *Browser) FindFirst(cssSelector string) (*FindResults, error) {
	return b.FindNth(cssSelector, 1)
}

// FindLast returns the last element matching the given CSS selector.
func (b *Browser) FindLast(cssSelector string) (*FindResults, error) {
	result, err := b.evaluateString(fmt.Sprintf(`(() => {
		const els = document.querySelectorAll(%s);
		if (els.length === 0) return JSON.stringify({error: 'no elements match', count: 0});
		const el = els[els.length - 1];
		return JSON.stringify({
			tag: el.tagName.toLowerCase(),
			text: (el.innerText || el.textContent || '').substring(0, 80),
			id: el.id || '',
			className: el.className || '',
			index: els.length,
			count: els.length
		});
	})()`, mustJSON(cssSelector)))
	if err != nil {
		return nil, fmt.Errorf("find last: %w", err)
	}

	items := []FindResult{
		{Role: "match", Name: result},
	}
	return &FindResults{
		Items: items,
		Text:  result,
	}, nil
}

// --- helpers ---

// walkTree walks the snapshot tree in display order and calls fn for each non-root node
// with its display ID.
func walkTree(nodes []*axtree.Node, fn func(displayID int, node *axtree.Node)) {
	counter := 0
	var walk func(nodes []*axtree.Node)
	walk = func(nodes []*axtree.Node) {
		for _, node := range nodes {
			if isRootRole(node.Role) {
				walk(node.Children)
				continue
			}
			counter++
			fn(counter, node)
			walk(node.Children)
		}
	}
	walk(nodes)
}

func isFormRole(role string) bool {
	switch strings.ToLower(role) {
	case "textbox", "searchbox", "checkbox", "radio", "combobox",
		"listbox", "switch", "slider", "spinbutton", "option":
		return true
	default:
		return false
	}
}

func formatFindResults(items []FindResult) *FindResults {
	var buf strings.Builder
	for _, item := range items {
		if item.Name != "" {
			buf.WriteString(fmt.Sprintf("%d: %s %q\n", item.ID, item.Role, item.Name))
		} else {
			buf.WriteString(fmt.Sprintf("%d: %s\n", item.ID, item.Role))
		}
	}
	return &FindResults{
		Items: items,
		Text:  buf.String(),
	}
}
