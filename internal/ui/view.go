package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/result"
	"github.com/Ryequ/redis-cockpit/internal/security"
)

const (
	navWidth              = 20
	headerRows            = 1
	footerRows            = 1
	panelVerticalOverhead = 4
)

func (m Model) View() string {
	width := maxInt(m.width, 80)
	header := m.renderHeader(width)
	if m.height < 8 {
		lines := []string{header}
		if m.height >= 2 {
			lines = append(lines, "Terminal too small")
		}
		return truncateToViewport(strings.Join(lines, "\n"), maxInt(m.height, 1), width)
	}
	height := m.height

	left := m.renderNav()
	center := m.renderWorkspace(width-navWidth-4, height-headerRows-footerRows-panelVerticalOverhead)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, center)
	footer := m.renderFooter(width)

	headerLineCount := len(strings.Split(header, "\n"))
	footerLineCount := len(strings.Split(footer, "\n"))
	bodyHeight := maxInt(height-headerLineCount-footerLineCount, 1)
	body = truncateToViewport(body, bodyHeight, width)

	screen := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.showHelpModal {
		screen = lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, m.renderHelpModal())
	}
	return truncateToViewport(screen, height, width)
}

func (m Model) renderHeader(width int) string {
	tabs := make([]string, 0, len(navItems))
	for i, view := range navItems {
		label := fmt.Sprintf("[%d %s]", i+1, view.String())
		if view == m.activeView {
			label = "▸" + label
			label = navActiveStyle.Render(label)
		} else {
			label = subtitleStyle.Render(label)
		}
		tabs = append(tabs, label)
	}
	joined := lipgloss.JoinHorizontal(
		lipgloss.Left,
		titleStyle.Render(" Redis Cockpit "),
		"  ",
		strings.Join(tabs, " "),
		"  ",
		subtitleStyle.Render("[Help]"),
	)
	return runewidth.Truncate(stripANSI(joined), maxInt(width-2, 1), "…")
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
		Height(maxInt(m.height-footerRows-headerRows, 1)).
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

	clampedHeight := maxInt(height, 1)
	clampedWidth := maxInt(width, 20)
	body = truncateToViewport(body, clampedHeight, clampedWidth)
	return panelStyle.
		Width(clampedWidth).
		Height(clampedHeight).
		Render(body)
}

func truncateToViewport(body string, maxLines, maxWidth int) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		plain := stripANSI(line)
		if runewidth.StringWidth(plain) > maxWidth {
			lines[i] = runewidth.Truncate(plain, maxWidth, "…")
		}
	}
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}
	lines = lines[:maxLines]
	marker := "  ... truncated"
	lines[len(lines)-1] = runewidth.Truncate(lines[len(lines)-1]+marker, maxWidth, "…")
	return strings.Join(lines, "\n")
}

