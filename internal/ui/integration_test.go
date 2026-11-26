package ui

import (
	"strings"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	miniredis "github.com/alicebob/miniredis/v2"
	tea "github.com/charmbracelet/bubbletea"
)

func TestConnectionViewListsConfiguredProfiles(t *testing.T) {
	cfg := config.Default()
	cfg.Profiles = append(cfg.Profiles, redis.ConnectionProfile{Name: "prod", Addr: "redis.example.com:6379"})

	m := NewModelWithConfig(cfg, false)
	if !strings.Contains(m.View(), "prod") {
		t.Fatalf("view does not list prod profile:\n%s", m.View())
	}
}

func TestConnectionsViewConnectMessageSetsExecutor(t *testing.T) {
	server := miniredis.RunT(t)
	cfg := config.Default()
	cfg.Profiles = []redis.ConnectionProfile{{Name: "local", Addr: server.Addr()}}
	cfg.ActiveProfile = "local"

	m := NewModelWithConfig(cfg, false)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter returned nil connect command")
	}

	msg := cmd()
	next, _ = next.Update(msg)
	got := next.(Model)
	if got.executor == nil {
		t.Fatal("executor was not set after connect")
	}
	if got.ActiveView() != ViewKeys {
		t.Fatalf("active view = %v, want ViewKeys", got.ActiveView())
	}
}

func TestWorkbenchExecutesCommandAndStoresResult(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := redis.NewClient(t.Context(), redis.ConnectionProfile{Name: "local", Addr: server.Addr()})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	m := NewModelWithConfig(config.Default(), false)
	m.executor = redis.NewExecutor(client)
	m.setView(ViewWorkbench)
	m.commandInput.SetValue("SET greeting hello")

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter returned nil execute command")
	}
	msg := cmd()
	next, _ = next.Update(msg)
	got := next.(Model)

	if got.lastResult.Display() != "OK" {
		t.Fatalf("result = %q, want OK", got.lastResult.Display())
	}
	if len(got.history.All()) != 1 || got.history.All()[0] != "SET greeting hello" {
		t.Fatalf("history = %#v", got.history.All())
	}
}

func TestWorkbenchBlocksWriteInReadOnlyMode(t *testing.T) {
	m := NewModelWithConfig(config.Default(), true)
	m.setView(ViewWorkbench)
	m.commandInput.SetValue("SET key value")

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("read-only command should not start execution")
	}
	got := next.(Model)
	if !strings.Contains(got.Status(), "read-only") {
		t.Fatalf("status = %q, want read-only message", got.Status())
	}
}
