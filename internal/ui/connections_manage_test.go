package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func TestConnectCmdExpandsPasswordEnvironmentVariable(t *testing.T) {
	server := miniredis.RunT(t)
	server.RequireAuth("s3cr3t")
	t.Setenv("REDIS_TUI_TEST_PASSWORD", "s3cr3t")

	profile := redis.ConnectionProfile{
		Name:     "auth",
		Addr:     server.Addr(),
		Password: "${REDIS_TUI_TEST_PASSWORD}",
	}
	msg := connectCmd(profile, 0)().(connectedMsg)
	if msg.err != nil {
		t.Fatalf("connectCmd returned error: %v", msg.err)
	}
	if msg.executor == nil {
		t.Fatal("connectCmd did not return executor")
	}
	msg.executor.Close()
}

func TestNewProfileFormOpensAndSaves(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	m := NewModelWithConfig(config.Default(), false)
	m.SetConfigPath(path)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := next.(Model)
	if got.profileForm == nil || got.profileForm.mode != profileCreate {
		t.Fatalf("profile form = %#v, want create form", got.profileForm)
	}

	got.profileForm.inputs[profileFieldName].SetValue("prod")
	got.profileForm.inputs[profileFieldAddr].SetValue("redis.example.com:6379")
	got.profileForm.inputs[profileFieldPassword].SetValue("${REDIS_TUI_PASSWORD}")

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = next.(Model)

	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	profile, ok := saved.FindProfile("prod")
	if !ok {
		t.Fatalf("profile prod was not saved: %#v", saved.Profiles)
	}
	if profile.Addr != "redis.example.com:6379" {
		t.Fatalf("addr = %q, want redis.example.com:6379", profile.Addr)
	}
	if profile.Password != "${REDIS_TUI_PASSWORD}" {
		t.Fatalf("password = %q, want env placeholder", profile.Password)
	}
}

func TestEditProfileFormUpdatesSelectedProfileAndSaves(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	m := NewModelWithConfig(config.Default(), false)
	m.SetConfigPath(path)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	got := next.(Model)
	if got.profileForm == nil || got.profileForm.mode != profileEdit {
		t.Fatalf("profile form = %#v, want edit form", got.profileForm)
	}
	got.profileForm.inputs[profileFieldAddr].SetValue("127.0.0.1:6380")

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = next.(Model)

	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if saved.Profiles[0].Addr != "127.0.0.1:6380" {
		t.Fatalf("addr = %q, want 127.0.0.1:6380", saved.Profiles[0].Addr)
	}
}

func TestDeleteProfileRequiresConfirmationAndSaves(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := config.Default()
	cfg.Profiles = append(cfg.Profiles, redis.ConnectionProfile{ID: "prod", Name: "prod", Addr: "redis.example.com:6379"})
	m := NewModelWithConfig(cfg, false)
	m.SetConfigPath(path)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := next.(Model)
	if got.pendingProfileDelete != "local" {
		t.Fatalf("pending delete = %q, want local", got.pendingProfileDelete)
	}

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	_ = next.(Model)

	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(saved.Profiles) != 1 || saved.Profiles[0].Name != "prod" {
		t.Fatalf("profiles = %#v, want only prod", saved.Profiles)
	}
}

func TestTestConnectionMessageUpdatesStatus(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	if cmd == nil {
		t.Fatal("test connection did not return a command")
	}
	got := next.(Model)

	next, _ = got.Update(profileTestedMsg{gen: got.generation, name: "local", err: nil})
	got = next.(Model)
	if !strings.Contains(got.Status(), "OK") {
		t.Fatalf("status = %q, want OK", got.Status())
	}
}

func TestNewProfilePersistsActiveProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	m := NewModelWithConfig(config.Default(), false)
	m.SetConfigPath(path)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := next.(Model)
	got.profileForm.inputs[profileFieldName].SetValue("prod")
	got.profileForm.inputs[profileFieldAddr].SetValue("redis.example.com:6379")

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = next.(Model)

	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if saved.ActiveProfile != "prod" {
		t.Fatalf("active profile = %q, want prod", saved.ActiveProfile)
	}
}

func TestDeleteProfileSaveFailureRollsBack(t *testing.T) {
	cfg := config.Default()
	cfg.Profiles = append(cfg.Profiles, redis.ConnectionProfile{ID: "prod", Name: "prod", Addr: "redis.example.com:6379"})
	m := NewModelWithConfig(cfg, false)
	m.SetConfigPath(t.TempDir())
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := next.(Model)
	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	got = next.(Model)
	if !strings.Contains(got.Status(), "Save failed") {
		t.Fatalf("status = %q, want save failure", got.Status())
	}
	if len(got.cfg.Profiles) != 2 || got.cfg.Profiles[0].Name != "local" {
		t.Fatalf("rollback failed: %#v", got.cfg.Profiles)
	}
}

