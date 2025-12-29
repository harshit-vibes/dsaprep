package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/tui/styles"
)

// SubmissionsModel is the submissions view model
type SubmissionsModel struct {
	width  int
	height int

	// Data
	submissions []cfapi.Submission
	table       table.Model

	// State
	loading bool
}

// NewSubmissionsModel creates a new submissions model
func NewSubmissionsModel() SubmissionsModel {
	columns := []table.Column{
		{Title: "Time", Width: 14},
		{Title: "Problem", Width: 10},
		{Title: "Name", Width: 35},
		{Title: "Verdict", Width: 8},
		{Title: "Time", Width: 8},
		{Title: "Language", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorSubtle).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(styles.ColorTextPrimary).
		Background(styles.ColorBgSelected).
		Bold(true)
	t.SetStyles(s)

	return SubmissionsModel{
		table: t,
	}
}

// SetSize sets the view dimensions
func (m *SubmissionsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.table.SetHeight(height - 6)
}

// SetSubmissions sets the submissions data
func (m *SubmissionsModel) SetSubmissions(submissions []cfapi.Submission) {
	m.submissions = submissions
	m.loading = false

	rows := make([]table.Row, len(submissions))
	for i, s := range submissions {
		timeStr := s.SubmissionTime().Format("Jan 02 15:04")

		name := s.Problem.Name
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		execTime := fmt.Sprintf("%dms", s.TimeConsumedMillis)

		rows[i] = table.Row{
			timeStr,
			s.Problem.ProblemID(),
			name,
			styles.GetVerdictShort(s.Verdict),
			execTime,
			s.ProgrammingLanguage,
		}
	}

	m.table.SetRows(rows)
}

// Init initializes the model
func (m SubmissionsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SubmissionsModel) Update(msg tea.Msg) (SubmissionsModel, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the submissions view
func (m SubmissionsModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("📤 Submissions"))
	b.WriteString("\n")
	b.WriteString(styles.SubtitleStyle.Render(fmt.Sprintf("  %d submissions", len(m.submissions))))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("  Loading submissions...")
		return b.String()
	}

	if len(m.submissions) == 0 {
		b.WriteString(styles.SubtitleStyle.Render("  No submissions yet"))
		return b.String()
	}

	// Render table with colored verdicts
	b.WriteString(m.renderTable())
	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("  ↑/↓ navigate • r refresh"))

	return b.String()
}

func (m SubmissionsModel) renderTable() string {
	// Custom render to apply verdict colors
	var b strings.Builder

	// Header
	header := fmt.Sprintf("  %-14s %-10s %-35s %-8s %-8s %s",
		"Time", "Problem", "Name", "Verdict", "Time", "Language",
	)
	b.WriteString(styles.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	start := 0
	end := len(m.submissions)
	if end > m.height-8 {
		end = m.height - 8
	}

	cursor := m.table.Cursor()

	for i := start; i < end; i++ {
		s := m.submissions[i]
		timeStr := s.SubmissionTime().Format("Jan 02 15:04")

		name := s.Problem.Name
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		execTime := fmt.Sprintf("%dms", s.TimeConsumedMillis)
		lang := s.ProgrammingLanguage
		if len(lang) > 18 {
			lang = lang[:18] + ".."
		}

		verdictColor := styles.GetVerdictColor(s.Verdict)
		verdictStyle := lipgloss.NewStyle().Foreground(verdictColor).Bold(true)

		row := fmt.Sprintf("  %-14s %-10s %-35s %-8s %-8s %s",
			timeStr,
			s.Problem.ProblemID(),
			name,
			verdictStyle.Render(styles.GetVerdictShort(s.Verdict)),
			execTime,
			lang,
		)

		if i == cursor {
			b.WriteString(styles.SelectedItemStyle.Render(row))
		} else if i%2 == 0 {
			b.WriteString(styles.TableRowStyle.Render(row))
		} else {
			b.WriteString(styles.TableRowAltStyle.Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatTimeAgo formats a time as "X ago"
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 02")
	}
}
