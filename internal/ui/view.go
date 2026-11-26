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
		label := m.btnLabel(fmt.Sprintf("[%d %s]", i+1, view.String()))
		if view == m.activeView {
			label = "▸" + label
		}
		tabs = append(tabs, button(label))
	}
	joined := lipgloss.JoinHorizontal(
		lipgloss.Left,
		titleStyle.Render(" Redis Cockpit "),
		"  ",
		strings.Join(tabs, " "),
	)
	return runewidth.Truncate(stripANSI(joined), maxInt(width-2, 1), "…")
}

func (m Model) renderNav() string {
	var b strings.Builder
	for i, view := range navItems {
		label := fmt.Sprintf("%d  %s", i+1, m.viewName(view))
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

func sectionSeparator(width int) string {
	return lipgloss.NewStyle().Foreground(blue).Render(strings.Repeat("─", maxInt(width, 20)))
}

func renderBottomHintLine(hint string, width int) string {
	hint = truncateToViewport(stripANSI(hint), 1, maxInt(width-2, 1))
	hint = hintStyle.Render(hint)
	hintWidth := runewidth.StringWidth(stripANSI(hint))
	return strings.Repeat(" ", maxInt(width-hintWidth, 0)) + hint
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
	case ViewSettings:
		body = m.renderSettingsBody()
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

	clampedHeight := maxInt(height, 2)
	clampedWidth := maxInt(width, 20)
	contentWidth := maxInt(clampedWidth-4, 10)
	bodyLines := strings.Split(truncateToViewport(body, maxInt(clampedHeight-1, 1), contentWidth), "\n")
	for len(bodyLines) < maxInt(clampedHeight-1, 1) {
		bodyLines = append(bodyLines, "")
	}
	hint := renderBottomHintLine(m.workspaceHintLine(), contentWidth)
	body = strings.Join(append(bodyLines, hint), "\n")
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
		return m.profileForm.ViewHeight(m.height - headerRows - footerRows - panelVerticalOverhead)
	}

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render(m.txt("Connections")),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		m.txt("Search:") + " " + m.profileSearch.View(),
		"",
		m.buttonRow("[Connect]", "[N new]", "[E edit]", "[D delete]", "[C clone]", "[I import]", "[X export]", "[T test]", "[Test all]", "[Search]", "[Sort]", "[Group]"),
		"",
		m.txt("Profiles:"),
	}

	if m.pendingProfileDelete != "" {
		lines = append(lines, m.renderConfirmBox(
			"Delete connection profile",
			m.pendingProfileDelete,
			"[Y] Delete",
			"[N] Cancel",
		), "")
	}

	if len(m.cfg.Profiles) == 0 {
		lines = append(lines, warnStyle.Render("  "+m.txt("No profiles configured")))
	} else if len(m.visibleProfileIndices()) == 0 {
		lines = append(lines, warnStyle.Render("  "+m.txt("No profiles match filter")))
	} else {
		lines = append(lines, m.renderProfileList())
	}

	if len(m.cfg.Profiles) > 0 {
		lines = append(lines, "", m.renderConnectionSummary())
		profile := m.cfg.Profiles[m.selectedProfile]
		lines = append(lines,
			"",
			subtitleStyle.Render(m.txt("Selected:")),
			"  "+m.txt("Name   :")+" "+profile.Name,
			"  "+m.txt("Addr   :")+" "+security.RedactText(profile.Addr),
			"  "+m.txt("DB     :")+" "+fmt.Sprint(profile.DB),
			"  "+m.txt("Auth   :")+" "+authLabel(profile.Username, profile.Password),
			"  "+m.txt("Mode   :")+" "+modeLabel(profile),
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
		m.commandInput.View() + "  " + m.buttonRow("[Input]", "[Execute]"),
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
			"  "+m.buttonRow("[Y] Execute", "[N] Cancel"),
		)
	} else {
		lines = append(lines, result.Render(m.lastResult, result.FormatRaw))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderKeysBody() string {
	if m.keyForm != nil {
		return m.keyForm.ViewHeight(m.height - headerRows - footerRows - panelVerticalOverhead)
	}
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Keys"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"Filter: " + m.filterInput.View(),
		"",
		m.buttonRow(keyViewToggleLabel(m), "[New key]", "[Edit value]", "[Del]", "[Next]", "[Refresh]"),
		"",
		"Keys:",
	}

	if len(m.keys) == 0 {
		lines = append(lines, subtitleStyle.Render("  (empty)"))
	} else {
		lines = append(lines, m.renderKeyRows())
	}

	separatorWidth := maxInt(m.width-navWidth-8, 20)
	lines = append(lines, "", sectionSeparator(separatorWidth), "", "Value:  "+m.button("[Copy value]"))
	if m.pendingDelete != nil {
		lines = append(lines, m.renderConfirmBox(
			"Confirm delete",
			strings.Join(m.pendingDelete, ", "),
			"[Y] Delete",
			"[N] Cancel",
		))
	} else {
		lines = append(lines, previewValue(m.selectedValue.Display(), 12, 1200))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderOpsBody() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Ops"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		m.buttonRow("[Refresh]"),
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
	)
	return strings.Join(lines, "\n")
}

func (m Model) renderStreamsBody() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render("Streams"),
		"",
		subtitleStyle.Render(m.workspaceHint()),
		"",
		"Channel: " + m.streamInput.View() + "  " + m.button("[Channel]"),
		"",
		m.buttonRow("[Sub]", "[Pub]", "[Monitor]", "[Stop]"),
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
		m.txt("Stream: 使用 Workbench 执行 XADD/XRANGE/XREAD 管理流条目。"),
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
		return m.txt("连接档案、认证、TLS、Cluster 与 Sentinel 将在此管理。")
	case ViewKeys:
		return m.txt("使用 SCAN 分页浏览 Key，查看类型、TTL 与值预览。")
	case ViewWorkbench:
		return m.txt("搜索命令、填写参数或执行原始 Redis 命令。")
	case ViewOps:
		return m.txt("INFO、CONFIG、CLIENT、SLOWLOG 与复制状态。")
	case ViewStreams:
		return m.txt("Pub/Sub、MONITOR 与 Stream 实时消息流。")
	case ViewSettings:
		return m.txt("设置语言、主题和查看快捷键帮助。")
	default:
		return ""
	}
}