func TestProfileFormAllowsSpaceInTextFields(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := next.(Model)
	got.profileForm.focus = profileFieldName

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeySpace})
	got = next.(Model)
	if got.profileForm.inputs[profileFieldName].Value() != " " {
		t.Fatalf("name = %q, want a space", got.profileForm.inputs[profileFieldName].Value())
	}
}

func TestEditProfileRenamePreservesIDAndRemovesOldName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := config.Default()
	cfg.Profiles[0].ID = "local-id"
	m := NewModelWithConfig(cfg, false)
	m.SetConfigPath(path)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	got := next.(Model)
	got.profileForm.inputs[profileFieldName].SetValue("renamed")

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = next.(Model)

	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if _, ok := saved.FindProfile("local"); ok {
		t.Fatal("old profile name still exists after rename")
	}
	profile, ok := saved.FindProfile("renamed")
	if !ok {
		t.Fatalf("renamed profile not found: %#v", saved.Profiles)
	}
	if profile.ID != "local-id" {
		t.Fatalf("id = %q, want local-id", profile.ID)
	}
}

func TestNewProfileRejectsDuplicateName(t *testing.T) {
	cfg := config.Default()
	cfg.Profiles = append(cfg.Profiles, redis.ConnectionProfile{ID: "prod", Name: "prod", Addr: "redis.example.com:6379"})
	m := NewModelWithConfig(cfg, false)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := next.(Model)
	got.profileForm.inputs[profileFieldName].SetValue("prod")
	got.profileForm.inputs[profileFieldAddr].SetValue("127.0.0.1:6379")

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = next.(Model)
	if got.profileForm == nil {
		t.Fatal("duplicate form was closed")
	}
	if !strings.Contains(got.Status(), "already exists") {
		t.Fatalf("status = %q, want duplicate error", got.Status())
	}
	if len(got.cfg.Profiles) != 2 || got.cfg.Profiles[0].Addr != "127.0.0.1:6379" {
		t.Fatalf("profiles were overwritten: %#v", got.cfg.Profiles)
	}
}

func TestSaveFailureRollsBackProfileChanges(t *testing.T) {
	cfg := config.Default()
	cfg.Profiles[0].ID = "local-id"
	m := NewModelWithConfig(cfg, false)
	m.SetConfigPath(t.TempDir())
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	got := next.(Model)
	got.profileForm.inputs[profileFieldName].SetValue("renamed")

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = next.(Model)
	if !strings.Contains(got.Status(), "Save failed") {
		t.Fatalf("status = %q, want save failure", got.Status())
	}

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got = next.(Model)
	if len(got.cfg.Profiles) != 1 || got.cfg.Profiles[0].Name != "local" || got.cfg.Profiles[0].ID != "local-id" {
		t.Fatalf("rollback failed: %#v", got.cfg.Profiles)
	}
}

func TestStaleProfileTestResultIsIgnored(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	before := m.Status()

	next, _ := m.Update(profileTestedMsg{gen: m.generation + 1, name: "local", err: nil})
	got := next.(Model)
	if got.Status() != before {
		t.Fatalf("status = %q, want %q", got.Status(), before)
	}
}

func TestProfileFormTabAndSpaceToggleFields(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := next.(Model)
	got.profileForm.focus = profileFieldCount + profileBoolTLS

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeySpace})
	got = next.(Model)
	if !got.profileForm.bools[profileBoolTLS] {
		t.Fatal("space did not toggle TLS")
	}

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyTab})
	got = next.(Model)
	if got.profileForm.focus != profileFieldCount+profileBoolInsecure {
		t.Fatalf("focus = %d, want skip TLS verify field", got.profileForm.focus)
	}
}

func TestProfileFormValidationBlocksSave(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := next.(Model)
	got.profileForm.inputs[profileFieldName].SetValue("prod")

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = next.(Model)
	if got.profileForm == nil {
		t.Fatal("invalid form was closed")
	}
	if !strings.Contains(got.Status(), "Invalid profile") {
		t.Fatalf("status = %q, want validation error", got.Status())
	}
}

func TestProfileFormCancelDiscardsChanges(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := next.(Model)
	got.profileForm.inputs[profileFieldName].SetValue("discard")

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got = next.(Model)
	if got.profileForm != nil {
		t.Fatal("profile form was not closed")
	}
	if len(got.cfg.Profiles) != 1 || got.cfg.Profiles[0].Name != "local" {
		t.Fatalf("profiles were changed: %#v", got.cfg.Profiles)
	}
}
