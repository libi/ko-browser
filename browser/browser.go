package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/libi/ko-browser/axtree"
)

type Browser struct {
	ctx          context.Context
	cancel       context.CancelFunc
	allocCtx     context.Context
	allocCanc    context.CancelFunc
	timeout      time.Duration
	lastSnap     *SnapshotResult
	networkState *networkState
	consoleState *consoleState
	tabs         []tabEntry
	activeTab    int

	// Phase 7
	trace     *traceState
	profiling bool
	recording *recordState

	// Phase 8.3: Global options
	downloadPath     string
	screenshotDir    string
	screenshotFormat string
}

type SnapshotResult struct {
	Text     string
	Nodes    []*axtree.Node
	IDMap    map[int]int64
	RawCount int
}

func New(opts Options) (*Browser, error) {
	opts = opts.normalized()

	allocOpts := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	if !opts.Headless {
		allocOpts = append(allocOpts, chromedp.Flag("headless", false))
	}
	if shouldDisableSandbox() {
		allocOpts = append(allocOpts,
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)
	}
	if opts.Profile != "" {
		allocOpts = append(allocOpts, chromedp.UserDataDir(opts.Profile))
	}

	// Phase 8.3: Global options
	if opts.UserAgent != "" {
		allocOpts = append(allocOpts, chromedp.UserAgent(opts.UserAgent))
	}
	if opts.Proxy != "" {
		allocOpts = append(allocOpts, chromedp.ProxyServer(opts.Proxy))
	}
	if opts.ProxyBypass != "" {
		allocOpts = append(allocOpts, chromedp.Flag("proxy-bypass-list", opts.ProxyBypass))
	}
	if opts.IgnoreHTTPSErrors {
		allocOpts = append(allocOpts, chromedp.Flag("ignore-certificate-errors", true))
	}
	if opts.AllowFileAccess {
		allocOpts = append(allocOpts, chromedp.Flag("allow-file-access-from-files", true))
	}
	for _, ext := range opts.Extensions {
		allocOpts = append(allocOpts, chromedp.Flag("load-extension", ext))
		// Extensions require non-headless or at least --headless=new
		allocOpts = append(allocOpts, chromedp.Flag("disable-extensions-except", ext))
	}
	for _, arg := range opts.ExtraArgs {
		allocOpts = append(allocOpts, chromedp.Flag(arg, true))
	}

	allocCtx, allocCanc := chromedp.NewExecAllocator(context.Background(), allocOpts...)

	ctxOpts := []chromedp.ContextOption{}
	if opts.Logf != nil {
		ctxOpts = append(ctxOpts, chromedp.WithLogf(opts.Logf))
	}

	ctx, cancel := chromedp.NewContext(allocCtx, ctxOpts...)

	b := &Browser{
		ctx:              ctx,
		cancel:           cancel,
		allocCtx:         allocCtx,
		allocCanc:        allocCanc,
		timeout:          opts.Timeout,
		downloadPath:     opts.DownloadPath,
		screenshotDir:    opts.ScreenshotDir,
		screenshotFormat: opts.ScreenshotFormat,
	}

	// If StatePath is set, import state after browser starts
	if opts.StatePath != "" {
		// Ensure browser is started by running a trivial action
		if err := chromedp.Run(ctx); err != nil {
			b.Close()
			return nil, fmt.Errorf("start browser: %w", err)
		}
		if err := b.ImportState(opts.StatePath); err != nil {
			b.Close()
			return nil, fmt.Errorf("import state: %w", err)
		}
	}

	return b, nil
}

func (b *Browser) Context() context.Context {
	return b.ctx
}

func (b *Browser) Close() {
	if b.allocCanc != nil {
		b.allocCanc()
	}
	if b.cancel != nil {
		b.cancel()
	}
}

func (b *Browser) resolveID(id int) (int64, error) {
	if b.lastSnap == nil {
		return 0, fmt.Errorf("no snapshot yet, call Snapshot() first")
	}
	bid, ok := b.lastSnap.IDMap[id]
	if !ok {
		return 0, fmt.Errorf("element %d not found (valid: 1-%d)", id, len(b.lastSnap.IDMap))
	}
	if bid == 0 {
		return 0, fmt.Errorf("element %d has no BackendDOMNodeID", id)
	}
	return bid, nil
}

func (b *Browser) operationContext() (context.Context, context.CancelFunc) {
	ctx := b.activeContext()
	return ctx, func() {}
}
