package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
	"github.com/harshit-vibes/dsaprep/internal/config"
)

// StatsModel is the model for the statistics view
type StatsModel struct {
	keys          KeyMap
	width         int
	height        int
	user          *codeforces.User
	submissions   []codeforces.Submission
	ratingHistory []codeforces.RatingChange
	tab           int // 0 = overview, 1 = rating, 2 = submissions
}

// NewStatsModel creates a new stats model
func NewStatsModel(keys KeyMap) *StatsModel {
	return &StatsModel{
		keys: keys,
		tab:  0,
	}
}

// SetSize sets the view dimensions
func (m *StatsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetUser sets the user data
func (m *StatsModel) SetUser(user codeforces.User) {
	m.user = &user
}

// SetSubmissions sets the submissions data
func (m *StatsModel) SetSubmissions(submissions []codeforces.Submission) {
	m.submissions = submissions
}

// SetRatingHistory sets the rating history
func (m *StatsModel) SetRatingHistory(history []codeforces.RatingChange) {
	m.ratingHistory = history
}

// Init implements tea.Model
func (m *StatsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.NextTab):
			m.tab = (m.tab + 1) % 3
		case key.Matches(msg, m.keys.PrevTab):
			m.tab = (m.tab + 2) % 3
		case key.Matches(msg, m.keys.Left):
			m.tab = (m.tab + 2) % 3
		case key.Matches(msg, m.keys.Right):
			m.tab = (m.tab + 1) % 3
		}

	case UserLoadedMsg:
		m.SetUser(msg.User)

	case UserStatsLoadedMsg:
		m.SetSubmissions(msg.Submissions)
		m.SetRatingHistory(msg.RatingHistory)
	}

	return m, nil
}

// View implements tea.Model
func (m *StatsModel) View() string {
	if m.user == nil {
		return m.renderNoUser()
	}

	var b strings.Builder

	// User header
	b.WriteString(m.renderUserHeader())
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Content based on selected tab
	switch m.tab {
	case 0:
		b.WriteString(m.renderOverview())
	case 1:
		b.WriteString(m.renderRatingHistory())
	case 2:
		b.WriteString(m.renderSubmissions())
	}

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(1, 2).
		Render(b.String())
}

func (m *StatsModel) renderNoUser() string {
	handle := config.GetCFHandle()

	var content string
	if handle == "" {
		content = `
No Codeforces handle configured.

Set your handle with:
  dsaprep config set cf_handle <your_handle>

Then restart the app to see your statistics.
`
	} else {
		content = fmt.Sprintf(`
Loading statistics for @%s...

If this takes too long, press [r] to refresh.
`, handle)
	}

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Height(m.height - 4).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ColorMuted).
		Render(content)
}

