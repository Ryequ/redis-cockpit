package ui

import (
	"strings"
	"testing"

	"github.com/Ryequ/redis-cockpit/internal/config"
)

func TestLongMultilineStatusKeepsHeaderClickable(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.width = 80
	m.height = 12
	m.setView(ViewConnections)
	m.status = "Connect failed: line one\nline two\nline three\n" + strings.Repeat("very long error ", 20)

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > m.height {
		t.Fatalf("view has %d lines, terminal has %d", len(lines), m.height)
	}
	if !strings.Contains(stripANSI(lines[0]), m.btnLabel("[1 Connections]")) {
		t.Fatalf("header missing with long status: %q", stripANSI(lines[0]))
	}
	if strings.Contains(m.renderFooter(80), "\n") {
		t.Fatal("footer rendered multiple lines for multiline status")
	}

	next, _ := m.Update(clickTokenMsg(t, m, "[3 Workbench]"))
	if next.(Model).ActiveView() != ViewWorkbench {
		t.Fatal("top tab click failed after long status")
	}
}

func TestSummaryPanelsHaveEqualHeightWithLongValue(t *testing.T) {
	server := renderMetricPanel("Server", [][2]string{
		{"Version", "7.2.4"},
		{"OS", strings.Repeat("Linux 6.1 ", 30)},
		{"PID", "1234"},
	}, 18)
	memory := renderMetricPanel("Memory", [][2]string{
		{"Used", "1M"}, {"Peak", "2M"}, {"Lua", "37K"},
	}, 18)
	stats := renderMetricPanel("Stats", [][2]string{
		{"Clients", "5"}, {"Connections", "99"}, {"Commands", "1024"},
	}, 18)
	if len(strings.Split(server, "\n")) != len(strings.Split(memory, "\n")) ||
		len(strings.Split(server, "\n")) != len(strings.Split(stats, "\n")) {
		t.Fatalf("summary panels have different heights:\n%s\n%s\n%s", server, memory, stats)
	}
}

func TestStatsClientsReadsClientsInfoSection(t *testing.T) {
	cfg := config.Default()
	cfg.Language = "en"
	m := NewModelWithConfig(cfg, false)
	m.opsSnapshot.Info = map[string]map[string]string{
		"Clients": {"connected_clients": "42"},
		"Stats": {
			"total_connections_received": "99",
			"total_commands_processed":   "1024",
		},
	}

	body := stripANSI(m.renderConnectionsBody())
	if !strings.Contains(body, "42") {
		t.Fatalf("Clients value missing from summary:\n%s", body)
	}
}

func TestConnectionsBodyChineseLocalization(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	m.setView(ViewConnections)
	body := stripANSI(m.renderConnectionsBody())
	for _, want := range []string{"搜索:", "连接列表:", "已选:", "名称", "地址", "认证", "模式"} {
		if !strings.Contains(body, want) {
			t.Fatalf("Chinese Connections body missing %q:\n%s", want, body)
		}
	}
	if m.profileSearch.Placeholder != "筛选连接档案" {
		t.Fatalf("search placeholder = %q, want 筛选连接档案", m.profileSearch.Placeholder)
	}
}

func TestConnectionsBodyEnglishLocalization(t *testing.T) {
	cfg := config.Default()
	cfg.Language = "en"
	m := NewModelWithConfig(cfg, false)
	m.setView(ViewConnections)
	body := stripANSI(m.renderConnectionsBody())
	for _, want := range []string{"Search:", "Profiles:", "Selected:", "Name", "Addr", "Auth", "Mode"} {
		if !strings.Contains(body, want) {
			t.Fatalf("English Connections body missing %q:\n%s", want, body)
		}
	}
	if m.profileSearch.Placeholder != "filter profiles" {
		t.Fatalf("search placeholder = %q, want filter profiles", m.profileSearch.Placeholder)
	}
}

func TestTopTabsUseUniformButtonStyle(t *testing.T) {
	m := NewModelWithConfig(config.Default(), false)
	header := m.renderHeader(140)
	if !strings.Contains(stripANSI(header), "▸"+m.btnLabel("[1 Connections]")) {
		t.Fatalf("active marker missing: %q", stripANSI(header))
	}
	if activeButtonStyle.GetBackground() == nil || buttonStyle.GetBackground() == nil {
		t.Fatal("button styles are missing theme colors")
	}
}
