package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/chromedp/cdproto/profiler"
	"github.com/chromedp/cdproto/tracing"
	"github.com/chromedp/chromedp"
)

// traceState tracks an in-progress trace recording.
type traceState struct {
	mu       sync.Mutex
	active   bool
	chunks   [][]byte // collected trace event chunks
	doneCh   chan struct{}
	listenFn context.CancelFunc
}

// TraceStart begins recording a Chrome trace.
// Categories is an optional comma-separated list of tracing categories.
// If empty, default categories are used.
func (b *Browser) TraceStart(categories ...string) error {
	if b.trace != nil && b.trace.active {
		return fmt.Errorf("trace already in progress; call TraceStop first")
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	ts := &traceState{
		active: true,
		doneCh: make(chan struct{}),
	}

	// Set up event listeners before starting trace
	chromedp.ListenTarget(b.activeContext(), func(ev interface{}) {
		switch e := ev.(type) {
		case *tracing.EventDataCollected:
			ts.mu.Lock()
			for _, v := range e.Value {
				ts.chunks = append(ts.chunks, []byte(v))
			}
			ts.mu.Unlock()
		case *tracing.EventTracingComplete:
			close(ts.doneCh)
		}
	})

	// Start tracing
	params := tracing.Start()
	if len(categories) > 0 && categories[0] != "" {
		params = params.WithTraceConfig(&tracing.TraceConfig{
			IncludedCategories: categories,
		})
	}

	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return params.Do(ctx)
	}))
	if err != nil {
		return fmt.Errorf("start tracing: %w", err)
	}

	b.trace = ts
	return nil
}

// TraceStop stops the trace recording and saves the trace data to the given path.
// The output is a JSON file that can be loaded in chrome://tracing or Perfetto UI.
func (b *Browser) TraceStop(outputPath string) error {
	if b.trace == nil || !b.trace.active {
		return fmt.Errorf("no trace in progress; call TraceStart first")
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	// End tracing
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return tracing.End().Do(ctx)
	}))
	if err != nil {
		b.trace.active = false
		return fmt.Errorf("end tracing: %w", err)
	}

	// Wait for TracingComplete event with timeout
	select {
	case <-b.trace.doneCh:
	case <-ctx.Done():
		b.trace.active = false
		return fmt.Errorf("timeout waiting for trace data")
	}

	// Assemble trace data
	b.trace.mu.Lock()
	chunks := b.trace.chunks
	b.trace.mu.Unlock()
	b.trace.active = false

	// Write trace JSON: {"traceEvents": [...all chunks...]}
	if outputPath != "" {
		if err := writeTraceFile(outputPath, chunks); err != nil {
			return fmt.Errorf("write trace: %w", err)
		}
	}

	b.trace = nil
	return nil
}

// ProfilerStart begins CPU profiling.
func (b *Browser) ProfilerStart() error {
	if b.profiling {
		return fmt.Errorf("profiler already running; call ProfilerStop first")
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		if err := profiler.Enable().Do(ctx); err != nil {
			return err
		}
		return profiler.Start().Do(ctx)
	}))
	if err != nil {
		return fmt.Errorf("start profiler: %w", err)
	}

	b.profiling = true
	return nil
}

// ProfilerStop stops CPU profiling and saves the result to the given path.
// The output is a JSON file in the Chrome DevTools CPU Profile format.
func (b *Browser) ProfilerStop(outputPath string) error {
	if !b.profiling {
		return fmt.Errorf("profiler not running; call ProfilerStart first")
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	var profile *profiler.Profile
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		profile, err = profiler.Stop().Do(ctx)
		if err != nil {
			return err
		}
		return profiler.Disable().Do(ctx)
	}))
	if err != nil {
		b.profiling = false
		return fmt.Errorf("stop profiler: %w", err)
	}

	b.profiling = false

	if outputPath != "" && profile != nil {
		data, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal profile: %w", err)
		}
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("write profile: %w", err)
		}
	}

	return nil
}

// writeTraceFile writes collected trace chunks as a valid JSON trace file.
func writeTraceFile(path string, chunks [][]byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write all chunks as a JSON array
	if _, err := f.WriteString("{\"traceEvents\":["); err != nil {
		return err
	}

	first := true
	for _, chunk := range chunks {
		// Each chunk is a JSON value from the trace event
		if !first {
			if _, err := f.WriteString(","); err != nil {
				return err
			}
		}
		if _, err := f.Write(chunk); err != nil {
			return err
		}
		first = false
	}

	if _, err := f.WriteString("]}"); err != nil {
		return err
	}

	return nil
}
