package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func inputColumnPositions(view string) []int {
	var positions []int
	for _, line := range strings.Split(stripANSI(view), "\n") {
		if idx := strings.Index(line, "> "); idx >= 0 {
			positions = append(positions, runewidth.StringWidth(line[:idx]))
		}
	}
	return positions
}

func TestChineseProfileFormInputColumnsAlign(t *testing.T) {
	form := newProfileFormWithLanguage(redis.ConnectionProfile{}, profileCreate, "zh")
	positions := inputColumnPositions(form.View())
	if len(positions) < 2 {
		t.Fatalf("expected multiple input rows, got %v", positions)
	}
	for _, pos := range positions[1:] {
		if pos != positions[0] {
			t.Fatalf("input columns not aligned: %v", positions)
		}
	}
}

func TestChineseKeyFormInputColumnsAlign(t *testing.T) {
	form := newKeyCreateFormWithLanguage("zh")
	view := stripANSI(form.View())
	var positions []int
	for _, line := range strings.Split(view, "\n") {
		for _, label := range form.labels {
			if idx := strings.Index(line, label+":"); idx >= 0 {
				positions = append(positions, runewidth.StringWidth(line[:idx]))
				break
			}
		}
	}
	if len(positions) < 2 {
		t.Fatalf("expected multiple labeled rows, got %v", positions)
	}
	for _, pos := range positions[1:] {
		if pos != positions[0] {
			t.Fatalf("label columns not aligned: %v", positions)
		}
	}
}

func TestProfileFormButtonsAtBottom(t *testing.T) {
	form := newProfileFormWithLanguage(redis.ConnectionProfile{}, profileCreate, "zh")
	view := stripANSI(form.View())
	if strings.Index(view, "[保存]") < strings.Index(view, "认证方式:") {
		t.Fatalf("save button is not below fields:\n%s", view)
	}
}

func TestKeyFormButtonsAtBottom(t *testing.T) {
	form := newKeyCreateFormWithLanguage("zh")
	view := stripANSI(form.View())
	if strings.Index(view, "[确认]") < strings.Index(view, "值:") {
		t.Fatalf("OK button is not below fields:\n%s", view)
	}
}

func TestShortTerminalStillShowsProfileFormButtons(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.height = 16
	m.setView(ViewConnections)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	view := stripANSI(next.(Model).View())
	if !strings.Contains(view, "[保存]") || !strings.Contains(view, "[取消]") {
		t.Fatalf("form buttons hidden in short terminal:\n%s", view)
	}
}

func TestPlainKeySearchUsesContainsPattern(t *testing.T) {
	if got := keyMatchPattern("user"); got != "*user*" {
		t.Fatalf("pattern = %q, want *user*", got)
	}
	if got := keyMatchPattern("user*"); got != "user*" {
		t.Fatalf("glob pattern changed: %q", got)
	}
	if got := keyMatchPattern(""); got != "*" {
		t.Fatalf("empty pattern = %q, want *", got)
	}
}

func TestFilterModeNumberDoesNotSwitchView(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.filterMode = true
	m.filterInput.SetValue("")
	m.filterInput.Focus()

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	got := next.(Model)
	if got.ActiveView() != ViewKeys {
		t.Fatalf("view = %v, want Keys", got.ActiveView())
	}
	if got.filterInput.Value() != "1" {
		t.Fatalf("filter value = %q, want 1", got.filterInput.Value())
	}
}

func TestProfileSearchModeNumberDoesNotSwitchView(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)
	m.profileSearchMode = true
	m.profileSearch.SetValue("")
	m.profileSearch.Focus()

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	got := next.(Model)
	if got.ActiveView() != ViewConnections {
		t.Fatalf("view = %v, want Connections", got.ActiveView())
	}
	if got.profileSearch.Value() != "2" {
		t.Fatalf("profile search value = %q, want 2", got.profileSearch.Value())
	}
}
