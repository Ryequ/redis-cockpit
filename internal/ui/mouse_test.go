package ui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func clickTokenMsg(t *testing.T, m Model, token string) tea.MouseMsg {
	t.Helper()
	x, y, ok := locateToken(m.View(), m.btnLabel(token))
	if !ok {
		x, y, ok = locateToken(m.View(), token)
	}
	if !ok {
		t.Fatalf("token %q not found in view", token)
	}
	return tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
}

func TestLocateTokenUsesCellWidthForUnicodeBorder(t *testing.T) {
	view := "│ [Y] Delete"
	x, y, ok := locateToken(view, "[Y] Delete")
	if !ok {
		t.Fatal("token not found")
	}
	if y != 0 {
		t.Fatalf("y = %d, want 0", y)
	}
	if x != 2 {
		t.Fatalf("x = %d, want 2 terminal cells", x)
	}
	if !clickToken(tea.MouseMsg{X: 2, Y: 0}, view, "[Y] Delete") {
		t.Fatal("click at visible token start did not hit")
	}
	if clickToken(tea.MouseMsg{X: 1, Y: 0}, view, "[Y] Delete") {
		t.Fatal("click before visible token must not hit")
	}
	if !clickToken(tea.MouseMsg{X: 4, Y: 0}, view, "[Y] Delete") {
		t.Fatal("byte-offset column should still be inside the visible token")
	}
}

func TestMouseClickWorkspaceTitleDoesNotSwitchView(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)

	x, y, ok := locateToken(m.View(), "│  Keys")
	if !ok {
		t.Fatal("workspace title not found")
	}
	next, _ := m.Update(tea.MouseMsg{X: x + 4, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	got := next.(Model)
	if got.ActiveView() != ViewKeys {
		t.Fatalf("active view = %v, want Keys", got.ActiveView())
	}
}

func TestMouseClickFocusedFieldWorks(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(clickTokenMsg(t, m, "[N new]"))
	got := next.(Model)
	got.profileForm.setFocus(profileFieldAddr)

	next, _ = got.Update(clickTokenMsg(t, got, got.btnLabel("Name:")))
	got = next.(Model)
	if got.profileForm.focus != profileFieldName {
		t.Fatalf("focus = %d, want name field", got.profileForm.focus)
	}
	if !got.profileForm.inputs[profileFieldName].Focused() {
		t.Fatal("name field was not focused")
	}
}

func TestMouseClickBooleanFieldToggles(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(clickTokenMsg(t, m, "[N new]"))
	got := next.(Model)

	next, _ = got.Update(clickTokenMsg(t, got, "TLS:"))
	got = next.(Model)
	if !got.profileForm.bools[profileBoolTLS] {
		t.Fatal("TLS was not toggled")
	}
	if got.profileForm.focus != profileFieldCount+profileBoolTLS {
		t.Fatalf("focus = %d, want TLS bool", got.profileForm.focus)
	}
}

func TestMouseClickProfileDeleteConfirmation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	m := NewModelWithConfig(config.Default(), false)
	m.SetConfigPath(path)
	m.setView(ViewConnections)
	m.pendingProfileDelete = "local"

	next, _ := m.Update(clickTokenMsg(t, m, "[Y] Delete"))
	got := next.(Model)
	if got.pendingProfileDelete != "" {
		t.Fatal("profile delete confirmation was not cleared")
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(saved.Profiles) != 0 {
		t.Fatalf("profiles = %#v, want empty", saved.Profiles)
	}
}

func TestMouseClickConnectButtonStartsConnection(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, cmd := m.Update(clickTokenMsg(t, m, "[Connect]"))
	got := next.(Model)
	if cmd == nil {
		t.Fatal("connect button did not return a command")
	}
	if !strings.Contains(got.Status(), "Connecting") {
		t.Fatalf("status = %q, want connecting", got.Status())
	}
}

func TestMouseClickHeaderWorkbenchTabSwitchesView(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)

	next, _ := m.Update(clickTokenMsg(t, m, "[3 Workbench]"))
	got := next.(Model)
	if got.ActiveView() != ViewWorkbench {
		t.Fatalf("active view = %v, want Workbench", got.ActiveView())
	}
}

func TestMouseClickNewProfileButtonOpensForm(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(clickTokenMsg(t, m, "[N new]"))
	got := next.(Model)
	if got.profileForm == nil || got.profileForm.mode != profileCreate {
		t.Fatalf("profile form = %#v, want create form", got.profileForm)
	}
}

