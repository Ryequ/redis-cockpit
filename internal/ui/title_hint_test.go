package ui

import (
	"strings"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/config"
)

func TestTitleRowDoesNotContainHint(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.height = 24
	m.setView(ViewConnections)
	lines := strings.Split(m.View(), "\n")
	if len(lines) < 4 {
		t.Fatalf("view too short: %#v", lines)
	}
	if !strings.Contains(stripANSI(lines[3]), m.txt("Connections")) {
		t.Fatalf("title missing: %q", stripANSI(lines[3]))
	}
	if strings.Contains(stripANSI(lines[3]), "↑↓") {
		t.Fatalf("hint still on title row: %q", stripANSI(lines[3]))
	}
}

func TestValueTextareaHasNoLineNumbers(t *testing.T) {
	form := newKeyCreateFormWithLanguage("zh")
	if form.inputs[keyCreateValueField].area.ShowLineNumbers {
		t.Fatal("textarea line numbers are enabled")
	}
}
