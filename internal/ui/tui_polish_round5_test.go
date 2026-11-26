package ui

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/config"
)

func TestMultilineValueContinuationLinesAlign(t *testing.T) {
	form := newKeyCreateFormWithLanguage("zh")
	form.focus = keyCreateValueField
	form.inputs[keyCreateValueField].Focus()
	form.inputs[keyCreateValueField].SetValue("{\n  \"a\": 1\n}")
	lines := strings.Split(stripANSI(form.View()), "\n")
	var positions []int
	for _, line := range lines {
		if idx := strings.Index(line, "> "); idx >= 0 {
			positions = append(positions, runewidth.StringWidth(line[:idx]))
		}
	}
	if len(positions) < 3 {
		t.Fatalf("expected multiline prompts, got %v", positions)
	}
	for _, pos := range positions[1:] {
		if pos != positions[0] {
			t.Fatalf("continuation lines not aligned: %v", positions)
		}
	}
}

func TestPasteInValueFormDoesNotSubmit(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	form := newKeyEditValueFormWithLanguage("k", "", 0, "zh")
	m.keyForm = &form
	m.keyForm.focus = keyFieldValue

	next, cmd := m.Update(tea.KeyMsg{
		Type: tea.KeyRunes, Paste: true, Runes: []rune("{\n  \"a\": 1\n}"),
	})
	got := next.(Model)
	if got.keyForm == nil {
		t.Fatal("pasting JSON submitted the form unexpectedly")
	}
	if cmd != nil {
		t.Fatal("paste should not create a write command")
	}
	if !strings.Contains(got.keyForm.inputs[keyFieldValue].Value(), "\n") {
		t.Fatalf("pasted newlines were lost: %q", got.keyForm.inputs[keyFieldValue].Value())
	}
}

func TestValuePreviewLimitsLongText(t *testing.T) {
	value := strings.Repeat("line\n", 100)
	preview := previewValue(value, 3, 80)
	if strings.Count(preview, "\n") > 3 {
		t.Fatalf("preview too long: %q", preview)
	}
	if !strings.Contains(preview, "…") {
		t.Fatalf("preview marker missing: %q", preview)
	}
}

func TestJSONValueIsFormattedForDisplay(t *testing.T) {
	value := `{"name":"redis","db":0}`
	formatted := formatValueForDisplay(value)
	var out any
	if err := json.Unmarshal([]byte(formatted), &out); err != nil {
		t.Fatalf("formatted JSON is invalid: %v\n%s", err, formatted)
	}
	if !strings.Contains(formatted, "\n") {
		t.Fatalf("JSON was not pretty printed: %q", formatted)
	}
}

func TestWorkspaceHintLineStaysNearBottom(t *testing.T) {
	m := manyProfilesModel(40)
	m.height = 24
	m.setView(ViewConnections)
	lines := strings.Split(m.View(), "\n")
	if len(lines) < 4 {
		t.Fatalf("view too short: %#v", lines)
	}
	hintIndex := -1
	for i, line := range lines {
		if strings.Contains(stripANSI(line), "↑↓") {
			hintIndex = i
			break
		}
	}
	if hintIndex < len(lines)-5 {
		t.Fatalf("hint is not near bottom: index=%d lines=%d", hintIndex, len(lines))
	}
}

func TestFooterSeparatesStatusAndShortcuts(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.status = strings.Repeat("错误", 100)
	plain := stripANSI(m.renderFooter(100))
	if !strings.HasPrefix(strings.TrimSpace(plain), "错误") {
		t.Fatalf("status is not left aligned: %q", plain)
	}
	if !strings.Contains(plain, "Tab/1-6") {
		t.Fatalf("shortcuts missing from footer: %q", plain)
	}
	if strings.Index(plain, "Tab/1-6") <= strings.Index(plain, "错误") {
		t.Fatalf("shortcuts are not to the right of status: %q", plain)
	}
}
