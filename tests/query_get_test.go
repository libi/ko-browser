package tests

import (
	"encoding/json"
	"testing"
)

func TestGet_PageInformation(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase2_query.html")
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	headingID := findID(t, snap.Text, `heading "Phase2 Query Test"`)
	usernameID := findID(t, snap.Text, `textbox "Username"`)
	submitID := findID(t, snap.Text, `button "Submit"`)

	t.Run("GetTitle", func(t *testing.T) {
		title, err := b.GetTitle()
		must(t, err)
		assertEqual(t, title, "Phase2 Query Test")
	})

	t.Run("GetURL", func(t *testing.T) {
		u, err := b.GetURL()
		must(t, err)
		assertContains(t, u, "phase2_query.html")
	})

	t.Run("GetText_heading", func(t *testing.T) {
		text, err := b.GetText(headingID)
		must(t, err)
		assertContains(t, text, "Phase2 Query Test")
	})

	t.Run("GetText_button", func(t *testing.T) {
		text, err := b.GetText(submitID)
		must(t, err)
		assertContains(t, text, "Submit")
	})

	t.Run("GetValue", func(t *testing.T) {
		val, err := b.GetValue(usernameID)
		must(t, err)
		assertEqual(t, val, "initial-value")
	})

	t.Run("GetAttr_id", func(t *testing.T) {
		attrVal, err := b.GetAttr(submitID, "id")
		must(t, err)
		assertEqual(t, attrVal, "submit-btn")
	})

	t.Run("GetAttr_type", func(t *testing.T) {
		attrVal, err := b.GetAttr(submitID, "type")
		must(t, err)
		assertEqual(t, attrVal, "button")
	})

	t.Run("GetCount", func(t *testing.T) {
		count, err := b.GetCount(".item")
		must(t, err)
		if count != 5 {
			t.Errorf("GetCount('.item') = %d, want 5", count)
		}
	})

	t.Run("GetCount_zero", func(t *testing.T) {
		count, err := b.GetCount(".nonexistent")
		must(t, err)
		if count != 0 {
			t.Errorf("GetCount('.nonexistent') = %d, want 0", count)
		}
	})

	t.Run("GetBox", func(t *testing.T) {
		box, err := b.GetBox(submitID)
		must(t, err)
		if box.Width <= 0 || box.Height <= 0 {
			t.Errorf("expected positive dimensions, got width=%f height=%f", box.Width, box.Height)
		}
	})

	t.Run("GetStyles", func(t *testing.T) {
		styles, err := b.GetStyles(submitID)
		must(t, err)
		if styles == "" {
			t.Fatal("GetStyles returned empty string")
		}
		// Should be valid JSON
		var m map[string]string
		if err := json.Unmarshal([]byte(styles), &m); err != nil {
			t.Fatalf("GetStyles returned invalid JSON: %v\nGot: %s", err, styles)
		}
		// Should contain display property
		if _, ok := m["display"]; !ok {
			t.Error("GetStyles missing 'display' property")
		}
	})

	// Try to find the info paragraph for GetText/GetHTML/GetAttr tests.
	// Search for the paragraph role node (not its text children) using ": paragraph"
	// to match the role precisely. The paragraph may or may not appear in the AX tree.
	infoID := tryFindID(snap.Text, `: paragraph`)

	if infoID > 0 {
		t.Run("GetText_paragraph", func(t *testing.T) {
			text, err := b.GetText(infoID)
			must(t, err)
			assertContains(t, text, "test")
			assertContains(t, text, "paragraph")
		})

		t.Run("GetHTML_paragraph", func(t *testing.T) {
			html, err := b.GetHTML(infoID)
			must(t, err)
			assertContains(t, html, "<strong>test</strong>")
		})

		t.Run("GetAttr_dataCustom", func(t *testing.T) {
			val, err := b.GetAttr(infoID, "data-custom")
			must(t, err)
			assertEqual(t, val, "hello-attr")
		})
	} else {
		t.Log("Info paragraph not found in AX tree, skipping paragraph-specific tests")
	}
}
