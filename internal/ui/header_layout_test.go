package ui

import (
	"strings"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/config"
)

func TestSmallWindowKeepsHeaderAfterConnectAndReturn(t *testing.T) {
	cfg := config.Default()
	m := NewModelWithConfig(cfg, false)
	m.width = 80
	m.height = 10
	m.setView(ViewConnections)

	m.setView(ViewKeys)
	if m.ActiveView() != ViewKeys {
		t.Fatalf("view after connect = %v, want Keys", m.ActiveView())
	}

	next, _ := m.Update(clickTokenMsg(t, m, "[1 Connections]"))
	m = next.(Model)
	if m.ActiveView() != ViewConnections {
		t.Fatalf("view after return = %v, want Connections", m.ActiveView())
	}

	lines := strings.Split(m.View(), "\n")
	if len(lines) > m.height {
		t.Fatalf("view has %d lines, terminal has %d", len(lines), m.height)
	}
	if !strings.Contains(stripANSI(lines[0]), "[1 Connections]") {
		t.Fatalf("header missing after return: %q", stripANSI(lines[0]))
	}
}

func TestSmallWindowRangeKeepsHeaderAndFooter(t *testing.T) {
	for height := 8; height <= 20; height++ {
		m := NewModelWithConfig(config.Default(), false)
		m.width = 80
		m.height = height
		m.setView(ViewConnections)

		lines := strings.Split(m.View(), "\n")
		if len(lines) > height {
			t.Fatalf("height %d: view has %d lines", height, len(lines))
		}
		if !strings.Contains(stripANSI(lines[0]), "[1 Connections]") {
			t.Fatalf("height %d: header missing: %q", height, stripANSI(lines[0]))
		}
		last := stripANSI(lines[len(lines)-1])
		if !strings.Contains(last, "Tab") {
			t.Fatalf("height %d: footer missing: %q", height, last)
		}
	}
}

func TestVerySmallWindowStillShowsHeaderTabs(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 100
	m.height = 5

	view := m.View()
	if !strings.Contains(stripANSI(view), "[1 Connections]") {
		t.Fatalf("header tabs missing in very small window: %q", stripANSI(view))
	}
}

func TestActiveHeaderTabHasVisibleMarker(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 100
	m.setView(ViewWorkbench)

	header := stripANSI(m.renderHeader(100))
	if !strings.Contains(header, "▸[3 Workbench]") {
		t.Fatalf("active tab marker missing: %q", header)
	}
	if strings.Contains(header, "▸[1 Connections]") {
		t.Fatalf("inactive tab was marked active: %q", header)
	}
}
