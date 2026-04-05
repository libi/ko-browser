package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/libi/ko-browser/browser"
)

func TestSnapshot_IncludesIframeContent(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "snapshot_iframe.html")

	snap := mustSnapshot(t, b)

	assertContains(t, snap.Text, "Host Document Heading")
	assertContains(t, snap.Text, "Host document text")
	assertContains(t, snap.Text, "Iframe Heading")
	assertContains(t, snap.Text, "Iframe Action")
	assertContains(t, snap.Text, "Iframe unique text")
}

func TestSnapshot_IncludesNestedIframeContent(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "snapshot_nested_iframe.html")

	snap := mustSnapshot(t, b)

	assertContains(t, snap.Text, "Host Snapshot Page")
	assertContains(t, snap.Text, `button "Host Action"`)
	assertContains(t, snap.Text, "Outer Frame Heading")
	assertContains(t, snap.Text, `button "Outer Action"`)
	assertContains(t, snap.Text, `textbox "Outer Input"`)
	assertContains(t, snap.Text, "Inner Frame Heading")
	assertContains(t, snap.Text, `button "Inner Action"`)
	assertContains(t, snap.Text, `textbox "Inner Input"`)
	assertContains(t, snap.Text, "Inner frame unique text")
}

func TestSnapshot_IframeIDsStableAcrossModes(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "snapshot_nested_iframe.html")

	base := mustSnapshot(t, b)
	expected := map[string]int{
		`button "Host Action"`:  findID(t, base.Text, `button "Host Action"`),
		`button "Outer Action"`: findID(t, base.Text, `button "Outer Action"`),
		`textbox "Outer Input"`: findID(t, base.Text, `textbox "Outer Input"`),
		`button "Inner Action"`: findID(t, base.Text, `button "Inner Action"`),
		`textbox "Inner Input"`: findID(t, base.Text, `textbox "Inner Input"`),
	}

	checkIDs := func(t *testing.T, snapText string) {
		t.Helper()
		for label, wantID := range expected {
			gotID := findID(t, snapText, label)
			if gotID != wantID {
				t.Fatalf("id changed for %s: got %d want %d\nSnapshot:\n%s", label, gotID, wantID, snapText)
			}
		}
	}

	t.Run("interactive_only", func(t *testing.T) {
		snap, err := b.Snapshot(browser.SnapshotOptions{InteractiveOnly: true})
		if err != nil {
			t.Fatalf("Snapshot interactive_only: %v", err)
		}
		checkIDs(t, snap.Text)
	})

	t.Run("compact", func(t *testing.T) {
		snap, err := b.Snapshot(browser.SnapshotOptions{Compact: true})
		if err != nil {
			t.Fatalf("Snapshot compact: %v", err)
		}
		checkIDs(t, snap.Text)
	})

	t.Run("max_depth", func(t *testing.T) {
		snap, err := b.Snapshot(browser.SnapshotOptions{MaxDepth: 8})
		if err != nil {
			t.Fatalf("Snapshot max_depth: %v", err)
		}
		checkIDs(t, snap.Text)
	})

	t.Run("cursor", func(t *testing.T) {
		innerInputID := expected[`textbox "Inner Input"`]
		must(t, b.Focus(innerInputID))

		snap, err := b.Snapshot(browser.SnapshotOptions{Cursor: true})
		if err != nil {
			t.Fatalf("Snapshot cursor: %v", err)
		}
		checkIDs(t, snap.Text)
		if !strings.Contains(snap.Text, `textbox "Inner Input" focused`) {
			t.Fatalf("expected focused iframe input in cursor snapshot, got:\n%s", snap.Text)
		}
	})
}

func TestSnapshot_IframeIDsSupportInteractions(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "snapshot_nested_iframe.html")

	snap := mustSnapshot(t, b)
	outerButtonID := findID(t, snap.Text, `button "Outer Action"`)
	innerButtonID := findID(t, snap.Text, `button "Inner Action"`)
	innerInputID := findID(t, snap.Text, `textbox "Inner Input"`)

	must(t, b.Click(outerButtonID))
	time.Sleep(150 * time.Millisecond)
	status, err := b.Eval(`document.getElementById('status').textContent`)
	if err != nil {
		t.Fatalf("Eval status after click: %v", err)
	}
	assertEqual(t, status, "outer-clicked")

	must(t, b.DblClick(innerButtonID))
	time.Sleep(150 * time.Millisecond)
	status, err = b.Eval(`document.getElementById('status').textContent`)
	if err != nil {
		t.Fatalf("Eval status after dblclick: %v", err)
	}
	assertEqual(t, status, "inner-double-clicked")

	must(t, b.Fill(innerInputID, "hello iframe"))
	time.Sleep(150 * time.Millisecond)
	value, err := b.GetValue(innerInputID)
	if err != nil {
		t.Fatalf("GetValue after fill: %v", err)
	}
	assertEqual(t, value, "hello iframe")

	status, err = b.Eval(`document.getElementById('status').textContent`)
	if err != nil {
		t.Fatalf("Eval status after fill: %v", err)
	}
	assertEqual(t, status, "inner:hello iframe")
}