func TestMouseClickFieldRowFocusesField(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(clickTokenMsg(t, m, "[N new]"))
	got := next.(Model)
	next, _ = got.Update(clickTokenMsg(t, got, got.btnLabel("Address:")))
	got = next.(Model)

	if got.profileForm.focus != profileFieldAddr {
		t.Fatalf("focus = %d, want address field", got.profileForm.focus)
	}
	if !got.profileForm.inputs[profileFieldAddr].Focused() {
		t.Fatal("address field was not focused")
	}
}

func TestMouseClickAnywhereOnEditFormRowFocusesField(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(clickTokenMsg(t, m, "[E edit]"))
	got := next.(Model)
	x, y, width, ok := locateLineContaining(got.View(), got.btnLabel("Address:"))
	if !ok {
		t.Fatal("address row not found")
	}
	next, _ = got.Update(tea.MouseMsg{X: x + width - 2, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	got = next.(Model)

	if got.profileForm.focus != profileFieldAddr {
		t.Fatalf("focus = %d, want address field", got.profileForm.focus)
	}
	if !got.profileForm.inputs[profileFieldAddr].Focused() {
		t.Fatal("address field was not focused")
	}
}

func TestMouseClickExecuteRunsWorkbenchCommand(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewWorkbench)
	m.commandInput.SetValue("PING")

	next, _ := m.Update(clickTokenMsg(t, m, "[Execute]"))
	got := next.(Model)
	if !strings.Contains(got.Status(), "Executing PING") {
		t.Fatalf("status = %q, want executing command", got.Status())
	}
}

func TestMouseClickYesConfirmsDelete(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.pendingDelete = []string{"sample:key"}

	next, _ := m.Update(clickTokenMsg(t, m, "[Y] Delete"))
	got := next.(Model)
	if got.pendingDelete != nil {
		t.Fatal("delete confirmation was not cleared")
	}
}

func TestMouseClickNoCancelsDelete(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.pendingDelete = []string{"sample:key"}

	next, _ := m.Update(clickTokenMsg(t, m, "[N] Cancel"))
	got := next.(Model)
	if got.pendingDelete != nil {
		t.Fatal("delete confirmation was not cancelled")
	}
	if !strings.Contains(got.Status(), "cancelled") {
		t.Fatalf("status = %q, want cancelled", got.Status())
	}
}

func TestMouseClickProfileRowSelectsProfile(t *testing.T) {
	cfg := config.Default()
	cfg.Profiles = append(cfg.Profiles, redis.ConnectionProfile{ID: "prod", Name: "prod", Addr: "redis.example.com:6379"})
	m := NewModelWithConfig(cfg, false)
	m.setView(ViewConnections)

	next, _ := m.Update(clickTokenMsg(t, m, "2. prod"))
	got := next.(Model)
	if got.selectedProfile != 1 {
		t.Fatalf("selected profile = %d, want 1", got.selectedProfile)
	}
}

func TestMouseClickEditProfileButtonOpensForm(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(clickTokenMsg(t, m, "[E edit]"))
	got := next.(Model)
	if got.profileForm == nil || got.profileForm.mode != profileEdit {
		t.Fatalf("profile form = %#v, want edit form", got.profileForm)
	}
}

func TestMouseClickYesExecutesCriticalCommand(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewWorkbench)
	m.pendingCommand = []string{"FLUSHALL"}

	next, cmd := m.Update(clickTokenMsg(t, m, "[Y] Execute"))
	got := next.(Model)
	if got.pendingCommand != nil {
		t.Fatal("critical command confirmation was not cleared")
	}
	if cmd == nil {
		t.Fatal("confirming critical command did not execute")
	}
}

func TestMouseClickRefreshOpsReturnsCommand(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewOps)

	_, cmd := m.Update(clickTokenMsg(t, m, "[Refresh]"))
	if cmd == nil {
		t.Fatal("refresh button did not return a command")
	}
}

func TestMouseClickSubscribeRoutesToStreamHandler(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewStreams)

	next, _ := m.Update(clickTokenMsg(t, m, "[Sub]"))
	got := next.(Model)
	if !strings.Contains(got.Status(), "Connect to Redis before subscribing") {
		t.Fatalf("status = %q, want subscribe handler message", got.Status())
	}
}

func TestMouseClickCloseHelpModal(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.showHelpModal = true

	next, _ := m.Update(clickTokenMsg(t, m, "[Close]"))
	got := next.(Model)
	if got.showHelpModal {
		t.Fatal("help modal was not closed")
	}
}
