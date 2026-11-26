package ui

import "github.com/charmbracelet/lipgloss"

var (
	accent  = lipgloss.Color("#00D4FF")
	violet  = lipgloss.Color("#7C3AED")
	blue    = lipgloss.Color("#2F81F7")
	bg      = lipgloss.Color("#0B1020")
	surface = lipgloss.Color("#111A2E")
	text    = lipgloss.Color("#E6EDF3")
	muted   = lipgloss.Color("#8B949E")
	green   = lipgloss.Color("#2EA043")
	red     = lipgloss.Color("#F85149")
	amber   = lipgloss.Color("#D29922")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(text).
			Background(violet).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(muted)

	navItemStyle = lipgloss.NewStyle().
			Foreground(muted).
			PaddingLeft(2)

	navActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			Background(surface).
			PaddingLeft(2)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blue).
			Background(bg).
			Foreground(text).
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(muted).
			Background(surface).
			Padding(0, 1)

	dangerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(red)

	okStyle = lipgloss.NewStyle().
		Foreground(green)

	warnStyle = lipgloss.NewStyle().
			Foreground(amber)
)
