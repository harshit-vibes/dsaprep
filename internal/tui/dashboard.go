package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
	"github.com/harshit-vibes/dsaprep/internal/config"
)

// DashboardModel is the model for the dashboard view
type DashboardModel struct {
	keys       KeyMap
	width      int
	height     int
	problems   []codeforces.Problem
	statistics []codeforces.ProblemStatistic
	user       *codeforces.User
	focused    int // 0 = quick actions, 1 = recent problems
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(keys KeyMap) *DashboardModel {
	return &DashboardModel{
		keys:    keys,
		focused: 0,
	}
}

// SetSize sets the view dimensions
func (m *DashboardModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetProblems sets the problems data
func (m *DashboardModel) SetProblems(problems []codeforces.Problem) {
	m.problems = problems
}

// SetStatistics sets the problem statistics
func (m *DashboardModel) SetStatistics(stats []codeforces.ProblemStatistic) {
	m.statistics = stats
}

// SetUser sets the user data
func (m *DashboardModel) SetUser(user codeforces.User) {
	m.user = &user
}

// Init implements tea.Model
func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.focused > 0 {
				m.focused--
			}
		case key.Matches(msg, m.keys.Down):
			if m.focused < 1 {
				m.focused++
			}
		case key.Matches(msg, m.keys.Enter):
			// Handle selection based on focused item
			if m.focused == 0 {
				// Quick actions
				return m, func() tea.Msg {
					return SwitchViewMsg{View: ViewProblems}
				}
			}
		case key.Matches(msg, m.keys.Random):
			return m, m.selectRandomProblem()
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *DashboardModel) View() string {
	var b strings.Builder

	// Welcome section
	b.WriteString(m.renderWelcome())
	b.WriteString("\n\n")

	// Stats summary
	b.WriteString(m.renderStats())
	b.WriteString("\n\n")

	// Quick actions
	b.WriteString(m.renderQuickActions())
	b.WriteString("\n\n")

	// Recent/suggested problems
	b.WriteString(m.renderSuggestedProblems())

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(1, 2).
		Render(b.String())
}

func (m *DashboardModel) renderWelcome() string {
	cfg := config.Get()

	greeting := "Welcome to DSA Prep!"
	if cfg.CFHandle != "" {
		greeting = fmt.Sprintf("Welcome, @%s!", cfg.CFHandle)
	}

	title := TitleStyle.Render(greeting)

	var userInfo string
	if m.user != nil {
		rankStyle := RankStyle(m.user.Rank)
		userInfo = fmt.Sprintf(
			"Rank: %s • Rating: %s • Max: %d",
			rankStyle.Render(strings.ToTitle(m.user.Rank)),
			RatingStyle(m.user.Rating).Render(fmt.Sprintf("%d", m.user.Rating)),
			m.user.MaxRating,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, userInfo)
}

func (m *DashboardModel) renderStats() string {
	cfg := config.Get()

	// Count problems in difficulty range
	var inRange int
	for _, p := range m.problems {
		if p.Rating >= cfg.Difficulty.Min && p.Rating <= cfg.Difficulty.Max {
			inRange++
		}
	}

	statsBox := BoxStyle.Render(fmt.Sprintf(
		"📊 Stats\n\n"+
			"Total Problems: %s\n"+
			"In Your Range (%d-%d): %s\n"+
			"Daily Goal: %s problems",
		TextBold.Render(fmt.Sprintf("%d", len(m.problems))),
		cfg.Difficulty.Min, cfg.Difficulty.Max,
		TextBold.Render(fmt.Sprintf("%d", inRange)),
		TextBold.Render(fmt.Sprintf("%d", cfg.DailyGoal)),
	))

	return statsBox
}

func (m *DashboardModel) renderQuickActions() string {
	title := SubtitleStyle.Render("⚡ Quick Actions")

	actions := []string{
		"[p] Browse Problems",
		"[R] Random Problem",
		"[s] View Stats",
		"[r] Refresh Data",
	}

	style := BoxStyle
	if m.focused == 0 {
		style = FocusedBoxStyle
	}

	actionsText := strings.Join(actions, "  •  ")
	return lipgloss.JoinVertical(lipgloss.Left, title, style.Render(actionsText))
}

func (m *DashboardModel) renderSuggestedProblems() string {
	title := SubtitleStyle.Render("🎯 Suggested Problems")

	if len(m.problems) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, TextMuted.Render("No problems loaded"))
	}

	cfg := config.Get()

	// Filter problems in range and take first 5
	var suggested []codeforces.Problem
	for _, p := range m.problems {
		if p.Rating >= cfg.Difficulty.Min && p.Rating <= cfg.Difficulty.Max {
			suggested = append(suggested, p)
			if len(suggested) >= 5 {
				break
			}
		}
	}

	if len(suggested) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, TextMuted.Render("No problems in your difficulty range"))
	}

	var rows []string
	for i, p := range suggested {
		rating := RatingStyle(p.Rating).Render(fmt.Sprintf("%4d", p.Rating))
		name := p.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		row := fmt.Sprintf("%d. [%s] %s", i+1, rating, name)
		rows = append(rows, row)
	}

	style := BoxStyle
	if m.focused == 1 {
		style = FocusedBoxStyle
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, style.Render(strings.Join(rows, "\n")))
}

func (m *DashboardModel) selectRandomProblem() tea.Cmd {
	return func() tea.Msg {
		if len(m.problems) == 0 {
			return StatusMsg{Message: "No problems available", IsError: true}
		}

		cfg := config.Get()

		// Filter by difficulty
		var eligible []codeforces.Problem
		for _, p := range m.problems {
			if p.Rating >= cfg.Difficulty.Min && p.Rating <= cfg.Difficulty.Max {
				eligible = append(eligible, p)
			}
		}

		if len(eligible) == 0 {
			return StatusMsg{Message: "No problems in your difficulty range", IsError: true}
		}

		// Pick a random one (simple selection for now)
		idx := len(eligible) / 2
		return ProblemSelectedMsg{Problem: eligible[idx]}
	}
}
