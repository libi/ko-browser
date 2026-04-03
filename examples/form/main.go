package main

import (
	"fmt"
	"log"
	"time"

	"github.com/libi/ko-browser/browser"
)

func main() {
	b, err := browser.New(browser.Options{
		Headless: true,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		log.Fatalf("create browser: %v", err)
	}
	defer b.Close()

	// Navigate to a page with forms
	if err := b.Open("https://httpbin.org/forms/post"); err != nil {
		log.Fatalf("open: %v", err)
	}

	// Take a snapshot to discover form elements
	snap, err := b.Snapshot()
	if err != nil {
		log.Fatalf("snapshot: %v", err)
	}
	fmt.Println("=== Form Snapshot ===")
	fmt.Println(snap.Text)

	// Find form elements by role
	textboxes, _ := b.FindRole("textbox", "")
	fmt.Printf("Found %d textbox(es)\n", len(textboxes.Items))

	// Find by label text
	custNameFields, _ := b.FindLabel("Customer name")
	if len(custNameFields.Items) > 0 {
		id := custNameFields.Items[0].ID
		// Fill clears existing content first, then types new text
		if err := b.Fill(id, "Alice"); err != nil {
			log.Printf("fill customer name: %v", err)
		}
		fmt.Println("Filled customer name with 'Alice'")
	}

	// Type vs Fill:
	// Fill(id, text)  = clear field + type text
	// Type(id, text)  = append text to existing content

	// Checkbox:
	// b.Check(id)   -- checks a checkbox
	// b.Uncheck(id) -- unchecks a checkbox

	// Select dropdown:
	// b.Select(id, "value1", "value2") -- selects option(s) in a <select>

	// Wait utilities:
	// b.WaitSelector("form.loaded")       -- wait for CSS selector
	// b.WaitText("Success")               -- wait for text in page
	// b.WaitURL("/dashboard")             -- wait for URL change
	// b.WaitFunc("window.ready === true") -- wait for JS condition

	// Element state queries
	if len(textboxes.Items) > 0 {
		id := textboxes.Items[0].ID
		visible, _ := b.IsVisible(id)
		enabled, _ := b.IsEnabled(id)
		fmt.Printf("Textbox [%d]: visible=%v, enabled=%v\n", id, visible, enabled)

		// Read value back
		val, _ := b.GetValue(id)
		fmt.Printf("Textbox [%d] value: %q\n", id, val)
	}

	fmt.Println("\nDone!")
}
