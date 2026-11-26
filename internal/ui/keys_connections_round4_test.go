package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
)

func adaptersKeyProfile(i int) redis.ConnectionProfile {
	name := fmt.Sprintf("profile-%02d", i)
	return redis.ConnectionProfile{ID: name, Name: name, Addr: "127.0.0.1:6379"}
}

func manyProfilesModel(count int) Model {
	cfg := config.Default()
	cfg.Profiles = nil
	for i := 0; i < count; i++ {
		cfg.Profiles = append(cfg.Profiles, adaptersKeyProfile(i))
	}
	m := NewModelWithConfig(cfg, false)
	m.setView(ViewConnections)
	m.height = 24
	return m
}

func TestConnectionsListScrollsWithManyProfiles(t *testing.T) {
	m := manyProfilesModel(40)
	body := stripANSI(m.renderConnectionsBody())
	if strings.Contains(body, "profile-39") {
		t.Fatal("all profiles were rendered; list must be windowed")
	}
	if !strings.Contains(body, "more") {
		t.Fatalf("scroll indicator missing:\n%s", body)
	}
	if !strings.Contains(body, m.txt("Selected:")) {
		t.Fatal("selected area was pushed out of view")
	}
}

func TestMouseWheelScrollsConnectionsList(t *testing.T) {
	m := manyProfilesModel(40)
	next, _ := m.Update(tea.MouseMsg{
		X: 30, Y: 12, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown,
	})
	got := next.(Model)
	if got.profileScroll != 1 {
		t.Fatalf("profileScroll = %d, want 1", got.profileScroll)
	}
}

func TestKeysLoadedAutomaticallyLoadsFirstValue(t *testing.T) {
	m, _ := keyEditorTestModel(t, nil)
	next, cmd := m.Update(keysLoadedMsg{
		gen:   m.generation,
		infos: []adapters.KeyInfo{{Name: "first", Type: "string"}},
	})
	if cmd == nil {
		t.Fatal("initial keys load did not return a value command")
	}
	got := next.(Model)
	if got.selectedKey != 0 {
		t.Fatalf("selectedKey = %d, want 0", got.selectedKey)
	}
}

func TestKeyboardDownLoadsSelectedValue(t *testing.T) {
	m, _ := keyEditorTestModel(t, map[string]string{"first": "one", "second": "two"})
	m.keys = []adapters.KeyInfo{{Name: "first", Type: "string"}, {Name: "second", Type: "string"}}
	m.selectedKey = 0

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd == nil {
		t.Fatal("keyboard movement did not load value")
	}
	msg := cmd().(valueLoadedMsg)
	next, _ = next.Update(msg)
	got := next.(Model)
	if got.selectedValue.Display() != "two" {
		t.Fatalf("value = %q, want two", got.selectedValue.Display())
	}
}

func TestTreeLeafClickSelectsAndLoadsValue(t *testing.T) {
	m, _ := keyEditorTestModel(t, map[string]string{"user:1:profile": "alice", "user:2:orders": "3"})
	m.keys = []adapters.KeyInfo{
		{Name: "user:1:profile", Type: "string"},
		{Name: "user:2:orders", Type: "string"},
	}
	m.keyTreeMode = true
	m.keyTreeExpanded = map[string]bool{"user": true, "user:1": true, "user:2": true}

	next, cmd := m.Update(clickTokenMsg(t, m, "orders"))
	if cmd == nil {
		t.Fatal("tree leaf click did not load value")
	}
	got := next.(Model)
	if got.keys[got.selectedKey].Name != "user:2:orders" {
		t.Fatalf("selected key = %q, want user:2:orders", got.keys[got.selectedKey].Name)
	}
}

func TestNextWithoutMoreShowsMessage(t *testing.T) {
	m, _ := keyEditorTestModel(t, nil)
	m.keyHasMore = false

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := next.(Model)
	if cmd != nil {
		t.Fatal("Next without more pages should not return a command")
	}
	if !strings.Contains(got.Status(), m.txt("No more keys")) {
		t.Fatalf("status = %q, want no-more-keys message", got.Status())
	}
}

func TestKeyValueFieldSupportsMultiline(t *testing.T) {
	form := newKeyCreateFormWithLanguage("zh")
	value := "{\n  \"name\": \"redis\"\n}"
	form.inputs[keyCreateValueField].SetValue(value)
	if got := form.inputs[keyCreateValueField].Value(); got != value {
		t.Fatalf("multiline value was not preserved: %q", got)
	}
	if !form.inputs[keyCreateValueField].multiline {
		t.Fatal("value field is not a multiline field")
	}
}
