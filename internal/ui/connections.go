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
}

func (m Model) renderConnectionSummary() string {
	server := m.opsSnapshot.Info["Server"]
	memory := m.opsSnapshot.Info["Memory"]
	stats := m.opsSnapshot.Info["Stats"]

	panelWidth := maxInt((m.width-navWidth-38)/3, 10)
	panels := []string{
		renderMetricPanel("Server", [][2]string{
			{"Version", metricValue(server, "redis_version")},
			{"OS", metricValue(server, "os")},
			{"PID", metricValue(server, "process_id")},
		}, panelWidth),
		renderMetricPanel("Memory", [][2]string{
			{"Used", metricValue(memory, "used_memory_human")},
			{"Peak", metricValue(memory, "used_memory_peak_human")},
			{"Lua", metricValue(memory, "used_memory_lua_human")},
		}, panelWidth),
		renderMetricPanel("Stats", [][2]string{
			{"Clients", metricValue(stats, "connected_clients")},
			{"Connections", metricValue(stats, "total_connections_received")},
			{"Commands", metricValue(stats, "total_commands_processed")},
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
		return "  FAIL"
	}
	version := ""
	if health.Version != "" && health.Version != "unknown" {
		version = " v" + health.Version
	}
	return fmt.Sprintf("  OK %dms%s", health.Latency.Milliseconds(), version)
}
