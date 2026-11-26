package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
	"github.com/Ryequ/redis-cockpit/internal/result"
)

func TestLargeKeyValueDoesNotScrollHeaderAway(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 80
	m.height = 24
	m.setView(ViewKeys)
	m.keys = []adapters.KeyInfo{{Name: "large:key", Type: "string", TTL: -1}}
	m.selectedValue = result.Normalize(strings.Repeat("line of value\n", 200))

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > m.height {
		t.Fatalf("view has %d lines, terminal has %d", len(lines), m.height)
	}
	if !strings.Contains(stripANSI(lines[0]), m.btnLabel("[3 Workbench]")) {
		t.Fatalf("header is missing from first line: %q", stripANSI(lines[0]))
	}
}

func TestHeaderTabsRemainClickableAfterLargeValue(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 80
	m.height = 24
	m.setView(ViewKeys)
	m.keys = []adapters.KeyInfo{{Name: "large:key", Type: "string", TTL: -1}}
	m.selectedValue = result.Normalize(strings.Repeat("line of value\n", 200))

	next, _ := m.Update(clickTokenMsg(t, m, m.btnLabel("[3 Workbench]")))
	got := next.(Model)
	if got.ActiveView() != ViewWorkbench {
		t.Fatalf("active view = %v, want Workbench", got.ActiveView())
	}
}

func TestFooterStaysOneLineAtNarrowWidth(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 80
	footer := m.renderFooter(80)
	if strings.Contains(footer, "\n") {
		t.Fatalf("footer wrapped to multiple lines: %q", stripANSI(footer))
	}
}

func TestSmallTerminalDoesNotScroll(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 80
	m.height = 10
	m.setView(ViewKeys)

	lines := strings.Split(m.View(), "\n")
	if len(lines) > m.height {
		t.Fatalf("view has %d lines, terminal has %d", len(lines), m.height)
	}
}

func TestSingleLongLineDoesNotScroll(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 80
	m.height = 24
	m.setView(ViewKeys)
	m.selectedValue = result.Normalize(strings.Repeat("x", 1000))

	lines := strings.Split(m.View(), "\n")
	if len(lines) > m.height {
		t.Fatalf("view has %d lines, terminal has %d", len(lines), m.height)
	}
}

func TestHeaderFitsNarrowTerminalWidth(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 80
	lines := strings.Split(m.View(), "\n")
	header := stripANSI(lines[0])
	if width := runewidth.StringWidth(header); width > m.width {
		t.Fatalf("header width = %d, want <= %d: %q", width, m.width, header)
	}
	if !strings.Contains(header, m.btnLabel("[3 Workbench]")) {
		t.Fatalf("workbench tab was truncated away: %q", header)
	}
}

func TestTruncateToViewportLimitsLongLineWidth(t *testing.T) {
	got := truncateToViewport(strings.Repeat("x", 100), 3, 12)
	if width := runewidth.StringWidth(stripANSI(got)); width > 12 {
		t.Fatalf("line width = %d, want <= 12", width)
	}
}

func TestTruncateToViewportKeepsTruncationMarker(t *testing.T) {
	body := strings.Join([]string{"one", "two", "three", "four"}, "\n")
	got := truncateToViewport(body, 2, 40)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %#v, want 2", lines)
	}
	if !strings.Contains(lines[1], "truncated") {
		t.Fatalf("last line = %q, want truncation marker", lines[1])
	}
}
