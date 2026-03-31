package browser

// IsVisible returns true if the element is visible (has non-zero size and not hidden).
func (b *Browser) IsVisible(id int) (bool, error) {
	result, err := b.evaluateOnElement(id, `function() {
		const cs = window.getComputedStyle(this);
		if (cs.display === 'none' || cs.visibility === 'hidden' || cs.opacity === '0') return 'false';
		const r = this.getBoundingClientRect();
		if (r.width === 0 && r.height === 0) return 'false';
		return 'true';
	}`)
	if err != nil {
		return false, err
	}
	return result == "true", nil
}

// IsEnabled returns true if the element is not disabled.
func (b *Browser) IsEnabled(id int) (bool, error) {
	result, err := b.evaluateOnElement(id, `function() {
		if ('disabled' in this) return this.disabled ? 'false' : 'true';
		return this.getAttribute('aria-disabled') === 'true' ? 'false' : 'true';
	}`)
	if err != nil {
		return false, err
	}
	return result == "true", nil
}

// IsChecked returns true if the element is checked.
func (b *Browser) IsChecked(id int) (bool, error) {
	result, err := b.evaluateOnElement(id, `function() {
		if ('checked' in this) return this.checked ? 'true' : 'false';
		const ariaChecked = this.getAttribute('aria-checked');
		return ariaChecked === 'true' ? 'true' : 'false';
	}`)
	if err != nil {
		return false, err
	}
	return result == "true", nil
}
