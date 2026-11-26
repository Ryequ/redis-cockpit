package ui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func featureTestModel() Model {
	cfg := config.Default()
	cfg.Profiles = append(cfg.Profiles,
		redis.ConnectionProfile{ID: "prod", Name: "prod", Addr: "redis.example.com:6379", TLS: true},
		redis.ConnectionProfile{ID: "dev", Name: "dev", Addr: "127.0.0.1:6380"},
	)
	return NewModelWithConfig(cfg, false)
}

func TestProfileSearchFiltersVisibleProfiles(t *testing.T) {
	m := featureTestModel()
	m.profileSearch.SetValue("prod")

	indices := m.visibleProfileIndices()
	if len(indices) != 1 || m.cfg.Profiles[indices[0]].Name != "prod" {
		t.Fatalf("visible profiles = %#v, want prod only", indices)
	}
}

func TestProfileSortByNameChangesVisibleOrder(t *testing.T) {
	m := featureTestModel()
	m.profileSort = profileSortName

	indices := m.visibleProfileIndices()
	names := make([]string, 0, len(indices))
	for _, idx := range indices {
		names = append(names, m.cfg.Profiles[idx].Name)
	}
	if strings.Join(names, ",") != "dev,local,prod" {
		t.Fatalf("order = %v, want dev,local,prod", names)
	}
}

func TestProfileGroupingAddsGroupHeaders(t *testing.T) {
	m := featureTestModel()
	m.profileGroupByMode = true

	view := m.renderConnectionsBody()
	if !strings.Contains(view, "Group: single+TLS") || !strings.Contains(view, "Group: single") {
		t.Fatalf("group headers missing:\n%s", view)
	}
}

func TestSelectedAndActiveProfilesAreMarked(t *testing.T) {
	m := featureTestModel()
	m.selectedProfile = 0
	view := m.renderConnectionsBody()

	if !strings.Contains(view, "▸") {
		t.Fatal("selected marker missing")
	}
	if !strings.Contains(view, "active") {
		t.Fatal("active marker missing")
	}
}

func TestConnectionsSummaryPanelsShowServerMemoryStats(t *testing.T) {
	m := featureTestModel()
	m.width = 140
	m.opsSnapshot.Info = map[string]map[string]string{
		"Server": {
			"redis_version": "7.2.4",
			"os":            "Linux 6.1",
			"process_id":    "1234",
		},
		"Memory": {
			"used_memory_human":      "1.20M",
			"used_memory_peak_human": "2.40M",
			"used_memory_lua_human":  "37.0K",
		},
		"Stats": {
			"connected_clients":          "5",
			"total_connections_received": "99",
			"total_commands_processed":   "1024",
		},
	}

	view := m.renderConnectionsBody()
	for _, want := range []string{
		"Server", "Memory", "Stats",
		"Version", "7.2.4", "OS", "Linux 6.1", "PID", "1234",
		"Used", "1.20M", "Peak", "2.40M", "Lua", "37.0K",
		"Clients", "5", "Connections", "99", "Commands", "1024",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("summary missing %q:\n%s", want, view)
		}
	}

	lines := strings.Split(view, "\n")
	found := false
	for _, line := range lines {
		plain := stripANSI(line)
		if strings.Contains(plain, "Server") && strings.Contains(plain, "Memory") && strings.Contains(plain, "Stats") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("summary panels are not rendered side by side")
	}
}

func TestProfileHealthShownInList(t *testing.T) {
	m := featureTestModel()
	m.profileHealth = map[string]ProfileHealth{
		"local": {Status: "OK", Latency: 12 * time.Millisecond, Version: "7.2.4"},
	}

	view := m.renderConnectionsBody()
	for _, want := range []string{"OK", "12ms", "v7.2.4"} {
		if !strings.Contains(view, want) {
			t.Fatalf("health missing %q:\n%s", want, view)
		}
	}
}

