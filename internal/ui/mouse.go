package ui

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func stripANSI(value string) string {
	return ansiPattern.ReplaceAllString(value, "")
}

func locateToken(view, token string) (int, int, bool) {
	for y, line := range strings.Split(stripANSI(view), "\n") {
		if idx := strings.Index(line, token); idx >= 0 {
			return runewidth.StringWidth(line[:idx]), y, true
		}
	}
	return 0, 0, false
}

func locateLineContaining(view, needle string) (int, int, int, bool) {
	for y, line := range strings.Split(stripANSI(view), "\n") {
		if idx := strings.Index(line, needle); idx >= 0 {
			return runewidth.StringWidth(line[:idx]), y, runewidth.StringWidth(line), true
		}
	}
	return 0, 0, 0, false
}

func clickToken(msg tea.MouseMsg, view, token string) bool {
	x, y, ok := locateToken(view, token)
	if !ok || msg.Y != y {
		return false
	}
	width := runewidth.StringWidth(token)
	return msg.X >= x && msg.X < x+width
}

func (m Model) clickAction(msg tea.MouseMsg, view, canonical string) bool {
	return clickToken(msg, view, m.btnLabel(canonical)) || clickToken(msg, view, canonical)
}

func clickLineContaining(msg tea.MouseMsg, view, needle string) bool {
	x, y, width, ok := locateLineContaining(view, needle)
	if !ok || msg.Y != y {
		return false
	}
	return msg.X >= x && msg.X < x+width
}
