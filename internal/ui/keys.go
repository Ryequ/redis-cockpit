package ui

import tea "github.com/charmbracelet/bubbletea"

var navItems = []ViewID{
	ViewConnections,
	ViewKeys,
	ViewWorkbench,
	ViewOps,
	ViewStreams,
	ViewSettings,
}

func viewFromNumber(r rune) (ViewID, bool) {
	if r < '1' || r > '6' {
		return ViewConnections, false
	}
	idx := int(r - '1')
	if idx < 0 || idx >= len(navItems) {
		return ViewConnections, false
	}
	return navItems[idx], true
}

func isQuitKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyCtrlQ
}

func isCycleForward(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyTab
}

func isCycleBackward(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyShiftTab
}
