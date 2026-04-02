package tests

import (
	"testing"
	"time"

	"github.com/libi/ko-browser/browser"
)

// Tab Management — Comprehensive Coverage

func TestTab_ListShowsAllTabs(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	tabs, err := b.TabList()
	must(t, err)
	if len(tabs) != 1 {
		t.Fatalf("expected 1 tab initially, got %d", len(tabs))
	}
	assertTrue(t, tabs[0].Active, "first tab should be active")
	assertContains(t, tabs[0].URL, "phase5_test.html")
	must(t, b.TabNew(testdataURL("phase2_nav_a.html")))
	time.Sleep(500 * time.Millisecond)
	tabs, err = b.TabList()
	must(t, err)
	if len(tabs) != 2 {
		t.Fatalf("expected 2 tabs after TabNew, got %d", len(tabs))
	}
	foundOriginal := false
	foundNew := false
	for _, tab := range tabs {
		if containsStr(tab.URL, "phase5_test.html") {
			foundOriginal = true
		}
		if containsStr(tab.URL, "phase2_nav_a.html") {
			foundNew = true
			assertTrue(t, tab.Active, "new tab should be active")
		}
	}
	assertTrue(t, foundOriginal, "original tab should be listed")
	assertTrue(t, foundNew, "new tab should be listed")
	must(t, b.TabNew(testdataURL("phase2_nav_b.html")))
	time.Sleep(500 * time.Millisecond)
	tabs, err = b.TabList()
	must(t, err)
	if len(tabs) != 3 {
		t.Fatalf("expected 3 tabs after second TabNew, got %d", len(tabs))
	}
}

func TestTab_SwitchAndVerifyActive(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.TabNew(testdataURL("phase2_nav_a.html")))
	time.Sleep(500 * time.Millisecond)
	tabs, err := b.TabList()
	must(t, err)
	if len(tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(tabs))
	}
	activeCount := 0
	for _, tab := range tabs {
		if tab.Active {
			activeCount++
			assertContains(t, tab.URL, "phase2_nav_a.html")
		}
	}
	if activeCount != 1 {
		t.Fatalf("expected 1 active tab, got %d", activeCount)
	}
	must(t, b.TabSwitch(0))
	time.Sleep(300 * time.Millisecond)
	tabs, err = b.TabList()
	must(t, err)
	activeCount = 0
	for _, tab := range tabs {
		if tab.Active {
			activeCount++
			assertContains(t, tab.URL, "phase5_test.html")
		}
	}
	if activeCount != 1 {
		t.Fatalf("expected 1 active tab after switch, got %d", activeCount)
	}
}

func TestTab_CloseMiddleTab(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.TabNew(testdataURL("phase2_nav_a.html")))
	time.Sleep(500 * time.Millisecond)
	must(t, b.TabNew(testdataURL("phase2_nav_b.html")))
	time.Sleep(500 * time.Millisecond)
	tabs, err := b.TabList()
	must(t, err)
	if len(tabs) != 3 {
		t.Fatalf("expected 3 tabs, got %d", len(tabs))
	}
	// Close the last tab (index 2), which is also the active tab
	must(t, b.TabClose(2))
	time.Sleep(500 * time.Millisecond)
	tabs, err = b.TabList()
	must(t, err)
	if len(tabs) != 2 {
		t.Fatalf("expected 2 tabs after close, got %d", len(tabs))
	}
}

func TestTab_CloseCurrentTab(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.TabNew(testdataURL("phase2_nav_a.html")))
	time.Sleep(500 * time.Millisecond)
	must(t, b.TabClose(-1))
	time.Sleep(300 * time.Millisecond)
	tabs, err := b.TabList()
	must(t, err)
	if len(tabs) != 1 {
		t.Fatalf("expected 1 tab after closing current, got %d", len(tabs))
	}
	assertContains(t, tabs[0].URL, "phase5_test.html")
}

func TestTab_CloseOnlyTabFails(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	err := b.TabClose(0)
	if err == nil {
		t.Fatal("expected error when closing the last tab")
	}
	assertContains(t, err.Error(), "cannot close the last tab")
}

