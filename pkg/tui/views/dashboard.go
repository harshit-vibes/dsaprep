// Package views provides TUI view models
package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/tui/styles"
)

// DashboardModel is the dashboard view model
type DashboardModel struct {
	width  int
	height int

	// Data
	user        *cfapi.User
	submissions []cfapi.Submission

	// Stats
	totalSolved  int
	recentSolved int
	streak       int
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() DashboardModel {
	return DashboardModel{}
}

// SetSize sets the view dimensions
func (m *DashboardModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetUser sets the user data
func (m *DashboardModel) SetUser(user *cfapi.User) {
	m.user = user
}

// SetSubmissions sets the submissions data
func (m *DashboardModel) SetSubmissions(submissions []cfapi.Submission) {
	m.submissions = submissions
	m.calculateStats()
}

// SetStats sets the statistics
func (m *DashboardModel) SetStats(totalSolved, recentSolved, streak int) {
	m.totalSolved = totalSolved
	m.recentSolved = recentSolved
	m.streak = streak
}

func (m *DashboardModel) calculateStats() {
	seen := make(map[string]bool)
	for _, s := range m.submissions {
		if s.IsAccepted() {
			key := s.Problem.ProblemID()
			if !seen[key] {
				seen[key] = true
				m.totalSolved++
			}
		}
	}
}

// Init initializes the model
func (m DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	var sections []string

	// Welcome section
	sections = append(sections, m.renderWelcome())

	// Stats cards
	sections = append(sections, m.renderStatsCards())

	// Recent activity
	sections = append(sections, m.renderRecentActivity())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m DashboardModel) renderWelcome() string {
	var b strings.Builder

	if m.user != nil {
		rankColor := styles.GetRankColor(m.user.Rating)
		rankStyle := lipgloss.NewStyle().Foreground(rankColor).Bold(true)

		b.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Welcome back, %s!", m.user.Handle)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s • Rating: %s",
			rankStyle.Render(m.user.Rank),
			rankStyle.Render(fmt.Sprintf("%d", m.user.Rating)),
		))
	} else {
		b.WriteString(styles.TitleStyle.Render("Welcome to cf!"))
		b.WriteString("\n")
		b.WriteString(styles.SubtitleStyle.Render("  Configure your handle with 'cf config set cf_handle <handle>'"))
	}

	return b.String()
}

func (m DashboardModel) renderStatsCards() string {
	cardWidth := (m.width - 10) / 3
	if cardWidth < 20 {
		cardWidth = 20
	}

	// Solved card
	solvedCard := m.renderCard(
		"📊 Problems Solved",
		fmt.Sprintf("%d", m.totalSolved),
		"unique problems",
		styles.ColorSuccess,
		cardWidth,
	)

	// Streak card
	streakText := "days"
	if m.streak == 1 {
		streakText = "day"
	}
	streakCard := m.renderCard(
		"🔥 Current Streak",
		fmt.Sprintf("%d", m.streak),
		streakText,
		styles.ColorWarning,
		cardWidth,
	)

	// Rating card
	ratingValue := "-"
	ratingSubtext := "not rated"
	var ratingColor lipgloss.Color = styles.ColorMuted
	if m.user != nil && m.user.Rating > 0 {
		ratingValue = fmt.Sprintf("%d", m.user.Rating)
		ratingSubtext = fmt.Sprintf("max: %d", m.user.MaxRating)
		ratingColor = styles.GetRankColor(m.user.Rating)
	}
	ratingCard := m.renderCard(
		"⭐ Rating",
		ratingValue,
		ratingSubtext,
		ratingColor,
		cardWidth,
	)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		solvedCard,
		"  ",
		streakCard,
		"  ",
		ratingCard,
	)
}

func (m DashboardModel) renderCard(title, value, subtext string, accentColor lipgloss.Color, width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextSecondary).
		MarginBottom(1)

	valueStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Width(width - 4).
		Align(lipgloss.Center)

	subtextStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Width(width - 4).
		Align(lipgloss.Center)

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render(title),
		valueStyle.Render(value),
		subtextStyle.Render(subtext),
	)

	return styles.CardStyle.Width(width).Render(content)
}

func (m DashboardModel) renderRecentActivity() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.TitleStyle.Render("📋 Recent Submissions"))
	b.WriteString("\n\n")

	if len(m.submissions) == 0 {
		b.WriteString(styles.SubtitleStyle.Render("  No recent submissions"))
		return b.String()
	}

	// Header
	header := fmt.Sprintf("  %-10s %-40s %-8s %s",
		"Problem",
		"Name",
		"Verdict",
		"Language",
	)
	b.WriteString(styles.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Show last 10 submissions
	count := 10
	if len(m.submissions) < count {
		count = len(m.submissions)
	}

	for i := 0; i < count; i++ {
		s := m.submissions[i]
		name := s.Problem.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}

		verdictColor := styles.GetVerdictColor(s.Verdict)
		verdictStyle := lipgloss.NewStyle().Foreground(verdictColor).Bold(true)

		row := fmt.Sprintf("  %-10s %-40s %-8s %s",
			s.Problem.ProblemID(),
			name,
			verdictStyle.Render(styles.GetVerdictShort(s.Verdict)),
			s.ProgrammingLanguage,
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
