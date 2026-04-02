package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/libi/ko-browser/browser"
)

func TestScreenshot_Basic(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase3_screenshot.html")

	tmpDir := t.TempDir()

	t.Run("viewport_png", func(t *testing.T) {
		path := filepath.Join(tmpDir, "viewport.png")
		err := b.Screenshot(path)
		if err != nil {
			t.Fatalf("Screenshot: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if info.Size() < 100 {
			t.Errorf("screenshot too small: %d bytes", info.Size())
		}
	})

	t.Run("fullpage_png", func(t *testing.T) {
		path := filepath.Join(tmpDir, "fullpage.png")
		err := b.Screenshot(path, browser.ScreenshotOptions{FullPage: true})
		if err != nil {
			t.Fatalf("Screenshot fullpage: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if info.Size() < 100 {
			t.Errorf("fullpage screenshot too small: %d bytes", info.Size())
		}
	})

	t.Run("jpeg_quality", func(t *testing.T) {
		path := filepath.Join(tmpDir, "viewport.jpg")
		err := b.Screenshot(path, browser.ScreenshotOptions{Quality: 80})
		if err != nil {
			t.Fatalf("Screenshot JPEG: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if info.Size() < 100 {
			t.Errorf("JPEG screenshot too small: %d bytes", info.Size())
		}
	})

	t.Run("to_bytes", func(t *testing.T) {
		data, err := b.ScreenshotToBytes()
		if err != nil {
			t.Fatalf("ScreenshotToBytes: %v", err)
		}
		if len(data) < 100 {
			t.Errorf("screenshot bytes too small: %d bytes", len(data))
		}
	})
}

func TestEval_Basic(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase3_screenshot.html")

	t.Run("document_title", func(t *testing.T) {
		result, err := b.Eval("document.title")
		if err != nil {
			t.Fatalf("Eval: %v", err)
		}
		assertEqual(t, result, "Phase 3 Test Page")
	})

	t.Run("arithmetic", func(t *testing.T) {
		result, err := b.Eval("1 + 2")
		if err != nil {
			t.Fatalf("Eval: %v", err)
		}
		assertEqual(t, result, "3")
	})

	t.Run("window_variable", func(t *testing.T) {
		result, err := b.Eval("window.testValue")
		if err != nil {
			t.Fatalf("Eval: %v", err)
		}
		assertEqual(t, result, "42")
	})

	t.Run("function_call", func(t *testing.T) {
		result, err := b.Eval("window.getTestData()")
		if err != nil {
			t.Fatalf("Eval: %v", err)
		}
		assertContains(t, result, "status")
		assertContains(t, result, "ok")
	})

	t.Run("querySelector", func(t *testing.T) {
		result, err := b.Eval("document.querySelector('#dynamic-text').textContent")
		if err != nil {
			t.Fatalf("Eval: %v", err)
		}
		assertEqual(t, result, "This is dynamic content area")
	})
}

func TestFind_Basic(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase3_screenshot.html")
	mustSnapshot(t, b)

	t.Run("FindRole_button", func(t *testing.T) {
		results, err := b.FindRole("button", "")
		if err != nil {
			t.Fatalf("FindRole: %v", err)
		}
		if len(results.Items) < 3 {
			t.Errorf("expected at least 3 buttons, got %d", len(results.Items))
		}
		assertContains(t, results.Text, "Search")
	})

	t.Run("FindRole_button_with_name", func(t *testing.T) {
		results, err := b.FindRole("button", "Submit")
		if err != nil {
			t.Fatalf("FindRole: %v", err)
		}
		if len(results.Items) != 1 {
			t.Errorf("expected 1 submit button, got %d: %s", len(results.Items), results.Text)
		}
		assertContains(t, results.Text, "Submit Form")
	})

	t.Run("FindRole_link", func(t *testing.T) {
		results, err := b.FindRole("link", "")
		if err != nil {
			t.Fatalf("FindRole: %v", err)
		}
		if len(results.Items) < 3 {
			t.Errorf("expected at least 3 links, got %d", len(results.Items))
		}
	})

	t.Run("FindText", func(t *testing.T) {
		results, err := b.FindText("Home")
		if err != nil {
			t.Fatalf("FindText: %v", err)
		}
		if len(results.Items) == 0 {
			t.Error("expected at least 1 result for 'Home'")
		}
		assertContains(t, results.Text, "Home")
	})

	t.Run("FindLabel_Username", func(t *testing.T) {
		results, err := b.FindLabel("Username")
		if err != nil {
			t.Fatalf("FindLabel: %v", err)
		}
		if len(results.Items) == 0 {
			t.Error("expected at least 1 result for label 'Username'")
		}
		found := false
		for _, item := range results.Items {
			if strings.ToLower(item.Role) == "textbox" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find a textbox for 'Username'")
		}
	})

	t.Run("FindLabel_Email", func(t *testing.T) {
		results, err := b.FindLabel("Email")
		if err != nil {
			t.Fatalf("FindLabel: %v", err)
		}
		if len(results.Items) == 0 {
			t.Error("expected at least 1 result for label 'Email'")
		}
	})

	t.Run("FindNth_card", func(t *testing.T) {
		results, err := b.FindNth(".card", 2)
		if err != nil {
			t.Fatalf("FindNth: %v", err)
		}
		if results.Text == "" {
			t.Error("expected non-empty result for FindNth")
		}
	})
}

func TestSnapshot_Options(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase3_screenshot.html")

	t.Run("default", func(t *testing.T) {
		snap, err := b.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot: %v", err)
		}
		assertContains(t, snap.Text, "heading")
		assertContains(t, snap.Text, "link")
		assertContains(t, snap.Text, "button")
	})

	t.Run("interactive_only", func(t *testing.T) {
		snap, err := b.Snapshot(browser.SnapshotOptions{
			InteractiveOnly: true,
		})
		if err != nil {
			t.Fatalf("Snapshot interactive: %v", err)
		}
		assertContains(t, snap.Text, "button")
		assertContains(t, snap.Text, "link")
		assertContains(t, snap.Text, "textbox")
		if strings.Contains(snap.Text, "heading") {
			t.Error("interactive-only snapshot should not contain 'heading'")
		}
	})

	t.Run("compact", func(t *testing.T) {
		snap, err := b.Snapshot(browser.SnapshotOptions{
			Compact: true,
		})
		if err != nil {
			t.Fatalf("Snapshot compact: %v", err)
		}
		assertContains(t, snap.Text, "button")
		assertContains(t, snap.Text, "link")
	})

	t.Run("max_depth", func(t *testing.T) {
		snapDeep, err := b.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot: %v", err)
		}
		snapShallow, err := b.Snapshot(browser.SnapshotOptions{
			MaxDepth: 2,
		})
		if err != nil {
			t.Fatalf("Snapshot depth: %v", err)
		}
		deepLines := strings.Count(snapDeep.Text, "\n")
		shallowLines := strings.Count(snapShallow.Text, "\n")
		if shallowLines >= deepLines {
			t.Errorf("expected fewer lines with MaxDepth=2: deep=%d shallow=%d", deepLines, shallowLines)
		}
	})
}
