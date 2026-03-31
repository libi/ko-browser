package tests

import (
	"testing"
	"time"
)

func TestPhase2_Wait(t *testing.T) {
	b := newBrowser(t)

	t.Run("WaitLoad", func(t *testing.T) {
		openPage(t, b, "phase2_query.html")
		must(t, b.WaitLoad())
	})

	t.Run("Wait_duration", func(t *testing.T) {
		start := time.Now()
		must(t, b.Wait(200*time.Millisecond))
		elapsed := time.Since(start)
		if elapsed < 150*time.Millisecond {
			t.Errorf("Wait returned too quickly: %v", elapsed)
		}
	})

	t.Run("WaitText", func(t *testing.T) {
		// Re-open page so the 1.5s setTimeout starts fresh
		openPage(t, b, "phase2_query.html")
		must(t, b.WaitText("Dynamic content loaded!", 10*time.Second))
	})

	t.Run("WaitFunc", func(t *testing.T) {
		// Re-open page so window.dynamicReady resets
		openPage(t, b, "phase2_query.html")
		must(t, b.WaitFunc("window.dynamicReady", 10*time.Second))
	})

	t.Run("WaitSelector", func(t *testing.T) {
		// Re-open page so #delayed-text is removed and re-added after 1.5s
		openPage(t, b, "phase2_query.html")
		must(t, b.WaitSelector("#delayed-text", 10*time.Second))
	})
}
