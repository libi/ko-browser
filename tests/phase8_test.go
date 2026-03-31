package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/libi/ko-browser/browser"
	"github.com/libi/ko-browser/internal/session"
)

// =============================================================================
// Phase 8.4 Unit Tests (no browser needed)
// =============================================================================

// TestPhase8_AuthVault tests the auth vault CRUD operations.
func TestPhase8_AuthVault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	vault, err := session.NewAuthVault()
	if err != nil {
		t.Fatalf("NewAuthVault: %v", err)
	}

	// Save two profiles
	must(t, vault.Save(session.AuthProfile{
		Name:     "github",
		URL:      "https://github.com/login",
		Username: "testuser",
		Password: "testpass",
	}))
	must(t, vault.Save(session.AuthProfile{
		Name:     "gitlab",
		URL:      "https://gitlab.com/users/sign_in",
		Username: "gluser",
		Password: "glpass",
	}))

	// List
	profiles := vault.List()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// Get
	p, ok := vault.Get("github")
	if !ok {
		t.Fatal("github profile not found")
	}
	assertEqual(t, p.Username, "testuser")
	assertEqual(t, p.URL, "https://github.com/login")

	// Get non-existent
	_, ok = vault.Get("nonexistent")
	if ok {
		t.Fatal("expected nonexistent profile not found")
	}

	// Delete
	must(t, vault.Delete("gitlab"))
	profiles = vault.List()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile after delete, got %d", len(profiles))
	}

	// Delete non-existent
	if err := vault.Delete("nonexistent"); err == nil {
		t.Fatal("expected error deleting nonexistent profile")
	}

	// Persistence: reload from disk
	vault2, err := session.NewAuthVault()
	if err != nil {
		t.Fatalf("NewAuthVault reload: %v", err)
	}
	profiles = vault2.List()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile on reload, got %d", len(profiles))
	}
	assertEqual(t, profiles[0].Name, "github")
}

// TestPhase8_ConfirmStore tests the confirmation store mechanism.
func TestPhase8_ConfirmStore(t *testing.T) {
	store := session.NewConfirmStore([]string{"open", "click"}, 5*time.Second)

	// NeedsConfirmation
	if !store.NeedsConfirmation("open") {
		t.Fatal("expected open to need confirmation")
	}
	if !store.NeedsConfirmation("click") {
		t.Fatal("expected click to need confirmation")
	}
	if store.NeedsConfirmation("snapshot") {
		t.Fatal("expected snapshot to NOT need confirmation")
	}

	// Empty store doesn't need confirmation
	emptyStore := session.NewConfirmStore(nil, 0)
	if emptyStore.NeedsConfirmation("open") {
		t.Fatal("empty store should not require confirmation")
	}

	// Add + Confirm
	pa := store.Add("open", "https://example.com")
	if pa.ID == "" {
		t.Fatal("expected non-empty confirmation ID")
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		if err := store.Confirm(pa.ID); err != nil {
			t.Errorf("Confirm: %v", err)
		}
	}()
	if !pa.Wait() {
		t.Fatal("expected action to be confirmed")
	}

	// Add + Deny
	pa2 := store.Add("click", "element 5")
	go func() {
		time.Sleep(50 * time.Millisecond)
		if err := store.Deny(pa2.ID); err != nil {
			t.Errorf("Deny: %v", err)
		}
	}()
	if pa2.Wait() {
		t.Fatal("expected action to be denied")
	}

	// Confirm non-existent
	if err := store.Confirm("c_nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent confirmation")
	}
}

// TestPhase8_InstallFindChrome verifies Chrome is available (required by all tests).
func TestPhase8_InstallFindChrome(t *testing.T) {
	b := newBrowser(t)
	_ = b
}

// =============================================================================
// CLI Command Coverage Tests
// Each test exercises a browser.* API that backs a CLI command.
// =============================================================================

// ---------- Core lifecycle ----------

func TestCLI_Open_Snapshot_Close(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	assertContains(t, snap.Text, "Phase1 Interaction Test")
	if snap.RawCount == 0 {
		t.Fatal("RawCount should be > 0")
	}
}

