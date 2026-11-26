package ui

import (
	"context"
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
)

func newBrowserModel(t *testing.T) (Model, context.Context) {
	t.Helper()
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	m := NewModelWithConfig(config.Default(), false)
	m.executor = redis.NewExecutor(client)
	m.adapter = adapters.New(m.executor)
	m.setView(ViewKeys)
	return m, context.Background()
}

func refreshKeys(t *testing.T, m Model) Model {
	t.Helper()
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("refresh key returned nil command")
	}
	next, _ = next.Update(cmd())
	return next.(Model)
}

func TestKeysViewLoadsKeys(t *testing.T) {
	m, ctx := newBrowserModel(t)
	if err := m.adapter.SetString(ctx, "user:1", "alice", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}

	got := refreshKeys(t, m)
	if !strings.Contains(got.View(), "user:1") {
		t.Fatalf("view does not contain loaded key:\n%s", got.View())
	}
}

func TestKeysViewLoadsValueForSelectedKey(t *testing.T) {
	m, ctx := newBrowserModel(t)
	if err := m.adapter.SetString(ctx, "user:1", "alice", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	loaded := refreshKeys(t, m)

	next, cmd := loaded.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter returned nil value-load command")
	}
	next, _ = next.Update(cmd())
	got := next.(Model)

	if !strings.Contains(got.View(), "alice") {
		t.Fatalf("view does not contain loaded value:\n%s", got.View())
	}
}

func TestKeysViewDeleteRequiresConfirmation(t *testing.T) {
	m, ctx := newBrowserModel(t)
	if err := m.adapter.SetString(ctx, "user:1", "alice", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	loaded := refreshKeys(t, m)

	next, cmd := loaded.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd != nil {
		t.Fatal("delete key should not execute before confirmation")
	}
	confirming := next.(Model)
	if !strings.Contains(confirming.View(), "Delete key?") {
		t.Fatalf("view does not show delete confirmation:\n%s", confirming.View())
	}

	next, cmd = confirming.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd != nil {
		t.Fatal("first confirmation must not delete the key")
	}
	second := next.(Model)
	next, cmd = second.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("second confirmation did not delete the key")
	}
	next, refreshCmd := next.Update(cmd())
	if refreshCmd != nil {
		next, _ = next.Update(refreshCmd())
	}
	got := next.(Model)
	if strings.Contains(got.View(), "user:1") {
		t.Fatalf("deleted key still visible:\n%s", got.View())
	}
}
