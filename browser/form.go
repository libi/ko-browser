package browser

import (
	"fmt"
	"strings"
)

func (b *Browser) Check(id int) error {
	return b.setCheckedState(id, true)
}

func (b *Browser) Uncheck(id int) error {
	return b.setCheckedState(id, false)
}

func (b *Browser) Select(id int, values ...string) error {
	if err := b.Focus(id); err != nil {
		return err
	}

	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			trimmed = append(trimmed, value)
		}
	}
	if len(trimmed) == 0 {
		return fmt.Errorf("at least one select value is required")
	}

	return b.evaluate(`(() => {
		const values = ` + mustJSON(trimmed) + `;
		const el = document.activeElement;
		if (!el || (el.tagName || '').toLowerCase() !== 'select') {
			throw new Error('target element is not a select');
		}
		const wanted = new Set(values);
		let firstMatchValue = null;
		for (const option of el.options) {
			const match = wanted.has(option.value) || wanted.has(option.text);
			option.selected = match;
			if (match && firstMatchValue === null) {
				firstMatchValue = option.value;
			}
		}
		if (!el.multiple && firstMatchValue !== null) {
			el.value = firstMatchValue;
		}
		el.dispatchEvent(new Event('input', { bubbles: true }));
		el.dispatchEvent(new Event('change', { bubbles: true }));
	})()`)
}

func (b *Browser) setCheckedState(id int, checked bool) error {
	if err := b.Focus(id); err != nil {
		return err
	}

	return b.evaluate(`(() => {
		const checked = ` + mustJSON(checked) + `;
		const el = document.activeElement;
		if (!el || !('checked' in el)) {
			throw new Error('target element is not checkable');
		}
		el.checked = checked;
		el.dispatchEvent(new Event('input', { bubbles: true }));
		el.dispatchEvent(new Event('change', { bubbles: true }));
	})()`)
}
