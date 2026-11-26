package ui

import (
	"context"
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codebuddy/redis-tui/internal/config"
	"github.com/codebuddy/redis-tui/internal/ops"
	"github.com/codebuddy/redis-tui/internal/redis"
)

func newAdvancedModel(t *testing.T) Model {
	t.Helper()
	server := miniredis.RunT(t)
	client, err := redis.NewClient(context.Background(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	m := NewModelWithConfig(config.Default(), false)
	m.executor = redis.NewExecutor(client)
	m.opsManager = ops.New(m.executor)
	m.setView(ViewOps)
	return m
}

func TestOpsViewLoadsSnapshot(t *testing.T) {
	m := newAdvancedModel(t)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("refresh returned nil command")
	}
	next, _ = next.Update(cmd())
	got := next.(Model)

	if !strings.Contains(got.Status(), "snapshot") {
		t.Fatalf("status = %q, want snapshot loaded", got.Status())
	}
	if !strings.Contains(got.View(), "INFO") {
		t.Fatalf("view does not show INFO section:\n%s", got.View())
	}
}

func TestStreamsViewRendersStreamControls(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewStreams)
	view := m.View()

	for _, want := range []string{"Pub/Sub", "MONITOR", "Stream"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
}
