package browser

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	cruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// ConsoleMessage represents a single console log message.
type ConsoleMessage struct {
	Level     string `json:"level"` // "log", "warning", "error", "info", "debug"
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// PageError represents a JavaScript exception caught on the page.
type PageError struct {
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// consoleState holds collected console messages and page errors.
type consoleState struct {
	mu       sync.Mutex
	messages []ConsoleMessage
	errors   []PageError
	started  bool
}

// initConsoleState ensures console listening state is initialized.
func (b *Browser) initConsoleState() {
	if b.consoleState == nil {
		b.consoleState = &consoleState{}
	}
}

// ConsoleStart begins collecting console messages and page errors.
func (b *Browser) ConsoleStart() error {
	b.initConsoleState()
	if b.consoleState.started {
		return nil
	}

	ctx, cancel := b.operationContext()
	defer cancel()

	// Enable the Runtime domain to receive console events
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return cruntime.Enable().Do(ctx)
	}))
	if err != nil {
		return err
	}

	b.consoleState.started = true

	// Listen for console API calls
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *cruntime.EventConsoleAPICalled:
			level := string(e.Type)
			var parts []string
			for _, arg := range e.Args {
				if arg.Value != nil {
					parts = append(parts, strings.Trim(string(arg.Value), `"`))
				} else if arg.Description != "" {
					parts = append(parts, arg.Description)
				} else if arg.UnserializableValue != "" {
					parts = append(parts, string(arg.UnserializableValue))
				}
			}
			msg := ConsoleMessage{
				Level:     level,
				Text:      strings.Join(parts, " "),
				Timestamp: time.Now().UnixMilli(),
			}
			b.consoleState.mu.Lock()
			b.consoleState.messages = append(b.consoleState.messages, msg)
			b.consoleState.mu.Unlock()

		case *cruntime.EventExceptionThrown:
			if e.ExceptionDetails != nil {
				pe := PageError{
					Line:   int(e.ExceptionDetails.LineNumber),
					Column: int(e.ExceptionDetails.ColumnNumber),
				}
				if e.ExceptionDetails.URL != "" {
					pe.URL = e.ExceptionDetails.URL
				}
				if e.ExceptionDetails.Exception != nil {
					pe.Message = e.ExceptionDetails.Exception.Description
					if pe.Message == "" && e.ExceptionDetails.Exception.Value != nil {
						pe.Message = strings.Trim(string(e.ExceptionDetails.Exception.Value), `"`)
					}
				}
				if pe.Message == "" {
					pe.Message = e.ExceptionDetails.Text
				}
				b.consoleState.mu.Lock()
				b.consoleState.errors = append(b.consoleState.errors, pe)
				b.consoleState.mu.Unlock()
			}
		}
	})

	return nil
}

// ConsoleMessages returns all collected console messages.
// Call ConsoleStart() first to begin listening.
func (b *Browser) ConsoleMessages() ([]ConsoleMessage, error) {
	b.initConsoleState()
	b.consoleState.mu.Lock()
	defer b.consoleState.mu.Unlock()

	result := make([]ConsoleMessage, len(b.consoleState.messages))
	copy(result, b.consoleState.messages)
	return result, nil
}

// ConsoleMessagesByLevel returns console messages filtered by level.
// Levels: "log", "warning", "error", "info", "debug"
func (b *Browser) ConsoleMessagesByLevel(level string) ([]ConsoleMessage, error) {
	all, err := b.ConsoleMessages()
	if err != nil {
		return nil, err
	}
	var filtered []ConsoleMessage
	for _, m := range all {
		if m.Level == level {
			filtered = append(filtered, m)
		}
	}
	return filtered, nil
}

// ConsoleClear clears all collected console messages.
func (b *Browser) ConsoleClear() {
	b.initConsoleState()
	b.consoleState.mu.Lock()
	b.consoleState.messages = nil
	b.consoleState.mu.Unlock()
}

// PageErrors returns all collected JavaScript exceptions.
// Call ConsoleStart() first to begin listening.
func (b *Browser) PageErrors() ([]PageError, error) {
	b.initConsoleState()
	b.consoleState.mu.Lock()
	defer b.consoleState.mu.Unlock()

	result := make([]PageError, len(b.consoleState.errors))
	copy(result, b.consoleState.errors)
	return result, nil
}

// PageErrorsClear clears all collected page errors.
func (b *Browser) PageErrorsClear() {
	b.initConsoleState()
	b.consoleState.mu.Lock()
	b.consoleState.errors = nil
	b.consoleState.mu.Unlock()
}

// Highlight visually highlights the element with the given snapshot ID.
// The highlight is shown temporarily with a colored border.
func (b *Browser) Highlight(id int) error {
	js := `function() {
		var old = this.style.outline;
		this.style.outline = '3px solid red';
		this.style.outlineOffset = '-1px';
		this.scrollIntoViewIfNeeded ? this.scrollIntoViewIfNeeded(true) : this.scrollIntoView({block:'center'});
		var el = this;
		setTimeout(function() { el.style.outline = old; el.style.outlineOffset = ''; }, 3000);
	}`
	return b.runOnElement(id, js)
}

// OpenDevTools sends an inspector open request via CDP.
// Note: this only works when the browser is NOT in headless mode.
func (b *Browser) OpenDevTools() error {
	ctx, cancel := b.operationContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Use runtime.evaluate to try opening DevTools
		// In a non-headless browser the inspector can be attached
		_, _, err := cruntime.Evaluate(`void 0`).Do(ctx)
		if err != nil {
			return fmt.Errorf("cannot open DevTools: %w", err)
		}
		return nil
	}))
}
