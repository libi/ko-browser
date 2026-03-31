package tests

import "testing"

func TestPhase2_StateQueries(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_query.html")
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	submitID := findID(t, snap.Text, `button "Submit"`)
	agreeID := findID(t, snap.Text, `checkbox "Agree"`)
	emailID := findID(t, snap.Text, `textbox "Email"`)

	t.Run("IsVisible_visible", func(t *testing.T) {
		vis, err := b.IsVisible(submitID)
		must(t, err)
		assertTrue(t, vis, "submit button should be visible")
	})

	t.Run("IsEnabled_enabled", func(t *testing.T) {
		enabled, err := b.IsEnabled(submitID)
		must(t, err)
		assertTrue(t, enabled, "submit button should be enabled")
	})

	t.Run("IsEnabled_disabled", func(t *testing.T) {
		enabled, err := b.IsEnabled(emailID)
		must(t, err)
		assertFalse(t, enabled, "email input should be disabled")
	})

	t.Run("IsChecked_unchecked", func(t *testing.T) {
		checked, err := b.IsChecked(agreeID)
		must(t, err)
		assertFalse(t, checked, "agree checkbox should be initially unchecked")
	})

	t.Run("IsChecked_afterCheck", func(t *testing.T) {
		must(t, b.Check(agreeID))
		checked, err := b.IsChecked(agreeID)
		must(t, err)
		assertTrue(t, checked, "agree checkbox should be checked after Check()")
	})
}
