package ui

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type keyFormMode int

const (
	keyFormCreate keyFormMode = iota
	keyFormEditValue
)

const keyFieldName = 0
const keyFieldValue = 0
const keyCreateValueField = 1

type keyFormField struct {
	multiline bool
	text      textinput.Model
	area      textarea.Model
}

func newKeyFormText(label, placeholder, value string, focus bool) keyFormField {
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = placeholder
	input.SetValue(value)
	input.CharLimit = 1024
	input.Width = 64
	if focus {
		input.Focus()
	}
	return keyFormField{text: input}
}

func setTextareaValue(area *textarea.Model, value string) {
	area.SetValue("")
	area.Focus()
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		area.InsertString(line)
		if i < len(lines)-1 {
			updated, _ := area.Update(tea.KeyMsg{Type: tea.KeyEnter})
			*area = updated
		}
	}
}

func newKeyFormArea(label, placeholder, value string, focus bool) keyFormField {
	area := textarea.New()
	area.Prompt = "> "
	area.Placeholder = placeholder
	area.ShowLineNumbers = false
	area.SetWidth(64)
	area.SetHeight(6)
	area.CharLimit = 1 << 20
	setTextareaValue(&area, value)
	if !focus {
		area.Blur()
	}
	return keyFormField{multiline: true, area: area}
}

func (f *keyFormField) Paste(value string) {
	if !f.multiline {
		f.text.SetValue(f.text.Value() + value)
		return
	}
	focused := f.area.Focused()
	f.area.Focus()
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		f.area.InsertString(line)
		if i < len(lines)-1 {
			updated, _ := f.area.Update(tea.KeyMsg{Type: tea.KeyEnter})
			f.area = updated
		}
	}
	if !focused {
		f.area.Blur()
	}
}

func indentContinuation(view string, width int) string {
	lines := strings.Split(view, "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", width) + lines[i]
	}
	return strings.Join(lines, "\n")
}

func previewValue(value string, maxLines, maxChars int) string {
	value = formatValueForDisplay(value)
	runes := []rune(value)
	truncated := len(runes) > maxChars
	if truncated {
		runes = runes[:maxChars]
	}
	lines := strings.Split(string(runes), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		truncated = true
	}
	result := strings.Join(lines, "\n")
	if truncated {
		result += " …"
	}
	return result
}

func formatValueForDisplay(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !json.Valid([]byte(trimmed)) {
		return value
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(trimmed), "", "  "); err != nil {
		return value
	}
	return buf.String()
}

func (f keyFormField) View() string {
	if f.multiline {
		return f.area.View()
	}
	return f.text.View()
}

func (f keyFormField) ViewPreview() string {
	if !f.multiline {
		return f.text.View()
	}
	if f.area.Focused() {
		return f.area.View()
	}
	return previewValue(f.area.Value(), 3, 240)
}

func (f *keyFormField) SetValue(value string) {
	if f.multiline {
		focused := f.area.Focused()
		setTextareaValue(&f.area, value)
		if !focused {
			f.area.Blur()
		}
		return
	}
	f.text.SetValue(value)
}

func (f keyFormField) Value() string {
	if f.multiline {
		return f.area.Value()
	}
	return f.text.Value()
}

func (f *keyFormField) Focus() tea.Cmd {
	if f.multiline {
		return f.area.Focus()
	}
	return f.text.Focus()
}

func (f *keyFormField) Blur() {
	if f.multiline {
		f.area.Blur()
		return
	}
	f.text.Blur()
}

func (f keyFormField) Focused() bool {
	if f.multiline {
		return f.area.Focused()
	}
	return f.text.Focused()
}

func (f *keyFormField) Update(msg tea.Msg) tea.Cmd {
	if f.multiline {
		var cmd tea.Cmd
		f.area, cmd = f.area.Update(msg)
		return cmd
	}
	var cmd tea.Cmd
	f.text, cmd = f.text.Update(msg)
	return cmd
}

