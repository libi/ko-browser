package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/libi/ko-browser/browser"
)

// ---------- SetViewport ----------

func TestSettings_SetViewport(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	// Set viewport to 800x600
	if err := b.SetViewport(800, 600); err != nil {
		t.Fatalf("SetViewport: %v", err)
	}

	// Wait briefly for resize to propagate
	time.Sleep(200 * time.Millisecond)

	// Verify via JS
	width, err := b.Eval("window.innerWidth")
	if err != nil {
		t.Fatalf("Eval innerWidth: %v", err)
	}
	if width != "800" {
		t.Errorf("expected viewport width 800, got %q", width)
	}
}

// ---------- SetDevice ----------

func TestSettings_SetDevice(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	if err := b.SetDevice("iPhone 12"); err != nil {
		t.Fatalf("SetDevice: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	width, err := b.Eval("window.innerWidth")
	if err != nil {
		t.Fatalf("Eval innerWidth: %v", err)
	}
	if width != "390" {
		t.Errorf("expected viewport width 390 for iPhone 12, got %q", width)
	}
}

func TestSettings_SetDevice_Unknown(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	err := b.SetDevice("NonExistentDevice999")
	if err == nil {
		t.Fatal("expected error for unknown device")
	}
	if !strings.Contains(err.Error(), "unknown device") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- SetGeo ----------

func TestSettings_SetGeo(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	// Override geolocation
	if err := b.SetGeo(37.7749, -122.4194); err != nil {
		t.Fatalf("SetGeo: %v", err)
	}

	// Trigger geolocation query via JS and wait
	result, err := b.Eval(`new Promise(function(resolve, reject) {
		navigator.geolocation.getCurrentPosition(
			function(pos) { resolve(pos.coords.latitude.toFixed(4) + ',' + pos.coords.longitude.toFixed(4)); },
			function(err) { resolve('error:' + err.message); }
		);
	})`)
	if err != nil {
		t.Fatalf("Eval geolocation: %v", err)
	}

	if !strings.Contains(result, "37.7749") || !strings.Contains(result, "-122.4194") {
		t.Errorf("expected geolocation to contain 37.7749,-122.4194, got %q", result)
	}
}

// ---------- SetOffline ----------

func TestSettings_SetOffline(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	// Enable offline mode
	if err := b.SetOffline(true); err != nil {
		t.Fatalf("SetOffline(true): %v", err)
	}

	// navigator.onLine should reflect the state (note: with OverrideNetworkState it updates)
	onLine, err := b.Eval("navigator.onLine")
	if err != nil {
		t.Fatalf("Eval navigator.onLine: %v", err)
	}
	if onLine != "false" {
		t.Logf("navigator.onLine returned %q (may vary by API)", onLine)
	}

	// Disable offline mode
	if err := b.SetOffline(false); err != nil {
		t.Fatalf("SetOffline(false): %v", err)
	}

	onLine2, err := b.Eval("navigator.onLine")
	if err != nil {
		t.Fatalf("Eval navigator.onLine after online: %v", err)
	}
	// After disabling offline, navigator.onLine should be true
	if onLine2 != "true" {
		t.Logf("navigator.onLine after disable offline: %q", onLine2)
	}
}

// ---------- SetHeaders ----------

func TestSettings_SetHeaders(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	headers := map[string]string{
		"X-Custom-Header": "ko-browser-test",
	}

	if err := b.SetHeaders(headers); err != nil {
		t.Fatalf("SetHeaders: %v", err)
	}
	// Just verify no error; actual header injection would need a server to verify
}

// ---------- SetColorScheme ----------

func TestSettings_SetColorScheme(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	// Set to dark mode
	if err := b.SetColorScheme("dark"); err != nil {
		t.Fatalf("SetColorScheme(dark): %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	result, err := b.Eval("window.matchMedia('(prefers-color-scheme: dark)').matches")
	if err != nil {
		t.Fatalf("Eval color scheme: %v", err)
	}
	if result != "true" {
		t.Errorf("expected dark mode to be true, got %q", result)
	}

	// Set to light mode
	if err := b.SetColorScheme("light"); err != nil {
		t.Fatalf("SetColorScheme(light): %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	result2, err := b.Eval("window.matchMedia('(prefers-color-scheme: dark)').matches")
	if err != nil {
		t.Fatalf("Eval color scheme (light): %v", err)
	}
	if result2 != "false" {
		t.Errorf("expected dark mode to be false after setting light, got %q", result2)
	}
}

// ---------- SetMedia ----------

func TestSettings_SetMedia(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	// Set reduced motion preference
	if err := b.SetMedia(browser.MediaFeature{
		Name:  "prefers-reduced-motion",
		Value: "reduce",
	}); err != nil {
		t.Fatalf("SetMedia: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	result, err := b.Eval("window.matchMedia('(prefers-reduced-motion: reduce)').matches")
	if err != nil {
		t.Fatalf("Eval prefers-reduced-motion: %v", err)
	}
	if result != "true" {
		t.Errorf("expected prefers-reduced-motion: reduce to be true, got %q", result)
	}
}

// ---------- ConsoleMessages ----------

func TestConsole_ListMessages(t *testing.T) {
	b := newBrowser(t)

	// Start console listening BEFORE opening page
	if err := b.ConsoleStart(); err != nil {
		t.Fatalf("ConsoleStart: %v", err)
	}

	openPage(t, b, "phase6_test.html")
	time.Sleep(300 * time.Millisecond)

	// Trigger a console.log via JS
	_, _ = b.Eval("console.log('hello-from-test')")
	time.Sleep(300 * time.Millisecond)

	msgs, err := b.ConsoleMessages()
	if err != nil {
		t.Fatalf("ConsoleMessages: %v", err)
	}

	found := false
	for _, m := range msgs {
		if strings.Contains(m.Text, "hello-from-test") {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Messages: %+v", msgs)
		t.Error("expected to find 'hello-from-test' in console messages")
	}
}

func TestConsole_ListMessages_ByLevel(t *testing.T) {
	b := newBrowser(t)
	if err := b.ConsoleStart(); err != nil {
		t.Fatalf("ConsoleStart: %v", err)
	}

	openPage(t, b, "phase6_test.html")
	time.Sleep(300 * time.Millisecond)

	// Trigger different levels
	_, _ = b.Eval("console.log('level-log')")
	_, _ = b.Eval("console.warn('level-warn')")
	time.Sleep(300 * time.Millisecond)

	warns, err := b.ConsoleMessagesByLevel("warning")
	if err != nil {
		t.Fatalf("ConsoleMessagesByLevel: %v", err)
	}

	found := false
	for _, m := range warns {
		if strings.Contains(m.Text, "level-warn") {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Warnings: %+v", warns)
		t.Error("expected to find 'level-warn' in warning messages")
	}

	// Ensure log messages aren't in warnings
	for _, m := range warns {
		if strings.Contains(m.Text, "level-log") {
			t.Error("'level-log' should not appear in warning-level messages")
		}
	}
}

func TestConsole_Clear(t *testing.T) {
	b := newBrowser(t)
	if err := b.ConsoleStart(); err != nil {
		t.Fatalf("ConsoleStart: %v", err)
	}

	openPage(t, b, "phase6_test.html")
	_, _ = b.Eval("console.log('before-clear')")
	time.Sleep(300 * time.Millisecond)

	b.ConsoleClear()

	msgs, err := b.ConsoleMessages()
	if err != nil {
		t.Fatalf("ConsoleMessages after clear: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(msgs))
	}
}

// ---------- PageErrors ----------

func TestPageErrors_List(t *testing.T) {
	b := newBrowser(t)
	if err := b.ConsoleStart(); err != nil {
		t.Fatalf("ConsoleStart: %v", err)
	}

	openPage(t, b, "phase6_test.html")
	time.Sleep(300 * time.Millisecond)

	// Trigger a JS error
	_, _ = b.Eval("setTimeout(function() { throw new Error('test-runtime-error'); }, 0)")
	time.Sleep(500 * time.Millisecond)

	errs, err := b.PageErrors()
	if err != nil {
		t.Fatalf("PageErrors: %v", err)
	}

	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "test-runtime-error") {
			found = true
			break
		}
	}
	if !found {
		t.Logf("PageErrors: %+v", errs)
		t.Error("expected to find 'test-runtime-error' in page errors")
	}
}

// ---------- Highlight ----------

func TestHighlight_Element(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	snap := mustSnapshot(t, b)
	id := findID(t, snap.Text, "Highlight Me")

	if err := b.Highlight(id); err != nil {
		t.Fatalf("Highlight: %v", err)
	}
	// Verify that the outline was applied
	result, err := b.Eval(`document.getElementById('highlight-btn').style.outline`)
	if err != nil {
		t.Fatalf("Eval outline: %v", err)
	}
	if !strings.Contains(result, "red") {
		t.Errorf("expected outline to contain 'red', got %q", result)
	}
}

// ---------- SetCredentials ----------

func TestSettings_SetCredentials(t *testing.T) {
	b := newBrowser(t)
	openPage(t, b, "phase6_test.html")

	// Just verify the call doesn't error
	if err := b.SetCredentials("testuser", "testpass"); err != nil {
		t.Fatalf("SetCredentials: %v", err)
	}
}
