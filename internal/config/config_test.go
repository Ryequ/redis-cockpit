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

func TestLoadExpandsPasswordEnvironmentVariable(t *testing.T) {
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
	if got.Profiles[0].Password != "s3cr3t" {
		t.Fatalf("password = %q, want %q", got.Profiles[0].Password, "s3cr3t")
	}
}
