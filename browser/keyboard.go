package browser

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

func (b *Browser) Type(id int, text string) error {
	if err := b.Focus(id); err != nil {
		return err
	}
	return b.KeyboardType(text)
}

func (b *Browser) Fill(id int, text string) error {
	if err := b.Focus(id); err != nil {
		return err
	}
	if err := b.evaluate(`(() => {
		const el = document.activeElement;
		if (!el) {
			throw new Error('no active element');
		}
		if (el.isContentEditable) {
			el.textContent = '';
			el.dispatchEvent(new Event('input', { bubbles: true }));
			return;
		}
		if ('value' in el) {
			el.value = '';
			el.dispatchEvent(new Event('input', { bubbles: true }));
			el.dispatchEvent(new Event('change', { bubbles: true }));
			return;
		}
		el.textContent = '';
		el.dispatchEvent(new Event('input', { bubbles: true }));
	})()`); err != nil {
		return err
	}
	return b.KeyboardType(text)
}

func (b *Browser) KeyboardType(text string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return input.InsertText(text).Do(ctx)
	}))
}

// KeyboardInsertText inserts text without triggering keyboard events (keydown/keypress/keyup).
// This is similar to pasting text and is useful for rich text editors.
func (b *Browser) KeyboardInsertText(text string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return input.InsertText(text).Do(ctx)
	}))
}

func (b *Browser) Press(key string) error {
	spec, err := parseKeySpec(key)
	if err != nil {
		return err
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		down := input.DispatchKeyEvent(input.KeyRawDown).
			WithModifiers(spec.Modifiers).
			WithKey(spec.Key).
			WithCode(spec.Code).
			WithWindowsVirtualKeyCode(spec.WindowsVirtualKeyCode)
		if spec.Text != "" {
			down = down.WithText(spec.Text).WithUnmodifiedText(spec.Text)
		}
		if err := down.Do(ctx); err != nil {
			return err
		}

		up := input.DispatchKeyEvent(input.KeyUp).
			WithModifiers(spec.Modifiers).
			WithKey(spec.Key).
			WithCode(spec.Code).
			WithWindowsVirtualKeyCode(spec.WindowsVirtualKeyCode)
		return up.Do(ctx)
	}))
}

type keySpec struct {
	Key                   string
	Code                  string
	Text                  string
	WindowsVirtualKeyCode int64
	Modifiers             input.Modifier
}

func parseKeySpec(value string) (keySpec, error) {
	parts := strings.Split(strings.TrimSpace(value), "+")
	if len(parts) == 0 {
		return keySpec{}, fmt.Errorf("key cannot be empty")
	}

	var modifiers input.Modifier
	for _, part := range parts[:len(parts)-1] {
		switch strings.ToLower(strings.TrimSpace(part)) {
		case "alt", "option":
			modifiers |= input.ModifierAlt
		case "control", "ctrl":
			modifiers |= input.ModifierCtrl
		case "meta", "cmd", "command":
			modifiers |= input.ModifierMeta
		case "shift":
			modifiers |= input.ModifierShift
		case "":
		default:
			return keySpec{}, fmt.Errorf("unsupported modifier %q", part)
		}
	}

	mainKey := strings.TrimSpace(parts[len(parts)-1])
	if mainKey == "" {
		return keySpec{}, fmt.Errorf("key cannot be empty")
	}

	lookup := map[string]keySpec{
		"enter":      {Key: "Enter", Code: "Enter", WindowsVirtualKeyCode: 13},
		"tab":        {Key: "Tab", Code: "Tab", WindowsVirtualKeyCode: 9},
		"escape":     {Key: "Escape", Code: "Escape", WindowsVirtualKeyCode: 27},
		"esc":        {Key: "Escape", Code: "Escape", WindowsVirtualKeyCode: 27},
		"backspace":  {Key: "Backspace", Code: "Backspace", WindowsVirtualKeyCode: 8},
		"delete":     {Key: "Delete", Code: "Delete", WindowsVirtualKeyCode: 46},
		"space":      {Key: " ", Code: "Space", Text: " ", WindowsVirtualKeyCode: 32},
		"arrowup":    {Key: "ArrowUp", Code: "ArrowUp", WindowsVirtualKeyCode: 38},
		"arrowdown":  {Key: "ArrowDown", Code: "ArrowDown", WindowsVirtualKeyCode: 40},
		"arrowleft":  {Key: "ArrowLeft", Code: "ArrowLeft", WindowsVirtualKeyCode: 37},
		"arrowright": {Key: "ArrowRight", Code: "ArrowRight", WindowsVirtualKeyCode: 39},
		"home":       {Key: "Home", Code: "Home", WindowsVirtualKeyCode: 36},
		"end":        {Key: "End", Code: "End", WindowsVirtualKeyCode: 35},
		"pageup":     {Key: "PageUp", Code: "PageUp", WindowsVirtualKeyCode: 33},
		"pagedown":   {Key: "PageDown", Code: "PageDown", WindowsVirtualKeyCode: 34},
	}

	if spec, ok := lookup[strings.ToLower(mainKey)]; ok {
		spec.Modifiers = modifiers
		if modifiers != input.ModifierNone {
			spec.Text = ""
		}
		return spec, nil
	}

	runes := []rune(mainKey)
	if len(runes) == 1 {
		r := runes[0]
		upper := strings.ToUpper(string(r))
		spec := keySpec{
			Key:                   string(r),
			Text:                  string(r),
			WindowsVirtualKeyCode: int64(r),
			Modifiers:             modifiers,
		}
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			spec.Code = "Key" + upper
			spec.WindowsVirtualKeyCode = int64(strings.ToUpper(string(r))[0])
			if modifiers != input.ModifierNone {
				spec.Text = ""
			}
			return spec, nil
		}
		if r >= '0' && r <= '9' {
			spec.Code = "Digit" + string(r)
			if modifiers != input.ModifierNone {
				spec.Text = ""
			}
			return spec, nil
		}
	}

	return keySpec{}, fmt.Errorf("unsupported key %q", value)
}
