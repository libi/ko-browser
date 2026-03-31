package browser

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func (b *Browser) Open(url string) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)
}

func (b *Browser) Back() error {
	return b.evaluate(`history.back()`)
}

func (b *Browser) Forward() error {
	return b.evaluate(`history.forward()`)
}

func (b *Browser) Reload() error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return page.Reload().Do(ctx)
	}), chromedp.WaitReady("body", chromedp.ByQuery))
}

func (b *Browser) Wait(d time.Duration) error {
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-b.ctx.Done():
		return b.ctx.Err()
	}
}
