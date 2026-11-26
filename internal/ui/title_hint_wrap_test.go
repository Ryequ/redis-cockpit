package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/config"
)

func TestBottomHintLineIsRightAligned(t *testing.T) {
	line := renderBottomHintLine("Enter 执行 • ↑↓ 历史", 40)
	plain := stripANSI(line)
	if runewidth.StringWidth(plain) != 40 {
		t.Fatalf("line width = %d, want 40: %q", runewidth.StringWidth(plain), plain)
	}
	if !strings.HasSuffix(strings.TrimRight(plain, " "), "Enter 执行 • ↑↓ 历史") {
		t.Fatalf("hint is not right aligned: %q", plain)
	}
}

func TestBottomHintLineTruncatesLongHint(t *testing.T) {
	line := renderBottomHintLine(strings.Repeat("very long operation hint ", 10), 40)
	plain := stripANSI(line)
	if runewidth.StringWidth(plain) != 40 {
		t.Fatalf("line width = %d, want 40: %q", runewidth.StringWidth(plain), plain)
	}
	if !strings.HasSuffix(strings.TrimRight(plain, " "), "…") {
		t.Fatalf("long hint was not truncated with marker: %q", plain)
	}
}

func TestWorkspaceHintSitsAtBottomRight(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)
	m.height = 24

	lines := strings.Split(m.View(), "\n")
	hintIndexes := []int{}
	for i, line := range lines {
		if strings.Contains(stripANSI(line), "↑↓") {
			hintIndexes = append(hintIndexes, i)
		}
	}
	if len(hintIndexes) != 1 {
		t.Fatalf("expected one hint line, got %v:\n%s", hintIndexes, strings.Join(lines, "\n"))
	}
	hintIndex := hintIndexes[0]
	if hintIndex < len(lines)-5 {
		t.Fatalf("hint is not near bottom: index=%d lines=%d", hintIndex, len(lines))
	}
	if hintIndex == 3 {
		t.Fatal("hint is still on the title row")
	}
}
