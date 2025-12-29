package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/tui/styles"
)

// ProfileModel is the profile view model
type ProfileModel struct {
	width  int
	height int

	// Data
	user          *cfapi.User
	ratingHistory []cfapi.RatingChange
}

// NewProfileModel creates a new profile model
func NewProfileModel() ProfileModel {
	return ProfileModel{}
}

// SetSize sets the view dimensions
func (m *ProfileModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetUser sets the user data
func (m *ProfileModel) SetUser(user *cfapi.User) {
	m.user = user
}

// SetRatingHistory sets the rating history
func (m *ProfileModel) SetRatingHistory(history []cfapi.RatingChange) {
	m.ratingHistory = history
}

// Init initializes the model
func (m ProfileModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ProfileModel) Update(msg tea.Msg) (ProfileModel, tea.Cmd) {
	return m, nil
}

// View renders the profile view
func (m ProfileModel) View() string {
	if m.user == nil {
		return styles.SubtitleStyle.Render("  Loading profile...")
	}

	var sections []string

	// User header
	sections = append(sections, m.renderUserHeader())

	// Stats
	sections = append(sections, m.renderStats())

	// Rating graph (sparkline)
	sections = append(sections, m.renderRatingGraph())

	// Recent rating changes
	sections = append(sections, m.renderRatingHistory())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m ProfileModel) renderUserHeader() string {
	var b strings.Builder

	rankColor := styles.GetRankColor(m.user.Rating)
	rankStyle := lipgloss.NewStyle().Foreground(rankColor).Bold(true)

	// Handle and rank
	b.WriteString(styles.TitleStyle.Render(fmt.Sprintf("👤 %s", m.user.Handle)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("   %s\n", rankStyle.Render(m.user.Rank)))
	b.WriteString("\n")

	return b.String()
}

func (m ProfileModel) renderStats() string {
	cardWidth := (m.width - 10) / 4
	if cardWidth < 18 {
		cardWidth = 18
	}

	rankColor := styles.GetRankColor(m.user.Rating)

	// Rating card
	ratingCard := m.renderStatCard(
		"Rating",
		fmt.Sprintf("%d", m.user.Rating),
		rankColor,
		cardWidth,
	)

	// Max Rating card
	maxRatingCard := m.renderStatCard(
		"Max Rating",
		fmt.Sprintf("%d", m.user.MaxRating),
		styles.GetRankColor(m.user.MaxRating),
		cardWidth,
	)

	// Contribution card
	contribColor := styles.ColorSuccess
	if m.user.Contribution < 0 {
		contribColor = styles.ColorDanger
	}
	contribCard := m.renderStatCard(
		"Contribution",
		fmt.Sprintf("%+d", m.user.Contribution),
		contribColor,
		cardWidth,
	)

	// Friends card
	friendsCard := m.renderStatCard(
		"Friends",
		fmt.Sprintf("%d", m.user.FriendOfCount),
		styles.ColorPrimary,
		cardWidth,
	)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		ratingCard, "  ",
		maxRatingCard, "  ",
		contribCard, "  ",
		friendsCard,
	)
}

func (m ProfileModel) renderStatCard(label, value string, color lipgloss.Color, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Width(width - 4).
		Align(lipgloss.Center)

	valueStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Width(width - 4).
		Align(lipgloss.Center)

	content := lipgloss.JoinVertical(lipgloss.Center,
		labelStyle.Render(label),
		valueStyle.Render(value),
	)

	return styles.CardStyle.Width(width).Render(content)
}

func (m ProfileModel) renderRatingGraph() string {
	if len(m.ratingHistory) < 2 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.TitleStyle.Render("📈 Rating History"))
	b.WriteString("\n\n")

	// Simple ASCII sparkline
	width := m.width - 10
	if width > 80 {
		width = 80
	}

	history := m.ratingHistory
	if len(history) > width {
		// Sample to fit width
		step := len(history) / width
		sampled := make([]cfapi.RatingChange, width)
		for i := 0; i < width; i++ {
			sampled[i] = history[i*step]
		}
		history = sampled
	}

	// Find min/max
	minRating := history[0].NewRating
	maxRating := history[0].NewRating
	for _, rc := range history {
		if rc.NewRating < minRating {
			minRating = rc.NewRating
		}
		if rc.NewRating > maxRating {
			maxRating = rc.NewRating
		}
	}

	// Render sparkline
	sparkChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	ratingRange := maxRating - minRating
	if ratingRange == 0 {
		ratingRange = 1
	}

	var spark strings.Builder
	for _, rc := range history {
		normalized := float64(rc.NewRating-minRating) / float64(ratingRange)
		idx := int(normalized * 7)
		if idx > 7 {
			idx = 7
		}
		color := styles.GetRankColor(rc.NewRating)
		charStyle := lipgloss.NewStyle().Foreground(color)
		spark.WriteString(charStyle.Render(string(sparkChars[idx])))
	}

	b.WriteString("   ")
	b.WriteString(spark.String())
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("   %d%s%d\n",
		minRating,
		strings.Repeat(" ", len(history)-8),
		maxRating,
	))

	return b.String()
}

func (m ProfileModel) renderRatingHistory() string {
	if len(m.ratingHistory) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(styles.TitleStyle.Render("🏆 Recent Contests"))
	b.WriteString("\n\n")

	// Header
	header := fmt.Sprintf("  %-12s %-50s %5s → %5s  %s",
		"Date", "Contest", "Old", "New", "Delta",
	)
	b.WriteString(styles.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Show last 10 contests
	start := 0
	if len(m.ratingHistory) > 10 {
		start = len(m.ratingHistory) - 10
	}

	for i := start; i < len(m.ratingHistory); i++ {
		rc := m.ratingHistory[i]

		date := time.Unix(rc.RatingUpdateTimeSeconds, 0).Format("Jan 02 2006")
		contestName := rc.ContestName
		if len(contestName) > 48 {
			contestName = contestName[:45] + "..."
		}

		delta := rc.RatingDelta()
		deltaStr := fmt.Sprintf("%+d", delta)
		deltaColor := styles.ColorSuccess
		if delta < 0 {
			deltaColor = styles.ColorDanger
		}
		deltaStyle := lipgloss.NewStyle().Foreground(deltaColor).Bold(true)

		row := fmt.Sprintf("  %-12s %-50s %5d → %5d  %s",
			date,
			contestName,
			rc.OldRating,
			rc.NewRating,
			deltaStyle.Render(deltaStr),
		)

		if i%2 == 0 {
			b.WriteString(styles.TableRowStyle.Render(row))
		} else {
			b.WriteString(styles.TableRowAltStyle.Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}
