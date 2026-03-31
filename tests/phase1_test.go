package tests

import (
	"testing"
	"time"
)

// TestPhase1_FormSubmission validates the full form workflow:
// Fill -> KeyboardType -> Type -> Select -> Check -> Press -> Click submit -> verify.
func TestPhase1_FormSubmission(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	nameID := findID(t, snap.Text, `textbox "Name"`)
	notesID := findID(t, snap.Text, `textbox "Notes"`)
	colorID := findID(t, snap.Text, `combobox "Color"`)
	termsID := findID(t, snap.Text, `checkbox "Terms"`)
	submitID := findID(t, snap.Text, `button "Submit"`)

	// Fill name field (clears first, then types)
	must(t, b.Fill(nameID, "Alice"))

	// Type into notes: Focus -> KeyboardType -> Type (appends)
	must(t, b.Focus(notesID))
	must(t, b.KeyboardType("Hello"))
	must(t, b.Type(notesID, " World"))

	// Select color "Blue"
	must(t, b.Select(colorID, "Blue"))

	// Check terms checkbox
	must(t, b.Check(termsID))

	// Press Tab (verify the command works without error)
	must(t, b.Press("Tab"))

	// Scroll to submit button (it is below a 1200px spacer)
	must(t, b.ScrollIntoView(submitID))

	// Click submit to trigger form submission
	must(t, b.Click(submitID))
	time.Sleep(200 * time.Millisecond)

	// Verify form data in status element
	snap2 := mustSnapshot(t, b)
	t.Logf("After submit:\n%s", snap2.Text)
	assertContains(t, snap2.Text, "submitted:Alice:Hello World:blue:true")
}

// TestPhase1_HoverAndDblClick validates hover and double-click interactions.
func TestPhase1_HoverAndDblClick(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)

	submitID := findID(t, snap.Text, `button "Submit"`)

	// Scroll to submit button first (below 1200px spacer)
	must(t, b.ScrollIntoView(submitID))

	// Hover -> status should become "hovered"
	must(t, b.Hover(submitID))
	time.Sleep(200 * time.Millisecond)

	snap2 := mustSnapshot(t, b)
	assertContains(t, snap2.Text, `"hovered"`)

	// DblClick -> status should become "double-clicked"
	must(t, b.DblClick(submitID))
	time.Sleep(200 * time.Millisecond)

	snap3 := mustSnapshot(t, b)
	assertContains(t, snap3.Text, `"double-clicked"`)
}

// TestPhase1_Scroll validates scroll and scroll-into-view.
func TestPhase1_Scroll(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	_ = mustSnapshot(t, b) // need snapshot for ID resolution

	// Scroll down
	must(t, b.Scroll("down", 300))

	// Scroll up
	must(t, b.Scroll("up", 100))

	// ScrollIntoView submit button
	snap := mustSnapshot(t, b)
	submitID := findID(t, snap.Text, `button "Submit"`)
	must(t, b.ScrollIntoView(submitID))
}

// TestPhase1_CheckUncheck validates check/uncheck toggle.
func TestPhase1_CheckUncheck(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase1_interaction.html")
	snap := mustSnapshot(t, b)

	termsID := findID(t, snap.Text, `checkbox "Terms"`)

	// Initially unchecked
	checked, err := b.IsChecked(termsID)
	must(t, err)
	assertFalse(t, checked, "terms should be initially unchecked")

	// Check
	must(t, b.Check(termsID))
	checked, err = b.IsChecked(termsID)
	must(t, err)
	assertTrue(t, checked, "terms should be checked after Check()")

	// Uncheck
	must(t, b.Uncheck(termsID))
	checked, err = b.IsChecked(termsID)
	must(t, err)
	assertFalse(t, checked, "terms should be unchecked after Uncheck()")
}
