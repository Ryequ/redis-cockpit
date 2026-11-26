package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codebuddy/redis-tui/internal/redis"
	"gopkg.in/yaml.v3"
)

// Config is the persisted application configuration.
type Config struct {
	ActiveProfile string                    `yaml:"active_profile" json:"active_profile"`
	PageSize      int                       `yaml:"page_size" json:"page_size"`
	MaxPreview    int                       `yaml:"max_preview" json:"max_preview"`
	Profiles      []redis.ConnectionProfile `yaml:"profiles" json:"profiles"`
}

func Default() Config {
	return Config{
		ActiveProfile: "local",
		PageSize:      100,
		MaxPreview:    4096,
		Profiles: []redis.ConnectionProfile{
			{
				ID:   "local",
				Name: "local",
				Addr: "127.0.0.1:6379",
				DB:   0,
			},
		},
	}
}

// DefaultPath resolves the platform-specific configuration file path.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(dir, "redis-tui", "config.yaml"), nil
}

// Load reads a config file; a missing file yields defaults without writing.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.PageSize <= 0 {
		cfg.PageSize = 100
	}
	if cfg.MaxPreview <= 0 {
		cfg.MaxPreview = 4096
	}
	if len(cfg.Profiles) == 0 {
		cfg = Default()
	}
	for i := range cfg.Profiles {
		cfg.Profiles[i].Password = os.ExpandEnv(cfg.Profiles[i].Password)
		cfg.Profiles[i].Username = os.ExpandEnv(cfg.Profiles[i].Username)
	}
	return cfg, nil
}

// Save writes the config file, creating parent directories as needed.
func Save(path string, cfg Config) error {
	if cfg.PageSize <= 0 {
		cfg.PageSize = 100
	}
	if cfg.MaxPreview <= 0 {
		cfg.MaxPreview = 4096
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

func (c Config) FindProfile(name string) (redis.ConnectionProfile, bool) {
	for _, profile := range c.Profiles {
		if profile.Name == name || profile.ID == name {
			return profile, true
		}
	}
	return redis.ConnectionProfile{}, false
}