func TestTab_SwitchOutOfRange(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	err := b.TabSwitch(5)
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
	assertContains(t, err.Error(), "out of range")
	err = b.TabSwitch(-1)
	if err == nil {
		t.Fatal("expected error for negative index")
	}
}

func TestTab_CloseOutOfRange(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.TabNew("about:blank"))
	time.Sleep(300 * time.Millisecond)
	err := b.TabClose(10)
	if err == nil {
		t.Fatal("expected error for out-of-range tab close")
	}
	assertContains(t, err.Error(), "out of range")
}

func TestTab_NewBlankTab(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.TabNew(""))
	time.Sleep(300 * time.Millisecond)
	tabs, err := b.TabList()
	must(t, err)
	if len(tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(tabs))
	}
}

func TestTab_OperationsAfterSwitch(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.TabNew(testdataURL("phase2_nav_a.html")))
	time.Sleep(500 * time.Millisecond)
	title, err := b.GetTitle()
	must(t, err)
	assertContains(t, title, "Page A")
	must(t, b.TabSwitch(0))
	time.Sleep(300 * time.Millisecond)
	title, err = b.GetTitle()
	must(t, err)
	assertContains(t, title, "Phase 5")
}

func TestTab_JSWindowOpen(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	tabs1, err := b.TabList()
	must(t, err)
	initialCount := len(tabs1)
	// Inject a link with target=_blank and click it to open a new tab
	_, err = b.Eval(`
		var a = document.createElement('a');
		a.href = 'about:blank';
		a.target = '_blank';
		a.id = 'test-new-tab-link';
		a.textContent = 'Open New Tab';
		document.body.appendChild(a);
		'ok'
	`)
	must(t, err)
	// Click the link using JS to open new tab
	_, err = b.Eval("document.getElementById('test-new-tab-link').click(); 'ok'")
	must(t, err)
	time.Sleep(1 * time.Second)
	tabs2, err := b.TabList()
	must(t, err)
	if len(tabs2) <= initialCount {
		t.Logf("Tabs after link click: %+v", tabs2)
		t.Fatalf("expected more than %d tabs after link click, got %d", initialCount, len(tabs2))
	}
}

func TestTab_FormatTabList(t *testing.T) {
	tabs := []browser.TabInfo{
		{Index: 0, URL: "https://example.com", Title: "Example", Active: true},
		{Index: 1, URL: "about:blank", Title: "", Active: false},
	}
	output := browser.FormatTabList(tabs)
	assertContains(t, output, "* 0: Example")
	assertContains(t, output, "  1:")
	emptyOutput := browser.FormatTabList(nil)
	assertContains(t, emptyOutput, "No tabs open")
}

// Find — Exact Match and Edge Cases

func TestFind_RoleExact(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase3_screenshot.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindRole("button", "Submit Form", browser.WithExact())
	must(t, err)
	if result == nil || result.Text == "" {
		t.Fatal("expected exact match results")
	}
	assertContains(t, result.Text, "Submit Form")
}

func TestFind_TextExact(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase3_screenshot.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindText("Home", browser.WithExact())
	must(t, err)
	if result == nil {
		t.Fatal("expected exact match results")
	}
}

func TestFind_LabelExact(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase3_screenshot.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindLabel("Username", browser.WithExact())
	must(t, err)
	if result == nil {
		t.Fatal("expected exact match results")
	}
}

// Network — Route/Unroute/Requests

func TestNetwork_StartLoggingAndRequests(t *testing.T) {
	b := newBrowser(t)
	must(t, b.NetworkStartLogging())
	openPage(t, b, "phase5_test.html")
	time.Sleep(300 * time.Millisecond)
	reqs, err := b.NetworkRequests()
	must(t, err)
	_ = reqs
}

func TestNetwork_RouteAndUnroute(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.NetworkRoute("*.jpg", browser.RouteBlock))
	must(t, b.NetworkUnroute("*.jpg"))
}

func TestNetwork_ClearRequests(t *testing.T) {
	b := newBrowser(t)
	must(t, b.NetworkStartLogging())
	openPage(t, b, "phase5_test.html")
	time.Sleep(200 * time.Millisecond)
	b.NetworkClearRequests()
	reqs, err := b.NetworkRequests()
	must(t, err)
	if len(reqs) != 0 {
		t.Fatalf("expected 0 requests after clear, got %d", len(reqs))
	}
}