func TestProfileTestedMessageUpdatesHealth(t *testing.T) {
	m := featureTestModel()

	next, _ := m.Update(profileTestedMsg{
		gen:     m.generation,
		name:    "local",
		latency: 9 * time.Millisecond,
		version: "7.2.4",
	})
	got := next.(Model)
	health := got.profileHealth["local"]
	if health.Status != "OK" || health.Latency != 9*time.Millisecond || health.Version != "7.2.4" {
		t.Fatalf("health = %#v, want OK/9ms/7.2.4", health)
	}
}

func TestBatchProfileTestReturnsCommands(t *testing.T) {
	m := featureTestModel()
	m.setView(ViewConnections)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	got := next.(Model)
	if cmd == nil {
		t.Fatal("batch test did not return a command")
	}
	if !strings.Contains(got.Status(), "Testing all") {
		t.Fatalf("status = %q, want testing all", got.Status())
	}
}

func TestRedisProbeReturnsLatencyAndVersion(t *testing.T) {
	server := miniredis.RunT(t)
	profile := redis.ConnectionProfile{Name: "local", Addr: server.Addr()}

	version, latency, err := redis.Probe(context.Background(), profile)
	if err != nil {
		t.Fatalf("Probe returned error: %v", err)
	}
	if latency <= 0 {
		t.Fatalf("latency = %v, want > 0", latency)
	}
	if version == "" {
		t.Fatal("version is empty")
	}
}

func TestCloneProfileShortcutSavesCopy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	m := featureTestModel()
	m.SetConfigPath(path)
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	_ = next.(Model)

	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if _, ok := saved.FindProfile("local-copy"); !ok {
		t.Fatalf("cloned profile not saved: %#v", saved.Profiles)
	}
}

func TestImportProfilesAddsProfiles(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	importPath := filepath.Join(dir, "import.yaml")
	source := config.Default()
	source.Profiles = []redis.ConnectionProfile{{ID: "imported", Name: "imported", Addr: "10.0.0.1:6379"}}
	if err := config.ExportProfiles(importPath, source); err != nil {
		t.Fatalf("ExportProfiles returned error: %v", err)
	}

	m := featureTestModel()
	m.SetConfigPath(configPath)
	m.setView(ViewConnections)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	got := next.(Model)
	if got.profileIOForm == nil || got.profileIOForm.mode != profileIOImport {
		t.Fatalf("import form = %#v, want import form", got.profileIOForm)
	}
	got.profileIOForm.pathInput.SetValue(importPath)
	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = next.(Model)

	saved, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if _, ok := saved.FindProfile("imported"); !ok {
		t.Fatalf("imported profile not saved: %#v", saved.Profiles)
	}
}

func TestExportProfilesWritesFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	exportPath := filepath.Join(dir, "export.yaml")

	m := featureTestModel()
	m.SetConfigPath(configPath)
	m.setView(ViewConnections)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	got := next.(Model)
	if got.profileIOForm == nil || got.profileIOForm.mode != profileIOExport {
		t.Fatalf("export form = %#v, want export form", got.profileIOForm)
	}
	got.profileIOForm.pathInput.SetValue(exportPath)
	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = next.(Model)

	profiles, err := config.ImportProfiles(exportPath)
	if err != nil {
		t.Fatalf("ImportProfiles returned error: %v", err)
	}
	if len(profiles) < 3 {
		t.Fatalf("exported profiles = %#v, want at least 3", profiles)
	}
}

func TestProfileFormExposesSSHAndCloudFields(t *testing.T) {
	form := newProfileForm(redis.ConnectionProfile{}, profileCreate)
	form.inputs[profileFieldName].SetValue("ssh-prod")
	form.inputs[profileFieldAddr].SetValue("10.0.0.5:6379")
	form.inputs[profileFieldSSHHost].SetValue("bastion.example.com")
	form.inputs[profileFieldSSHUser].SetValue("ubuntu")
	form.inputs[profileFieldSSHKeyPath].SetValue("keys/id_ed25519")
	form.inputs[profileFieldCloudProvider].SetValue("aws")
	form.inputs[profileFieldAuthMechanism].SetValue("iam")

	profile, err := form.profile()
	if err != nil {
		t.Fatalf("profile returned error: %v", err)
	}
	if profile.SSHHost != "bastion.example.com" || profile.SSHUser != "ubuntu" || profile.SSHKeyPath != "keys/id_ed25519" {
		t.Fatalf("ssh fields not copied: %#v", profile)
	}
	if profile.CloudProvider != "aws" || profile.AuthMechanism != "iam" {
		t.Fatalf("cloud/auth fields not copied: %#v", profile)
	}
}

