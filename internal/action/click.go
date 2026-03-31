package action

import (
	"context"
	"fmt"

	cdp "github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// Click performs a click on the DOM element identified by its BackendDOMNodeID.
//
// Steps:
//  1. Resolve BackendNodeID → RemoteObject (to validate it still exists)
//  2. Scroll the element into view
//  3. Call element.click() via JS
//
// This approach works for links, buttons, checkboxes, and most interactive elements.
func Click(ctx context.Context, backendNodeID int64) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Resolve backend node to a live remote object
		remoteObj, err := dom.ResolveNode().WithBackendNodeID(cdp.BackendNodeID(backendNodeID)).Do(ctx)
		if err != nil {
			return fmt.Errorf("resolve node (backendID=%d): %w", backendNodeID, err)
		}

		// Scroll element into view so it's visible
		_, _, err = runtime.CallFunctionOn(`function() {
			if (this.scrollIntoViewIfNeeded) {
				this.scrollIntoViewIfNeeded(true);
			} else {
				this.scrollIntoView({block: 'center', inline: 'center'});
			}
		}`).WithObjectID(remoteObj.ObjectID).Do(ctx)
		if err != nil {
			// Non-fatal: some elements can't be scrolled
			fmt.Printf("  [warn] scroll into view failed: %v\n", err)
		}

		// Click the element via JS
		_, exceptionDetails, err := runtime.CallFunctionOn(`function() { this.click(); }`).
			WithObjectID(remoteObj.ObjectID).
			Do(ctx)
		if err != nil {
			return fmt.Errorf("click element (backendID=%d): %w", backendNodeID, err)
		}
		if exceptionDetails != nil {
			return fmt.Errorf("click threw exception: %s", exceptionDetails.Text)
		}

		return nil
	}))
}
