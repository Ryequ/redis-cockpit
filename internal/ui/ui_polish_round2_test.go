package ui

import (
	"strings"
	"testing"

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
