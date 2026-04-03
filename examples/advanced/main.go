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

	_ = b.Open("https://example.com")

	// Tab management
	tabs, _ := b.TabList()
	fmt.Printf("Tabs: %d\n", len(tabs))

	// Open a new tab
	if err := b.TabNew("https://example.org"); err != nil {
		log.Printf("new tab: %v", err)
	}

	// List tabs again
	tabs, _ = b.TabList()
	for _, t := range tabs {
		fmt.Printf("  Tab %d: %s (active=%v)\n", t.Index, t.URL, t.Active)
	}

	// Switch between tabs
	_ = b.TabSwitch(0)

	// Close the second tab
	_ = b.TabClose(1)

	// Viewport and device emulation
	_ = b.SetViewport(1280, 720)
	_ = b.SetDevice("iPhone 12")

	// Screenshot options
	// Full-page screenshot
	_ = b.Screenshot("/tmp/ko-fullpage.png", browser.ScreenshotOptions{FullPage: true})

	// JPEG with quality
	_ = b.Screenshot("/tmp/ko-quality.jpg", browser.ScreenshotOptions{Quality: 80})

	// Screenshot a specific element (by snapshot ID)
	snap, _ := b.Snapshot()
	if len(snap.IDMap) > 0 {
		_ = b.Screenshot("/tmp/ko-element.png", browser.ScreenshotOptions{ElementID: 1})
	}

	// Get screenshot as bytes (for in-memory processing)
	data, _ := b.ScreenshotToBytes()
	fmt.Printf("Screenshot bytes: %d\n", len(data))

	// Cookies and storage
	cookies, _ := b.CookiesGet()
	fmt.Printf("Cookies: %d\n", len(cookies))

	// Set a cookie
	_ = b.CookieSet(browser.CookieInfo{
		Name:  "session",
		Value: "abc123",
	})

	// LocalStorage
	_ = b.StorageSet("local", "myKey", "myValue")
	val, _ := b.StorageGet("local", "myKey")
	fmt.Printf("localStorage.myKey = %q\n", val)

	// Console capture
	_ = b.ConsoleStart()
	_, _ = b.Eval("console.log('hello from ko-browser')")
	time.Sleep(200 * time.Millisecond)
	msgs, _ := b.ConsoleMessages()
	for _, m := range msgs {
		fmt.Printf("  console.%s: %s\n", m.Level, m.Text)
	}

	// Network request logging
	_ = b.NetworkStartLogging()
	_ = b.Open("https://example.com")
	reqs, _ := b.NetworkRequests()
	fmt.Printf("Network requests captured: %d\n", len(reqs))

	// Snapshot with options
	// Interactive-only: only show clickable / fillable elements
	snapInteractive, _ := b.Snapshot(browser.SnapshotOptions{InteractiveOnly: true})
	fmt.Println("\n=== Interactive Elements Only ===")
	fmt.Println(snapInteractive.Text)

	// Compact mode: omit unnamed structural wrappers
	snapCompact, _ := b.Snapshot(browser.SnapshotOptions{Compact: true})
	fmt.Println("=== Compact Snapshot ===")
	fmt.Println(snapCompact.Text)

	// Scoped to a CSS selector
	snapScoped, _ := b.Snapshot(browser.SnapshotOptions{Selector: "body > div"})
	fmt.Println("=== Scoped Snapshot ===")
	fmt.Println(snapScoped.Text)

	fmt.Println("Done!")
}