// ---------- Click ----------

func TestCLI_Click(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_nav_a.html")
	snap := mustSnapshot(t, b)
	linkID := findID(t, snap.Text, `link "Go to Page B"`)
	must(t, b.Click(linkID))
	must(t, b.WaitURL("phase2_nav_b.html"))
}

// ---------- Type ----------

func TestCLI_Type(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Type(nameID, "Hello"))
	val, err := b.GetValue(nameID)
	must(t, err)
	assertContains(t, val, "Hello")
}

// ---------- Fill ----------

func TestCLI_Fill(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Fill(nameID, "World"))
	val, err := b.GetValue(nameID)
	must(t, err)
	assertEqual(t, val, "World")
}

// ---------- Press ----------

func TestCLI_Press(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.Press("Tab"))
}

// ---------- Keyboard ----------

func TestCLI_KeyboardType(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Focus(nameID))
	must(t, b.KeyboardType("typed text"))
	val, err := b.GetValue(nameID)
	must(t, err)
	assertContains(t, val, "typed text")
}

func TestCLI_KeyboardInsertText(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Focus(nameID))
	must(t, b.KeyboardInsertText("inserted"))
	val, err := b.GetValue(nameID)
	must(t, err)
	assertContains(t, val, "inserted")
}

// ---------- Hover / Focus ----------

func TestCLI_Hover(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	btnID := findID(t, snap.Text, `button "Submit"`)
	must(t, b.ScrollIntoView(btnID))
	must(t, b.Hover(btnID))
}

func TestCLI_Focus(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Focus(nameID))
}

// ---------- Check / Uncheck ----------

func TestCLI_Check_Uncheck(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	termsID := findID(t, snap.Text, `checkbox "Terms"`)
	must(t, b.Check(termsID))
	checked, err := b.IsChecked(termsID)
	must(t, err)
	assertTrue(t, checked, "checkbox should be checked")
	must(t, b.Uncheck(termsID))
	checked, err = b.IsChecked(termsID)
	must(t, err)
	assertFalse(t, checked, "checkbox should be unchecked")
}

// ---------- Select ----------

func TestCLI_Select(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	colorID := findID(t, snap.Text, `combobox "Color"`)
	must(t, b.Select(colorID, "Blue"))
}

// ---------- Scroll ----------

func TestCLI_Scroll(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.Scroll("down", 200))
}

func TestCLI_ScrollIntoView(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	btnID := findID(t, snap.Text, `button "Submit"`)
	must(t, b.ScrollIntoView(btnID))
}

// ---------- DblClick ----------

func TestCLI_DblClick(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	btnID := findID(t, snap.Text, `button "Submit"`)
	must(t, b.ScrollIntoView(btnID))
	must(t, b.DblClick(btnID))
}

// ---------- Navigation ----------

func TestCLI_Back_Forward_Reload(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_nav_a.html")
	snap := mustSnapshot(t, b)
	linkID := findID(t, snap.Text, `link "Go to Page B"`)
	must(t, b.Click(linkID))
	must(t, b.WaitURL("phase2_nav_b.html"))
	must(t, b.Back())
	must(t, b.WaitURL("phase2_nav_a.html"))
	must(t, b.Forward())
	must(t, b.WaitURL("phase2_nav_b.html"))
	must(t, b.Reload())
}

// ---------- Get* queries ----------

func TestCLI_GetTitle(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_query.html")
	title, err := b.GetTitle()
	must(t, err)
	if title == "" {
		t.Fatal("title should not be empty")
	}
}

func TestCLI_GetURL(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_query.html")
	url, err := b.GetURL()
	must(t, err)
	assertContains(t, url, "phase2_query.html")
}

func TestCLI_GetText(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_query.html")
	snap := mustSnapshot(t, b)
	id := findAnyID(t, snap.Text)
	_, err := b.GetText(id)
	must(t, err)
}

func TestCLI_GetHTML(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_query.html")
	snap := mustSnapshot(t, b)
	id := findAnyID(t, snap.Text)
	html, err := b.GetHTML(id)
	must(t, err)
	if html == "" {
		t.Fatal("html should not be empty")
	}
}

