package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestKeysViewLoadsNextPage(t *testing.T) {
	m, ctx := newBrowserModel(t)
	m.cfg.PageSize = 1
	if err := m.adapter.SetString(ctx, "key:1", "one", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	if err := m.adapter.SetString(ctx, "key:2", "two", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}

	first := refreshKeys(t, m)
	if !strings.Contains(first.View(), "key:1") {
		t.Fatalf("first page missing key:1:\n%s", first.View())
	}
	first.keyHasMore = true

	next, cmd := first.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatal("next page returned nil command")
	}
	next, _ = next.Update(cmd())
	got := next.(Model)
	if !strings.Contains(got.View(), "key:2") {
		t.Fatalf("second page missing key:2:\n%s", got.View())
	}
}
