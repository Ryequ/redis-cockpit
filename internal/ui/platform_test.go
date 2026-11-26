package ui

import (
	"strings"
	"testing"

	"github.com/codebuddy/redis-tui/internal/config"
)

func TestDetectMouseSupportReturnsFalseForDumbTerminal(t *testing.T) {
	t.Setenv("TERM", "dumb")
	if DetectMouseSupport() {
		t.Fatal("DetectMouseSupport returned true for dumb terminal")
	}
}

func TestMouseDisabledHidesFooterHint(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.SetMouseEnabled(false)
	if strings.Contains(m.View(), "Mouse Click") {
		t.Fatalf("footer still shows mouse hint:\n%s", m.View())
	}
}