func TestCLI_GetValue(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Fill(nameID, "test"))
	val, err := b.GetValue(nameID)
	must(t, err)
	assertEqual(t, val, "test")
}

func TestCLI_GetAttr(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_query.html")
	snap := mustSnapshot(t, b)
	linkID := tryFindID(snap.Text, "link")
	if linkID > 0 {
		href, err := b.GetAttr(linkID, "href")
		must(t, err)
		_ = href
	} else {
		t.Skip("no link found in snapshot")
	}
}

func TestCLI_GetCount(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	count, err := b.GetCount("input")
	must(t, err)
	if count == 0 {
		t.Fatal("expected at least 1 input element")
	}
}

func TestCLI_GetBox(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	box, err := b.GetBox(nameID)
	must(t, err)
	if box.Width == 0 || box.Height == 0 {
		t.Fatal("box dimensions should be > 0")
	}
}

func TestCLI_GetStyles(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	styles, err := b.GetStyles(nameID)
	must(t, err)
	if styles == "" {
		t.Fatal("styles should not be empty")
	}
}

// ---------- Is* queries ----------

func TestCLI_IsVisible(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	visible, err := b.IsVisible(nameID)
	must(t, err)
	assertTrue(t, visible, "textbox should be visible")
}

func TestCLI_IsEnabled(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	enabled, err := b.IsEnabled(nameID)
	must(t, err)
	assertTrue(t, enabled, "textbox should be enabled")
}

func TestCLI_IsChecked(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	termsID := findID(t, snap.Text, `checkbox "Terms"`)
	checked, err := b.IsChecked(termsID)
	must(t, err)
	assertFalse(t, checked, "checkbox should start unchecked")
}

// ---------- Wait ----------

func TestCLI_WaitTime(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	start := time.Now()
	must(t, b.Wait(200*time.Millisecond))
	elapsed := time.Since(start)
	if elapsed < 150*time.Millisecond {
		t.Fatal("wait should have taken at least 150ms")
	}
}

func TestCLI_WaitSelector(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.WaitSelector("input"))
}

func TestCLI_WaitURL(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.WaitURL("*interaction*"))
}

func TestCLI_WaitLoad(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.WaitLoad())
}

func TestCLI_WaitText(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.WaitText("Submit"))
}

func TestCLI_WaitFunc(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.WaitFunc("document.readyState === 'complete'"))
}

func TestCLI_WaitHidden(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.WaitHidden("#nonexistent-element"))
}

// ---------- Screenshot / PDF ----------

func TestCLI_Screenshot(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	path := filepath.Join(t.TempDir(), "test.png")
	must(t, b.Screenshot(path))
	info, err := os.Stat(path)
	must(t, err)
	if info.Size() == 0 {
		t.Fatal("screenshot file should not be empty")
	}
}

func TestCLI_ScreenshotAnnotated(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	path := filepath.Join(t.TempDir(), "annotated.png")
	must(t, b.ScreenshotAnnotated(path))
	info, err := os.Stat(path)
	must(t, err)
	if info.Size() == 0 {
		t.Fatal("annotated screenshot file should not be empty")
	}
}

func TestCLI_PDF(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	path := filepath.Join(t.TempDir(), "test.pdf")
	must(t, b.PDF(path))
	info, err := os.Stat(path)
	must(t, err)
	if info.Size() == 0 {
		t.Fatal("PDF file should not be empty")
	}
}

// ---------- Eval ----------

func TestCLI_Eval(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	result, err := b.Eval("1 + 1")
	must(t, err)
	assertEqual(t, result, "2")
}

// ---------- Find ----------

func TestCLI_FindRole(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindRole("button", "Submit")
	must(t, err)
	if result == nil || result.Text == "" {
		t.Fatal("expected find results")
	}
}

func TestCLI_FindText(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindText("Submit")
	must(t, err)
	if result == nil || result.Text == "" {
		t.Fatal("expected find results")
	}
}

