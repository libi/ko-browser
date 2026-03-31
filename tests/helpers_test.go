package tests

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/libi/ko-browser/browser"
)

// projectRoot returns the absolute path to the project root directory.
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filename))
}

// testdataURL returns a file:// URL for a file in the testdata directory.
func testdataURL(name string) string {
	return "file://" + filepath.Join(projectRoot(), "testdata", name)
}

// newBrowser creates a headless browser and registers cleanup.
func newBrowser(t *testing.T) *browser.Browser {
	t.Helper()
	b, err := browser.New(browser.Options{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("browser.New: %v", err)
	}
	t.Cleanup(func() { b.Close() })
	return b
}

// openPage navigates to a testdata HTML file.
func openPage(t *testing.T, b *browser.Browser, testdataFile string) {
	t.Helper()
	if err := b.Open(testdataURL(testdataFile)); err != nil {
		t.Fatalf("Open(%s): %v", testdataFile, err)
	}
}

// mustSnapshot takes an accessibility snapshot and fatals on error.
func mustSnapshot(t *testing.T, b *browser.Browser) *browser.SnapshotResult {
	t.Helper()
	snap, err := b.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	return snap
}

// findID searches snapshot text for a line containing substr and returns its display ID.
// Fatals if not found.
func findID(t *testing.T, snapText, substr string) int {
	t.Helper()
	id := tryFindID(snapText, substr)
	if id == 0 {
		t.Fatalf("element not found in snapshot for %q\nSnapshot:\n%s", substr, snapText)
	}
	return id
}

// tryFindID searches snapshot text for a line containing substr and returns its display ID.
// Returns 0 if not found.
func tryFindID(snapText, substr string) int {
	for _, line := range strings.Split(snapText, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, substr) {
			var id int
			if _, err := fmt.Sscanf(trimmed, "%d:", &id); err == nil {
				return id
			}
		}
	}
	return 0
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected string to contain %q, got:\n%s", want, got)
	}
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertTrue(t *testing.T, val bool, msg string) {
	t.Helper()
	if !val {
		t.Errorf("expected true: %s", msg)
	}
}

func assertFalse(t *testing.T, val bool, msg string) {
	t.Helper()
	if val {
		t.Errorf("expected false: %s", msg)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// evalVoid evaluates a JS expression and discards the result. Fatals on error.
func evalVoid(t *testing.T, b *browser.Browser, expression string) {
	t.Helper()
	if _, err := b.Eval(expression); err != nil {
		t.Fatalf("Eval(%q): %v", expression, err)
	}
}
