package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
)

func treeTestModel(keys []adapters.KeyInfo) Model {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.keys = keys
	return m
}

func TestBuildKeyTreeGroupsPrefixSegments(t *testing.T) {
	keys := []adapters.KeyInfo{
		{Name: "user:2:profile", Type: "string"},
		{Name: "user:1:profile", Type: "string"},
		{Name: "user:1:orders", Type: "string"},
		{Name: "standalone", Type: "string"},
	}
	expanded := map[string]bool{"user": true, "user:1": true, "user:2": true}

	rows := buildKeyTree(keys).visibleRows(expanded)
	got := make([]string, 0, len(rows))
	for _, row := range rows {
		got = append(got, strings.Repeat("  ", row.depth)+row.label)
	}
	want := []string{
		"user",
		"  1",
		"    orders",
		"    profile",
		"  2",
		"    profile",
		"standalone",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("tree rows = %#v, want %#v", got, want)
	}
}

func TestTreeToggleShortcutBuildsExpandedTree(t *testing.T) {
	m := treeTestModel([]adapters.KeyInfo{{Name: "user:1:profile", Type: "string"}})

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	got := next.(Model)
	if !got.keyTreeMode {
		t.Fatal("tree mode was not enabled")
	}
	if !got.keyTreeExpanded["user"] || !got.keyTreeExpanded["user:1"] {
		t.Fatalf("tree branches are not expanded: %#v", got.keyTreeExpanded)
	}
	body := got.renderKeysBody()
	if !strings.Contains(body, "user") || !strings.Contains(body, "profile") {
		t.Fatalf("tree body missing nodes:\n%s", body)
	}
}

func TestClickTreeBranchTogglesExpansion(t *testing.T) {
	m := treeTestModel([]adapters.KeyInfo{
		{Name: "user:1:profile", Type: "string"},
		{Name: "user:1:orders", Type: "string"},
	})
	m.keyTreeMode = true
	m.keyTreeExpanded = map[string]bool{"user": true, "user:1": true}

	next, _ := m.Update(clickTokenMsg(t, m, "1/"))
	got := next.(Model)
	if got.keyTreeExpanded["user:1"] {
		t.Fatal("branch was not collapsed")
	}
	if strings.Contains(got.renderKeysBody(), "profile") {
		t.Fatal("collapsed branch still shows children")
	}
}

func TestMouseWheelScrollsKeyList(t *testing.T) {
	keys := make([]adapters.KeyInfo, 30)
	for i := range keys {
		keys[i] = adapters.KeyInfo{Name: "key:" + string(rune('a'+i%26)), Type: "string"}
	}
	m := treeTestModel(keys)
	m.height = 12

	next, _ := m.Update(tea.MouseMsg{
		X: 30, Y: 8, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown,
	})
	got := next.(Model)
	if got.keyScroll != 1 {
		t.Fatalf("keyScroll = %d, want 1", got.keyScroll)
	}

	next, _ = got.Update(tea.MouseMsg{
		X: 30, Y: 8, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelUp,
	})
	got = next.(Model)
	if got.keyScroll != 0 {
		t.Fatalf("keyScroll = %d, want 0", got.keyScroll)
	}
}

func TestEnsureKeyVisibleAdjustsScroll(t *testing.T) {
	keys := make([]adapters.KeyInfo, 30)
	for i := range keys {
		keys[i] = adapters.KeyInfo{Name: "key:" + string(rune('a'+i%26)), Type: "string"}
	}
	m := treeTestModel(keys)
	m.height = 12
	m.selectedKey = 20

	m.ensureKeyVisible()
	if m.keyScroll <= 0 {
		t.Fatalf("keyScroll = %d, want > 0", m.keyScroll)
	}
}

func TestKeyListShowsScrollIndicator(t *testing.T) {
	keys := make([]adapters.KeyInfo, 30)
	for i := range keys {
		keys[i] = adapters.KeyInfo{Name: "key:" + string(rune('a'+i%26)), Type: "string"}
	}
	m := treeTestModel(keys)
	m.height = 12

	body := m.renderKeysBody()
	if !strings.Contains(body, "more") {
		t.Fatalf("scroll indicator missing:\n%s", body)
	}
}
