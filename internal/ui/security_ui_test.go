package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/ops"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/result"
	"github.com/Ryequ/redis-cockpit/internal/security"
	"github.com/Ryequ/redis-cockpit/internal/stream"
)

func TestWorkbenchCriticalCommandRequiresConfirmation(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewWorkbench)
	m.commandInput.SetValue("FLUSHALL")

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("critical command executed before confirmation")
	}
	confirming := next.(Model)
	if confirming.pendingCommand == nil {
		t.Fatal("pendingCommand was not set")
	}
	if !strings.Contains(confirming.View(), "Confirm critical command") {
		t.Fatalf("view does not show confirmation:\n%s", confirming.View())
	}

	next, cmd = confirming.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd != nil {
		t.Fatal("cancelling critical command should not execute")
	}
	if next.(Model).pendingCommand != nil {
		t.Fatal("pendingCommand was not cleared after cancel")
	}
}

func TestStreamsPublishBlockedInReadOnly(t *testing.T) {
	m := NewModelWithConfig(config.Default(), true)
	m.setView(ViewStreams)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cmd != nil {
		t.Fatal("publish should not run in read-only mode")
	}
	got := next.(Model)
	if !strings.Contains(got.Status(), "read-only") {
		t.Fatalf("status = %q, want read-only message", got.Status())
	}
}

func TestStoppingStreamsInvalidatesPendingHandshake(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewStreams)
	oldGen := m.generation

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	next, cmd := next.Update(subscribedMsg{gen: oldGen})
	if cmd != nil {
		t.Fatal("stale subscribe message restarted message loop")
	}
	got := next.(Model)
	if got.subscription != nil {
		t.Fatal("stale subscribe message reattached subscription")
	}
}

func TestStaleAsyncMessagesAreIgnored(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.generation = 2

	next, _ := m.Update(connectedMsg{gen: 1, executor: redis.NewExecutor(nil)})
	got := next.(Model)
	if got.executor != nil {
		t.Fatal("stale connectedMsg replaced current connection")
	}

	next, _ = got.Update(commandExecutedMsg{gen: 1, value: result.Normalize("stale")})
	got = next.(Model)
	if got.lastResult.Display() != "(nil)" {
		t.Fatalf("stale command result was applied: %q", got.lastResult.Display())
	}
}

func TestCloseClientClearsTransientState(t *testing.T) {
	m, ctx := newBrowserModel(t)
	if err := m.adapter.SetString(ctx, "temp:key", "value", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	m = refreshKeys(t, m)
	m.pendingCommand = []string{"FLUSHALL"}
	m.pendingDelete = []string{"temp:key"}
	m.opsSnapshot.Config = map[string]string{"maxmemory": "0"}
	m.streamMessages = append(m.streamMessages, stream.Message{Channel: "test", Payload: "value"})

	m.closeClient()
	got := m
	if got.executor != nil || got.adapter != nil || got.opsManager != nil || got.streamHub != nil {
		t.Fatal("connection managers were not cleared")
	}
	if len(got.keys) != 0 || got.pendingCommand != nil || got.pendingDelete != nil || len(got.streamMessages) != 0 {
		t.Fatal("transient state was not cleared")
	}
}

func TestCancelledCriticalCommandIsNotAudited(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	m := NewModelWithConfig(config.Default(), false)
	m.SetAuditLog(security.NewAuditLog(path))
	m.setView(ViewWorkbench)
	m.commandInput.SetValue("FLUSHALL")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	_ = next.(Model)

	data, err := os.ReadFile(path)
	if err == nil && strings.Contains(string(data), "FLUSHALL") {
		t.Fatalf("cancelled command was audited: %s", data)
	}
}

func TestCommandExecutedMessageRedactsSensitiveArguments(t *testing.T) {
	msg := executeCmd(nil, []string{"AUTH", "supersecret"}, 0)().(commandExecutedMsg)
	if !strings.Contains(msg.line, "AUTH") {
		t.Fatalf("message line was not populated: %q", msg.line)
	}
	if strings.Contains(msg.line, "supersecret") {
		t.Fatalf("message line leaked secret: %q", msg.line)
	}
}

func TestWorkbenchHistoryRedactsSensitiveArguments(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewWorkbench)
	m.commandInput.SetValue("CONFIG SET requirepass secret")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(Model)
	for _, entry := range got.history.All() {
		if strings.Contains(entry, "secret") {
			t.Fatalf("history leaked secret: %#v", got.history.All())
		}
	}
}

func TestOpsViewRedactsSensitiveValues(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewOps)
	m.opsSnapshot.Config = map[string]string{"requirepass": "secret", "masterauth": "replsecret"}
	m.opsSnapshot.Slowlog = []ops.SlowlogEntry{{ID: 1, Command: []string{"AUTH", "secret"}}}

	view := m.View()
	if strings.Contains(view, "secret") || strings.Contains(view, "replsecret") {
		t.Fatalf("ops view leaked secret:\n%s", view)
	}
}

func TestKeysViewDeleteBlockedInReadOnly(t *testing.T) {
	m, ctx := newBrowserModel(t)
	m.readOnly = true
	if err := m.adapter.SetString(ctx, "locked:key", "value", 0); err != nil {
		t.Fatalf("SetString returned error: %v", err)
	}
	loaded := refreshKeys(t, m)

	next, cmd := loaded.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd != nil {
		t.Fatal("delete should not run in read-only mode")
	}
	got := next.(Model)
	if got.pendingDelete != nil {
		t.Fatal("delete confirmation should not appear in read-only mode")
	}
	if !strings.Contains(got.Status(), "read-only") {
		t.Fatalf("status = %q, want read-only message", got.Status())
	}
}
