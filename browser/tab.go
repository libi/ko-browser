package browser

import (
	"context"
	"fmt"
	"slices"
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
	ctx      context.Context
	cancel   context.CancelFunc
	targetID target.ID
}

// initTabs initializes tab tracking if needed,
// registering the current (first) tab.
func (b *Browser) initTabs() {
	if b.tabs != nil {
		return
	}
	var tid target.ID
	_ = chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		c := chromedp.FromContext(ctx)
		if c != nil && c.Target != nil {
			tid = c.Target.TargetID
		}
		return nil
	}))
	b.tabs = []tabEntry{{ctx: b.ctx, cancel: nil, targetID: tid}}
	b.activeTab = 0
}

// getPageTargets returns all "page" type targets from the browser via CDP.
// Uses b.ctx (the first tab) to query so that all tabs in the same browser
// context are visible.
func (b *Browser) getPageTargets() ([]*target.Info, error) {
	var allTargets []*target.Info
	if err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		allTargets, err = target.GetTargets().Do(ctx)
		return err
	})); err != nil {
		return nil, err
	}
	var pages []*target.Info
	for _, t := range allTargets {
		if t.Type == "page" {
			pages = append(pages, t)
		}
	}
	return pages, nil
}

// orderedPageTargets returns page targets in the same stable order that TabList
// exposes to callers: tracked tabs first, then any untracked tabs.
func (b *Browser) orderedPageTargets() ([]*target.Info, error) {
	pages, err := b.getPageTargets()
	if err != nil {
		return nil, err
	}

	byID := make(map[target.ID]*target.Info, len(pages))
	for _, p := range pages {
		byID[p.TargetID] = p
	}

	ordered := make([]*target.Info, 0, len(pages))
	for _, entry := range b.tabs {
		if p, ok := byID[entry.targetID]; ok {
			ordered = append(ordered, p)
			delete(byID, entry.targetID)
		}
	}

	var remainder []*target.Info
	for _, p := range pages {
		if _, ok := byID[p.TargetID]; ok {
			remainder = append(remainder, p)
		}
	}
	slices.SortFunc(remainder, func(a, b *target.Info) int {
		if a.URL != b.URL {
			if a.URL < b.URL {
				return -1
			}
			return 1
		}
		if a.TargetID < b.TargetID {
			return -1
		}
		if a.TargetID > b.TargetID {
			return 1
		}
		return 0
	})

	return append(ordered, remainder...), nil
}

// TabList returns information about all open tabs.
// It queries the browser via CDP to discover ALL actual tabs,
// not just the ones tracked in b.tabs.
func (b *Browser) TabList() ([]TabInfo, error) {
	b.initTabs()

	pages, err := b.orderedPageTargets()
	if err != nil {
		return nil, fmt.Errorf("list tabs: %w", err)
	}

	activeTID := b.tabs[b.activeTab].targetID

	var tabs []TabInfo
	for i, p := range pages {
		tabs = append(tabs, TabInfo{
			Index:  i,
			URL:    p.URL,
			Title:  p.Title,
			Active: p.TargetID == activeTID,
		})
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

	// Create a new tab within the same browser context by using b.ctx as parent.
	// Using b.ctx (instead of b.allocCtx) ensures the new tab is in the same
	// browser context, so target.GetTargets() can discover all tabs.
	newCtx, newCancel := chromedp.NewContext(b.ctx)

	// Navigate the new tab to the URL
	if err := chromedp.Run(newCtx, chromedp.Navigate(url)); err != nil {
		newCancel()
		return err
	}

	// Small wait for the target to stabilize
	time.Sleep(100 * time.Millisecond)

	// Get the target ID of the new tab
	var tid target.ID
	_ = chromedp.Run(newCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		c := chromedp.FromContext(ctx)
		if c != nil && c.Target != nil {
			tid = c.Target.TargetID
		}
		return nil
	}))

	b.tabs = append(b.tabs, tabEntry{ctx: newCtx, cancel: newCancel, targetID: tid})
	b.activeTab = len(b.tabs) - 1

	// Clear snapshot cache since we're on a new page
	b.lastSnap = nil

	return nil
}

// TabClose closes the tab at the given index (from TabList).
// If index is -1, closes the current tab.
func (b *Browser) TabClose(index int) error {
	b.initTabs()

	pages, err := b.orderedPageTargets()
	if err != nil {
		return err
	}

	if index == -1 {
		// Find the index of the active tab in the page targets
		activeTID := b.tabs[b.activeTab].targetID
		for i, p := range pages {
			if p.TargetID == activeTID {
				index = i
				break
			}
		}
		if index == -1 {
			return fmt.Errorf("could not determine current tab")
		}
	}

	if index < 0 || index >= len(pages) {
		return fmt.Errorf("tab index %d out of range (0-%d)", index, len(pages)-1)
	}
	if len(pages) <= 1 {
		return fmt.Errorf("cannot close the last tab")
	}

	closeTID := pages[index].TargetID

	for i, entry := range b.tabs {
		if entry.targetID == closeTID {
			if entry.cancel != nil {
				if err := chromedp.Cancel(entry.ctx); err != nil {
					return err
				}
			} else {
				if err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					return target.CloseTarget(closeTID).Do(ctx)
				})); err != nil {
					return err
				}
			}
			b.tabs = append(b.tabs[:i], b.tabs[i+1:]...)
			switch {
			case len(b.tabs) == 0:
				b.activeTab = 0
			case i < b.activeTab:
				b.activeTab--
			case i == b.activeTab:
				if b.activeTab >= len(b.tabs) {
					b.activeTab = len(b.tabs) - 1
				}
			}
			b.lastSnap = nil
			return nil
		}
	}

	if err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return target.CloseTarget(closeTID).Do(ctx)
	})); err != nil {
		return err
	}
	b.lastSnap = nil
	return nil
}

// TabSwitch switches to the tab at the given index (from TabList).
func (b *Browser) TabSwitch(index int) error {
	b.initTabs()

	pages, err := b.orderedPageTargets()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(pages) {
		return fmt.Errorf("tab index %d out of range (0-%d)", index, len(pages)-1)
	}

	switchTID := pages[index].TargetID

	// Check if we already track this target
	for i, entry := range b.tabs {
		if entry.targetID == switchTID {
			b.activeTab = i
			b.lastSnap = nil
			return chromedp.Run(entry.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
				return target.ActivateTarget(switchTID).Do(ctx)
			}))
		}
	}

	// Target not tracked — create a new context attached to this target
	newCtx, newCancel := chromedp.NewContext(b.ctx, chromedp.WithTargetID(switchTID))
	if err := chromedp.Run(newCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		return nil
	})); err != nil {
		newCancel()
		return fmt.Errorf("attach to tab: %w", err)
	}

	b.tabs = append(b.tabs, tabEntry{ctx: newCtx, cancel: newCancel, targetID: switchTID})
	b.activeTab = len(b.tabs) - 1
	b.lastSnap = nil

	return chromedp.Run(newCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		return target.ActivateTarget(switchTID).Do(ctx)
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
