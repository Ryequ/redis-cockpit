package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codebuddy/redis-tui/internal/config"
	"github.com/codebuddy/redis-tui/internal/security"
	"github.com/codebuddy/redis-tui/internal/ui"
)

// Options is the application-level configuration supplied by the CLI layer.
type Options struct {
	ConfigPath string
	Profile    string
	ReadOnly   bool
	Debug      bool
}

// Run starts the Bubble Tea program with alternative screen and mouse support.
func Run(opts Options) error {
	if opts.Debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
	slog.Debug("starting redis-tui", "config", opts.ConfigPath, "profile", opts.Profile, "readonly", opts.ReadOnly)

	path := opts.ConfigPath
	if path == "" {
		if defaultPath, err := config.DefaultPath(); err == nil {
			path = defaultPath
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		if opts.ConfigPath != "" {
			return err
		}
		slog.Warn("failed to load config, using defaults", "path", path, "error", err)
		cfg = config.Default()
	}

	if opts.Profile != "" {
		profile, ok := cfg.FindProfile(opts.Profile)
		if !ok {
			return fmt.Errorf("profile %q not found", opts.Profile)
		}
		cfg.ActiveProfile = profile.Name
	}

	readOnly := opts.ReadOnly
	if profile, ok := cfg.FindProfile(cfg.ActiveProfile); ok {
		readOnly = readOnly || profile.ReadOnly
	}

	model := ui.NewModelWithConfig(cfg, readOnly)

	auditDir := filepath.Dir(path)
	if path == "" {
		if defaultPath, err := config.DefaultPath(); err == nil {
			auditDir = filepath.Dir(defaultPath)
		}
	}
	if auditDir != "" {
		model.SetAuditLog(security.NewAuditLog(filepath.Join(auditDir, "audit.jsonl")))
	}

	mouseSupported := ui.DetectMouseSupport()
	model.SetMouseEnabled(mouseSupported)
	programOptions := []tea.ProgramOption{tea.WithAltScreen()}
	if mouseSupported {
		programOptions = append(programOptions, tea.WithMouseCellMotion())
	}
	program := tea.NewProgram(model, programOptions...)

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run terminal UI: %w", err)
	}
	return nil
}
