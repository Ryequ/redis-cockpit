package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/result"
	"github.com/Ryequ/redis-cockpit/internal/security"
)

const (
	navWidth   = 20
	headerRows = 2
	footerRows = 2
)

func (m Model) View() string {
	width := maxInt(m.width, 80)
	height := maxInt(m.height, 24)

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		titleStyle.Render(" Redis TUI "),
		"  ",
		subtitleStyle.Render("Redis command workbench"),
	)

	left := m.renderNav()
	center := m.renderWorkspace(width-navWidth-4, height-headerRows-footerRows)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, center)
	footer := m.renderFooter(width)

	screen := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.showHelpModal {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, m.renderHelpModal())
	}
	return screen
}

func (m Model) renderNav() string {
	var b strings.Builder
	for i, view := range navItems {
		label := fmt.Sprintf("%d  %s", i+1, view.String())
		if view == m.activeView {
			b.WriteString(navActiveStyle.Render("▸ " + label))
		} else {
			b.WriteString(navItemStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}
	return lipgloss.NewStyle().
		Width(navWidth).
		Height(maxInt(m.height-footerRows-headerRows, 10)).
		Render(b.String())
}

func (m Model) renderWorkspace(width, height int) string {
	var body string
	switch m.activeView {
	case ViewConnections:
		body = m.renderConnectionsBody()
	case ViewKeys:
		body = m.renderKeysBody()
	case ViewWorkbench:
		body = m.renderWorkbenchBody()
	case ViewOps:
		body = m.renderOpsBody()
	case ViewStreams:
		body = m.renderStreamsBody()
	default:
		body = lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(accent).Render(m.activeView.String()),
			"",
			subtitleStyle.Render(m.workspaceHint()),
			"",
			okStyle.Render("● ready"),
			"  "+subtitleStyle.Render("模块初始化完成，按数字键或点击左侧导航切换页面。"),
		)
	}

	return panelStyle.
		Width(maxInt(width, 40)).
		Height(maxInt(height, 10)).
		Render(body)
}

func (m Model) renderConnectionsBody() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Connections"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"Profiles:",
	}

	if len(m.cfg.Profiles) == 0 {
		lines = append(lines, warnStyle.Render("  No profiles configured"))
	}
	for i, profile := range m.cfg.Profiles {
		marker := "  "
		if i == m.selectedProfile {
			marker = okStyle.Render("▸ ")
		}
		topology := "single"
		if profile.Cluster {
			topology = "cluster"
		} else if profile.Sentinel {
			topology = "sentinel"
		}
		if profile.TLS {
			topology += "+TLS"
		}
		lines = append(lines, fmt.Sprintf("%s%d. %s  %s  [%s]", marker, i+1, profile.Name, profile.Addr, topology))
	}

	if len(m.cfg.Profiles) > 0 {
		profile := m.cfg.Profiles[m.selectedProfile]
		lines = append(lines,
			"",
			subtitleStyle.Render("Selected:"),
			"  Name   : "+profile.Name,
			"  Addr   : "+security.RedactText(profile.Addr),
			"  DB     : "+fmt.Sprint(profile.DB),
			"  Auth   : "+authLabel(profile.Username, profile.Password),
			"  Mode   : "+modeLabel(profile),
			"",
			"↑↓/j/k 选择  •  Enter 连接  •  鼠标点击列表项选择",
		)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderWorkbenchBody() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Workbench"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"Command:",
		m.commandInput.View(),
		"",
		"History:",
	}

	history := m.history.All()
	if len(history) == 0 {
		lines = append(lines, subtitleStyle.Render("  (empty)"))
	}
	for i := len(history) - 1; i >= 0 && i >= len(history)-5; i-- {
		lines = append(lines, "  "+security.RedactText(history[i]))
	}

	lines = append(lines, "", "Result:")
	if m.pendingCommand != nil {
		redacted := strings.Join(security.RedactCommand(m.pendingCommand), " ")
		lines = append(lines, dangerStyle.Render("  Confirm critical command: "+redacted), warnStyle.Render("  y = execute, n/Esc = cancel"))
	} else {
		lines = append(lines, result.Render(m.lastResult, result.FormatRaw))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderKeysBody() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Keys"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"Filter: " + m.filterInput.View(),
		"",
		"Keys:",
	}

	if len(m.keys) == 0 {
		lines = append(lines, subtitleStyle.Render("  (empty)"))
	}
	for i, key := range m.keys {
		marker := "  "
		if i == m.selectedKey {
			marker = okStyle.Render("▸ ")
		}
		lines = append(lines, fmt.Sprintf("%s%s  [%s]  ttl=%s", marker, key.Name, key.Type, ttlLabel(key.TTL)))
	}

	lines = append(lines, "", "Value:")
	if m.pendingDelete != nil {
		lines = append(lines,
			dangerStyle.Render("  Confirm delete: "+strings.Join(m.pendingDelete, ", ")),
			warnStyle.Render("  y = delete, n/Esc = cancel"),
		)
	} else {
		lines = append(lines, result.Render(m.selectedValue, result.FormatRaw))
	}

	nextPageHint := ""
	if m.keyHasMore {
		nextPageHint = "  •  n 下一页"
	}
	lines = append(lines, "", "↑↓/j/k 选择  •  Enter 查看值  •  / 筛选  •  r 刷新"+nextPageHint+"  •  d 删除")
	return strings.Join(lines, "\n")
}

