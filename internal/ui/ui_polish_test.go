package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/result"
)

func TestConfigDefaultsIncludeLanguageAndTheme(t *testing.T) {
	cfg := config.Default()
	if cfg.Language != "zh" {
		t.Fatalf("language = %q, want zh", cfg.Language)
	}
	if cfg.Theme != "aurora" {
		t.Fatalf("theme = %q, want aurora", cfg.Theme)
	}
}

func TestSettingsViewIsNavigable(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewSettings)

	view := m.renderWorkspace(60, 20)
	for _, want := range []string{"设置", "语言", "主题", "帮助"} {
		if !strings.Contains(stripANSI(view), want) {
			t.Fatalf("settings view missing %q:\n%s", want, stripANSI(view))
		}
	}
	if view, ok := viewFromNumber('6'); !ok || view != ViewSettings {
		t.Fatal("number 6 does not select Settings")
	}
}

func TestEnglishLanguageRendersEnglishChrome(t *testing.T) {
	cfg := config.Default()
	cfg.Language = "en"
	m := NewModelWithConfig(cfg, false)
	m.setView(ViewSettings)

	view := stripANSI(m.View())
	for _, want := range []string{"Settings", "Language", "Theme", "Help", "Shortcuts"} {
		if !strings.Contains(view, want) {
			t.Fatalf("English UI missing %q:\n%s", want, view)
		}
	}
	for _, unwanted := range []string{"设置", "语言", "主题", "帮助"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("English UI still contains %q", unwanted)
		}
	}
}

func TestHeaderNoLongerHasHelpButton(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	header := stripANSI(m.renderHeader(120))
	if strings.Contains(header, "[Help]") {
		t.Fatalf("header still contains Help: %q", header)
	}
}

func TestClickLanguageButtonSwitchesEnglishAndPersists(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	m := NewModelWithConfig(config.Default(), false)
	m.SetConfigPath(path)
	m.setView(ViewSettings)

	next, _ := m.Update(clickTokenMsg(t, m, "[English]"))
	got := next.(Model)
	if got.cfg.Language != "en" {
		t.Fatalf("language = %q, want en", got.cfg.Language)
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if saved.Language != "en" {
		t.Fatalf("saved language = %q, want en", saved.Language)
	}
	if !strings.Contains(stripANSI(got.renderSettingsBody()), "Language") {
		t.Fatalf("settings body not translated:\n%s", stripANSI(got.renderSettingsBody()))
	}
}

func TestClickThemeButtonPersistsAndAppliesTheme(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	m := NewModelWithConfig(config.Default(), false)
	m.SetConfigPath(path)
	m.setView(ViewSettings)

	next, _ := m.Update(clickTokenMsg(t, m, "[Ember]"))
	got := next.(Model)
	if got.cfg.Theme != "ember" {
		t.Fatalf("theme = %q, want ember", got.cfg.Theme)
	}
	if themePaletteFor(got.cfg.Theme).accent != themePaletteFor("ember").accent {
		t.Fatal("theme palette was not applied")
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if saved.Theme != "ember" {
		t.Fatalf("saved theme = %q, want ember", saved.Theme)
	}
}

func TestButtonRendererUsesSolidStyle(t *testing.T) {
	rendered := button("[OK]")
	if rendered == "[OK]" {
		t.Fatal("button was not styled")
	}
	if strings.TrimSpace(stripANSI(rendered)) != "[OK]" {
		t.Fatalf("button text changed: %q", stripANSI(rendered))
	}
	if buttonStyle.GetBackground() == nil || buttonStyle.GetForeground() == nil {
		t.Fatal("button style has no foreground/background colors")
	}
}

func TestKeysActionOrderPlacesRefreshLastAndNextBeforeIt(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	body := stripANSI(next.(Model).renderKeysBody())
	treeAt := strings.Index(body, m.btnLabel("[Tree]"))
	nextAt := strings.Index(body, m.btnLabel("[Next]"))
	refreshAt := strings.Index(body, m.btnLabel("[Refresh]"))
	if treeAt < 0 || nextAt < 0 || refreshAt < 0 {
		t.Fatalf("action buttons missing:\n%s", body)
	}
	if treeAt > nextAt || nextAt > refreshAt {
		t.Fatalf("unexpected action order: tree=%d next=%d refresh=%d", treeAt, nextAt, refreshAt)
	}
}

func TestCopyValueButtonCopiesSelectedValue(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.selectedValue = result.Normalize("hello")

	next, cmd := m.Update(clickTokenMsg(t, m, "[Copy value]"))
	got := next.(Model)
	if cmd == nil {
		t.Fatal("copy button did not return a command")
	}
	if !strings.Contains(got.Status(), "Copied") {
		t.Fatalf("status = %q, want copied", got.Status())
	}
}
