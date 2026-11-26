package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
	"github.com/Ryequ/redis-cockpit/internal/result"
)

func polishTestModel() Model {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewKeys)
	m.keys = []adapters.KeyInfo{{Name: "greeting", Type: "string"}}
	m.selectedValue = result.Normalize("hello")
	return m
}

func TestTopTabsUseChineseLabelsAndSolidButtons(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	header := m.renderHeader(140)
	plain := stripANSI(header)
	if !strings.Contains(plain, "[1 连接管理]") {
		t.Fatalf("Chinese top tab missing: %q", plain)
	}
	if strings.Contains(plain, "[1 Connections]") {
		t.Fatalf("English top tab still present: %q", plain)
	}
	if buttonStyle.GetBackground() == nil {
		t.Fatal("top tab buttons are not using solid button style")
	}
}

func TestClickLocalizedTopTabSwitchesView(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	next, _ := m.Update(clickTokenMsg(t, m, "[3 Workbench]"))
	if next.(Model).ActiveView() != ViewWorkbench {
		t.Fatalf("view = %v, want Workbench", next.(Model).ActiveView())
	}
}

func TestKeyActionButtonsUseChineseLabels(t *testing.T) {
	m := polishTestModel()
	body := stripANSI(m.renderKeysBody())
	for _, want := range []string{"[树]", "[新建 Key]", "[编辑 Value]", "[删除]", "[下一个]", "[刷新]"} {
		if !strings.Contains(body, want) {
			t.Fatalf("localized button %q missing:\n%s", want, body)
		}
	}
}

func TestEditValueRequiresFirstConfirmation(t *testing.T) {
	m := polishTestModel()

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	got := next.(Model)
	if got.keyForm != nil {
		t.Fatal("edit form opened without confirmation")
	}
	if got.pendingEditConfirm != "greeting" {
		t.Fatalf("pending edit confirm = %q, want greeting", got.pendingEditConfirm)
	}
	if !strings.Contains(stripANSI(got.renderKeysBody()), "[确认]") {
		t.Fatalf("edit confirmation missing:\n%s", stripANSI(got.renderKeysBody()))
	}
}

func TestEditValueConfirmationOpensForm(t *testing.T) {
	m := polishTestModel()
	m.pendingEditConfirm = "greeting"

	next, _ := m.Update(clickTokenMsg(t, m, "[Confirm]"))
	got := next.(Model)
	if got.pendingEditConfirm != "" {
		t.Fatal("edit confirmation was not cleared")
	}
	if got.keyForm == nil || got.keyForm.mode != keyFormEditValue {
		t.Fatalf("key form = %#v, want edit value form", got.keyForm)
	}
}

func TestDeleteKeyRequiresTwoConfirmations(t *testing.T) {
	m := polishTestModel()
	m.pendingDelete = []string{"greeting"}
	m.deleteConfirmStage = 0

	next, cmd := m.Update(clickTokenMsg(t, m, "[Confirm]"))
	got := next.(Model)
	if cmd != nil {
		t.Fatal("first confirmation must not delete the key")
	}
	if got.pendingDelete == nil || got.deleteConfirmStage != 1 {
		t.Fatal("delete did not advance to second confirmation")
	}

	next, cmd = got.Update(clickTokenMsg(t, got, "[Y] Delete"))
	got = next.(Model)
	if got.pendingDelete != nil {
		t.Fatal("delete confirmation was not cleared")
	}
	if cmd == nil {
		t.Fatal("second confirmation did not delete the key")
	}
}
