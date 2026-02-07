package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name       string
	Accent     string
	AccentDark string
	Ink        string
	Muted      string
	Background string
}

var Themes = []Theme{
	{
		Name:       "Ocean",
		Accent:     "#34B3A0",
		AccentDark: "#0F2E2B",
		Ink:        "#E6EDF3",
		Muted:      "#8AA1A8",
		Background: "#0B1115",
	},
	{
		Name:       "Sand",
		Accent:     "#D7A86E",
		AccentDark: "#332819",
		Ink:        "#F2E8D5",
		Muted:      "#B8A387",
		Background: "#1A140D",
	},
	{
		Name:       "Day",
		Accent:     "#3B82F6",
		AccentDark: "#E6EEF9",
		Ink:        "#0B1220",
		Muted:      "#506072",
		Background: "#F7FAFF",
	},
}

type Styles struct {
	Header      lipgloss.Style
	ActiveTab   lipgloss.Style
	InactiveTab lipgloss.Style
	DisabledTab lipgloss.Style
	Footer      lipgloss.Style
	Summary     lipgloss.Style
	Info        lipgloss.Style
	ContentBox  lipgloss.Style
	Overflow    lipgloss.Style
	Accent      lipgloss.Color
	AccentDark  lipgloss.Color
	Ink         lipgloss.Color
	Muted       lipgloss.Color
	Background  lipgloss.Color
}

func BuildStyles(index int) Styles {
	if index < 0 || index >= len(Themes) {
		index = 0
	}
	t := Themes[index]
	
	s := Styles{}
	s.Accent = lipgloss.Color(t.Accent)
	s.AccentDark = lipgloss.Color(t.AccentDark)
	s.Ink = lipgloss.Color(t.Ink)
	s.Muted = lipgloss.Color(t.Muted)
	s.Background = lipgloss.Color(t.Background)

	s.Header = lipgloss.NewStyle().Foreground(s.Ink).Background(s.Background).Padding(0, 1)
	s.ActiveTab = lipgloss.NewStyle().Foreground(s.Background).Background(s.Accent).Bold(true).Padding(0, 1)
	s.InactiveTab = lipgloss.NewStyle().Foreground(s.Muted).Background(s.Background).Padding(0, 1)
	s.DisabledTab = lipgloss.NewStyle().Foreground(s.Muted).Background(s.Background).Faint(true).Padding(0, 1)
	s.Footer = lipgloss.NewStyle().Foreground(s.Muted).Background(s.Background).Padding(0, 1)
	s.Summary = lipgloss.NewStyle().Foreground(s.Ink).Background(s.AccentDark).Padding(0, 1)
	s.Info = lipgloss.NewStyle().Foreground(s.Ink).Background(s.Background).Padding(0, 1)
	s.ContentBox = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(s.Muted).
		Padding(0, 1)
	s.Overflow = lipgloss.NewStyle().Foreground(s.Muted).Background(s.Background).Padding(0, 1)

	return s
}
