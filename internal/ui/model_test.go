package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelRendersWorkbenchChrome(t *testing.T) {
	m := NewModel()
	view := m.View()

	for _, want := range []string{"Redis Cockpit", "连接管理", "键浏览", "命令工作台", "运维监控", "流式任务", "Tab", "鼠标"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
}

func TestNewModelHandlesWindowSize(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	got := next.(Model)

	if got.Width() != 120 || got.Height() != 40 {
		t.Fatalf("window size = %dx%d, want 120x40", got.Width(), got.Height())
	}
}

func TestNewModelQuitsOnCtrlC(t *testing.T) {
	m := NewModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c returned nil command, want quit command")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c produced %T, want tea.QuitMsg", msg)
	}
}

func TestNewModelSwitchesViewWithNumberKeys(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	got := next.(Model)

	if got.ActiveView() != ViewWorkbench {
		t.Fatalf("active view = %v, want %v", got.ActiveView(), ViewWorkbench)
	}
}

func TestNewModelHandlesMouseClickOnNavigation(t *testing.T) {
	m := NewModel()
	next, _ := m.Update(tea.MouseMsg{
		X:      2,
		Y:      2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	got := next.(Model)

	if got.ActiveView() != ViewKeys {
		t.Fatalf("active view = %v, want %v", got.ActiveView(), ViewKeys)
	}
}
