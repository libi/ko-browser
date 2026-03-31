package browser

import (
	"context"
	"encoding/json"
	"fmt"

	cdp "github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func (b *Browser) resolveBackendID(id int) (cdp.BackendNodeID, error) {
	backendID, err := b.resolveID(id)
	if err != nil {
		return 0, err
	}
	return cdp.BackendNodeID(backendID), nil
}

func (b *Browser) resolveRemoteObject(ctx context.Context, id int) (*runtime.RemoteObject, cdp.BackendNodeID, error) {
	backendID, err := b.resolveBackendID(id)
	if err != nil {
		return nil, 0, err
	}

	var remoteObj *runtime.RemoteObject
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var e error
		remoteObj, e = dom.ResolveNode().WithBackendNodeID(backendID).Do(ctx)
		return e
	}))
	if err != nil {
		return nil, 0, fmt.Errorf("resolve node (backendID=%d): %w", backendID, err)
	}

	return remoteObj, backendID, nil
}

func marshalCallArguments(values ...any) ([]*runtime.CallArgument, error) {
	if len(values) == 0 {
		return nil, nil
	}

	args := make([]*runtime.CallArgument, 0, len(values))
	for _, value := range values {
		buf, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		args = append(args, &runtime.CallArgument{Value: buf})
	}
	return args, nil
}

func callFunctionOnObject(ctx context.Context, objectID runtime.RemoteObjectID, function string, args ...any) (*runtime.RemoteObject, error) {
	callArgs, err := marshalCallArguments(args...)
	if err != nil {
		return nil, err
	}

	params := runtime.CallFunctionOn(function).WithObjectID(objectID)
	if len(callArgs) > 0 {
		params = params.WithArguments(callArgs)
	}

	result, exceptionDetails, err := params.Do(ctx)
	if err != nil {
		return nil, err
	}
	if exceptionDetails != nil {
		return nil, fmt.Errorf("javascript exception: %s", exceptionDetails.Text)
	}
	return result, nil
}

func (b *Browser) runOnElement(id int, function string, args ...any) error {
	ctx, cancel := b.operationContext()
	defer cancel()

	remoteObj, _, err := b.resolveRemoteObject(ctx, id)
	if err != nil {
		return err
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		_, err := callFunctionOnObject(ctx, remoteObj.ObjectID, function, args...)
		return err
	}))
}

func (b *Browser) focusNode(ctx context.Context, backendID cdp.BackendNodeID) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		if err := dom.Focus().WithBackendNodeID(backendID).Do(ctx); err != nil {
			return fmt.Errorf("focus node (backendID=%d): %w", backendID, err)
		}
		return nil
	}))
}

func nodeCenter(ctx context.Context, backendID cdp.BackendNodeID) (float64, float64, error) {
	var x, y float64
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		boxModel, err := dom.GetBoxModel().WithBackendNodeID(backendID).Do(ctx)
		if err != nil {
			return fmt.Errorf("get box model (backendID=%d): %w", backendID, err)
		}
		quad := boxModel.Content
		if len(quad) < 8 {
			return fmt.Errorf("unexpected content quad length: %d", len(quad))
		}
		for i := 0; i < 8; i += 2 {
			x += quad[i]
			y += quad[i+1]
		}
		x /= 4
		y /= 4
		return nil
	}))
	return x, y, err
}
