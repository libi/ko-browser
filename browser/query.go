package browser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// GetTitle returns the current page title.
func (b *Browser) GetTitle() (string, error) {
	return b.evaluateString(`document.title`)
}

// GetURL returns the current page URL.
func (b *Browser) GetURL() (string, error) {
	return b.evaluateString(`location.href`)
}

// GetText returns the inner text of the element with the given display ID.
func (b *Browser) GetText(id int) (string, error) {
	return b.evaluateOnElement(id, `function() { return this.innerText || this.textContent || ''; }`)
}

// GetHTML returns the inner HTML of the element with the given display ID.
func (b *Browser) GetHTML(id int) (string, error) {
	return b.evaluateOnElement(id, `function() { return this.innerHTML || ''; }`)
}

// GetValue returns the value of a form element with the given display ID.
func (b *Browser) GetValue(id int) (string, error) {
	return b.evaluateOnElement(id, `function() {
		if ('value' in this) return this.value;
		if (this.isContentEditable) return this.textContent || '';
		return '';
	}`)
}

// GetAttr returns the value of an HTML attribute on the element.
func (b *Browser) GetAttr(id int, name string) (string, error) {
	return b.evaluateOnElement(id, `function(name) {
		const v = this.getAttribute(name);
		return v === null ? '' : v;
	}`, name)
}

// GetCount returns the number of elements matching a CSS selector.
func (b *Browser) GetCount(cssSelector string) (int, error) {
	result, err := b.evaluateString(fmt.Sprintf(
		`document.querySelectorAll(%s).length.toString()`, mustJSON(cssSelector)))
	if err != nil {
		return 0, err
	}
	var n int
	if _, err := fmt.Sscanf(result, "%d", &n); err != nil {
		return 0, fmt.Errorf("parse count %q: %w", result, err)
	}
	return n, nil
}

// BoxResult holds the bounding box of an element.
type BoxResult struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// GetBox returns the bounding box of the element with the given display ID.
func (b *Browser) GetBox(id int) (*BoxResult, error) {
	ctx, cancel := b.operationContext()
	defer cancel()

	backendID, err := b.resolveBackendID(id)
	if err != nil {
		// fallback: try via JS
		raw, jsErr := b.evaluateOnElement(id, `function() {
			const r = this.getBoundingClientRect();
			return JSON.stringify({ x: r.x, y: r.y, width: r.width, height: r.height });
		}`)
		if jsErr != nil {
			return nil, err
		}
		var box BoxResult
		if e := json.Unmarshal([]byte(raw), &box); e != nil {
			return nil, e
		}
		return &box, nil
	}

	var box BoxResult
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		model, e := dom.GetBoxModel().WithBackendNodeID(backendID).Do(ctx)
		if e != nil {
			return e
		}
		quad := model.Content
		if len(quad) < 8 {
			return fmt.Errorf("unexpected quad length: %d", len(quad))
		}
		box.X = quad[0]
		box.Y = quad[1]
		box.Width = quad[2] - quad[0]
		box.Height = quad[5] - quad[1]
		return nil
	}))
	if err != nil {
		return nil, err
	}
	return &box, nil
}

// GetStyles returns selected computed styles of the element as a JSON string.
func (b *Browser) GetStyles(id int) (string, error) {
	return b.evaluateOnElement(id, `function() {
		const cs = window.getComputedStyle(this);
		const props = [
			'display', 'visibility', 'opacity', 'position',
			'width', 'height', 'color', 'backgroundColor',
			'fontSize', 'fontFamily', 'fontWeight',
			'margin', 'padding', 'border',
			'overflow', 'zIndex'
		];
		const result = {};
		for (const p of props) { result[p] = cs.getPropertyValue(p.replace(/[A-Z]/g, m => '-' + m.toLowerCase())); }
		return JSON.stringify(result);
	}`)
}

// --- helper: evaluate JS and return string result ---

func (b *Browser) evaluateString(expression string) (string, error) {
	ctx, cancel := b.operationContext()
	defer cancel()

	var result string
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		val, exception, e := runtime.Evaluate(expression).Do(ctx)
		if e != nil {
			return e
		}
		if exception != nil {
			return fmt.Errorf("javascript exception: %s", exception.Text)
		}
		if val == nil {
			return nil
		}
		// Try to unmarshal the JSON value as a string
		if err := json.Unmarshal(val.Value, &result); err != nil {
			// Fallback: use the raw description or string value
			result = val.Description
		}
		return nil
	}))
	return result, err
}

// evaluateOnElement resolves a display ID to a remote object and calls a JS function on it.
// Returns the string result.
func (b *Browser) evaluateOnElement(id int, function string, args ...any) (string, error) {
	ctx, cancel := b.operationContext()
	defer cancel()

	remoteObj, _, err := b.resolveRemoteObject(ctx, id)
	if err != nil {
		// Fallback: use interactiveOrdinal if available
		return b.evaluateOnElementFallback(id, function, args...)
	}

	var result string
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		res, e := callFunctionOnObject(ctx, remoteObj.ObjectID, function, args...)
		if e != nil {
			return e
		}
		if res != nil && res.Value != nil {
			if e := json.Unmarshal(res.Value, &result); e != nil {
				result = res.Description
			}
		}
		return nil
	}))
	return result, err
}

// evaluateOnElementFallback uses querySelectorAll to find the element.
func (b *Browser) evaluateOnElementFallback(id int, function string, args ...any) (string, error) {
	// Build args JSON array for JS
	argsJSON := "[]"
	if len(args) > 0 {
		buf, err := json.Marshal(args)
		if err != nil {
			return "", err
		}
		argsJSON = string(buf)
	}

	ordinal, err := b.interactiveOrdinal(id)
	if err != nil {
		// Not interactive, try all elements via tree walk
		return "", fmt.Errorf("element %d not found for query", id)
	}

	expression := `(() => {
		const elements = Array.from(document.querySelectorAll(` + mustJSON(interactiveQuery) + `));
		const el = elements[` + mustJSON(ordinal) + `];
		if (!el) throw new Error('element not found');
		const fn = ` + function + `;
		const args = ` + argsJSON + `;
		return fn.apply(el, args);
	})()`

	return b.evaluateString(expression)
}
