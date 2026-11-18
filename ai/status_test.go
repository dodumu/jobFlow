package ai

import "testing"

func TestEnabledFlag(t *testing.T) {
	t.Cleanup(func() { SetEnabled(false) }) // leave things as we found them

	SetEnabled(false)
	if Enabled() {
		t.Fatal("expected AI to be disabled")
	}

	SetEnabled(true)
	if !Enabled() {
		t.Fatal("expected AI to be enabled")
	}
}
