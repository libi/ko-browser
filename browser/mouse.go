package browser

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

func (b *Browser) Hover(id int) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	backendID, err := b.resolveBackendID(id)
	if err != nil {
		return b.evaluateOnInteractiveElement(id, `
			el.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));
			el.dispatchEvent(new MouseEvent('mouseenter', { bubbles: true }));
			el.dispatchEvent(new MouseEvent('mousemove', { bubbles: true }));`)
	}

	x, y, err := nodeCenter(ctx, backendID)
	if err != nil {
		return b.evaluateOnInteractiveElement(id, `
			el.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));
			el.dispatchEvent(new MouseEvent('mouseenter', { bubbles: true }));
			el.dispatchEvent(new MouseEvent('mousemove', { bubbles: true }));`)
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		if err := input.DispatchMouseEvent(input.MouseMoved, x, y).Do(ctx); err != nil {
			return fmt.Errorf("hover element (backendID=%d): %w", backendID, err)
		}
		return nil
	}))
}

func (b *Browser) DblClick(id int) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	backendID, err := b.resolveBackendID(id)
	if err != nil {
		return b.evaluateOnInteractiveElement(id, `
			if (typeof el.click === 'function') { el.click(); el.click(); }
			el.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }));`)
	}

	x, y, err := nodeCenter(ctx, backendID)
	if err != nil {
		return b.evaluateOnInteractiveElement(id, `
			if (typeof el.click === 'function') { el.click(); el.click(); }
			el.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }));`)
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		if err := input.DispatchMouseEvent(input.MouseMoved, x, y).Do(ctx); err != nil {
			return err
		}
		if err := input.DispatchMouseEvent(input.MousePressed, x, y).WithButton(input.Left).WithButtons(1).WithClickCount(1).Do(ctx); err != nil {
			return err
		}
		if err := input.DispatchMouseEvent(input.MouseReleased, x, y).WithButton(input.Left).WithButtons(1).WithClickCount(1).Do(ctx); err != nil {
			return err
		}
		if err := input.DispatchMouseEvent(input.MousePressed, x, y).WithButton(input.Left).WithButtons(1).WithClickCount(2).Do(ctx); err != nil {
			return err
		}
		if err := input.DispatchMouseEvent(input.MouseReleased, x, y).WithButton(input.Left).WithButtons(1).WithClickCount(2).Do(ctx); err != nil {
			return err
		}
		return nil
	}))
}

func (b *Browser) Focus(id int) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	backendID, err := b.resolveBackendID(id)
	if err != nil {
		return b.evaluateOnInteractiveElement(id, `if (typeof el.focus === 'function') { el.focus(); }`)
	}

	if err := b.focusNode(ctx, backendID); err != nil {
		return b.evaluateOnInteractiveElement(id, `if (typeof el.focus === 'function') { el.focus(); }`)
	}
	return nil
}

// --- Low-level mouse operations ---

// MouseButton specifies a mouse button.
type MouseButton string

const (
	MouseLeft   MouseButton = "left"
	MouseRight  MouseButton = "right"
	MouseMiddle MouseButton = "middle"
)

// MouseOptions configures a mouse action.
type MouseOptions struct {
	Button MouseButton // button for down/up (default: left)
}

func toInputButton(mb MouseButton) input.MouseButton {
	switch mb {
	case MouseRight:
		return input.Right
	case MouseMiddle:
		return input.Middle
	default:
		return input.Left
	}
}

func buttonMask(mb MouseButton) int64 {
	switch mb {
	case MouseRight:
		return 2
	case MouseMiddle:
		return 4
	default:
		return 1
	}
}

// MouseMove moves the mouse to the given page coordinates (x, y).
func (b *Browser) MouseMove(x, y float64) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return input.DispatchMouseEvent(input.MouseMoved, x, y).Do(ctx)
	}))
}

// MouseDown presses a mouse button at the current position (or the given coordinates).
func (b *Browser) MouseDown(x, y float64, opts ...MouseOptions) error {
	btn := input.Left
	mask := int64(1)
	if len(opts) > 0 && opts[0].Button != "" {
		btn = toInputButton(opts[0].Button)
		mask = buttonMask(opts[0].Button)
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return input.DispatchMouseEvent(input.MousePressed, x, y).
			WithButton(btn).
			WithButtons(mask).
			WithClickCount(1).
			Do(ctx)
	}))
}

// MouseUp releases a mouse button at the given coordinates.
func (b *Browser) MouseUp(x, y float64, opts ...MouseOptions) error {
	btn := input.Left
	if len(opts) > 0 && opts[0].Button != "" {
		btn = toInputButton(opts[0].Button)
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return input.DispatchMouseEvent(input.MouseReleased, x, y).
			WithButton(btn).
			WithButtons(0).
			WithClickCount(1).
			Do(ctx)
	}))
}

// MouseWheel dispatches a mouse wheel event at the given coordinates.
// deltaX scrolls horizontally (positive = right), deltaY scrolls vertically (positive = down).
func (b *Browser) MouseWheel(x, y float64, deltaX, deltaY float64) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return input.DispatchMouseEvent(input.MouseWheel, x, y).
			WithDeltaX(deltaX).
			WithDeltaY(deltaY).
			Do(ctx)
	}))
}

// MouseClick performs a full click (move + down + up) at the given coordinates.
func (b *Browser) MouseClick(x, y float64, opts ...MouseOptions) error {
	btn := input.Left
	mask := int64(1)
	if len(opts) > 0 && opts[0].Button != "" {
		btn = toInputButton(opts[0].Button)
		mask = buttonMask(opts[0].Button)
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		if err := input.DispatchMouseEvent(input.MouseMoved, x, y).Do(ctx); err != nil {
			return err
		}
		if err := input.DispatchMouseEvent(input.MousePressed, x, y).
			WithButton(btn).
			WithButtons(mask).
			WithClickCount(1).
			Do(ctx); err != nil {
			return err
		}
		return input.DispatchMouseEvent(input.MouseReleased, x, y).
			WithButton(btn).
			WithButtons(0).
			WithClickCount(1).
			Do(ctx)
	}))
}