func TestProfileFormCloudPresetButtonAppliesAWS(t *testing.T) {
	m := featureTestModel()
	m.setView(ViewConnections)
	next, _ := m.Update(clickTokenMsg(t, m, "[N new]"))
	got := next.(Model)

	next, _ = got.Update(clickTokenMsg(t, got, "[AWS]"))
	got = next.(Model)
	got.profileForm.inputs[profileFieldName].SetValue("aws-prod")
	got.profileForm.inputs[profileFieldAddr].SetValue("cache.example.com:6379")
	profile, err := got.profileForm.profile()
	if err != nil {
		t.Fatalf("profile returned error: %v", err)
	}
	if !profile.TLS || profile.CloudProvider != "aws" || profile.AuthMechanism != "iam" {
		t.Fatalf("AWS preset not applied: %#v", profile)
	}
}

func TestSearchHidingSelectedProfileMovesSelectionToVisible(t *testing.T) {
	m := featureTestModel()
	m.setView(ViewConnections)
	m.selectedProfile = 0
	m.profileSearch.SetValue("prod")
	m.profileSearchMode = true

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(Model)
	if got.cfg.Profiles[got.selectedProfile].Name != "prod" {
		t.Fatalf("selected profile = %q, want prod", got.cfg.Profiles[got.selectedProfile].Name)
	}
}

func TestBatchTestStatusTracksProgress(t *testing.T) {
	m := featureTestModel()
	m.setView(ViewConnections)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = next.(Model)

	next, _ = m.Update(profileTestedMsg{gen: m.generation, profileGen: m.profileGeneration, name: "local"})
	m = next.(Model)
	next, _ = m.Update(profileTestedMsg{gen: m.generation, profileGen: m.profileGeneration, name: "prod"})
	m = next.(Model)

	if !strings.Contains(m.Status(), "2/3") {
		t.Fatalf("status = %q, want batch progress 2/3", m.Status())
	}
}

func TestRenameProfileMovesHealthEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	m := featureTestModel()
	m.SetConfigPath(path)
	m.setView(ViewConnections)
	m.profileHealth = map[string]ProfileHealth{"local": {Status: "OK", Version: "7.2.4"}}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	got := next.(Model)
	got.profileForm.inputs[profileFieldName].SetValue("renamed")
	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = next.(Model)

	if _, ok := got.profileHealth["local"]; ok {
		t.Fatal("old health entry was not removed")
	}
	if _, ok := got.profileHealth["renamed"]; !ok {
		t.Fatal("health entry was not moved to renamed profile")
	}
}

func TestFormButtonsVisibleInShortTerminal(t *testing.T) {
	m := featureTestModel()
	m.height = 16
	m.setView(ViewConnections)
	next, _ := m.Update(clickTokenMsg(t, m, "[N new]"))
	got := next.(Model)

	view := got.View()
	if !strings.Contains(view, "[Save]") || !strings.Contains(view, "[Cancel]") {
		t.Fatalf("form buttons missing in short terminal:\n%s", view)
	}
}

func TestSummaryPanelsFitNarrowTerminal(t *testing.T) {
	m := featureTestModel()
	m.width = 80
	summary := stripANSI(m.renderConnectionSummary())
	for _, line := range strings.Split(summary, "\n") {
		if width := runewidth.StringWidth(line); width > m.width-navWidth {
			t.Fatalf("summary line width = %d, want <= %d: %q", width, m.width-navWidth, line)
		}
	}
}

func TestProfileSearchKeyEntersSearchMode(t *testing.T) {
	m := featureTestModel()
	m.setView(ViewConnections)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	got := next.(Model)
	if !got.profileSearchMode {
		t.Fatal("search mode was not enabled")
	}
}
