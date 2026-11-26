package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/config"
)

func TestTitleHintFitsOnOneLine(t *testing.T) {
	lines := renderTitleWithHint("Workbench", "Enter 执行 • ↑↓ 历史", 40)
	if len(lines) != 1 {
		t.Fatalf("lines = %#v, want one line", lines)
	}
	plain := stripANSI(lines[0])
	if runewidth.StringWidth(plain) != 40 {
		t.Fatalf("line width = %d, want 40: %q", runewidth.StringWidth(plain), plain)
	}
	if !strings.HasSuffix(plain, "Enter 执行 • ↑↓ 历史") {
		t.Fatalf("hint is not right aligned: %q", plain)
	}
}

func TestWorkspaceHintDoesNotSplitAtActualPanelWidth(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	lines := strings.Split(stripANSI(m.renderWorkspace(56, 18)), "\n")
	hintLine := -1
	for i, line := range lines {
		if strings.Contains(line, "↑↓/j/k") {
			hintLine = i
			break
		}
	}
	if hintLine < 0 {
		t.Fatalf("hint line missing:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.Contains(lines[hintLine], "…") && !strings.Contains(lines[hintLine], m.workspaceHintLine()) {
		t.Fatalf("hint was cut without ellipsis: %q", lines[hintLine])
	}
	for i := hintLine + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], "↑↓/j/k") || strings.Contains(lines[i], "d 删除") || strings.Contains(lines[i], "鼠标点击") {
			t.Fatalf("hint fragment wrapped to line %d: %q", i, lines[i])
		}
	}
}

func TestTitleHintMovesEntirelyToNextLine(t *testing.T) {
	hint := strings.Repeat("very long operation hint ", 10)
	lines := renderTitleWithHint("Connections", hint, 40)
	if len(lines) != 2 {
		t.Fatalf("lines = %#v, want two lines", lines)
	}
	if strings.Contains(stripANSI(lines[0]), "very long") {
		t.Fatalf("hint was split onto title line: %q", stripANSI(lines[0]))
	}
	plain := stripANSI(lines[1])
	if runewidth.StringWidth(plain) != 40 {
		t.Fatalf("hint line width = %d, want 40: %q", runewidth.StringWidth(plain), plain)
	}
	if !strings.HasSuffix(plain, "…") {
		t.Fatalf("long hint was not truncated with marker: %q", plain)
	}
}
