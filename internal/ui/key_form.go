package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

type keyFormMode int

const (
	keyFormCreate keyFormMode = iota
	keyFormEditValue
)

const (
	keyFieldName int = iota
	keyCreateValueField
)

const keyFieldValue = 0

type keyForm struct {
	mode   keyFormMode
	key    string
	ttl    time.Duration
	inputs []textinput.Model
	labels []string
	focus  int
}

func newKeyCreateForm() keyForm {
	keyInput := textinput.New()
	keyInput.Placeholder = "app:config:timeout"
	keyInput.CharLimit = 1024
	keyInput.Focus()

	valueInput := textinput.New()
	valueInput.Placeholder = "optional value"
	valueInput.CharLimit = 1 << 20

	return keyForm{
		mode:   keyFormCreate,
		inputs: []textinput.Model{keyInput, valueInput},
		labels: []string{"Key", "Value"},
	}
}

func newKeyEditValueForm(key, value string, ttl time.Duration) keyForm {
	valueInput := textinput.New()
	valueInput.SetValue(value)
	valueInput.Placeholder = "value"
	valueInput.CharLimit = 1 << 20
	valueInput.Focus()

	return keyForm{
		mode:   keyFormEditValue,
		key:    key,
		ttl:    ttl,
		inputs: []textinput.Model{valueInput},
		labels: []string{"Value"},
	}
}

func (f *keyForm) nextFocus() {
	f.focus = (f.focus + 1) % len(f.inputs)
	f.inputs[f.focus].Focus()
}

func (f *keyForm) prevFocus() {
	f.focus = (f.focus - 1 + len(f.inputs)) % len(f.inputs)
	f.inputs[f.focus].Focus()
}

func (f keyForm) View() string {
	title := "New Key"
	if f.mode == keyFormEditValue {
		title = "Edit Value: " + f.key
	}
	lines := []string{title, "", buttonRow("[OK]", "[Cancel]"), ""}
	for i, label := range f.labels {
		marker := "  "
		if i == f.focus {
			marker = "▸ "
		}
		lines = append(lines, fmt.Sprintf("%s%-10s %s", marker, label+":", f.inputs[i].View()))
	}
	lines = append(lines, "", "Tab/Shift+Tab 切换字段  •  Enter 确认  •  Esc 取消")
	return strings.Join(lines, "\n")
}
