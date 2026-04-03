package main

import (
	"fmt"
	"log"
	"time"

	"github.com/libi/ko-browser/browser"
)

func main() {
	// 1. Create a headless browser
	b, err := browser.New(browser.Options{
		Headless: true,
		Timeout:  30 * time.Second,
		// Other useful options:
		// Profile:   "/tmp/my-chrome-profile",  // persistent session
		// UserAgent: "MyBot/1.0",
		// Proxy:     "http://proxy:8080",
	})
	if err != nil {
		log.Fatalf("create browser: %v", err)
	}
	defer b.Close()

	// 2. Navigate to a page
	if err := b.Open("https://example.com"); err != nil {
		log.Fatalf("open page: %v", err)
	}

	// 3. Query basic page info
	title, _ := b.GetTitle()
	pageURL, _ := b.GetURL()
	fmt.Printf("Title: %s\nURL:   %s\n\n", title, pageURL)

	// 4. Take a snapshot (accessibility tree)
	// The snapshot assigns each element a numeric display ID like [1], [2], ...
	// You use these IDs in all subsequent interactions.
	snap, err := b.Snapshot()
	if err != nil {
		log.Fatalf("snapshot: %v", err)
	}
	fmt.Println("=== Accessibility Tree ===")
	fmt.Println(snap.Text)

	// 5. Evaluate JavaScript
	result, err := b.Eval("document.title")
	if err != nil {
		log.Fatalf("eval: %v", err)
	}
	fmt.Printf("document.title via Eval: %s\n\n", result)

	// 6. Screenshot
	if err := b.Screenshot("/tmp/ko-browser-example.png"); err != nil {
		log.Fatalf("screenshot: %v", err)
	}
	fmt.Println("Screenshot saved to /tmp/ko-browser-example.png")

	// 7. Find elements in the snapshot
	// FindRole searches by ARIA role; FindText searches by visible text.
	links, _ := b.FindRole("link", "")
	fmt.Printf("\nFound %d links\n", len(links.Items))
	for _, item := range links.Items {
		fmt.Printf("  [%d] %s %q\n", item.ID, item.Role, item.Name)
	}

	// 8. Click an element by its snapshot ID
	// After a Snapshot(), you can interact by element ID.
	// For example, if Snapshot shows: [3] link "More information..."
	// you would call: b.Click(3)
	if len(links.Items) > 0 {
		firstLink := links.Items[0]
		fmt.Printf("\nClicking link [%d] %q ...\n", firstLink.ID, firstLink.Name)
		if err := b.Click(firstLink.ID); err != nil {
			log.Printf("click: %v", err)
		}
		// Wait for navigation
		_ = b.Wait(1 * time.Second)
		newURL, _ := b.GetURL()
		fmt.Printf("After click, URL: %s\n", newURL)
	}

	fmt.Println("\nDone!")
}
