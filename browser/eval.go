package browser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// Eval evaluates a JavaScript expression in the page context and returns the result as a string.
// For complex objects, the result is JSON-encoded.
func (b *Browser) Eval(expression string) (string, error) {
	ctx, cancel := b.operationContext()
	defer cancel()

	var result string
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		val, exception, e := runtime.Evaluate(expression).
			WithReturnByValue(true).
			WithAwaitPromise(true).
			Do(ctx)
		if e != nil {
			return e
		}
		if exception != nil {
			return fmt.Errorf("javascript exception: %s", exception.Text)
		}
		if val == nil {
			return nil
		}

		// Determine result type and convert appropriately
		switch val.Type {
		case "undefined":
			result = "undefined"
		case "object":
			if val.Subtype == "null" {
				result = "null"
			} else if val.Value != nil {
				result = string(val.Value)
			} else {
				result = val.Description
			}
		default:
			if val.Value != nil {
				var str string
				if err := json.Unmarshal(val.Value, &str); err == nil {
					result = str
				} else {
					result = string(val.Value)
				}
			} else {
				result = val.Description
			}
		}
		return nil
	}))
	return result, err
}
