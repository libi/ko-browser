package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func (b *Browser) Scroll(direction string, pixels int) error {
	if pixels <= 0 {
		return fmt.Errorf("pixels must be positive")
	}

	dx := 0
	dy := 0
	switch strings.ToLower(strings.TrimSpace(direction)) {
	case "down":
		dy = pixels
	case "up":
		dy = -pixels
	case "left":
		dx = -pixels
	case "right":
		dx = pixels
	default:
		return fmt.Errorf("unsupported scroll direction %q", direction)
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, err := runtime.Evaluate(fmt.Sprintf("window.scrollBy(%d, %d)", dx, dy)).Do(ctx)
		return err
	}))
}

func (b *Browser) ScrollIntoView(id int) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	scrollFallback := func() error {
		return b.evaluateOnInteractiveElement(id, `
			if (el.scrollIntoViewIfNeeded) {
				el.scrollIntoViewIfNeeded(true);
			} else {
				el.scrollIntoView({ block: 'center', inline: 'center' });
			}`)
	}

	backendID, err := b.resolveBackendID(id)
	if err != nil {
		return scrollFallback()
	}
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return dom.ScrollIntoViewIfNeeded().WithBackendNodeID(backendID).Do(ctx)
	}))
	if err != nil {
		return scrollFallback()
	}
	return nil
}

func (b *Browser) evaluate(expression string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	var exception *runtime.ExceptionDetails
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var e error
		_, exception, e = runtime.Evaluate(expression).Do(ctx)
		return e
	}))
	if err != nil {
		return err
	}
	if exception != nil {
		return fmt.Errorf("javascript exception: %s", exception.Text)
	}
	return nil
}

func mustJSON(value any) string {
	buf, _ := json.Marshal(value)
	return string(buf)
}
