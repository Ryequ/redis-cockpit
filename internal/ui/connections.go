package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/redis"
)

type ProfileHealth struct {
	Status    string
	Latency   time.Duration
	Version   string
	Error     string
	CheckedAt time.Time
}

type profileSortKey int

const (
	profileSortDefault profileSortKey = iota
	profileSortName
	profileSortAddr
	profileSortMode
)

func profileSortLabel(sortKey profileSortKey) string {
	switch sortKey {
	case profileSortName:
		return "name"
	case profileSortAddr:
		return "address"
	case profileSortMode:
		return "mode"
	default:
		return "config order"
	}
}

func (m Model) profileModeKey(profile redis.ConnectionProfile) string {
	mode := "single"
	if profile.Cluster {
		mode = "cluster"
	} else if profile.Sentinel {
		mode = "sentinel"
	}
	mode = m.txt(mode)
	if profile.TLS {
		mode += "+TLS"
	}
	return mode
}

func (m Model) visibleProfileIndices() []int {
	indices := make([]int, 0, len(m.cfg.Profiles))
	for i := range m.cfg.Profiles {
		indices = append(indices, i)
	}

	query := strings.ToLower(strings.TrimSpace(m.profileSearch.Value()))
	if query != "" {
		filtered := indices[:0]
		for _, i := range indices {
			profile := m.cfg.Profiles[i]
			haystack := strings.ToLower(strings.Join([]string{
				profile.Name,
				profile.ID,
				profile.Addr,
				m.profileModeKey(profile),
			}, " "))
			if strings.Contains(haystack, query) {
				filtered = append(filtered, i)
			}
		}
		indices = filtered
	}

	switch m.profileSort {
	case profileSortName:
		sort.SliceStable(indices, func(i, j int) bool {
			return m.cfg.Profiles[indices[i]].Name < m.cfg.Profiles[indices[j]].Name
		})
	case profileSortAddr:
		sort.SliceStable(indices, func(i, j int) bool {
			return m.cfg.Profiles[indices[i]].Addr < m.cfg.Profiles[indices[j]].Addr
		})
	case profileSortMode:
		sort.SliceStable(indices, func(i, j int) bool {
			left := m.profileModeKey(m.cfg.Profiles[indices[i]])
			right := m.profileModeKey(m.cfg.Profiles[indices[j]])
			if left == right {
				return m.cfg.Profiles[indices[i]].Name < m.cfg.Profiles[indices[j]].Name
			}
			return left < right
		})
	}
	if m.profileGroupByMode && m.profileSort != profileSortMode {
		sort.SliceStable(indices, func(i, j int) bool {
			left := m.profileModeKey(m.cfg.Profiles[indices[i]])
			right := m.profileModeKey(m.cfg.Profiles[indices[j]])
			if left == right {
				return m.cfg.Profiles[indices[i]].Name < m.cfg.Profiles[indices[j]].Name
			}
			return left < right
		})
	}
	return indices
}

type profileListRow struct {
	text         string
	profileIndex int
	group        bool
}

func (m Model) profileListHeight() int {
	reserved := headerRows + footerRows + panelVerticalOverhead + 18
	return maxInt(m.height-reserved, 3)
}

func (m Model) profileListRows() []profileListRow {
	indices := m.visibleProfileIndices()
	rows := make([]profileListRow, 0, len(indices))
	lastGroup := ""
	for _, idx := range indices {
		profile := m.cfg.Profiles[idx]
		group := m.profileModeKey(profile)
		if m.profileGroupByMode && group != lastGroup {
			rows = append(rows, profileListRow{
				text:  m.txt("Group: ") + group,
				group: true,
			})
			lastGroup = group
		}
		marker := "  "
		if idx == m.selectedProfile {
			marker = okStyle.Render("▸ ")
		}
		active := ""
		if profile.Name == m.cfg.ActiveProfile || profile.ID == m.cfg.ActiveProfile {
			active = m.txt("  active")
		}
		rows = append(rows, profileListRow{
			text:         fmt.Sprintf("%s%d. %s  %s  [%s]%s%s", marker, idx+1, profile.Name, profile.Addr, group, active, m.healthLabel(profile.Name)),
			profileIndex: idx,
		})
	}
	return rows
}

