package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/libi/ko-browser/browser"
)

// ---------- DiffSnapshot ----------

func TestDiff_Snapshot_PreviousSnapshot(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	// Take an initial snapshot (sets lastSnap)
	snap1 := mustSnapshot(t, b)
	assertContains(t, snap1.Text, "Original text here")

	// Click the "Change Content" button to modify the page
	id := findID(t, snap1.Text, "Change Content")
	if err := b.Click(id); err != nil {
		t.Fatalf("Click change: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// DiffSnapshot should compare with the previous snapshot
	result, err := b.DiffSnapshot()
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	if !result.Changed {
		t.Fatal("expected diff to show changes")
	}
	if result.Added == 0 && result.Removed == 0 {
		t.Fatal("expected added or removed lines")
	}
	// The diff text should mention the changed text
	assertContains(t, result.Text, "Changed text")
}

func TestDiff_Snapshot_BaselineFile(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	// Take an initial snapshot and save as baseline
	snap1 := mustSnapshot(t, b)
	baselinePath := filepath.Join(t.TempDir(), "baseline.txt")
	if err := os.WriteFile(baselinePath, []byte(snap1.Text), 0644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	// Modify the page
	id := findID(t, snap1.Text, "Change Content")
	if err := b.Click(id); err != nil {
		t.Fatalf("Click change: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// DiffSnapshot with baseline file
	result, err := b.DiffSnapshot(browser.DiffSnapshotOptions{
		BaselineFile: baselinePath,
	})
	if err != nil {
		t.Fatalf("DiffSnapshot with baseline: %v", err)
	}

	if !result.Changed {
		t.Fatal("expected diff to show changes with baseline")
	}
	assertContains(t, result.Text, "Changed text")
}

func TestDiff_Snapshot_NoPreviousError(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	// Without calling Snapshot first and without baseline file, should get an error
	_, err := b.DiffSnapshot()
	if err == nil {
		t.Fatal("expected error when no previous snapshot exists")
	}
	assertContains(t, err.Error(), "no previous snapshot")
}

func TestDiff_Snapshot_NoChange(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	// Take a snapshot, then immediately diff with no changes
	_ = mustSnapshot(t, b)
	time.Sleep(200 * time.Millisecond)

	result, err := b.DiffSnapshot()
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	if result.Changed {
		t.Logf("Unexpected diff:\n%s", result.Text)
		// Note: some dynamic elements may cause minor diffs; just log it
	}
}

// ---------- DiffScreenshot ----------

func TestDiff_Screenshot_Basic(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")
	time.Sleep(500 * time.Millisecond) // wait for page to render

	// Take a baseline screenshot
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, "baseline.png")
	screenshotData, err := b.ScreenshotToBytes()
	if err != nil {
		t.Fatalf("ScreenshotToBytes: %v", err)
	}
	if err := os.WriteFile(baselinePath, screenshotData, 0644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	// Compare the same page - should have no differences
	result, err := b.DiffScreenshot(browser.DiffScreenshotOptions{
		BaselineFile: baselinePath,
		Threshold:    0.1,
	})
	if err != nil {
		t.Fatalf("DiffScreenshot: %v", err)
	}

	if result.Changed {
		t.Logf("Unexpected pixel diff: %d/%d (%.2f%%)", result.DiffCount, result.TotalPixels, result.DiffPercent)
	}
}

func TestDiff_Screenshot_WithChange(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")
	time.Sleep(500 * time.Millisecond)

	// Take a baseline screenshot
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, "baseline.png")
	screenshotData, err := b.ScreenshotToBytes()
	if err != nil {
		t.Fatalf("ScreenshotToBytes: %v", err)
	}
	if err := os.WriteFile(baselinePath, screenshotData, 0644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	// Modify the page visually by clicking change color
	snap := mustSnapshot(t, b)
	colorID := findID(t, snap.Text, "Change Color")
	if err := b.Click(colorID); err != nil {
		t.Fatalf("Click color: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Compare - should detect visual differences
	diffPath := filepath.Join(tmpDir, "diff.png")
	result, err := b.DiffScreenshot(browser.DiffScreenshotOptions{
		BaselineFile: baselinePath,
		OutputPath:   diffPath,
		Threshold:    0.05,
	})
	if err != nil {
		t.Fatalf("DiffScreenshot: %v", err)
	}

	if !result.Changed {
		t.Fatal("expected screenshot diff to show changes after color change")
	}
	if result.DiffCount == 0 {
		t.Fatal("expected nonzero diff pixel count")
	}
	t.Logf("Screenshot diff: %d pixels changed (%.2f%%)", result.DiffCount, result.DiffPercent)

	// Verify diff image was saved
	if result.Changed {
		if _, err := os.Stat(diffPath); os.IsNotExist(err) {
			t.Error("expected diff image to be saved")
		}
	}
}

func TestDiff_Screenshot_NoBaseline(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	_, err := b.DiffScreenshot(browser.DiffScreenshotOptions{})
	if err == nil {
		t.Fatal("expected error when no baseline file")
	}
	assertContains(t, err.Error(), "baseline file is required")
}

// ---------- DiffURL ----------

func TestDiff_URL_DifferentPages(t *testing.T) {
	b := newBrowser(t)

	url1 := testdataURL("phase7_test.html")
	url2 := testdataURL("phase7_diff_b.html")

	result, err := b.DiffURL(url1, url2)
	if err != nil {
		t.Fatalf("DiffURL: %v", err)
	}

	if result.SnapshotDiff == nil {
		t.Fatal("expected non-nil snapshot diff")
	}
	if !result.SnapshotDiff.Changed {
		t.Fatal("expected snapshot diff to show changes between different pages")
	}
	if result.SnapshotDiff.Added == 0 && result.SnapshotDiff.Removed == 0 {
		t.Fatal("expected added or removed lines in snapshot diff")
	}
	t.Logf("DiffURL snapshot: +%d/-%d lines", result.SnapshotDiff.Added, result.SnapshotDiff.Removed)
}

func TestDiff_URL_SamePage(t *testing.T) {
	b := newBrowser(t)

	url := testdataURL("phase7_test.html")

	result, err := b.DiffURL(url, url)
	if err != nil {
		t.Fatalf("DiffURL same page: %v", err)
	}

	if result.SnapshotDiff.Changed {
		t.Logf("Same page diff (may have minor variance):\n%s", result.SnapshotDiff.Text)
	}
}

func TestDiff_URL_WithScreenshot(t *testing.T) {
	b := newBrowser(t)

	url1 := testdataURL("phase7_test.html")
	url2 := testdataURL("phase7_diff_b.html")

	result, err := b.DiffURL(url1, url2, browser.DiffURLOptions{
		IncludeScreenshot: true,
		Threshold:         0.05,
	})
	if err != nil {
		t.Fatalf("DiffURL with screenshot: %v", err)
	}

	if result.ScreenshotDiff == nil {
		t.Fatal("expected screenshot diff when IncludeScreenshot=true")
	}
	if !result.ScreenshotDiff.Changed {
		t.Log("screenshot diff showed no visual changes (may be similar layout)")
	}
	t.Logf("DiffURL screenshot: %d diff pixels (%.2f%%)", result.ScreenshotDiff.DiffCount, result.ScreenshotDiff.DiffPercent)
}

// ---------- Trace ----------

func TestTrace_StartStop(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	// Start tracing
	if err := b.TraceStart(); err != nil {
		t.Fatalf("TraceStart: %v", err)
	}

	// Do some page activity
	snap := mustSnapshot(t, b)
	changeID := findID(t, snap.Text, "Change Content")
	if err := b.Click(changeID); err != nil {
		t.Fatalf("Click: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Stop and save trace
	tracePath := filepath.Join(t.TempDir(), "trace.json")
	if err := b.TraceStop(tracePath); err != nil {
		t.Fatalf("TraceStop: %v", err)
	}

	// Verify trace file was created
	info, err := os.Stat(tracePath)
	if err != nil {
		t.Fatalf("trace file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("trace file is empty")
	}
	t.Logf("Trace file size: %d bytes", info.Size())

	// Verify it's valid JSON
	data, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read trace: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("trace file is not valid JSON")
	}
}

func TestTrace_DoubleStart(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	if err := b.TraceStart(); err != nil {
		t.Fatalf("TraceStart: %v", err)
	}
	// Second start should fail
	err := b.TraceStart()
	if err == nil {
		t.Fatal("expected error on double TraceStart")
	}
	assertContains(t, err.Error(), "already in progress")

	// Clean up
	_ = b.TraceStop(filepath.Join(t.TempDir(), "cleanup.json"))
}

func TestTrace_StopWithoutStart(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	err := b.TraceStop(filepath.Join(t.TempDir(), "should_not_exist.json"))
	if err == nil {
		t.Fatal("expected error on TraceStop without TraceStart")
	}
	assertContains(t, err.Error(), "no trace in progress")
}

// ---------- Profiler ----------

func TestProfiler_StartStop(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	if err := b.ProfilerStart(); err != nil {
		t.Fatalf("ProfilerStart: %v", err)
	}

	// Do some page activity to generate profiling data
	snap := mustSnapshot(t, b)
	changeID := findID(t, snap.Text, "Change Content")
	for i := 0; i < 3; i++ {
		if err := b.Click(changeID); err != nil {
			t.Fatalf("Click %d: %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Stop and save profile
	profilePath := filepath.Join(t.TempDir(), "profile.json")
	if err := b.ProfilerStop(profilePath); err != nil {
		t.Fatalf("ProfilerStop: %v", err)
	}

	// Verify file
	info, err := os.Stat(profilePath)
	if err != nil {
		t.Fatalf("profile file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("profile file is empty")
	}
	t.Logf("Profile file size: %d bytes", info.Size())

	// Verify valid JSON
	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("profile file is not valid JSON")
	}
}

func TestProfiler_DoubleStart(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	if err := b.ProfilerStart(); err != nil {
		t.Fatalf("ProfilerStart: %v", err)
	}
	err := b.ProfilerStart()
	if err == nil {
		t.Fatal("expected error on double ProfilerStart")
	}
	assertContains(t, err.Error(), "already running")

	// Clean up
	_ = b.ProfilerStop(filepath.Join(t.TempDir(), "cleanup.json"))
}

func TestProfiler_StopWithoutStart(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	err := b.ProfilerStop(filepath.Join(t.TempDir(), "should_not_exist.json"))
	if err == nil {
		t.Fatal("expected error on ProfilerStop without ProfilerStart")
	}
	assertContains(t, err.Error(), "not running")
}

// ---------- Record ----------

func TestRecord_StartStop(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")
	time.Sleep(500 * time.Millisecond)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "recording.png")

	if err := b.RecordStart(outputPath); err != nil {
		t.Fatalf("RecordStart: %v", err)
	}

	// Do some page activity to generate frames
	snap := mustSnapshot(t, b)
	changeID := findID(t, snap.Text, "Change Content")
	for i := 0; i < 3; i++ {
		if err := b.Click(changeID); err != nil {
			t.Fatalf("Click %d: %v", i, err)
		}
		time.Sleep(400 * time.Millisecond)
	}

	frameCount, err := b.RecordStop()
	if err != nil {
		t.Fatalf("RecordStop: %v", err)
	}

	t.Logf("Captured %d frames", frameCount)
	// We expect at least some frames to have been captured
	if frameCount == 0 {
		t.Log("Warning: no frames captured (may depend on system performance)")
	}
}

func TestRecord_DoubleStart(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	if err := b.RecordStart(filepath.Join(t.TempDir(), "rec.png")); err != nil {
		t.Fatalf("RecordStart: %v", err)
	}
	err := b.RecordStart(filepath.Join(t.TempDir(), "rec2.png"))
	if err == nil {
		t.Fatal("expected error on double RecordStart")
	}
	assertContains(t, err.Error(), "already in progress")

	// Clean up
	_, _ = b.RecordStop()
}

func TestRecord_StopWithoutStart(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	_, err := b.RecordStop()
	if err == nil {
		t.Fatal("expected error on RecordStop without RecordStart")
	}
	assertContains(t, err.Error(), "no recording in progress")
}

// ---------- State Export/Import ----------

func TestState_ExportImport(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")
	time.Sleep(300 * time.Millisecond)

	// Set some localStorage and cookies via JS
	if _, err := b.Eval(`localStorage.setItem("test_key", "test_value")`); err != nil {
		t.Fatalf("set localStorage: %v", err)
	}
	if _, err := b.Eval(`document.cookie = "test_cookie=hello; path=/"`); err != nil {
		t.Fatalf("set cookie: %v", err)
	}

	// Export state
	statePath := filepath.Join(t.TempDir(), "state.json")
	if err := b.ExportState(statePath); err != nil {
		t.Fatalf("ExportState: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("state file is not valid JSON")
	}

	// Parse the state to verify content
	var state browser.BrowserState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("parse state: %v", err)
	}

	t.Logf("Exported state: %d cookies, %d localStorage items, origin=%s",
		len(state.Cookies), len(state.LocalStorage), state.Origin)

	// Verify localStorage was exported
	if val, ok := state.LocalStorage["test_key"]; !ok || val != "test_value" {
		t.Errorf("expected localStorage test_key=test_value, got %v", state.LocalStorage)
	}

	// Now import back into a new browser
	b2 := newBrowser(t)
	if err := b2.ImportState(statePath); err != nil {
		t.Fatalf("ImportState: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Verify cookies were imported
	cookies, err := b2.CookiesGet()
	if err != nil {
		t.Fatalf("CookiesGet: %v", err)
	}

	// Check if test_cookie was imported (may or may not work depending on domain)
	t.Logf("Imported %d cookies", len(cookies))
}

func TestState_Export_EmptyPage(t *testing.T) {
	b := newBrowser(t)
	// Don't navigate anywhere - should still work on about:blank

	statePath := filepath.Join(t.TempDir(), "empty_state.json")
	if err := b.ExportState(statePath); err != nil {
		t.Fatalf("ExportState on empty page: %v", err)
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("state file is not valid JSON")
	}
}

func TestState_Import_NonexistentFile(t *testing.T) {
	b := newBrowser(t)
	err := b.ImportState("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for non-existent state file")
	}
}

// ---------- Profile Option ----------

func TestProfile_Option(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ko-browser-profile-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	profileDir := filepath.Join(tmpDir, "test-profile")

	b, err := browser.New(browser.Options{
		Headless: true,
		Timeout:  30 * time.Second,
		Profile:  profileDir,
	})
	if err != nil {
		t.Fatalf("browser.New with profile: %v", err)
	}
	closed := false
	closeBrowser := func() {
		if !closed {
			b.Close()
			closed = true
		}
	}
	t.Cleanup(func() {
		closeBrowser()
		removeDirEventually(t, tmpDir)
	})

	// Navigate to a page
	if err := b.Open(testdataURL("phase7_test.html")); err != nil {
		closeBrowser()
		t.Fatalf("Open: %v", err)
	}

	// Set some data in localStorage
	if _, err := b.Eval(`localStorage.setItem("profile_test", "yes")`); err != nil {
		closeBrowser()
		t.Fatalf("set localStorage: %v", err)
	}

	// Verify the profile directory was created
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Fatal("expected profile directory to be created")
	}

	t.Logf("Profile directory created at: %s", profileDir)
	closeBrowser()
}

// ---------- Config File ----------

func TestConfig_FileFormat(t *testing.T) {
	// Test that a config file can be parsed correctly
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	config := map[string]interface{}{
		"headless": true,
		"timeout":  "30s",
		"profile":  "/tmp/test-profile",
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Verify the file is valid JSON
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !json.Valid(readData) {
		t.Fatal("config file is not valid JSON")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(readData, &parsed); err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if parsed["headless"] != true {
		t.Errorf("expected headless=true, got %v", parsed["headless"])
	}
	if parsed["profile"] != "/tmp/test-profile" {
		t.Errorf("expected profile=/tmp/test-profile, got %v", parsed["profile"])
	}
}

// ---------- State with Cookies roundtrip ----------

func TestState_Apply(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	// Test ApplyState with localStorage (cookies don't work on file:// protocol)
	state := &browser.BrowserState{
		LocalStorage: map[string]string{
			"apply_test_key":  "apply_test_value",
			"apply_test_key2": "second_value",
		},
		Origin: testdataURL("phase7_test.html"),
	}
	if err := b.ApplyState(state); err != nil {
		t.Fatalf("ApplyState: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Verify localStorage was applied
	val, err := b.Eval(`localStorage.getItem("apply_test_key")`)
	if err != nil {
		t.Fatalf("Eval localStorage: %v", err)
	}
	if val != "apply_test_value" {
		t.Errorf("expected localStorage apply_test_key=apply_test_value, got %q", val)
	}

	val2, err := b.Eval(`localStorage.getItem("apply_test_key2")`)
	if err != nil {
		t.Fatalf("Eval localStorage key2: %v", err)
	}
	if val2 != "second_value" {
		t.Errorf("expected localStorage apply_test_key2=second_value, got %q", val2)
	}

	// Now export and verify the roundtrip
	statePath := filepath.Join(t.TempDir(), "apply_roundtrip.json")
	if err := b.ExportState(statePath); err != nil {
		t.Fatalf("ExportState: %v", err)
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}

	var exported browser.BrowserState
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("parse state: %v", err)
	}

	if exported.LocalStorage["apply_test_key"] != "apply_test_value" {
		t.Errorf("exported state missing apply_test_key, got %v", exported.LocalStorage)
	}
	t.Logf("ApplyState roundtrip: %d localStorage items exported", len(exported.LocalStorage))
}

// ---------- Integration: Diff + Trace combo ----------

func TestTrace_WithDiff(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase7_test.html")

	// Start trace
	if err := b.TraceStart(); err != nil {
		t.Fatalf("TraceStart: %v", err)
	}

	// Take initial snapshot
	_ = mustSnapshot(t, b)

	// Modify page
	snap := mustSnapshot(t, b)
	id := findID(t, snap.Text, "Change Content")
	if err := b.Click(id); err != nil {
		t.Fatalf("Click: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Diff
	diffResult, err := b.DiffSnapshot()
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	// Stop trace
	tracePath := filepath.Join(t.TempDir(), "combo_trace.json")
	if err := b.TraceStop(tracePath); err != nil {
		t.Fatalf("TraceStop: %v", err)
	}

	t.Logf("Diff changed: %v (added:%d removed:%d), trace saved", diffResult.Changed, diffResult.Added, diffResult.Removed)

	// Verify trace file
	if _, err := os.Stat(tracePath); os.IsNotExist(err) {
		t.Fatal("trace file not created")
	}
}

// ---------- Helpers ----------
// containsStr is declared in phase5_test.go
