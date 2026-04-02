package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/libi/ko-browser/browser"
)

// ---------- Drag ----------

func TestDrag_ByElement(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")
	snap := mustSnapshot(t, b)

	srcID := findID(t, snap.Text, `"Drag Me"`)
	dstID := findID(t, snap.Text, `"Drop Here"`)

	if err := b.Drag(srcID, dstID); err != nil {
		t.Fatalf("Drag(%d, %d): %v", srcID, dstID, err)
	}

	// Wait a moment for the drop event to be processed
	time.Sleep(300 * time.Millisecond)

	// Verify the drag was completed by checking the status text
	result, err := b.Eval(`document.getElementById('drag-status').textContent`)
	if err != nil {
		t.Fatalf("Eval drag-status: %v", err)
	}
	// The status should indicate either "dropped" (HTML5 DnD or mouse-based)
	assertContains(t, result, "dropped")
}

// ---------- Upload ----------

func TestUpload_Files(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")

	// Create a temp file for upload
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-upload.txt")
	if err := os.WriteFile(testFile, []byte("hello ko-browser"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	t.Run("single_file_by_id", func(t *testing.T) {
		snap := mustSnapshot(t, b)

		// The file input may appear as "Choose file:" label in the AX tree
		// Use findID to get the label, then Upload should auto-find the associated input
		fileInputID := findID(t, snap.Text, `"Choose file:"`)

		if err := b.Upload(fileInputID, testFile); err != nil {
			t.Fatalf("Upload: %v", err)
		}

		time.Sleep(200 * time.Millisecond)

		// Check the upload status
		status, err := b.Eval(`document.getElementById('upload-status').textContent`)
		if err != nil {
			t.Fatalf("Eval upload-status: %v", err)
		}
		assertContains(t, status, "test-upload.txt")
	})

	t.Run("multiple_files_css", func(t *testing.T) {
		testFile2 := filepath.Join(tmpDir, "test-upload2.txt")
		if err := os.WriteFile(testFile2, []byte("second file"), 0644); err != nil {
			t.Fatalf("write test file2: %v", err)
		}

		// Use CSS selector for multi-file upload
		if err := b.UploadCSS("#multi-file-input", testFile, testFile2); err != nil {
			t.Fatalf("UploadCSS multiple: %v", err)
		}

		time.Sleep(200 * time.Millisecond)

		status, err := b.Eval(`document.getElementById('multi-upload-status').textContent`)
		if err != nil {
			t.Fatalf("Eval multi-upload-status: %v", err)
		}
		assertContains(t, status, "test-upload.txt")
		assertContains(t, status, "test-upload2.txt")
	})

	t.Run("upload_css", func(t *testing.T) {
		testFile3 := filepath.Join(tmpDir, "css-upload.txt")
		if err := os.WriteFile(testFile3, []byte("css upload test"), 0644); err != nil {
			t.Fatalf("write test file3: %v", err)
		}

		if err := b.UploadCSS("#file-input", testFile3); err != nil {
			t.Fatalf("UploadCSS: %v", err)
		}

		time.Sleep(200 * time.Millisecond)

		status, err := b.Eval(`document.getElementById('upload-status').textContent`)
		if err != nil {
			t.Fatalf("Eval upload-status: %v", err)
		}
		assertContains(t, status, "css-upload.txt")
	})
}

// ---------- Mouse Operations ----------

func TestMouse_Move(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")
	mustSnapshot(t, b)

	// Scroll the mouse area into view
	evalVoid(t, b, `document.getElementById('mouse-area').scrollIntoView({block: 'center'})`)
	time.Sleep(100 * time.Millisecond)

	// Get exact center coordinates of the mouse area
	x, y := getElementCenter(t, b, "mouse-area")

	// Move mouse to the mouse area center
	if err := b.MouseMove(x, y); err != nil {
		t.Fatalf("MouseMove: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Verify mouse position was tracked
	pos, err := b.Eval(`document.getElementById('mouse-pos').textContent`)
	if err != nil {
		t.Fatalf("Eval mouse-pos: %v", err)
	}
	// Should contain coordinates now
	assertContains(t, pos, "x:")
	t.Logf("Mouse position: %s", pos)
}

func TestMouse_DownAndUp(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")
	mustSnapshot(t, b)

	// Scroll mouse area into view and reset events
	evalVoid(t, b, `document.getElementById('mouse-area').scrollIntoView({block: 'center'})`)
	time.Sleep(100 * time.Millisecond)
	evalVoid(t, b, `document.getElementById('mouse-events').textContent = 'Events: none'`)

	x, y := getElementCenter(t, b, "mouse-area")

	// Move to area first
	must(t, b.MouseMove(x, y))
	time.Sleep(50 * time.Millisecond)

	// Mouse down
	must(t, b.MouseDown(x, y))
	time.Sleep(50 * time.Millisecond)

	// Mouse up
	must(t, b.MouseUp(x, y))
	time.Sleep(200 * time.Millisecond)

	// Check events were recorded
	events, err := b.Eval(`document.getElementById('mouse-events').textContent`)
	if err != nil {
		t.Fatalf("Eval mouse-events: %v", err)
	}
	assertContains(t, events, "mousedown")
	assertContains(t, events, "mouseup")
}

func TestMouse_Wheel(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")
	mustSnapshot(t, b)

	// Scroll mouse area into view and reset events
	evalVoid(t, b, `document.getElementById('mouse-area').scrollIntoView({block: 'center'})`)
	time.Sleep(100 * time.Millisecond)
	evalVoid(t, b, `document.getElementById('mouse-events').textContent = 'Events: none'`)

	x, y := getElementCenter(t, b, "mouse-area")

	// Move to area first
	must(t, b.MouseMove(x, y))
	time.Sleep(50 * time.Millisecond)

	// Dispatch wheel event
	must(t, b.MouseWheel(x, y, 0, 100))
	time.Sleep(200 * time.Millisecond)

	// Check wheel event was recorded
	events, err := b.Eval(`document.getElementById('mouse-events').textContent`)
	if err != nil {
		t.Fatalf("Eval mouse-events: %v", err)
	}
	assertContains(t, events, "wheel")
}

func TestMouse_Click(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")
	mustSnapshot(t, b)

	// Scroll mouse area into view and reset events
	evalVoid(t, b, `document.getElementById('mouse-area').scrollIntoView({block: 'center'})`)
	time.Sleep(100 * time.Millisecond)
	evalVoid(t, b, `document.getElementById('mouse-events').textContent = 'Events: none'`)

	x, y := getElementCenter(t, b, "mouse-area")

	// Full click
	must(t, b.MouseClick(x, y))
	time.Sleep(200 * time.Millisecond)

	// Check that click events were recorded (mousedown + mouseup + click)
	events, err := b.Eval(`document.getElementById('mouse-events').textContent`)
	if err != nil {
		t.Fatalf("Eval mouse-events: %v", err)
	}
	assertContains(t, events, "mousedown")
	assertContains(t, events, "mouseup")
	assertContains(t, events, "click")
}

func TestDrag_ByCoordinates(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase4_advanced.html")

	// Get source and target element centers via JS
	srcX, err := b.Eval(`(() => { const r = document.getElementById('drag-item').getBoundingClientRect(); return r.left + r.width/2; })()`)
	if err != nil {
		t.Fatalf("get srcX: %v", err)
	}
	srcY, err := b.Eval(`(() => { const r = document.getElementById('drag-item').getBoundingClientRect(); return r.top + r.height/2; })()`)
	if err != nil {
		t.Fatalf("get srcY: %v", err)
	}
	dstX, err := b.Eval(`(() => { const r = document.getElementById('target-zone').getBoundingClientRect(); return r.left + r.width/2; })()`)
	if err != nil {
		t.Fatalf("get dstX: %v", err)
	}
	dstY, err := b.Eval(`(() => { const r = document.getElementById('target-zone').getBoundingClientRect(); return r.top + r.height/2; })()`)
	if err != nil {
		t.Fatalf("get dstY: %v", err)
	}

	var sx, sy, dx, dy float64
	parseFloat(t, srcX, &sx)
	parseFloat(t, srcY, &sy)
	parseFloat(t, dstX, &dx)
	parseFloat(t, dstY, &dy)

	if err := b.DragCoords(sx, sy, dx, dy); err != nil {
		t.Fatalf("DragCoords: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	result, err := b.Eval(`document.getElementById('drag-status').textContent`)
	if err != nil {
		t.Fatalf("Eval drag-status: %v", err)
	}
	assertContains(t, result, "dropped")
}

// ---------- Helpers ----------

func parseFloat(t *testing.T, s string, out *float64) {
	t.Helper()
	var v float64
	if _, err := fmt.Sscanf(s, "%f", &v); err != nil {
		t.Fatalf("parseFloat(%q): %v", s, err)
	}
	*out = v
}

// getElementCenter returns the viewport-relative center coordinates of a DOM element by its CSS id.
func getElementCenter(t *testing.T, b *browser.Browser, elementID string) (float64, float64) {
	t.Helper()
	js := fmt.Sprintf(`(() => {
		const r = document.getElementById('%s').getBoundingClientRect();
		return JSON.stringify({x: r.left + r.width/2, y: r.top + r.height/2});
	})()`, elementID)
	result, err := b.Eval(js)
	if err != nil {
		t.Fatalf("getElementCenter(%s): %v", elementID, err)
	}
	var coords struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	if err := json.Unmarshal([]byte(result), &coords); err != nil {
		t.Fatalf("getElementCenter parse %q: %v", result, err)
	}
	t.Logf("Element %q center: (%.1f, %.1f)", elementID, coords.X, coords.Y)
	return coords.X, coords.Y
}
