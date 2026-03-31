package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// TabInfo represents information about a single browser tab.
type TabInfo struct {
	Index  int    `json:"index"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Active bool   `json:"active"`
}

// tabEntry tracks a chromedp context for a tab.
type tabEntry struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// initTabs initializes tab tracking if needed,
// registering the current (first) tab.
func (b *Browser) initTabs() {
	if b.tabs != nil {
		return
	}
	b.tabs = []tabEntry{{ctx: b.ctx, cancel: nil}} // first tab — never cancelled by us
	b.activeTab = 0
}

// TabList returns information about all open tabs.
func (b *Browser) TabList() ([]TabInfo, error) {
	b.initTabs()

	var tabs []TabInfo
	for i, entry := range b.tabs {
		info := TabInfo{
			Index:  i,
			Active: i == b.activeTab,
		}

		// Try to get URL and title from the tab's context
		_ = chromedp.Run(entry.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			c := chromedp.FromContext(ctx)
			if c != nil && c.Target != nil {
				tid := c.Target.TargetID
				// Use GetTargetInfo to get URL/title for this target
				targetInfo, err := target.GetTargetInfo().WithTargetID(tid).Do(ctx)
				if err == nil {
					info.URL = targetInfo.URL
					info.Title = targetInfo.Title
				}
			}
			return nil
		}))

		tabs = append(tabs, info)
	}
	return tabs, nil
}

// TabNew opens a new tab with the given URL and switches to it.
// If url is empty, opens about:blank.
func (b *Browser) TabNew(url string) error {
	b.initTabs()

	if url == "" {
		url = "about:blank"
	}

	// Create a new chromedp context (= new tab) from the allocator context
	newCtx, newCancel := chromedp.NewContext(b.allocCtx)

	// Navigate the new tab to the URL
	if err := chromedp.Run(newCtx, chromedp.Navigate(url)); err != nil {
		newCancel()
		return err
	}

	// Small wait for the target to stabilize
	time.Sleep(100 * time.Millisecond)

	b.tabs = append(b.tabs, tabEntry{ctx: newCtx, cancel: newCancel})
	b.activeTab = len(b.tabs) - 1

	// Clear snapshot cache since we're on a new page
	b.lastSnap = nil

	return nil
}

// TabClose closes the tab at the given index.
// If index is -1, closes the current tab.
func (b *Browser) TabClose(index int) error {
	b.initTabs()

	if index == -1 {
		index = b.activeTab
	}
	if index < 0 || index >= len(b.tabs) {
		return fmt.Errorf("tab index %d out of range (0-%d)", index, len(b.tabs)-1)
	}
	if len(b.tabs) <= 1 {
		return fmt.Errorf("cannot close the last tab")
	}

	// Cancel the chromedp context for this tab
	entry := b.tabs[index]
	if entry.cancel != nil {
		entry.cancel()
	} else {
		// First tab — use chromedp.Cancel
		_ = chromedp.Cancel(entry.ctx)
	}

	// Remove from our tab list
	b.tabs = append(b.tabs[:index], b.tabs[index+1:]...)

	// Adjust active tab index
	if b.activeTab >= len(b.tabs) {
		b.activeTab = len(b.tabs) - 1
	}
	if b.activeTab < 0 {
		b.activeTab = 0
	}

	b.lastSnap = nil
	return nil
}

// TabSwitch switches to the tab at the given index.
func (b *Browser) TabSwitch(index int) error {
	b.initTabs()

	if index < 0 || index >= len(b.tabs) {
		return fmt.Errorf("tab index %d out of range (0-%d)", index, len(b.tabs)-1)
	}

	b.activeTab = index
	b.lastSnap = nil

	// Activate the target in the browser
	tabCtx := b.tabs[index].ctx
	return chromedp.Run(tabCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		tid := chromedp.FromContext(ctx).Target.TargetID
		return target.ActivateTarget(tid).Do(ctx)
	}))
}

// activeContext returns the context for the active tab.
// This should be used by all browser operations instead of b.ctx directly.
func (b *Browser) activeContext() context.Context {
	if b.tabs == nil || len(b.tabs) == 0 {
		return b.ctx
	}
	return b.tabs[b.activeTab].ctx
}

// FormatTabList formats the tab list as human-readable text.
func FormatTabList(tabs []TabInfo) string {
	if len(tabs) == 0 {
		return "No tabs open\n"
	}
	var out string
	for _, t := range tabs {
		marker := " "
		if t.Active {
			marker = "*"
		}
		out += fmt.Sprintf("%s %d: %s (%s)\n", marker, t.Index, t.Title, t.URL)
	}
	return out
}
