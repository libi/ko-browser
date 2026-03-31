package browser

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

func (b *Browser) Click(id int) error {
	backendID, err := b.resolveID(id)
	if err != nil {
		return err
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		remoteObj, _, err := b.resolveRemoteObject(ctx, id)
		if err != nil {
			return b.evaluateOnInteractiveElement(id, `if (typeof el.click === 'function') { el.click(); }`)
		}

		_, err = callFunctionOnObject(ctx, remoteObj.ObjectID, `function() {
			if (this.scrollIntoViewIfNeeded) {
				this.scrollIntoViewIfNeeded(true);
			} else {
				this.scrollIntoView({block: 'center', inline: 'center'});
			}
		}`)
		if err != nil {
			return fmt.Errorf("scroll element (backendID=%d): %w", backendID, err)
		}

		_, err = callFunctionOnObject(ctx, remoteObj.ObjectID, `function() { this.click(); }`)
		if err != nil {
			if fallbackErr := b.evaluateOnInteractiveElement(id, `if (typeof el.click === 'function') { el.click(); }`); fallbackErr == nil {
				return nil
			}
			return fmt.Errorf("click element (backendID=%d): %w", backendID, err)
		}

		return nil
	}))
}

// ClickNewTab clicks a link element and opens it in a new tab.
// This works by setting target="_blank" temporarily and clicking.
func (b *Browser) ClickNewTab(id int) error {
	backendID, err := b.resolveID(id)
	if err != nil {
		return err
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		remoteObj, _, err := b.resolveRemoteObject(ctx, id)
		if err != nil {
			return fmt.Errorf("resolve element for new-tab click (backendID=%d): %w", backendID, err)
		}

		_, err = callFunctionOnObject(ctx, remoteObj.ObjectID, `function() {
			if (this.scrollIntoViewIfNeeded) {
				this.scrollIntoViewIfNeeded(true);
			} else {
				this.scrollIntoView({block: 'center', inline: 'center'});
			}
			// If it's a link, set target="_blank" and click
			if (this.tagName === 'A' && this.href) {
				const oldTarget = this.target;
				this.target = '_blank';
				this.click();
				this.target = oldTarget;
			} else {
				// For non-links, try Ctrl+Click behavior
				const evt = new MouseEvent('click', {
					bubbles: true,
					cancelable: true,
					ctrlKey: true,
					metaKey: true
				});
				this.dispatchEvent(evt);
			}
		}`)
		if err != nil {
			return fmt.Errorf("click new-tab element (backendID=%d): %w", backendID, err)
		}
		return nil
	}))
}
