package tests

import (
	"testing"
)

// TestPhase9_GetTextOnFormElements tests that GetText correctly returns
// meaningful content for form elements (input, textarea, select) where
// innerText/textContent is always empty.
// This was a bug: "kbr get text <id>" returned empty for textbox elements.
func TestPhase9_GetTextOnFormElements(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase9_get_test.html")
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	searchID := findID(t, snap.Text, `textbox "Search"`)
	emptyInputID := findID(t, snap.Text, `textbox "Empty Input"`)
	commentID := findID(t, snap.Text, `textbox "Comment"`)
	colorID := findID(t, snap.Text, `combobox "Color"`)
	buttonID := findID(t, snap.Text, `button "Click Me"`)
	passwordID := findID(t, snap.Text, `textbox "Password"`)

	t.Run("GetText_input_with_value", func(t *testing.T) {
		// Input with value="hello world" should return "hello world", not ""
		text, err := b.GetText(searchID)
		must(t, err)
		assertEqual(t, text, "hello world")
	})

	t.Run("GetText_input_empty_value_with_placeholder", func(t *testing.T) {
		// Empty value but has placeholder should return placeholder
		text, err := b.GetText(emptyInputID)
		must(t, err)
		assertEqual(t, text, "请输入内容")
	})

	t.Run("GetText_textarea", func(t *testing.T) {
		text, err := b.GetText(commentID)
		must(t, err)
		assertContains(t, text, "This is a comment")
	})

	t.Run("GetText_select", func(t *testing.T) {
		// Select returns value ("green") since value is non-empty
		text, err := b.GetText(colorID)
		must(t, err)
		assertContains(t, text, "green")
	})

	t.Run("GetText_button", func(t *testing.T) {
		text, err := b.GetText(buttonID)
		must(t, err)
		assertEqual(t, text, "Click Me")
	})

	t.Run("GetText_password_input", func(t *testing.T) {
		// Password input has value="secret123"
		text, err := b.GetText(passwordID)
		must(t, err)
		assertEqual(t, text, "secret123")
	})
}

// TestPhase9_GetHTMLOnFormElements tests that GetHTML returns meaningful
// content for void/form elements (where innerHTML is always empty).
func TestPhase9_GetHTMLOnFormElements(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase9_get_test.html")
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	searchID := findID(t, snap.Text, `textbox "Search"`)
	buttonID := findID(t, snap.Text, `button "Click Me"`)

	t.Run("GetHTML_input_returns_outerHTML", func(t *testing.T) {
		// Input is a void element, innerHTML is always "".
		// Should fall back to outerHTML.
		html, err := b.GetHTML(searchID)
		must(t, err)
		if html == "" {
			t.Fatal("GetHTML on input returned empty string (the original bug)")
		}
		assertContains(t, html, "input")
		assertContains(t, html, `value="hello world"`)
	})

	t.Run("GetHTML_button_returns_innerHTML", func(t *testing.T) {
		html, err := b.GetHTML(buttonID)
		must(t, err)
		assertEqual(t, html, "Click Me")
	})
}

// TestPhase9_GetValueOnFormElements tests that GetValue works on all form elements.
func TestPhase9_GetValueOnFormElements(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase9_get_test.html")
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	searchID := findID(t, snap.Text, `textbox "Search"`)
	emptyInputID := findID(t, snap.Text, `textbox "Empty Input"`)
	commentID := findID(t, snap.Text, `textbox "Comment"`)
	colorID := findID(t, snap.Text, `combobox "Color"`)
	passwordID := findID(t, snap.Text, `textbox "Password"`)

	t.Run("GetValue_input_with_value", func(t *testing.T) {
		val, err := b.GetValue(searchID)
		must(t, err)
		assertEqual(t, val, "hello world")
	})

	t.Run("GetValue_empty_input", func(t *testing.T) {
		val, err := b.GetValue(emptyInputID)
		must(t, err)
		assertEqual(t, val, "")
	})

	t.Run("GetValue_textarea", func(t *testing.T) {
		val, err := b.GetValue(commentID)
		must(t, err)
		assertEqual(t, val, "This is a comment")
	})

	t.Run("GetValue_select", func(t *testing.T) {
		val, err := b.GetValue(colorID)
		must(t, err)
		assertEqual(t, val, "green")
	})

	t.Run("GetValue_password", func(t *testing.T) {
		val, err := b.GetValue(passwordID)
		must(t, err)
		assertEqual(t, val, "secret123")
	})
}

