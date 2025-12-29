package styles

import "github.com/charmbracelet/lipgloss"

// App dimensions
const (
	DefaultWidth  = 120
	DefaultHeight = 40
	HeaderHeight  = 3
	FooterHeight  = 2
	TabHeight     = 1
)

// Base styles
var (
	// App container
	AppStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTextPrimary).
			Background(ColorBgHighlight).
			Padding(0, 2).
			MarginBottom(1)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	// Tab bar
	TabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorSubtle).
			MarginBottom(1)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTextPrimary).
			Background(ColorPrimary).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorTextSecondary).
			Padding(0, 2)

	// Footer
	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			MarginTop(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// Content area
	ContentStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Cards and boxes
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(1, 2)

	SelectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(1, 2)

	// List items
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorTextPrimary).
				Background(ColorBgSelected).
				Bold(true).
				PaddingLeft(2)

	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTextPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorTextSecondary).
			Italic(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary)

	// Status indicators
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorTextSecondary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorSubtle)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary)

	TableRowAltStyle = lipgloss.NewStyle().
				Foreground(ColorTextPrimary).
				Background(lipgloss.Color("#1E1E2E"))

	// Progress bar
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorSubtle)

	// Badges
	BadgeStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Background(ColorPrimary).
			Foreground(ColorTextPrimary)

	BadgeSuccessStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(ColorSuccess).
				Foreground(ColorTextPrimary)

	BadgeWarningStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(ColorWarning).
				Foreground(ColorTextPrimary)

	BadgeDangerStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(ColorDanger).
				Foreground(ColorTextPrimary)

	// Spinner
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	// Dialog
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Width(60)
)

// Helper functions

// RenderRating renders a rating with appropriate color
func RenderRating(rating int) string {
	color := GetRankColor(rating)
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	return style.Render(lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Render(
		lipgloss.NewStyle().Render(string(rune('0'+rating/1000)) + string(rune('0'+(rating/100)%10)) + string(rune('0'+(rating/10)%10)) + string(rune('0'+rating%10))),
	))
}

// RenderVerdict renders a verdict with appropriate color
func RenderVerdict(verdict string) string {
	color := GetVerdictColor(verdict)
	short := GetVerdictShort(verdict)
	return lipgloss.NewStyle().Foreground(color).Bold(true).Width(4).Render(short)
}

// RenderProgressBar renders a simple progress bar
func RenderProgressBar(percent float64, width int) string {
	filled := int(float64(width) * percent)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += ProgressBarStyle.Render("█")
		} else {
			bar += ProgressEmptyStyle.Render("░")
		}
	}
	return bar
}

// Truncate truncates a string to a max length with ellipsis
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Pad pads a string to a specific width
func Pad(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Render(s)
}

// PadRight pads a string to the right
func PadRight(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Left).Render(s)
}

// PadLeft pads a string to the left
func PadLeft(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(s)
}

// Center centers a string
func Center(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
}