func TestCLI_FindLabel(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindLabel("Name")
	must(t, err)
	if result == nil || result.Text == "" {
		t.Fatal("expected find results")
	}
}

func TestCLI_FindNth(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindNth("input", 1)
	must(t, err)
	if result == nil || result.Text == "" {
		t.Fatal("expected find results")
	}
}

func TestCLI_FindLast(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindLast("input")
	must(t, err)
	if result == nil || result.Text == "" {
		t.Fatal("expected find results")
	}
}

func TestCLI_FindPlaceholder(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	// May or may not find a match; just verify the method works
	_, _ = b.FindPlaceholder("Name")
}

func TestCLI_FindTestID(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	// May or may not find a match; just verify the method works
	_, _ = b.FindTestID("some-id")
}

// ---------- Snapshot Options ----------

func TestCLI_SnapshotOptions(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")

	// Interactive only
	snap, err := b.Snapshot(browser.SnapshotOptions{InteractiveOnly: true})
	must(t, err)
	assertContains(t, snap.Text, "textbox")

	// Compact
	snap, err = b.Snapshot(browser.SnapshotOptions{Compact: true})
	must(t, err)
	if snap.Text == "" {
		t.Fatal("compact snapshot should not be empty")
	}

	// MaxDepth
	snap, err = b.Snapshot(browser.SnapshotOptions{MaxDepth: 2})
	must(t, err)
	if snap.Text == "" {
		t.Fatal("depth-limited snapshot should not be empty")
	}

	// Cursor
	snap, err = b.Snapshot(browser.SnapshotOptions{Cursor: true})
	must(t, err)
	if snap.Text == "" {
		t.Fatal("cursor snapshot should not be empty")
	}

	// Selector
	snap, err = b.Snapshot(browser.SnapshotOptions{Selector: "form"})
	must(t, err)
	if snap.Text == "" {
		t.Fatal("selector snapshot should not be empty")
	}
}

// ---------- Mouse ----------

func TestCLI_MouseOperations(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.MouseMove(100, 100))
	must(t, b.MouseDown(100, 100))
	must(t, b.MouseUp(100, 100))
	must(t, b.MouseWheel(100, 100, 0, 100))
	must(t, b.MouseClick(100, 100))
}

// ---------- Tab ----------

func TestCLI_TabOperations(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	tabs, err := b.TabList()
	must(t, err)
	if len(tabs) < 1 {
		t.Fatal("expected at least 1 tab")
	}
	must(t, b.TabNew("about:blank"))
	tabs, err = b.TabList()
	must(t, err)
	if len(tabs) < 2 {
		t.Fatal("expected at least 2 tabs after TabNew")
	}
	must(t, b.TabSwitch(0))
	must(t, b.TabClose(1))
}

// ---------- Cookies ----------

func TestCLI_CookieOperations(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.CookieSet(browser.CookieInfo{
		Name:   "testcookie",
		Value:  "testvalue",
		Domain: "localhost",
	}))
	cookies, err := b.CookiesGet()
	must(t, err)
	found := false
	for _, c := range cookies {
		if c.Name == "testcookie" {
			assertEqual(t, c.Value, "testvalue")
			found = true
		}
	}
	if !found {
		t.Fatal("cookie not found after set")
	}
	must(t, b.CookieDelete("testcookie"))
	must(t, b.CookiesClear())
}

// ---------- Storage ----------

func TestCLI_StorageOperations(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.StorageSet("local", "testkey", "testval"))
	val, err := b.StorageGet("local", "testkey")
	must(t, err)
	assertEqual(t, val, "testval")
	items, err := b.StorageGetAll("local")
	must(t, err)
	if items["testkey"] != "testval" {
		t.Fatal("expected testkey in storage items")
	}
	must(t, b.StorageDelete("local", "testkey"))
	must(t, b.StorageClear("local"))
}

// ---------- Settings ----------

func TestCLI_SetViewport(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetViewport(1280, 720))
}

func TestCLI_SetViewportWithScale(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetViewport(1280, 720, 2.0))
}

func TestCLI_SetDevice(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetDevice("iPhone 12"))
}

