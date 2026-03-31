package tests

import "testing"

func TestPhase0_Lifecycle(t *testing.T) {
	b := newBrowser(t)

	// Open Page A (has a link to Page B)
	openPage(t, b, "phase2_nav_a.html")

	// Snapshot should contain page content
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	assertContains(t, snap.Text, "Page A")

	if len(snap.IDMap) == 0 {
		t.Fatal("IDMap is empty")
	}
	if snap.RawCount == 0 {
		t.Fatal("RawCount is 0")
	}

	// Click link to navigate to Page B
	linkID := findID(t, snap.Text, `link "Go to Page B"`)
	must(t, b.Click(linkID))

	// Wait for navigation
	must(t, b.WaitURL("phase2_nav_b.html"))
	must(t, b.WaitLoad())

	// Verify on Page B
	snap2 := mustSnapshot(t, b)
	assertContains(t, snap2.Text, "Page B")
}
