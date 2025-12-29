package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/tui/styles"
)

// ProblemsModel is the problems browser view model
type ProblemsModel struct {
	width  int
	height int

	// Data
	problems []cfapi.Problem
	table    table.Model

	// State
	loading bool
}

// NewProblemsModel creates a new problems model
func NewProblemsModel() ProblemsModel {
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 45},
		{Title: "Rating", Width: 8},
		{Title: "Tags", Width: 35},
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

	return ProblemsModel{
		table: t,
	}
}

// SetSize sets the view dimensions
func (m *ProblemsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.table.SetHeight(height - 6)
}

// SetProblems sets the problems data
func (m *ProblemsModel) SetProblems(problems []cfapi.Problem) {
	m.problems = problems
	m.loading = false

	rows := make([]table.Row, len(problems))
	for i, p := range problems {
		ratingStr := "-"
		if p.Rating > 0 {
			ratingStr = fmt.Sprintf("%d", p.Rating)
		}

		tags := strings.Join(p.Tags, ", ")
		if len(tags) > 33 {
			tags = tags[:30] + "..."
		}

		name := p.Name
		if len(name) > 43 {
			name = name[:40] + "..."
		}

		rows[i] = table.Row{
			p.ProblemID(),
			name,
			ratingStr,
			tags,
		}
	}

	m.table.SetRows(rows)
}

// Init initializes the model
func (m ProblemsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ProblemsModel) Update(msg tea.Msg) (ProblemsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "o":
			// Open selected problem in browser
			if len(m.problems) > 0 && m.table.Cursor() < len(m.problems) {
				p := m.problems[m.table.Cursor()]
				// Could use exec.Command to open browser
				_ = p.URL()
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the problems view
func (m ProblemsModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("📝 Problem Browser"))
	b.WriteString("\n")
	b.WriteString(styles.SubtitleStyle.Render(fmt.Sprintf("  %d problems loaded", len(m.problems))))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("  Loading problems...")
		return b.String()
	}

	if len(m.problems) == 0 {
		b.WriteString(styles.SubtitleStyle.Render("  Press 'r' to load problems"))
		return b.String()
	}

	b.WriteString(m.table.View())
	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("  ↑/↓ navigate • enter select • o open in browser • r refresh"))

	return b.String()
}