type keyForm struct {
	mode           keyFormMode
	language       string
	key            string
	ttl            time.Duration
	inputs         []keyFormField
	labels         []string
	focus          int
	scroll         int
	pendingConfirm bool
}

func newKeyCreateForm() keyForm {
	return newKeyCreateFormWithLanguage("zh")
}

func newKeyCreateFormWithLanguage(language string) keyForm {
	return keyForm{
		mode:     keyFormCreate,
		language: language,
		inputs: []keyFormField{
			newKeyFormText("Key", "app:config:timeout", "", true),
			newKeyFormArea("Value", tr(language, "optional value"), "", false),
		},
		labels: []string{tr(language, "Key"), tr(language, "Value")},
	}
}

func newKeyEditValueForm(key, value string, ttl time.Duration) keyForm {
	return newKeyEditValueFormWithLanguage(key, value, ttl, "zh")
}

func newKeyEditValueFormWithLanguage(key, value string, ttl time.Duration, language string) keyForm {
	return keyForm{
		mode:     keyFormEditValue,
		language: language,
		key:      key,
		ttl:      ttl,
		inputs: []keyFormField{
			newKeyFormArea("Value", tr(language, "value"), value, true),
		},
		labels: []string{tr(language, "Value")},
	}
}

func (f *keyForm) setFocus(next int) {
	if next < 0 || next >= len(f.inputs) {
		return
	}
	if f.focus < len(f.inputs) {
		f.inputs[f.focus].Blur()
	}
	f.focus = next
	f.inputs[next].Focus()
}

func (f *keyForm) nextFocus() {
	if len(f.inputs) == 0 {
		return
	}
	f.inputs[f.focus].Blur()
	f.focus = (f.focus + 1) % len(f.inputs)
	f.inputs[f.focus].Focus()
}

func (f *keyForm) prevFocus() {
	if len(f.inputs) == 0 {
		return
	}
	f.inputs[f.focus].Blur()
	f.focus = (f.focus - 1 + len(f.inputs)) % len(f.inputs)
	f.inputs[f.focus].Focus()
}

func (f keyForm) title() string {
	if f.mode == keyFormEditValue {
		return tr(f.language, "Edit Value: ") + f.key
	}
	return tr(f.language, "New Key")
}

func (f keyForm) contentLines() []string {
	lines := make([]string, 0, len(f.labels))
	for i, label := range f.labels {
		marker := "  "
		if i == f.focus {
			marker = "▸ "
		}
		prefix := marker + padToDisplayWidth(label+":", 12) + " "
		fieldView := indentContinuation(f.inputs[i].ViewPreview(), runewidth.StringWidth(prefix))
		lines = append(lines, prefix+fieldView)
	}
	return lines
}

func (f keyForm) footerLines() []string {
	lines := []string{""}
	if f.pendingConfirm {
		lines = append(lines, tr(f.language, "Confirm edit value?"))
	}
	return append(lines,
		buttonRowForLanguage(f.language, "[OK]", "[Cancel]"),
		tr(f.language, "Tab/Shift+Tab 切换字段  •  Enter 确认  •  Esc 取消"),
	)
}

func (f keyForm) View() string {
	lines := []string{f.title(), ""}
	lines = append(lines, f.contentLines()...)
	lines = append(lines, f.footerLines()...)
	return strings.Join(lines, "\n")
}

func (f keyForm) ViewHeight(maxLines int) string {
	all := f.contentLines()
	footer := f.footerLines()
	available := maxInt(maxLines-2-len(footer), 1)
	start := f.scroll
	if f.focus < start {
		start = f.focus
	}
	if f.focus >= start+available {
		start = f.focus - available + 1
	}
	maxStart := maxInt(len(all)-available, 0)
	if start > maxStart {
		start = maxStart
	}
	if start < 0 {
		start = 0
	}
	end := minInt(start+available, len(all))
	lines := []string{f.title(), ""}
	lines = append(lines, all[start:end]...)
	lines = append(lines, footer...)
	return strings.Join(lines, "\n")
}