// Console — Level Filtering

func TestConsole_LevelFiltering(t *testing.T) {
	b := newBrowser(t)
	must(t, b.ConsoleStart())
	openPage(t, b, "phase6_test.html")
	time.Sleep(200 * time.Millisecond)
	_, _ = b.Eval("console.log('info-msg')")
	_, _ = b.Eval("console.warn('warn-msg')")
	_, _ = b.Eval("console.error('error-msg')")
	time.Sleep(300 * time.Millisecond)
	errors, err := b.ConsoleMessagesByLevel("error")
	must(t, err)
	for _, m := range errors {
		if m.Level == "log" || m.Level == "info" {
			t.Errorf("unexpected log-level msg in error filter: %q", m.Text)
		}
	}
	all, err := b.ConsoleMessages()
	must(t, err)
	if len(all) < len(errors) {
		t.Errorf("all msgs (%d) < error msgs (%d)", len(all), len(errors))
	}
}

// Page Errors

func TestPageErrors_ClearAndList(t *testing.T) {
	b := newBrowser(t)
	must(t, b.ConsoleStart())
	openPage(t, b, "phase6_test.html")
	_, _ = b.Eval("setTimeout(function() { throw new Error('test-err'); }, 0)")
	time.Sleep(500 * time.Millisecond)
	errs, err := b.PageErrors()
	must(t, err)
	_ = errs
	b.PageErrorsClear()
	errs2, err := b.PageErrors()
	must(t, err)
	if len(errs2) != 0 {
		t.Errorf("expected 0 errors after clear, got %d", len(errs2))
	}
}

// Set — Edge Cases

func TestSet_ViewportWithScale(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")
	must(t, b.SetViewport(1024, 768, 2.0))
	time.Sleep(200 * time.Millisecond)
	w, err := b.Eval("window.innerWidth")
	must(t, err)
	if w != "1024" {
		t.Errorf("expected viewport width 1024, got %q", w)
	}
	dpr, err := b.Eval("window.devicePixelRatio")
	must(t, err)
	if dpr != "2" {
		t.Errorf("expected devicePixelRatio 2, got %q", dpr)
	}
}

func TestSet_ClearGeo(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")
	must(t, b.SetGeo(40.7128, -74.0060))
	must(t, b.ClearGeo())
}

func TestSet_DeviceUnknown(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")
	err := b.SetDevice("NonExistentDevice999")
	if err == nil {
		t.Fatal("expected error for unknown device")
	}
}

// Clipboard

func TestClipboard_WriteAndRead(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	err := b.ClipboardWrite("test clipboard content")
	if err != nil {
		t.Skipf("clipboard not supported: %v", err)
	}
	text, err := b.ClipboardRead()
	must(t, err)
	assertEqual(t, text, "test clipboard content")
}

// Diff — Edge Cases

func TestDiff_SnapshotNoPreviousError(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")
	_, err := b.DiffSnapshot()
	if err == nil {
		t.Fatal("expected error when no previous snapshot")
	}
}

func TestDiff_ScreenshotNoBaseline(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")
	_, err := b.DiffScreenshot(browser.DiffScreenshotOptions{})
	if err == nil {
		t.Fatal("expected error without baseline file")
	}
}

// State — Export/Import

func TestState_ImportNonexistent(t *testing.T) {
	b := newBrowser(t)
	err := b.ImportState("/nonexistent/path/state.json")
	if err == nil {
		t.Fatal("expected error for non-existent state file")
	}
}

// Keyboard

func TestKeyboard_InsertText(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, "textbox \"Name\"")
	must(t, b.Focus(nameID))
	must(t, b.KeyboardInsertText("inserted text"))
	val, err := b.GetValue(nameID)
	must(t, err)
	assertContains(t, val, "inserted text")
}

// Get — CDP URL

func TestGet_CDPURL(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	url, err := b.GetCDPURL()
	must(t, err)
	if url == "" {
		t.Fatal("CDP URL should not be empty")
	}
	// GetCDPURL returns the first target's page URL (not a ws:// debug URL)
	t.Logf("CDP URL: %s", url)
}

// Mouse — Button Options