func (m Model) renderConnectionsBody() string {
	if m.profileIOForm != nil {
		return m.profileIOForm.View()
	}
	if m.profileForm != nil {
		return m.profileForm.View()
	}

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Connections"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"Search: " + m.profileSearch.View(),
		"",
		subtitleStyle.Render("[Connect] [N new] [E edit] [D delete] [C clone] [I import] [X export] [T test] [Test all] [Search] [Sort] [Group]"),
		"",
		"Profiles:",
	}

	if m.pendingProfileDelete != "" {
		lines = append(lines, renderConfirmBox(
			"Delete connection profile",
			m.pendingProfileDelete,
			"[Y] Delete",
			"[N] Cancel",
		), "")
	}

	visible := m.visibleProfileIndices()
	if len(m.cfg.Profiles) == 0 {
		lines = append(lines, warnStyle.Render("  No profiles configured"))
	}
	if len(visible) == 0 && len(m.cfg.Profiles) > 0 {
		lines = append(lines, warnStyle.Render("  No profiles match filter"))
	}
	lastGroup := ""
	for _, idx := range visible {
		profile := m.cfg.Profiles[idx]
		group := m.profileModeKey(profile)
		if m.profileGroupByMode && group != lastGroup {
			lines = append(lines, subtitleStyle.Render("Group: "+group))
			lastGroup = group
		}
		marker := "  "
		if idx == m.selectedProfile {
			marker = okStyle.Render("▸ ")
		}
		active := ""
		if profile.Name == m.cfg.ActiveProfile || profile.ID == m.cfg.ActiveProfile {
			active = "  active"
		}
		lines = append(lines, fmt.Sprintf("%s%d. %s  %s  [%s]%s%s", marker, idx+1, profile.Name, profile.Addr, group, active, m.healthLabel(profile.Name)))
	}

	if len(m.cfg.Profiles) > 0 {
		lines = append(lines, "", m.renderConnectionSummary())
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
			"↑↓/j/k 选择  •  Enter 连接  •  n 新增  •  e 编辑  •  d 删除  •  t 测试  •  鼠标点击列表项选择",
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
		m.commandInput.View() + "  " + subtitleStyle.Render("[Input]") + "  " + okStyle.Render("[Execute]"),
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
		lines = append(lines,
			dangerStyle.Render("  Confirm critical command: "+redacted),
			warnStyle.Render("  [Y] Execute   [N] Cancel"),
		)
	} else {
		lines = append(lines, result.Render(m.lastResult, result.FormatRaw))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderKeysBody() string {
	if m.keyForm != nil {
		return m.keyForm.View()
	}
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Keys"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"Filter: " + m.filterInput.View(),
		"",
		subtitleStyle.Render("[Refresh] [Next] [New key] [Edit value] [Del] " + keyViewToggleLabel(m)),
		"",
		"Keys:",
	}

	if len(m.keys) == 0 {
		lines = append(lines, subtitleStyle.Render("  (empty)"))
	} else {
		lines = append(lines, m.renderKeyRows())
	}

	lines = append(lines, "", "Value:")
	if m.pendingDelete != nil {
		lines = append(lines, renderConfirmBox(
			"Confirm delete",
			strings.Join(m.pendingDelete, ", "),
			"[Y] Delete",
			"[N] Cancel",
		))
	} else {
		lines = append(lines, result.Render(m.selectedValue, result.FormatRaw))
	}

	nextPageHint := ""
	if m.keyHasMore {
		nextPageHint = "  •  n 下一页"
	}
	lines = append(lines, "", "↑↓/j/k 选择  •  Enter 查看值  •  a 新增 Key  •  v 修改 Value  •  t 树/平铺  •  ←/→ 折叠/展开  •  滚轮/PgUp/PgDn 滚动  •  / 筛选  •  r 刷新"+nextPageHint+"  •  d 删除")
	return strings.Join(lines, "\n")
}

func (m Model) renderOpsBody() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Ops"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		subtitleStyle.Render("[Refresh]"),
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
		"Channel: " + m.streamInput.View() + "  " + subtitleStyle.Render("[Channel]"),
		"",
		subtitleStyle.Render("[Sub] [Pub] [Monitor] [Stop]"),
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
	footerWidth := maxInt(width, 80)
	footerText := stripANSI(statusPrefix(m.status) + text)
	footerText = runewidth.Truncate(footerText, maxInt(footerWidth-2, 1), "…")
	return footerStyle.Width(footerWidth).Render(footerText)
}

func (m Model) renderHelpModal() string {
	lines := []string{
		"Redis Cockpit 快捷键",
		"",
		"1-5 / Tab / Shift+Tab   切换页面",
		"↑↓ / j k                移动选择",
		"Enter                   确认 / 执行",
		"Esc                     返回 / 关闭",
		"?                       显示帮助",
		"Ctrl+C / Ctrl+Q         退出",
		"",
		"鼠标：点击顶部标签、左侧导航、按钮和列表项；滚轮滚动长结果。",
		"",
		"[Close]",
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

func renderConfirmBox(title, detail, yes, no string) string {
	body := strings.Join([]string{
		dangerStyle.Render(title),
		detail,
		warnStyle.Render(yes + "   " + no),
	}, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(red).
		Padding(0, 1).
		Render(body)
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