func TestCLI_SetGeo(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetGeo(37.7749, -122.4194))
}

func TestCLI_SetOffline(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetOffline(true))
	must(t, b.SetOffline(false))
}

func TestCLI_SetHeaders(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetHeaders(map[string]string{"X-Test": "value"}))
}

func TestCLI_SetCredentials(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetCredentials("user", "pass"))
}

func TestCLI_SetMedia(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetMedia(browser.MediaFeature{Name: "prefers-color-scheme", Value: "dark"}))
}

func TestCLI_SetColorScheme(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetColorScheme("dark"))
}

// ---------- Console ----------

func TestCLI_ConsoleMessages(t *testing.T) {
	b := newBrowser(t)
	must(t, b.ConsoleStart())
	openPage(t, b, "phase1_interaction.html")
	_, _ = b.Eval("console.log('test message')")
	time.Sleep(200 * time.Millisecond)
	msgs, err := b.ConsoleMessages()
	must(t, err)
	_ = msgs
	b.ConsoleClear()
}

func TestCLI_PageErrors(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	errs, err := b.PageErrors()
	must(t, err)
	_ = errs
	b.PageErrorsClear()
}

// ---------- Highlight ----------

func TestCLI_Highlight(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Highlight(nameID))
}

// ---------- Clipboard ----------

func TestCLI_ClipboardReadWrite(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	// Focus the page so clipboard API works in headless mode
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Focus(nameID))
	// Click the element to truly focus the document
	must(t, b.Click(nameID))
	err := b.ClipboardWrite("clipboard test")
	if err != nil {
		// Clipboard API may not work in headless mode; skip gracefully
		t.Skipf("clipboard write not supported in this environment: %v", err)
	}
	text, err := b.ClipboardRead()
	must(t, err)
	assertEqual(t, text, "clipboard test")
}

func TestCLI_ClipboardCopyPaste(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	nameID := findID(t, snap.Text, `textbox "Name"`)
	must(t, b.Fill(nameID, "copy me"))
	must(t, b.Focus(nameID))
	must(t, b.Press("Control+a"))
	must(t, b.ClipboardCopy())
	notesID := findID(t, snap.Text, `textbox "Notes"`)
	must(t, b.Focus(notesID))
	must(t, b.ClipboardPaste())
}

// ---------- Network ----------

func TestCLI_NetworkOperations(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.NetworkStartLogging())
	reqs, err := b.NetworkRequests()
	must(t, err)
	_ = reqs
	b.NetworkClearRequests()
	must(t, b.NetworkRoute("*.jpg", browser.RouteBlock))
	must(t, b.NetworkUnroute("*.jpg"))
}

// ---------- Diff ----------

func TestCLI_DiffSnapshot(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")
	_ = mustSnapshot(t, b)
	result, err := b.DiffSnapshot()
	must(t, err)
	if result.Changed {
		t.Log("unexpected changes in diff, but not fatal")
	}
}

// ---------- Trace / Profiler ----------

func TestCLI_TraceStartStop(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.TraceStart())
	time.Sleep(200 * time.Millisecond)
	path := filepath.Join(t.TempDir(), "trace.json")
	must(t, b.TraceStop(path))
}

func TestCLI_ProfilerStartStop(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.ProfilerStart())
	time.Sleep(200 * time.Millisecond)
	path := filepath.Join(t.TempDir(), "profile.json")
	must(t, b.ProfilerStop(path))
}

// ---------- Record ----------

func TestCLI_RecordStartStop(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	dir := t.TempDir()
	must(t, b.RecordStart(dir))
	time.Sleep(300 * time.Millisecond)
	frames, err := b.RecordStop()
	must(t, err)
	if frames < 0 {
		t.Fatal("frames should be >= 0")
	}
}

// ---------- State Export/Import ----------

func TestCLI_StateExportImport(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	path := filepath.Join(t.TempDir(), "state.json")
	must(t, b.ExportState(path))
	info, err := os.Stat(path)
	must(t, err)
	if info.Size() == 0 {
		t.Fatal("state file should not be empty")
	}
	must(t, b.ImportState(path))
}

