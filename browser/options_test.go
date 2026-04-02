package browser

import "testing"

func TestShouldDisableSandbox(t *testing.T) {
	t.Setenv("KO_BROWSER_NO_SANDBOX", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("CI", "")

	if shouldDisableSandbox() {
		t.Fatal("expected sandbox to stay enabled by default")
	}

	if testing.GOOS == "linux" {
		t.Setenv("CI", "true")
		if !shouldDisableSandbox() {
			t.Fatal("expected sandbox to be disabled on linux CI")
		}
	}
}