func (m Model) renderOpsBody() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Ops"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"INFO:",
	}

	for _, section := range []string{"Server", "Memory", "Replication"} {
		values := m.opsSnapshot.Info[section]
		if len(values) == 0 {
			continue
		}
		lines = append(lines, "  ["+section+"]")
		keys := sortedKeys(values)
		for i, key := range keys {
			if i >= 5 {
				break
			}
			lines = append(lines, fmt.Sprintf("    %s: %s", key, values[key]))
		}
	}

	lines = append(lines, "", "CONFIG:")
	configKeys := sortedKeys(m.opsSnapshot.Config)
	for i, key := range configKeys {
		if i >= 5 {
			break
		}
		value := m.opsSnapshot.Config[key]
		if isSensitiveConfigKey(key) {
			value = "***"
		}
		lines = append(lines, fmt.Sprintf("  %s = %s", key, value))
	}
	if len(configKeys) == 0 {
		lines = append(lines, subtitleStyle.Render("  (empty)"))
	}

	lines = append(lines, "", "CLIENT:")
	for i, client := range m.opsSnapshot.Clients {
		if i >= 5 {
			break
		}
		lines = append(lines, fmt.Sprintf("  id=%s addr=%s db=%s cmd=%s", client.ID, client.Addr, client.DB, client.Cmd))
	}
	if len(m.opsSnapshot.Clients) == 0 {
		lines = append(lines, subtitleStyle.Render("  (empty)"))
	}

	lines = append(lines, "", "SLOWLOG:")
	for i, entry := range m.opsSnapshot.Slowlog {
		if i >= 3 {
			break
		}
		command := strings.Join(security.RedactCommand(entry.Command), " ")
		lines = append(lines, fmt.Sprintf("  #%d %dms %s", entry.ID, entry.Duration, command))
	}
	if len(m.opsSnapshot.Slowlog) == 0 {
		lines = append(lines, subtitleStyle.Render("  (empty)"))
	}

	lines = append(lines,
		"",
		"Modules: "+strings.Join(m.opsSnapshot.Modules, ", "),
		"Role: "+m.opsSnapshot.Replication.Role,
		"",
		"r 刷新快照  •  脚本与事务可在 Workbench 中执行",
	)
	return strings.Join(lines, "\n")
}