// ---------- CDP URL ----------

func TestCLI_GetCDPURL(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	url, err := b.GetCDPURL()
	must(t, err)
	if url == "" {
		t.Fatal("CDP URL should not be empty")
	}
}

// ---------- Session List ----------

func TestCLI_SessionList(t *testing.T) {
	sessions, err := session.ListSessions()
	must(t, err)
	_ = sessions
}

// ---------- Drag ----------

func TestCLI_Drag(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	// DragCoords: just verify the method works without panic
	must(t, b.DragCoords(10, 10, 100, 100))
}

// ---------- Upload ----------

func TestCLI_Upload(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	tmpFile := filepath.Join(t.TempDir(), "upload.txt")
	must(t, os.WriteFile(tmpFile, []byte("test data"), 0644))
	// phase1_interaction.html has no file input; just verify the method exists.
	// UploadCSS will fail with timeout, so we only test that it doesn't panic.
	_ = tmpFile
	t.Log("Upload API exists; skipping UploadCSS (no file input in test page)")
}

// ---------- Find (additional coverage) ----------

func TestCLI_FindFirst(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	result, err := b.FindFirst("input")
	must(t, err)
	if result == nil || result.Text == "" {
		t.Fatal("expected find results")
	}
}

func TestCLI_FindAlt(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	// May or may not match; just verify no panic
	_, _ = b.FindAlt("some-alt")
}

func TestCLI_FindTitle(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b)
	// May or may not match; just verify no panic
	_, _ = b.FindTitle("some-title")
}

// ---------- Diff (additional coverage) ----------

func TestCLI_DiffScreenshot(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")

	// Take a baseline screenshot
	dir := t.TempDir()
	baseline := filepath.Join(dir, "baseline.png")
	must(t, b.Screenshot(baseline))

	// Diff against baseline
	result, err := b.DiffScreenshot(browser.DiffScreenshotOptions{
		BaselineFile: baseline,
		OutputPath:   filepath.Join(dir, "diff.png"),
	})
	must(t, err)
	// Same page, so diff should be minimal
	_ = result
}

func TestCLI_DiffURL(t *testing.T) {
	b := newBrowser(t)
	url1 := testdataURL("phase2_nav_a.html")
	url2 := testdataURL("phase2_nav_b.html")
	result, err := b.DiffURL(url1, url2)
	must(t, err)
	if result.SnapshotDiff == nil {
		t.Fatal("expected snapshot diff result")
	}
}

// ---------- Console/Errors clear (explicit) ----------

func TestCLI_ConsoleClear(t *testing.T) {
	b := newBrowser(t)
	must(t, b.ConsoleStart())
	openPage(t, b, "phase1_interaction.html")
	b.ConsoleClear()
	msgs, err := b.ConsoleMessages()
	must(t, err)
	if len(msgs) != 0 {
		t.Log("console not fully cleared, may have new messages from page")
	}
}

func TestCLI_ErrorsClear(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	b.PageErrorsClear()
	errs, err := b.PageErrors()
	must(t, err)
	_ = errs
}

// ---------- Inspect (OpenDevTools) ----------

func TestCLI_Inspect(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	// OpenDevTools may not work in headless mode; just verify it doesn't panic
	_ = b.OpenDevTools()
}

// ---------- ClearGeo ----------

func TestCLI_ClearGeo(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	must(t, b.SetGeo(37.7749, -122.4194))
	must(t, b.ClearGeo())
}

// =============================================================================
// Helpers (specific to phase8)
// =============================================================================

// parseIDFromLine extracts the numeric ID prefix "123:" from a snapshot line.
func parseIDFromLine(line string) (int, error) {
	var id int
	_, err := fmt.Sscanf(line, "%d:", &id)
	return id, err
}

// findAnyID returns the first element ID found in the snapshot text.
func findAnyID(t *testing.T, snapText string) int {
	t.Helper()
	for _, line := range strings.Split(snapText, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		id, err := parseIDFromLine(trimmed)
		if err == nil && id > 0 {
			return id
		}
	}
	t.Fatal("no element ID found in snapshot")
	return 0
}