func (m *StatsModel) renderUserHeader() string {
	rankStyle := RankStyle(m.user.Rank)
	ratingStyle := RatingStyle(m.user.Rating)

	title := TitleStyle.Render(fmt.Sprintf("@%s", m.user.Handle))

	info := fmt.Sprintf(
		"Rank: %s • Rating: %s • Max Rating: %s (%s)",
		rankStyle.Render(strings.ToUpper(m.user.Rank)),
		ratingStyle.Render(fmt.Sprintf("%d", m.user.Rating)),
		RatingStyle(m.user.MaxRating).Render(fmt.Sprintf("%d", m.user.MaxRating)),
		m.user.MaxRank,
	)

	var location string
	if m.user.Country != "" {
		location = fmt.Sprintf("Location: %s", m.user.Country)
		if m.user.City != "" {
			location += ", " + m.user.City
		}
	}

	var org string
	if m.user.Organization != "" {
		org = fmt.Sprintf("Organization: %s", m.user.Organization)
	}

	parts := []string{title, info}
	if location != "" {
		parts = append(parts, TextMuted.Render(location))
	}
	if org != "" {
		parts = append(parts, TextMuted.Render(org))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *StatsModel) renderTabs() string {
	tabs := []string{"Overview", "Rating History", "Recent Submissions"}
	var tabViews []string

	for i, tab := range tabs {
		style := InactiveTabStyle
		if i == m.tab {
			style = ActiveTabStyle
		}
		tabViews = append(tabViews, style.Render(tab))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
}

func (m *StatsModel) renderOverview() string {
	// Calculate statistics
	contestsParticipated := len(m.ratingHistory)

	var solved int
	solvedProblems := make(map[string]bool)
	for _, sub := range m.submissions {
		if sub.Verdict == "OK" {
			pid := sub.Problem.ID()
			if !solvedProblems[pid] {
				solvedProblems[pid] = true
				solved++
			}
		}
	}

	// Verdicts breakdown
	verdicts := make(map[string]int)
	for _, sub := range m.submissions {
		verdicts[sub.Verdict]++
	}

	// Rating change
	var ratingChange int
	if len(m.ratingHistory) > 0 {
		latest := m.ratingHistory[len(m.ratingHistory)-1]
		ratingChange = latest.NewRating - latest.OldRating
	}

	ratingChangeStr := fmt.Sprintf("%+d", ratingChange)
	ratingChangeStyle := SuccessStyle
	if ratingChange < 0 {
		ratingChangeStyle = ErrorStyle
	}

	// Build overview
	stats := []string{
		fmt.Sprintf("Contests Participated: %s", TextBold.Render(fmt.Sprintf("%d", contestsParticipated))),
		fmt.Sprintf("Problems Solved: %s", TextBold.Render(fmt.Sprintf("%d", solved))),
		fmt.Sprintf("Total Submissions: %s", TextBold.Render(fmt.Sprintf("%d", len(m.submissions)))),
		fmt.Sprintf("Last Rating Change: %s", ratingChangeStyle.Render(ratingChangeStr)),
		"",
		SubtitleStyle.Render("Verdict Breakdown:"),
	}

	// Add verdict breakdown
	verdictOrder := []string{"OK", "WRONG_ANSWER", "TIME_LIMIT_EXCEEDED", "RUNTIME_ERROR", "COMPILATION_ERROR"}
	for _, v := range verdictOrder {
		if count := verdicts[v]; count > 0 {
			percentage := float64(count) / float64(len(m.submissions)) * 100
			bar := ProgressBar(count, len(m.submissions), 20)
			stats = append(stats, fmt.Sprintf("  %-20s %s %3d (%.1f%%)", v, bar, count, percentage))
		}
	}

	return BoxStyle.Width(m.width - 8).Render(strings.Join(stats, "\n"))
}

func (m *StatsModel) renderRatingHistory() string {
	if len(m.ratingHistory) == 0 {
		return TextMuted.Render("No rating history available")
	}

	var rows []string

	// Header
	header := fmt.Sprintf("%-40s  %8s  %8s  %8s  %4s",
		"Contest", "Rank", "Old", "New", "Δ")
	rows = append(rows, TableHeaderStyle.Render(header))

	// Show last 10 contests (most recent first)
	start := len(m.ratingHistory) - 10
	if start < 0 {
		start = 0
	}

	for i := len(m.ratingHistory) - 1; i >= start; i-- {
		rc := m.ratingHistory[i]

		contestName := rc.ContestName
		if len(contestName) > 38 {
			contestName = contestName[:35] + "..."
		}

		delta := rc.NewRating - rc.OldRating
		deltaStr := fmt.Sprintf("%+d", delta)
		deltaStyle := SuccessStyle
		if delta < 0 {
			deltaStyle = ErrorStyle
		}

		row := fmt.Sprintf("%-40s  %8d  %8d  %8d  %s",
			contestName,
			rc.Rank,
			rc.OldRating,
			rc.NewRating,
			deltaStyle.Render(fmt.Sprintf("%4s", deltaStr)),
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

func (m *StatsModel) renderSubmissions() string {
	if len(m.submissions) == 0 {
		return TextMuted.Render("No submissions available")
	}

	var rows []string

	// Header
	header := fmt.Sprintf("%-12s  %-35s  %-12s  %-20s",
		"Problem", "Name", "Verdict", "Time")
	rows = append(rows, TableHeaderStyle.Render(header))

	// Show last 15 submissions
	count := 15
	if len(m.submissions) < count {
		count = len(m.submissions)
	}

	for i := 0; i < count; i++ {
		sub := m.submissions[i]

		pid := sub.Problem.ID()
		name := sub.Problem.Name
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		verdictStyle := TextMuted
		switch sub.Verdict {
		case "OK":
			verdictStyle = SuccessStyle
		case "WRONG_ANSWER", "RUNTIME_ERROR":
			verdictStyle = ErrorStyle
		case "TIME_LIMIT_EXCEEDED", "MEMORY_LIMIT_EXCEEDED":
			verdictStyle = WarningStyle
		}

		timeAgo := formatTimeAgo(time.Unix(sub.CreationTimeSeconds, 0))

		row := fmt.Sprintf("%-12s  %-35s  %s  %-20s",
			pid,
			name,
			verdictStyle.Render(fmt.Sprintf("%-12s", sub.Verdict)),
			timeAgo,
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)

	if diff < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	}
	if diff < 30*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	}
	if diff < 365*24*time.Hour {
		return fmt.Sprintf("%d months ago", int(diff.Hours()/(24*30)))
	}
	return fmt.Sprintf("%d years ago", int(diff.Hours()/(24*365)))
}