func (m Model) renderStreamsBody() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Streams"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"Channel: " + m.streamInput.View(),
		"",
		"Pub/Sub:",
	}

	if m.subscription == nil {
		lines = append(lines, subtitleStyle.Render("  (not subscribed)"))
	}
	for i := len(m.streamMessages) - 1; i >= 0 && i >= len(m.streamMessages)-5; i-- {
		message := m.streamMessages[i]
		lines = append(lines, fmt.Sprintf("  %s: %s", message.Channel, message.Payload))
	}

	lines = append(lines, "", "MONITOR:")
	for i := len(m.monitorEvents) - 1; i >= 0 && i >= len(m.monitorEvents)-5; i-- {
		lines = append(lines, "  "+m.monitorEvents[i].Payload)
	}
	if len(m.monitorEvents) == 0 {
		lines = append(lines, subtitleStyle.Render("  (not running)"))
	}

	lines = append(lines,
		"",
		"Stream: 使用 Workbench 执行 XADD/XRANGE/XREAD 管理流条目。",
		"",
		"s 订阅  •  p 发布测试消息  •  m 启动 MONITOR  •  x 停止  •  输入频道名",
	)
	return strings.Join(lines, "\n")
}

func isSensitiveConfigKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "pass") || strings.Contains(lower, "auth") || strings.Contains(lower, "acl") || strings.Contains(lower, "secret")
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ttlLabel(ttl time.Duration) string {
	if ttl <= 0 {
		return "-"
	}
	return ttl.String()
}

func authLabel(username, password string) string {
	switch {
	case username != "" && password != "":
		return "username+password"
	case username != "":
		return "username"
	case password != "":
		return "password"
	default:
		return "none"
	}
}

func modeLabel(profile redis.ConnectionProfile) string {
	modes := []string{}
	if profile.Cluster {
		modes = append(modes, "cluster")
	}
	if profile.Sentinel {
		modes = append(modes, "sentinel")
	}
	if profile.TLS {
		modes = append(modes, "TLS")
	}
	if profile.ReadOnly {
		modes = append(modes, "read-only")
	}
	if len(modes) == 0 {
		return "default"
	}
	return strings.Join(modes, ", ")
}

func (m Model) workspaceHint() string {
	switch m.activeView {
	case ViewConnections:
		return "连接档案、认证、TLS、Cluster 与 Sentinel 将在此管理。"
	case ViewKeys:
		return "使用 SCAN 分页浏览 Key，查看类型、TTL 与值预览。"
	case ViewWorkbench:
		return "搜索命令、填写参数或执行原始 Redis 命令。"
	case ViewOps:
		return "INFO、CONFIG、CLIENT、SLOWLOG 与复制状态。"
	case ViewStreams:
		return "Pub/Sub、MONITOR 与 Stream 实时消息流。"
	default:
		return ""
	}
}

func (m Model) renderFooter(width int) string {
	mouseHint := ""
	if m.mouse {
		mouseHint = "  •  Mouse Click 选择"
	}
	text := "Tab/1-5 切换页面  •  ↑↓/hjkl 导航  •  Enter 确认  •  ? 帮助" + mouseHint + "  •  Ctrl+C 退出"
	return footerStyle.Width(maxInt(width, 80)).Render(statusPrefix(m.status) + text)
}

func (m Model) renderHelpModal() string {
	lines := []string{
		"Redis TUI 快捷键",
		"",
		"1-5 / Tab / Shift+Tab   切换页面",
		"↑↓ / j k                移动选择",
		"Enter                   确认 / 执行",
		"Esc                     返回 / 关闭",
		"?                       显示帮助",
		"Ctrl+C / Ctrl+Q         退出",
		"",
		"鼠标：点击左侧导航或列表项；滚轮滚动长结果。",
	}
	modal := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(violet).
		Background(surface).
		Foreground(text).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
	return dangerStyle.Render(modal)
}

func statusPrefix(status string) string {
	if status == "" {
		return ""
	}
	return warnStyle.Render(status) + "  |  "
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