func (m Model) renderProfileList() string {
	rows := m.profileListRows()
	height := m.profileListHeight()
	start := maxInt(m.profileScroll, 0)
	maxStart := maxInt(len(rows)-height, 0)
	if start > maxStart {
		start = maxStart
	}
	end := minInt(start+height, len(rows))

	lines := make([]string, 0, end-start+2)
	if start > 0 {
		lines = append(lines, subtitleStyle.Render("  ▲ more above"))
	}
	for _, row := range rows[start:end] {
		if row.group {
			lines = append(lines, subtitleStyle.Render(row.text))
			continue
		}
		lines = append(lines, row.text)
	}
	if end < len(rows) {
		lines = append(lines, subtitleStyle.Render("  ▼ more below"))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) ensureProfileVisible() {
	rows := m.profileListRows()
	position := -1
	for i, row := range rows {
		if row.profileIndex == m.selectedProfile {
			position = i
			break
		}
	}
	if position < 0 {
		return
	}
	height := m.profileListHeight()
	maxScroll := maxInt(len(rows)-height, 0)
	if m.profileScroll > maxScroll {
		m.profileScroll = maxScroll
	}
	if position < m.profileScroll {
		m.profileScroll = position
		return
	}
	if position >= m.profileScroll+height {
		m.profileScroll = position - height + 1
	}
}

func (m *Model) scrollProfileList(delta int) {
	maxScroll := maxInt(len(m.profileListRows())-m.profileListHeight(), 0)
	m.profileScroll += delta
	if m.profileScroll < 0 {
		m.profileScroll = 0
	}
	if m.profileScroll > maxScroll {
		m.profileScroll = maxScroll
	}
}

func uniqueProfileName(cfg config.Config, base string) string {
	name := base + "-copy"
	for suffix := 2; ; suffix++ {
		if _, exists := cfg.FindProfile(name); !exists {
			return name
		}
		name = fmt.Sprintf("%s-copy-%d", base, suffix)
	}
}

func (m *Model) normalizeProfileSelection() {
	indices := m.visibleProfileIndices()
	for _, idx := range indices {
		if idx == m.selectedProfile {
			return
		}
	}
	if len(indices) > 0 {
		m.selectedProfile = indices[0]
		m.ensureProfileVisible()
		return
	}
	m.selectedProfile = 0
}

func (m *Model) selectPreviousProfile() {
	indices := m.visibleProfileIndices()
	position := -1
	for i, idx := range indices {
		if idx == m.selectedProfile {
			position = i
			break
		}
	}
	if position <= 0 {
		return
	}
	m.selectedProfile = indices[position-1]
	m.ensureProfileVisible()
}

func (m *Model) selectNextProfile() {
	indices := m.visibleProfileIndices()
	position := -1
	for i, idx := range indices {
		if idx == m.selectedProfile {
			position = i
			break
		}
	}
	if position < 0 || position >= len(indices)-1 {
		return
	}
	m.selectedProfile = indices[position+1]
	m.ensureProfileVisible()
}

func (m Model) renderConnectionSummary() string {
	server := m.opsSnapshot.Info["Server"]
	memory := m.opsSnapshot.Info["Memory"]
	clients := m.opsSnapshot.Info["Clients"]
	stats := m.opsSnapshot.Info["Stats"]

	panelWidth := maxInt((m.width-navWidth-18)/3, 10)
	panels := []string{
		renderMetricPanel(m.txt("Server"), [][2]string{
			{m.txt("Version"), metricValue(server, "redis_version")},
			{m.txt("OS"), metricValue(server, "os")},
			{m.txt("PID"), metricValue(server, "process_id")},
		}, panelWidth),
		renderMetricPanel(m.txt("Memory"), [][2]string{
			{m.txt("Used"), metricValue(memory, "used_memory_human")},
			{m.txt("Peak"), metricValue(memory, "used_memory_peak_human")},
			{m.txt("Lua"), metricValue(memory, "used_memory_lua_human")},
		}, panelWidth),
		renderMetricPanel(m.txt("Stats"), [][2]string{
			{m.txt("Clients"), metricValue(clients, "connected_clients")},
			{m.txt("Metric Connections"), metricValue(stats, "total_connections_received")},
			{m.txt("Commands"), metricValue(stats, "total_commands_processed")},
		}, panelWidth),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

func renderMetricPanel(title string, rows [][2]string, width int) string {
	lines := []string{lipgloss.NewStyle().Bold(true).Foreground(accent).Render(title)}
	for _, row := range rows {
		line := fmt.Sprintf("%-10s %s", row[0]+":", row[1])
		lines = append(lines, runewidth.Truncate(line, maxInt(width-4, 1), "…"))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(blue).
		Padding(0, 1).
		Width(width).
		Height(4).
		Render(strings.Join(lines, "\n"))
}

func metricValue(values map[string]string, key string) string {
	if value := strings.TrimSpace(values[key]); value != "" {
		return value
	}
	return "-"
}

func (m Model) healthLabel(name string) string {
	health, ok := m.profileHealth[name]
	if !ok {
		return ""
	}
	if health.Status != "OK" {
		return "  " + m.txt("FAIL")
	}
	version := ""
	if health.Version != "" && health.Version != "unknown" {
		version = " v" + health.Version
	}
	return fmt.Sprintf("  %s %dms%s", m.txt("OK"), health.Latency.Milliseconds(), version)
}
