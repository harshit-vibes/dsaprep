package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
)

// PracticeModel is the model for the practice session view
type PracticeModel struct {
	keys      KeyMap
	width     int
	height    int
	problem   *codeforces.Problem
	startTime time.Time
	elapsed   time.Duration
	isPaused  bool
	isActive  bool
}

// NewPracticeModel creates a new practice model
func NewPracticeModel(keys KeyMap) *PracticeModel {
	return &PracticeModel{
		keys: keys,
	}
}

// SetSize sets the view dimensions
func (m *PracticeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetProblem sets the current problem for practice
func (m *PracticeModel) SetProblem(problem codeforces.Problem) {
	m.problem = &problem
	m.startTime = time.Now()
	m.elapsed = 0
	m.isPaused = false
	m.isActive = true
}

// Init implements tea.Model
func (m *PracticeModel) Init() tea.Cmd {
	return m.tickCmd()
}

func (m *PracticeModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// Update implements tea.Model
func (m *PracticeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Open):
			if m.problem != nil {
				return m, func() tea.Msg {
					return OpenURLMsg{URL: m.problem.URL()}
				}
			}
		case key.Matches(msg, m.keys.Copy):
			if m.problem != nil {
				return m, func() tea.Msg {
					return CopyToClipboardMsg{Text: m.problem.URL()}
				}
			}
		case key.Matches(msg, m.keys.MarkSolved):
			if m.problem != nil {
				m.isActive = false
				return m, func() tea.Msg {
					return ProblemMarkedSolvedMsg{ProblemID: m.problem.ID()}
				}
			}
		case key.Matches(msg, m.keys.Skip):
			if m.problem != nil {
				m.isActive = false
				return m, func() tea.Msg {
					return SwitchViewMsg{View: ViewProblems}
				}
			}
		case key.Matches(msg, m.keys.Enter):
			// Toggle pause
			m.isPaused = !m.isPaused
			if !m.isPaused {
				m.startTime = time.Now().Add(-m.elapsed)
			}
		case key.Matches(msg, m.keys.Random):
			// Get another random problem
			return m, func() tea.Msg {
				return RefreshMsg{}
			}
		}

	case TickMsg:
		if m.isActive && !m.isPaused {
			m.elapsed = time.Since(m.startTime)
		}
		return m, m.tickCmd()

	case ProblemSelectedMsg:
		m.SetProblem(msg.Problem)
	}

	return m, nil
}

// View implements tea.Model
func (m *PracticeModel) View() string {
	if m.problem == nil {
		return m.renderEmpty()
	}

	var b strings.Builder

	// Problem header
	b.WriteString(m.renderProblemHeader())
	b.WriteString("\n\n")

	// Timer
	b.WriteString(m.renderTimer())
	b.WriteString("\n\n")

	// Problem details
	b.WriteString(m.renderProblemDetails())
	b.WriteString("\n\n")

	// Actions
	b.WriteString(m.renderActions())

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(1, 2).
		Render(b.String())
}

func (m *PracticeModel) renderEmpty() string {
	content := `
No problem selected.

Press [p] to browse problems
Press [R] to get a random problem
`
	return lipgloss.NewStyle().
		Width(m.width - 4).
		Height(m.height - 4).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ColorMuted).
		Render(content)
}

func (m *PracticeModel) renderProblemHeader() string {
	title := TitleStyle.Render(fmt.Sprintf("Practice: %s", m.problem.Name))

	info := fmt.Sprintf(
		"ID: %s • Contest: %d • Index: %s",
		m.problem.ID(),
		m.problem.ContestID,
		m.problem.Index,
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, TextMuted.Render(info))
}

func (m *PracticeModel) renderTimer() string {
	hours := int(m.elapsed.Hours())
	minutes := int(m.elapsed.Minutes()) % 60
	seconds := int(m.elapsed.Seconds()) % 60

	timeStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	timerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Background(ColorBgLight).
		Padding(1, 4).
		Align(lipgloss.Center)

	if m.isPaused {
		timerStyle = timerStyle.Foreground(ColorWarning)
		timeStr += " (PAUSED)"
	}

	timer := timerStyle.Render(timeStr)

	statusText := "Timer running • Press Enter to pause"
	if m.isPaused {
		statusText = "Timer paused • Press Enter to resume"
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		timer,
		TextMuted.Render(statusText),
	)
}

func (m *PracticeModel) renderProblemDetails() string {
	// Rating with color
	rating := m.problem.Rating
	ratingStr := "Unrated"
	if rating > 0 {
		ratingStr = fmt.Sprintf("%d", rating)
	}
	ratingView := fmt.Sprintf("Rating: %s", RatingStyle(rating).Render(ratingStr))

	// Rank
	rankView := fmt.Sprintf("Difficulty: %s", m.problem.RankName())

	// Tags
	var tagsView string
	if len(m.problem.Tags) > 0 {
		var tagStrs []string
		for _, tag := range m.problem.Tags {
			tagStrs = append(tagStrs, TagStyle.Render(tag))
		}
		tagsView = "Tags: " + strings.Join(tagStrs, " ")
	} else {
		tagsView = "Tags: None"
	}

	// URL
	urlView := TextMuted.Render(fmt.Sprintf("URL: %s", m.problem.URL()))

	details := lipgloss.JoinVertical(
		lipgloss.Left,
		ratingView,
		rankView,
		tagsView,
		"",
		urlView,
	)

	return BoxStyle.Width(m.width - 8).Render(details)
}

func (m *PracticeModel) renderActions() string {
	actions := []string{
		"[o] Open in browser",
		"[c] Copy link",
		"[m] Mark as solved",
		"[n] Skip to next",
		"[Enter] Pause/Resume timer",
	}

	return HelpStyle.Render(strings.Join(actions, "  •  "))
}
