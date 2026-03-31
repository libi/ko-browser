package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

// Drag performs a drag-and-drop from the source element to the destination element.
// Both srcID and dstID are display IDs from the latest snapshot.
func (b *Browser) Drag(srcID, dstID int) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	srcBackend, err := b.resolveBackendID(srcID)
	if err != nil {
		return fmt.Errorf("resolve source element %d: %w", srcID, err)
	}

	dstBackend, err := b.resolveBackendID(dstID)
	if err != nil {
		return fmt.Errorf("resolve destination element %d: %w", dstID, err)
	}

	srcX, srcY, err := nodeCenter(ctx, srcBackend)
	if err != nil {
		return fmt.Errorf("get source center: %w", err)
	}

	dstX, dstY, err := nodeCenter(ctx, dstBackend)
	if err != nil {
		return fmt.Errorf("get destination center: %w", err)
	}

	return b.dragCoords(ctx, srcX, srcY, dstX, dstY)
}

// DragCoords performs a drag-and-drop from (srcX, srcY) to (dstX, dstY) in page coordinates.
func (b *Browser) DragCoords(srcX, srcY, dstX, dstY float64) error {
	ctx, cancel := b.operationContext()
	defer cancel()
	return b.dragCoords(ctx, srcX, srcY, dstX, dstY)
}

func (b *Browser) dragCoords(ctx context.Context, srcX, srcY, dstX, dstY float64) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// 1. Move to source position
		if err := input.DispatchMouseEvent(input.MouseMoved, srcX, srcY).Do(ctx); err != nil {
			return fmt.Errorf("move to source: %w", err)
		}
		time.Sleep(50 * time.Millisecond)

		// 2. Press mouse button at source
		if err := input.DispatchMouseEvent(input.MousePressed, srcX, srcY).
			WithButton(input.Left).
			WithButtons(1).
			WithClickCount(1).
			Do(ctx); err != nil {
			return fmt.Errorf("mousedown at source: %w", err)
		}
		time.Sleep(100 * time.Millisecond)

		// 3. Move in steps to destination (smooth drag)
		steps := 10
		for i := 1; i <= steps; i++ {
			fraction := float64(i) / float64(steps)
			x := srcX + (dstX-srcX)*fraction
			y := srcY + (dstY-srcY)*fraction
			if err := input.DispatchMouseEvent(input.MouseMoved, x, y).
				WithButton(input.Left).
				WithButtons(1).
				Do(ctx); err != nil {
				return fmt.Errorf("mousemove step %d: %w", i, err)
			}
			time.Sleep(20 * time.Millisecond)
		}

		// 4. Small pause at destination before release
		time.Sleep(50 * time.Millisecond)

		// 5. Release mouse button at destination
		if err := input.DispatchMouseEvent(input.MouseReleased, dstX, dstY).
			WithButton(input.Left).
			WithButtons(0).
			WithClickCount(1).
			Do(ctx); err != nil {
			return fmt.Errorf("mouseup at destination: %w", err)
		}

		return nil
	}))
}
