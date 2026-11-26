package ui

import (
	"strings"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/config"
)

func TestHintLineMovedToTitleRow(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.height = 24
	m.setView(ViewConnections)
	lines := strings.Split(m.View(), "\n")
	if len(lines) < 3 {
		t.Fatalf("view too short: %#v", lines)
	}
	titleLine := stripANSI(lines[3])
	if !strings.Contains(titleLine, m.txt("Connections")) {
		t.Fatalf("title missing: %q", titleLine)
	}
	found := false
	for i := 3; i <= 4 && i < len(lines); i++ {
		if strings.Contains(stripANSI(lines[i]), "↑↓/j/k") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("hint not on title row or immediately below:\n%s", strings.Join(lines[3:minInt(len(lines), 6)], "\n"))
	}
}

func TestValueTextareaHasNoLineNumbers(t *testing.T) {
	form := newKeyCreateFormWithLanguage("zh")
	if form.inputs[keyCreateValueField].area.ShowLineNumbers {
		t.Fatal("textarea line numbers are enabled")
	}
}
