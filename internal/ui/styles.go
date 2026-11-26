package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var bracketButtonPattern = regexp.MustCompile(`\[[^\]\n]+\]`)

type themePalette struct {
	accent   lipgloss.Color
	violet   lipgloss.Color
	blue     lipgloss.Color
	bg       lipgloss.Color
	surface  lipgloss.Color
	text     lipgloss.Color
	muted    lipgloss.Color
	green    lipgloss.Color
	red      lipgloss.Color
	amber    lipgloss.Color
	buttonBG lipgloss.Color
	buttonFG lipgloss.Color
}

const (
	themeAurora   = "aurora"
	themeEmber    = "ember"
	themeOcean    = "ocean"
	themeGraphite = "graphite"
)

var (
	accent  lipgloss.Color
	violet  lipgloss.Color
	blue    lipgloss.Color
	bg      lipgloss.Color
	surface lipgloss.Color
	text    lipgloss.Color
	muted   lipgloss.Color
	green   lipgloss.Color
	red     lipgloss.Color
	amber   lipgloss.Color

	titleStyle        lipgloss.Style
	subtitleStyle     lipgloss.Style
	navItemStyle      lipgloss.Style
	navActiveStyle    lipgloss.Style
	panelStyle        lipgloss.Style
	footerStyle       lipgloss.Style
	dangerStyle       lipgloss.Style
	okStyle           lipgloss.Style
	warnStyle         lipgloss.Style
	buttonStyle       lipgloss.Style
	activeButtonStyle lipgloss.Style
)

func themePaletteFor(name string) themePalette {
	switch strings.ToLower(name) {
	case themeEmber:
		return themePalette{
			accent: "#FFB454", violet: "#C2410C", blue: "#F97316",
			bg: "#160F0B", surface: "#24170F", text: "#FFF4E6", muted: "#C9A882",
			green: "#7FBF5A", red: "#FF6B5E", amber: "#FFC857",
			buttonBG: "#FFB454", buttonFG: "#24170F",
		}
	case themeOcean:
		return themePalette{
			accent: "#38BDF8", violet: "#0EA5E9", blue: "#2563EB",
			bg: "#071426", surface: "#0E2238", text: "#E6F6FF", muted: "#8FB3C9",
			green: "#34D399", red: "#FB7185", amber: "#FBBF24",
			buttonBG: "#38BDF8", buttonFG: "#071426",
		}
	case themeGraphite:
		return themePalette{
			accent: "#D4D4D8", violet: "#71717A", blue: "#A1A1AA",
			bg: "#111111", surface: "#1C1C1F", text: "#FAFAFA", muted: "#A1A1AA",
			green: "#86EFAC", red: "#FCA5A5", amber: "#FDE68A",
			buttonBG: "#D4D4D8", buttonFG: "#111111",
		}
	default:
		return themePalette{
			accent: "#00D4FF", violet: "#7C3AED", blue: "#2F81F7",
			bg: "#0B1020", surface: "#111A2E", text: "#E6EDF3", muted: "#8B949E",
			green: "#2EA043", red: "#F85149", amber: "#D29922",
			buttonBG: "#00D4FF", buttonFG: "#0B1020",
		}
	}
}

func applyTheme(name string) {
	p := themePaletteFor(name)
	accent, violet, blue = p.accent, p.violet, p.blue
	bg, surface, text, muted = p.bg, p.surface, p.text, p.muted
	green, red, amber = p.green, p.red, p.amber

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(text).Background(violet).Padding(0, 1)
	subtitleStyle = lipgloss.NewStyle().Foreground(muted)
	navItemStyle = lipgloss.NewStyle().Foreground(muted).PaddingLeft(2)
	navActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(accent).Background(surface).PaddingLeft(2)
	panelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(blue).Background(bg).Foreground(text).Padding(1, 2)
	footerStyle = lipgloss.NewStyle().Foreground(muted).Background(surface).Padding(0, 1)
	dangerStyle = lipgloss.NewStyle().Bold(true).Foreground(red)
	okStyle = lipgloss.NewStyle().Foreground(green)
	warnStyle = lipgloss.NewStyle().Foreground(amber)
	buttonStyle = lipgloss.NewStyle().Bold(true).Foreground(p.buttonFG).Background(p.buttonBG).Padding(0, 1)
	activeButtonStyle = lipgloss.NewStyle().Bold(true).Foreground(p.buttonBG).Background(violet).Padding(0, 1)
}

func padToDisplayWidth(value string, width int) string {
	current := runewidth.StringWidth(value)
	if current >= width {
		return value
	}
	return value + strings.Repeat(" ", width-current)
}

func button(label string) string {
	return buttonStyle.Render(label)
}

func activeButton(label string) string {
	return activeButtonStyle.Render(label)
}

func styleButtons(line string) string {
	return bracketButtonPattern.ReplaceAllStringFunc(line, button)
}

func styleButtonsForLanguage(language, line string) string {
	return bracketButtonPattern.ReplaceAllStringFunc(line, func(label string) string {
		return button(buttonLabel(language, label))
	})
}

func buttonRow(labels ...string) string {
	parts := make([]string, 0, len(labels))
	for _, label := range labels {
		parts = append(parts, button(label))
	}
	return strings.Join(parts, " ")
}

func init() {
	applyTheme(themeAurora)
}
