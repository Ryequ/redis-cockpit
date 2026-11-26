package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func TestLoadReturnsDefaultConfigWhenFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(cfg.Profiles) != 1 || cfg.Profiles[0].Name != "local" {
		t.Fatalf("profiles = %#v, want one local profile", cfg.Profiles)
	}
	if cfg.PageSize != 100 {
		t.Fatalf("page size = %d, want 100", cfg.PageSize)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	want := Config{
		ActiveProfile: "prod",
		PageSize:      50,
		MaxPreview:    2048,
		Profiles: []redis.ConnectionProfile{
			{Name: "prod", Addr: "redis.example.com:6379", DB: 2, TLS: true},
		},
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got.ActiveProfile != want.ActiveProfile || got.PageSize != want.PageSize || got.MaxPreview != want.MaxPreview {
		t.Fatalf("loaded config = %#v, want %#v", got, want)
	}
	if len(got.Profiles) != 1 || got.Profiles[0].Addr != want.Profiles[0].Addr {
		t.Fatalf("loaded profiles = %#v, want %#v", got.Profiles, want.Profiles)
	}
}

func TestLoadPreservesEmptyProfilesFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := Save(path, Config{PageSize: 100, MaxPreview: 4096}); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(got.Profiles) != 0 {
		t.Fatalf("profiles = %#v, want empty", got.Profiles)
	}
}

func TestLoadPreservesPasswordEnvironmentVariable(t *testing.T) {
	t.Setenv("REDIS_TUI_TEST_PASSWORD", "s3cr3t")
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := Default()
	cfg.Profiles[0].Password = "${REDIS_TUI_TEST_PASSWORD}"
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got.Profiles[0].Password != "${REDIS_TUI_TEST_PASSWORD}" {
		t.Fatalf("password = %q, want %q", got.Profiles[0].Password, "${REDIS_TUI_TEST_PASSWORD}")
	}
}

func TestCloneProfileCopiesFieldsWithNewIdentity(t *testing.T) {
	original := redis.ConnectionProfile{
		ID: "local", Name: "local", Addr: "127.0.0.1:6379", Password: "${PASSWORD}", TLS: true,
	}
	clone, err := CloneProfile(original, "local-copy")
	if err != nil {
		t.Fatalf("CloneProfile returned error: %v", err)
	}
	if clone.ID != "local-copy" || clone.Name != "local-copy" {
		t.Fatalf("clone identity = %q/%q", clone.ID, clone.Name)
	}
	if clone.Addr != original.Addr || clone.Password != original.Password || !clone.TLS {
		t.Fatalf("clone fields were not copied: %#v", clone)
	}
}

func TestExportProfilesRefusesToOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profiles.yaml")
	if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := ExportProfiles(path, Default()); err == nil {
		t.Fatal("ExportProfiles overwrote existing file")
	}
}

func TestImportProfilesRejectsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")
	if _, err := ImportProfiles(path); err == nil {
		t.Fatal("ImportProfiles returned nil error for missing file")
	}
}

func TestImportProfilesRejectsInvalidProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.yaml")
	data := "profiles:\n  - id: bad\n    name: bad\n    addr: ''\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if _, err := ImportProfiles(path); err == nil {
		t.Fatal("ImportProfiles accepted invalid profile")
	}
}

func TestExportImportProfilesRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profiles.yaml")
	cfg := Default()
	cfg.Profiles[0].Password = "${PASSWORD}"

	if err := ExportProfiles(path, cfg); err != nil {
		t.Fatalf("ExportProfiles returned error: %v", err)
	}
	profiles, err := ImportProfiles(path)
	if err != nil {
		t.Fatalf("ImportProfiles returned error: %v", err)
	}
	if len(profiles) != 1 || profiles[0].Name != "local" || profiles[0].Password != "${PASSWORD}" {
		t.Fatalf("imported profiles = %#v", profiles)
	}
}
