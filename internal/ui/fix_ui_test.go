package ui

import (
	"bytes"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
	"github.com/Ryequ/redis-cockpit/internal/result"
)

func TestCopyValueUsesOSC52WhenNoClipboardCommand(t *testing.T) {
	oldCommand := newClipboardCommand
	oldWriter := clipboardWriter
	defer func() {
		newClipboardCommand = oldCommand
		clipboardWriter = oldWriter
	}()
	newClipboardCommand = func(string) *clipboardCommandSpec { return nil }
	var buf bytes.Buffer
	clipboardWriter = &buf

	msg := copyValueCmd("hello", 7)().(valueCopiedMsg)
	if msg.err != nil {
		t.Fatalf("copyValueCmd returned error: %v", msg.err)
	}
	if !strings.Contains(buf.String(), "\x1b]52;c;") {
		t.Fatalf("OSC52 clipboard sequence not written: %q", buf.String())
	}
}

func TestProfileFormLocalizesChineseLabelsAndButtons(t *testing.T) {
	form := newProfileFormWithLanguage(redis.ConnectionProfile{}, profileCreate, "zh")
	view := stripANSI(form.View())
	for _, want := range []string{"新建连接档案", "名称:", "地址:", "保存", "取消", "模板:"} {
		if !strings.Contains(view, want) {
			t.Fatalf("Chinese profile form missing %q:\n%s", want, view)
		}
	}
}

func TestProfileFormLocalizesEnglishLabelsAndButtons(t *testing.T) {
	form := newProfileFormWithLanguage(redis.ConnectionProfile{}, profileCreate, "en")
	view := stripANSI(form.View())
	for _, want := range []string{"New Connection Profile", "Name:", "Address:", "Save", "Cancel", "Templates:"} {
		if !strings.Contains(view, want) {
			t.Fatalf("English profile form missing %q:\n%s", want, view)
		}
	}
}

func TestKeyFormLocalizesChineseLabelsAndButtons(t *testing.T) {
	form := newKeyCreateFormWithLanguage("zh")
	view := stripANSI(form.View())
	for _, want := range []string{"新建 Key", "键名:", "值:", "确认", "取消"} {
		if !strings.Contains(view, want) {
			t.Fatalf("Chinese key form missing %q:\n%s", want, view)
		}
	}
}

func TestKeyFormLocalizesEnglishLabelsAndButtons(t *testing.T) {
	form := newKeyCreateFormWithLanguage("en")
	view := stripANSI(form.View())
	for _, want := range []string{"New Key", "Key:", "Value:", "OK", "Cancel"} {
		if !strings.Contains(view, want) {
			t.Fatalf("English key form missing %q:\n%s", want, view)
		}
	}
}

func TestEditKeyOpensFormDirectly(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.keys = []adapters.KeyInfo{{Name: "greeting", Type: "string"}}
	m.selectedValue = result.Normalize("hello")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	got := next.(Model)
	if got.keyForm == nil || got.keyForm.mode != keyFormEditValue {
		t.Fatalf("edit form was not opened directly: %#v", got.keyForm)
	}
}

func TestEditKeyConfirmsWhenFormOKPressed(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.keys = []adapters.KeyInfo{{Name: "greeting", Type: "string"}}
	m.selectedValue = result.Normalize("hello")
	form := newKeyEditValueFormWithLanguage("greeting", "hello", 0, "zh")
	m.keyForm = &form

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(Model)
	if cmd != nil {
		t.Fatal("first form confirmation must not write")
	}
	if got.keyForm == nil || !got.keyForm.pendingConfirm {
		t.Fatal("edit confirmation was not shown in form")
	}

	next, cmd = got.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got = next.(Model)
	if cmd == nil {
		t.Fatal("second form confirmation did not write")
	}
}

func TestDeleteKeyUsesSingleConfirmation(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.pendingDelete = []string{"greeting"}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	got := next.(Model)
	if got.pendingDelete != nil {
		t.Fatal("delete confirmation was not cleared")
	}
	if cmd == nil {
		t.Fatal("single confirmation did not delete the key")
	}
}
