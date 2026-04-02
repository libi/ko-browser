package tests

import (
	"testing"
	"time"

	"github.com/libi/ko-browser/browser"
)

// ---------- Tab Management ----------

func TestTab_List(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	tabs, err := b.TabList()
	if err != nil {
		t.Fatalf("TabList: %v", err)
	}

	if len(tabs) < 1 {
		t.Fatalf("expected at least 1 tab, got %d", len(tabs))
	}

	// The first tab should be active and have our test page
	found := false
	for _, tab := range tabs {
		if tab.Active {
			found = true
			assertContains(t, tab.URL, "phase5_test.html")
			break
		}
	}
	if !found {
		t.Fatal("no active tab found")
	}
}

func TestTab_New(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// Get initial tab count
	tabs1, err := b.TabList()
	if err != nil {
		t.Fatalf("TabList: %v", err)
	}
	initialCount := len(tabs1)

	// Open new tab
	if err := b.TabNew(testdataURL("phase2_nav_a.html")); err != nil {
		t.Fatalf("TabNew: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Verify tab count increased
	tabs2, err := b.TabList()
	if err != nil {
		t.Fatalf("TabList after new: %v", err)
	}
	if len(tabs2) != initialCount+1 {
		t.Fatalf("expected %d tabs, got %d", initialCount+1, len(tabs2))
	}

	// Check that a tab has the new URL
	found := false
	for _, tab := range tabs2 {
		if tab.URL != "" && (tab.URL == testdataURL("phase2_nav_a.html") ||
			containsStr(tab.URL, "phase2_nav_a.html")) {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Tabs: %+v", tabs2)
		t.Fatal("new tab with phase2_nav_a.html not found")
	}
}

func TestTab_Switch(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// Open a second tab
	if err := b.TabNew(testdataURL("phase2_nav_a.html")); err != nil {
		t.Fatalf("TabNew: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Switch back to first tab (index 0)
	if err := b.TabSwitch(0); err != nil {
		t.Fatalf("TabSwitch(0): %v", err)
	}
	time.Sleep(300 * time.Millisecond)
}

func TestTab_Close(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// Open a second tab
	if err := b.TabNew(testdataURL("phase2_nav_a.html")); err != nil {
		t.Fatalf("TabNew: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	tabs1, err := b.TabList()
	if err != nil {
		t.Fatalf("TabList: %v", err)
	}
	if len(tabs1) < 2 {
		t.Fatalf("expected at least 2 tabs, got %d", len(tabs1))
	}
	countBefore := len(tabs1)

	// Close the last tab by index
	if err := b.TabClose(countBefore - 1); err != nil {
		t.Fatalf("TabClose: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	tabs2, err := b.TabList()
	if err != nil {
		t.Fatalf("TabList after close: %v", err)
	}
	if len(tabs2) != countBefore-1 {
		t.Fatalf("expected %d tabs after close, got %d", countBefore-1, len(tabs2))
	}
}

// ---------- Cookies ----------

func TestCookie_SetAndGet(t *testing.T) {
	b := newBrowser(t)
	// Navigate to an http page (file:// URLs don't support cookies well)
	// Use a simple html page with a server, or test via JS cookies
	openPage(t, b, "phase5_test.html")

	// Set a cookie via the browser API
	err := b.CookieSet(browser.CookieInfo{
		Name:   "testCookie",
		Value:  "hello123",
		Domain: "localhost",
		Path:   "/",
	})
	if err != nil {
		t.Fatalf("CookieSet: %v", err)
	}

	// Get cookies
	cookies, err := b.CookiesGet()
	if err != nil {
		t.Fatalf("CookiesGet: %v", err)
	}

	found := false
	for _, c := range cookies {
		if c.Name == "testCookie" && c.Value == "hello123" {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Cookies: %+v", cookies)
		t.Fatal("testCookie not found in cookies")
	}
}

func TestCookie_Delete(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// Set and then delete a cookie
	must(t, b.CookieSet(browser.CookieInfo{
		Name:   "toDelete",
		Value:  "deleteMe",
		Domain: "localhost",
		Path:   "/",
	}))

	must(t, b.CookieDelete("toDelete"))

	cookies, err := b.CookiesGet()
	if err != nil {
		t.Fatalf("CookiesGet: %v", err)
	}

	for _, c := range cookies {
		if c.Name == "toDelete" {
			t.Fatal("cookie 'toDelete' should have been deleted")
		}
	}
}

func TestCookie_Clear(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// Set a cookie
	must(t, b.CookieSet(browser.CookieInfo{
		Name:   "clearMe",
		Value:  "val",
		Domain: "localhost",
		Path:   "/",
	}))

	// Clear all cookies
	must(t, b.CookiesClear())

	cookies, err := b.CookiesGet()
	if err != nil {
		t.Fatalf("CookiesGet: %v", err)
	}

	if len(cookies) != 0 {
		t.Fatalf("expected 0 cookies after clear, got %d: %+v", len(cookies), cookies)
	}
}

// ---------- Storage ----------

func TestStorage_SetAndGet(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// Set localStorage
	must(t, b.StorageSet("local", "myKey", "myValue"))

	// Get localStorage
	val, err := b.StorageGet("local", "myKey")
	if err != nil {
		t.Fatalf("StorageGet: %v", err)
	}
	assertEqual(t, val, "myValue")
}

func TestStorage_GetPreset(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// The test page pre-sets 'presetKey' = 'presetValue'
	val, err := b.StorageGet("local", "presetKey")
	if err != nil {
		t.Fatalf("StorageGet: %v", err)
	}
	assertEqual(t, val, "presetValue")
}

func TestStorage_SessionStorage(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// Set session storage via API
	must(t, b.StorageSet("session", "sessAPIKey", "sessAPIVal"))

	val, err := b.StorageGet("session", "sessAPIKey")
	if err != nil {
		t.Fatalf("StorageGet session: %v", err)
	}
	assertEqual(t, val, "sessAPIVal")
}

func TestStorage_Delete(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	must(t, b.StorageSet("local", "delKey", "delVal"))
	must(t, b.StorageDelete("local", "delKey"))

	val, err := b.StorageGet("local", "delKey")
	if err != nil {
		t.Fatalf("StorageGet after delete: %v", err)
	}
	// After deletion, value should be empty/null
	if val != "" && val != "null" {
		t.Fatalf("expected empty after delete, got %q", val)
	}
}

func TestStorage_Clear(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// Clear local storage
	must(t, b.StorageClear("local"))

	// Verify it's empty
	items, err := b.StorageGetAll("local")
	if err != nil {
		t.Fatalf("StorageGetAll: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty storage after clear, got %d items: %+v", len(items), items)
	}
}

func TestStorage_GetAll(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase5_test.html")

	// The page pre-sets presetKey and anotherKey
	items, err := b.StorageGetAll("local")
	if err != nil {
		t.Fatalf("StorageGetAll: %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("expected at least 2 items in localStorage, got %d: %+v", len(items), items)
	}
	if items["presetKey"] != "presetValue" {
		t.Errorf("presetKey: got %q, want %q", items["presetKey"], "presetValue")
	}
	if items["anotherKey"] != "anotherValue" {
		t.Errorf("anotherKey: got %q, want %q", items["anotherKey"], "anotherValue")
	}
}

// Helper: check if string contains substring (non-fatal)
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
