package tests

import "testing"

func TestPhase2_Navigation(t *testing.T) {
	b := newBrowser(t)

	// Start on Page A
	openPage(t, b, "phase2_nav_a.html")
	snap := mustSnapshot(t, b)
	assertContains(t, snap.Text, "Page A")

	// Click link to navigate to Page B
	linkID := findID(t, snap.Text, `link "Go to Page B"`)
	must(t, b.Click(linkID))
	must(t, b.WaitURL("phase2_nav_b.html"))
	must(t, b.WaitLoad())

	snap = mustSnapshot(t, b)
	assertContains(t, snap.Text, "Page B")

	// Back -> Page A
	must(t, b.Back())
	must(t, b.WaitURL("phase2_nav_a.html"))
	must(t, b.WaitLoad())

	snap = mustSnapshot(t, b)
	assertContains(t, snap.Text, "Page A")

	// Forward -> Page B
	must(t, b.Forward())
	must(t, b.WaitURL("phase2_nav_b.html"))
	must(t, b.WaitLoad())

	snap = mustSnapshot(t, b)
	assertContains(t, snap.Text, "Page B")

	// Reload -> still on Page B
	must(t, b.Reload())
	snap = mustSnapshot(t, b)
	assertContains(t, snap.Text, "Page B")

	// Verify URL
	u, err := b.GetURL()
	must(t, err)
	assertContains(t, u, "phase2_nav_b.html")
}