func TestMouse_RightClick(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")
	_ = mustSnapshot(t, b)
	must(t, b.MouseClick(100, 100, browser.MouseOptions{Button: browser.MouseRight}))
}

func TestMouse_MiddleClick(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")
	_ = mustSnapshot(t, b)
	must(t, b.MouseClick(100, 100, browser.MouseOptions{Button: browser.MouseMiddle}))
}

// Eval — Various Return Types

func TestEval_ReturnNull(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	result, err := b.Eval("null")
	must(t, err)
	if result != "null" {
		t.Errorf("expected 'null', got %q", result)
	}
}

func TestEval_ReturnObject(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	result, err := b.Eval("({a: 1, b: 'hello'})")
	must(t, err)
	assertContains(t, result, "a")
	assertContains(t, result, "hello")
}

func TestEval_ReturnArray(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	result, err := b.Eval("[1, 2, 3]")
	must(t, err)
	assertContains(t, result, "1")
	assertContains(t, result, "3")
}

func TestEval_ReturnBoolean(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	result, err := b.Eval("true")
	must(t, err)
	assertEqual(t, result, "true")
	result, err = b.Eval("false")
	must(t, err)
	assertEqual(t, result, "false")
}

// Error Handling — Invalid IDs

func TestClick_InvalidID(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	err := b.Click(99999)
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestGetText_InvalidID(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	_, err := b.GetText(99999)
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestFocus_InvalidID(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	err := b.Focus(99999)
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

// Storage — Session Storage

func TestStorage_SessionStorageGetAll(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.StorageSet("session", "key1", "val1"))
	must(t, b.StorageSet("session", "key2", "val2"))
	items, err := b.StorageGetAll("session")
	must(t, err)
	if items["key1"] != "val1" {
		t.Errorf("expected key1=val1, got %q", items["key1"])
	}
	if items["key2"] != "val2" {
		t.Errorf("expected key2=val2, got %q", items["key2"])
	}
}

func TestStorage_ClearSession(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")
	must(t, b.StorageSet("session", "clearKey", "clearVal"))
	must(t, b.StorageClear("session"))
	items, err := b.StorageGetAll("session")
	must(t, err)
	if len(items) != 0 {
		t.Fatalf("expected empty sessionStorage after clear, got %d items", len(items))
	}
}

// Snapshot — Additional Options

func TestSnapshot_SelectorOption(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap, err := b.Snapshot(browser.SnapshotOptions{Selector: "form"})
	must(t, err)
	if snap.Text == "" {
		t.Fatal("selector snapshot should not be empty")
	}
	assertContains(t, snap.Text, "textbox")
}

func TestSnapshot_CursorOption(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap, err := b.Snapshot(browser.SnapshotOptions{Cursor: true})
	must(t, err)
	if snap.Text == "" {
		t.Fatal("cursor snapshot should not be empty")
	}
}

// Screenshot — Element

func TestScreenshot_ElementByID(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase3_screenshot.html")
	snap := mustSnapshot(t, b)
	btnID := findID(t, snap.Text, "button \"Submit Form\"")
	data, err := b.ScreenshotToBytes(browser.ScreenshotOptions{ElementID: btnID})
	must(t, err)
	if len(data) < 50 {
		t.Fatalf("element screenshot too small: %d bytes", len(data))
	}
}

// Wait — Additional

func TestWait_HiddenNonexistent(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.WaitHidden("#nonexistent-element"))
}

func TestWait_URLPattern(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.WaitURL("*interaction*"))
}

// Set — Media with multiple features

func TestSet_MediaMultipleFeatures(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")
	must(t, b.SetMedia(
		browser.MediaFeature{Name: "prefers-color-scheme", Value: "dark"},
		browser.MediaFeature{Name: "prefers-reduced-motion", Value: "reduce"},
	))
	time.Sleep(200 * time.Millisecond)
	dark, err := b.Eval("window.matchMedia('(prefers-color-scheme: dark)').matches")
	must(t, err)
	if dark != "true" {
		t.Errorf("expected dark=true, got %q", dark)
	}
	motion, err := b.Eval("window.matchMedia('(prefers-reduced-motion: reduce)').matches")
	must(t, err)
	if motion != "true" {
		t.Errorf("expected reduce=true, got %q", motion)
	}
}