func (m Model) workspaceHintLine() string {
	switch m.activeView {
	case ViewConnections:
		return m.txt("↑↓ 选  ↵ 连  n 增  e 改  d 删  t 测")
	case ViewKeys:
		return m.txt("↑↓ 选/读  a 增  v 改  d 删  t 树  / 筛  r 刷")
	case ViewWorkbench:
		return m.txt("↵ 执行  ↑↓ 历史  ? 设置")
	case ViewOps:
		return m.txt("r 刷新")
	case ViewStreams:
		return m.txt("s 订阅  p 发布  m 监控  x 停止")
	case ViewSettings:
		return m.txt("点击切换  ? 帮助")
	default:
		return ""
	}
}

func (m Model) renderFooter(width int) string {
	footerWidth := maxInt(width, 80)
	shortcuts := m.txt("鼠标点击  •  Tab/1-6 切换  •  Enter 确认  •  ? 设置  •  Ctrl+C 退出")
	statusText := strings.Join(strings.Fields(stripANSI(m.status)), " ")
	shortcuts = strings.Join(strings.Fields(stripANSI(shortcuts)), " ")

	contentWidth := maxInt(footerWidth-2, 10)
	statusWidth := maxInt(contentWidth/3, 20)
	shortcutWidth := maxInt(contentWidth-statusWidth-1, 10)
	statusText = runewidth.Truncate(statusText, statusWidth, "…")
	shortcuts = runewidth.Truncate(shortcuts, shortcutWidth, "…")

	left := warnStyle.Background(surface).Render(padToDisplayWidth(statusText, statusWidth))
	rightPad := strings.Repeat(" ", maxInt(shortcutWidth-runewidth.StringWidth(shortcuts), 0))
	right := subtitleStyle.Render(rightPad + shortcuts)
	return footerStyle.Width(footerWidth).Render(left + " " + right)
}

func (m Model) renderHelpModal() string {
	lines := []string{
		m.txt("Redis Cockpit 快捷键"),
		"",
		m.txt("1-6 / Tab / Shift+Tab   切换页面"),
		m.txt("↑↓ / j k                移动选择"),
		m.txt("Enter                   确认 / 执行"),
		m.txt("Esc                     返回 / 关闭"),
		m.txt("Ctrl+C / Ctrl+Q         退出"),
		"",
		m.txt("鼠标：点击顶部标签、左侧导航、按钮和列表项；滚轮滚动长结果。"),
		"",
		m.button("[Close]"),
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

func (m Model) renderSettingsBody() string {
	languageButtons := []string{m.button("[中文]"), m.button("[English]")}
	if m.language == "zh" {
		languageButtons[0] = m.activeButton("▸[中文]")
	} else {
		languageButtons[1] = m.activeButton("▸[English]")
	}
	themeButtons := make([]string, 0, 4)
	for _, name := range []string{themeAurora, themeEmber, themeOcean, themeGraphite} {
		label := "[" + strings.ToUpper(name[:1]) + name[1:] + "]"
		if m.cfg.Theme == name {
			themeButtons = append(themeButtons, m.activeButton("▸"+label))
		} else {
			themeButtons = append(themeButtons, m.button(label))
		}
	}

	return strings.Join([]string{
		lipgloss.NewStyle().Bold(true).Foreground(accent).Render(m.txt("Settings")),
		"",
		m.txt("Language") + ": " + strings.Join(languageButtons, " "),
		m.txt("当前语言") + ": " + map[string]string{"zh": m.txt("中文"), "en": m.txt("英文")}[m.language],
		"",
		m.txt("Theme") + ": " + strings.Join(themeButtons, " "),
		m.txt("当前主题") + ": " + strings.ToUpper(m.cfg.Theme),
		"",
		m.txt("帮助") + " / " + m.txt("Shortcuts"),
		m.txt("1-6 / Tab / Shift+Tab   切换页面"),
		m.txt("↑↓ / j k                移动选择"),
		m.txt("Enter                   确认 / 执行"),
		m.txt("Esc                     返回 / 关闭"),
		m.txt("Ctrl+C / Ctrl+Q         退出"),
		"",
		m.txt("鼠标：点击顶部标签、左侧导航、按钮和列表项；滚轮滚动长结果。"),
	}, "\n")
}

func (m Model) renderConfirmBox(title, detail, yes, no string) string {
	body := strings.Join([]string{
		dangerStyle.Render(title),
		detail,
		m.button(yes) + "   " + m.button(no),
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