// TestPhase9_GetTextOnNonInteractiveElements tests GetText on elements
// that are not interactive (heading, paragraph, link, status div).
func TestPhase9_GetTextOnNonInteractiveElements(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase9_get_test.html")
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	headingID := findID(t, snap.Text, `heading "Get Command Test"`)
	linkID := findID(t, snap.Text, `link "Visit Example"`)

	t.Run("GetText_heading", func(t *testing.T) {
		text, err := b.GetText(headingID)
		must(t, err)
		assertEqual(t, text, "Get Command Test")
	})

	t.Run("GetText_link", func(t *testing.T) {
		text, err := b.GetText(linkID)
		must(t, err)
		assertContains(t, text, "Visit")
		assertContains(t, text, "Example")
	})

	statusID := tryFindID(snap.Text, `status`)
	if statusID > 0 {
		t.Run("GetText_status_div", func(t *testing.T) {
			text, err := b.GetText(statusID)
			must(t, err)
			assertContains(t, text, "Status: OK")
		})
	}
}

// TestPhase9_GetHTMLOnNonInteractiveElements tests GetHTML on regular elements
// and void elements like img.
func TestPhase9_GetHTMLOnNonInteractiveElements(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase9_get_test.html")
	snap := mustSnapshot(t, b)
	t.Logf("Snapshot:\n%s", snap.Text)

	linkID := findID(t, snap.Text, `link "Visit Example"`)

	t.Run("GetHTML_link_with_nested_content", func(t *testing.T) {
		html, err := b.GetHTML(linkID)
		must(t, err)
		assertContains(t, html, "Visit")
		assertContains(t, html, "<em>Example</em>")
	})

	imgID := tryFindID(snap.Text, `image "Logo"`)
	if imgID == 0 {
		imgID = tryFindID(snap.Text, `image`)
	}
	if imgID > 0 {
		t.Run("GetHTML_img_returns_outerHTML", func(t *testing.T) {
			html, err := b.GetHTML(imgID)
			must(t, err)
			if html == "" {
				t.Fatal("GetHTML on img returned empty string")
			}
			assertContains(t, html, "img")
			assertContains(t, html, `alt="Logo"`)
		})
	} else {
		t.Log("image not found in snapshot, skipping img test")
	}
}

// TestPhase9_GetAfterDOMMutation tests that get commands still work
// after the DOM has been mutated (stale backend node IDs scenario).
func TestPhase9_GetAfterDOMMutation(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase9_get_test.html")

	snap := mustSnapshot(t, b)
	searchID := findID(t, snap.Text, `textbox "Search"`)
	buttonID := findID(t, snap.Text, `button "Click Me"`)

	// Mutate the DOM to potentially invalidate backend node IDs
	_, err := b.Eval(`document.getElementById('info-para').remove()`)
	must(t, err)

	t.Run("GetText_input_after_DOM_mutation", func(t *testing.T) {
		text, err := b.GetText(searchID)
		must(t, err)
		assertEqual(t, text, "hello world")
	})

	t.Run("GetText_button_after_DOM_mutation", func(t *testing.T) {
		text, err := b.GetText(buttonID)
		must(t, err)
		assertEqual(t, text, "Click Me")
	})

	t.Run("GetHTML_input_after_DOM_mutation", func(t *testing.T) {
		html, err := b.GetHTML(searchID)
		must(t, err)
		if html == "" {
			t.Fatal("GetHTML on input returned empty after DOM mutation")
		}
		assertContains(t, html, "input")
	})

	t.Run("GetValue_input_after_DOM_mutation", func(t *testing.T) {
		val, err := b.GetValue(searchID)
		must(t, err)
		assertEqual(t, val, "hello world")
	})
}
