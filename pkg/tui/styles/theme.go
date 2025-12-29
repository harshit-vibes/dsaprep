// Package styles provides theming and styling for the TUI
package styles

import "github.com/charmbracelet/lipgloss"

// Codeforces rank colors
var (
	// Rank colors (official CF colors)
	ColorNewbie       = lipgloss.Color("#808080") // Gray
	ColorPupil        = lipgloss.Color("#008000") // Green
	ColorSpecialist   = lipgloss.Color("#03A89E") // Cyan
	ColorExpert       = lipgloss.Color("#0000FF") // Blue
	ColorCandidateMaster = lipgloss.Color("#AA00AA") // Violet
	ColorMaster       = lipgloss.Color("#FF8C00") // Orange
	ColorIntlMaster   = lipgloss.Color("#FF8C00") // Orange
	ColorGrandmaster  = lipgloss.Color("#FF0000") // Red
	ColorIntlGM       = lipgloss.Color("#FF0000") // Red
	ColorLegendaryGM  = lipgloss.Color("#FF0000") // Red (with black first letter)

	// Verdict colors
	ColorAccepted     = lipgloss.Color("#27AE60") // Green
	ColorWrongAnswer  = lipgloss.Color("#E74C3C") // Red
	ColorTLE          = lipgloss.Color("#F39C12") // Yellow/Orange
	ColorMLE          = lipgloss.Color("#F39C12") // Yellow/Orange
	ColorRuntimeError = lipgloss.Color("#E74C3C") // Red
	ColorCompileError = lipgloss.Color("#9B59B6") // Purple
	ColorPending      = lipgloss.Color("#95A5A6") // Gray

	// UI Colors
	ColorPrimary    = lipgloss.Color("#3498DB") // Blue
	ColorSecondary  = lipgloss.Color("#2ECC71") // Green
	ColorAccent     = lipgloss.Color("#9B59B6") // Purple
	ColorWarning    = lipgloss.Color("#F39C12") // Orange
	ColorDanger     = lipgloss.Color("#E74C3C") // Red
	ColorSuccess    = lipgloss.Color("#27AE60") // Green
	ColorMuted      = lipgloss.Color("#95A5A6") // Gray
	ColorSubtle     = lipgloss.Color("#7F8C8D") // Darker gray

	// Background colors
	ColorBgDark       = lipgloss.Color("#1A1A2E") // Dark blue-gray
	ColorBgLight      = lipgloss.Color("#16213E") // Slightly lighter
	ColorBgHighlight  = lipgloss.Color("#0F3460") // Highlight background
	ColorBgSelected   = lipgloss.Color("#E94560") // Selected item

	// Text colors
	ColorTextPrimary   = lipgloss.Color("#FFFFFF")
	ColorTextSecondary = lipgloss.Color("#B0B0B0")
	ColorTextMuted     = lipgloss.Color("#666666")
)

// GetRankColor returns the appropriate color for a CF rating
func GetRankColor(rating int) lipgloss.Color {
	switch {
	case rating >= 3000:
		return ColorLegendaryGM
	case rating >= 2600:
		return ColorIntlGM
	case rating >= 2400:
		return ColorGrandmaster
	case rating >= 2300:
		return ColorIntlMaster
	case rating >= 2100:
		return ColorMaster
	case rating >= 1900:
		return ColorCandidateMaster
	case rating >= 1600:
		return ColorExpert
	case rating >= 1400:
		return ColorSpecialist
	case rating >= 1200:
		return ColorPupil
	default:
		return ColorNewbie
	}
}

// GetRankName returns the rank name for a rating
func GetRankName(rating int) string {
	switch {
	case rating >= 3000:
		return "Legendary Grandmaster"
	case rating >= 2600:
		return "International Grandmaster"
	case rating >= 2400:
		return "Grandmaster"
	case rating >= 2300:
		return "International Master"
	case rating >= 2100:
		return "Master"
	case rating >= 1900:
		return "Candidate Master"
	case rating >= 1600:
		return "Expert"
	case rating >= 1400:
		return "Specialist"
	case rating >= 1200:
		return "Pupil"
	default:
		return "Newbie"
	}
}

// GetVerdictColor returns the color for a verdict
func GetVerdictColor(verdict string) lipgloss.Color {
	switch verdict {
	case "OK":
		return ColorAccepted
	case "WRONG_ANSWER":
		return ColorWrongAnswer
	case "TIME_LIMIT_EXCEEDED":
		return ColorTLE
	case "MEMORY_LIMIT_EXCEEDED":
		return ColorMLE
	case "RUNTIME_ERROR":
		return ColorRuntimeError
	case "COMPILATION_ERROR":
		return ColorCompileError
	default:
		return ColorPending
	}
}

// GetVerdictShort returns a short form of the verdict
func GetVerdictShort(verdict string) string {
	switch verdict {
	case "OK":
		return "AC"
	case "WRONG_ANSWER":
		return "WA"
	case "TIME_LIMIT_EXCEEDED":
		return "TLE"
	case "MEMORY_LIMIT_EXCEEDED":
		return "MLE"
	case "RUNTIME_ERROR":
		return "RE"
	case "COMPILATION_ERROR":
		return "CE"
	case "TESTING":
		return "..."
	default:
		return verdict
	}
}
