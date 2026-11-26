package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
)

func redisProfileForHint(name string) redis.ConnectionProfile {
	return redis.ConnectionProfile{ID: name, Name: name, Addr: "127.0.0.1:6379"}
}

func TestWorkspaceHintsAreShort(t *testing.T) {
	for _, language := range []string{"zh", "en"} {
		cfg := config.Default()
		cfg.Language = language
		m := NewModelWithConfig(cfg, false)
		for _, view := range navItems {
			m.activeView = view
			hint := m.workspaceHintLine()
			if width := runewidth.StringWidth(hint); width > 56 {
				t.Fatalf("%s hint for %v is too long (%d): %q", language, view, width, hint)
			}
		}
	}
}

func TestHintStyleUsesAccentBackground(t *testing.T) {
	if hintStyle.GetBackground() == nil || hintStyle.GetForeground() == nil {
		t.Fatal("hint style must have foreground and background colors")
	}
}

func TestBottomHintStaysAtPanelBottom(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)

	indices := []int{}
	for _, profiles := range [][]string{{"one"}, {"one", "two", "three", "four"}} {
		cfg := config.Default()
		cfg.Profiles = nil
		for _, name := range profiles {
			cfg.Profiles = append(cfg.Profiles, redisProfileForHint(name))
		}
		testModel := NewModelWithConfig(cfg, false)
		testModel.setView(ViewConnections)
		lines := strings.Split(stripANSI(testModel.renderWorkspace(56, 18)), "\n")
		idx := -1
		for i, line := range lines {
			if strings.Contains(line, "↑↓") {
				idx = i
				break
			}
		}
		indices = append(indices, idx)
	}
	if indices[0] != indices[1] {
		t.Fatalf("hint moved with content: %v", indices)
	}
	if indices[0] < 0 {
		t.Fatal("hint line not found")
	}
	_ = m
}
