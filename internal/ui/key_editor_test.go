package ui

import (
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	tea "github.com/charmbracelet/bubbletea"
	goredis "github.com/redis/go-redis/v9"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
	"github.com/Ryequ/redis-cockpit/internal/result"
)

func keyEditorTestModel(t *testing.T, keys map[string]string) (Model, *redis.Executor) {
	t.Helper()
	server := miniredis.RunT(t)
	for key, value := range keys {
		server.Set(key, value)
	}
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	executor := redis.NewExecutor(client)
	m := NewModelWithConfig(config.Default(), false)
	m.adapter = adapters.New(executor)
	m.setView(ViewKeys)
	return m, executor
}

func TestAddKeyOpensForm(t *testing.T) {
	m, _ := keyEditorTestModel(t, nil)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	got := next.(Model)
	if got.keyForm == nil || got.keyForm.mode != keyFormCreate {
		t.Fatalf("key form = %#v, want create form", got.keyForm)
	}
}

func TestMouseClickNewKeyButtonOpensForm(t *testing.T) {
	m, _ := keyEditorTestModel(t, nil)

	next, _ := m.Update(clickTokenMsg(t, m, "[New key]"))
	got := next.(Model)
	if got.keyForm == nil || got.keyForm.mode != keyFormCreate {
		t.Fatalf("key form = %#v, want create form", got.keyForm)
	}
}

func TestMouseClickEditValueButtonOpensForm(t *testing.T) {
	m, _ := keyEditorTestModel(t, map[string]string{"greeting": "hello"})
	m.keys = []adapters.KeyInfo{{Name: "greeting", Type: "string"}}
	m.selectedValue = result.Normalize("hello")

	next, _ := m.Update(clickTokenMsg(t, m, "[Edit value]"))
	confirming := next.(Model)
	if confirming.pendingEditConfirm != "greeting" {
		t.Fatalf("edit confirmation = %q, want greeting", confirming.pendingEditConfirm)
	}
	next, _ = confirming.Update(clickTokenMsg(t, confirming, "[Confirm]"))
	got := next.(Model)
	if got.keyForm == nil || got.keyForm.mode != keyFormEditValue {
		t.Fatalf("key form = %#v, want edit form", got.keyForm)
	}
}

func TestCreateKeySelectsNewKeyAfterRefresh(t *testing.T) {
	m, executor := keyEditorTestModel(t, nil)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = next.(Model)
	m.keyForm.inputs[keyFieldName].SetValue("new:key")

	nextModel, writeCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nextModel.(Model)
	if writeCmd == nil {
		t.Fatal("create key did not return a command")
	}
	written := writeCmd().(keyWrittenMsg)
	nextModel, _ = m.Update(written)
	m = nextModel.(Model)
	if written.err != nil {
		t.Fatalf("write command failed: %v", written.err)
	}
	if m.pendingKeyFocus != "new:key" {
		t.Fatalf("pending focus = %q, want new:key", m.pendingKeyFocus)
	}

	refreshCmd := m.refreshKeysCmd()
	loaded := refreshCmd().(keysLoadedMsg)
	nextModel, valueCmd := m.Update(loaded)
	m = nextModel.(Model)
	if valueCmd == nil {
		t.Fatal("new key selection did not load value")
	}
	if m.selectedKey < 0 || m.selectedKey >= len(m.keys) || m.keys[m.selectedKey].Name != "new:key" {
		t.Fatalf("selected key = %#v, want new:key", m.keys)
	}
	if m.pendingKeyFocus != "" {
		t.Fatal("pending focus was not cleared")
	}
	_ = executor
}

func TestEditKeyValueOpensFormWithCurrentValue(t *testing.T) {
	m, _ := keyEditorTestModel(t, map[string]string{"greeting": "hello"})
	m.keys = []adapters.KeyInfo{{Name: "greeting", Type: "string"}}
	m.selectedValue = result.Normalize("hello")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	confirming := next.(Model)
	next, _ = confirming.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	got := next.(Model)
	if got.keyForm == nil || got.keyForm.mode != keyFormEditValue {
		t.Fatalf("key form = %#v, want edit value form", got.keyForm)
	}
	if got.keyForm.inputs[keyFieldValue].Value() != "hello" {
		t.Fatalf("value input = %q, want hello", got.keyForm.inputs[keyFieldValue].Value())
	}
}

func TestEditKeyValueWritesRedis(t *testing.T) {
	m, executor := keyEditorTestModel(t, map[string]string{"greeting": "hello"})
	m.keys = []adapters.KeyInfo{{Name: "greeting", Type: "string"}}
	m.selectedValue = result.Normalize("hello")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	confirming := next.(Model)
	next, _ = confirming.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = next.(Model)
	m.keyForm.inputs[keyFieldValue].SetValue("world")

	nextModel, writeCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nextModel.(Model)
	if writeCmd == nil {
		t.Fatal("edit value did not return a command")
	}
	written := writeCmd().(keyWrittenMsg)
	nextModel, _ = m.Update(written)
	m = nextModel.(Model)
	if written.err != nil {
		t.Fatalf("write command failed: %v", written.err)
	}
	value, err := executor.Client().Get(t.Context(), "greeting").Result()
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if value != "world" {
		t.Fatalf("value = %q, want world", value)
	}
	if !strings.Contains(m.Status(), "Updated") {
		t.Fatalf("status = %q, want updated", m.Status())
	}
}
