package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
	"github.com/harshit-vibes/dsaprep/internal/config"
)

// App is the main application model
type App struct {
	// Core
	client *codeforces.Client
	keys   KeyMap
	help   help.Model

	// UI state
	currentView View
	width       int
	height      int
	showHelp    bool
	isLoading   bool
	loadingMsg  string
	statusMsg   string
	isError     bool

	// Spinner for loading states
	spinner spinner.Model

	// View models
	dashboard *DashboardModel
	problems  *ProblemsModel
	practice  *PracticeModel
	stats     *StatsModel
}

// NewApp creates a new application instance
func NewApp() *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	h := help.New()
	h.ShowAll = false

	return &App{
		client:      codeforces.NewClient(),
		keys:        DefaultKeyMap(),
		help:        h,
		spinner:     s,
		currentView: ViewDashboard,
		width:       80,
		height:      24,
	}
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		a.loadInitialData(),
	)
}

// loadInitialData loads the initial data for the app
func (a *App) loadInitialData() tea.Cmd {
	return func() tea.Msg {
		a.isLoading = true
		a.loadingMsg = "Loading problems..."

		ctx := context.Background()

		// Load problems
		result, err := a.client.GetProblems(ctx)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return ProblemsLoadedMsg{
			Problems:   result.Problems,
			Statistics: result.ProblemStatistics,
		}
	}
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handlers
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit

		case key.Matches(msg, a.keys.Help):
			a.showHelp = !a.showHelp
			return a, nil

		case key.Matches(msg, a.keys.Dashboard):
			a.currentView = ViewDashboard
			return a, nil

		case key.Matches(msg, a.keys.Problems):
			a.currentView = ViewProblems
			return a, nil

		case key.Matches(msg, a.keys.Practice):
			a.currentView = ViewPractice
			return a, nil

		case key.Matches(msg, a.keys.Stats):
			if a.stats == nil {
				// Load user stats when first accessing
				return a, a.loadUserStats()
			}
			a.currentView = ViewStats
			return a, nil

		case key.Matches(msg, a.keys.Refresh):
			return a, a.refreshData()
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = msg.Width

		// Update child view dimensions
		if a.dashboard != nil {
			a.dashboard.SetSize(msg.Width, msg.Height-4)
		}
		if a.problems != nil {
			a.problems.SetSize(msg.Width, msg.Height-4)
		}
		if a.practice != nil {
			a.practice.SetSize(msg.Width, msg.Height-4)
		}
		if a.stats != nil {
			a.stats.SetSize(msg.Width, msg.Height-4)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case ProblemsLoadedMsg:
		a.isLoading = false
		a.loadingMsg = ""
		a.initializeViews(msg.Problems, msg.Statistics)

	case UserLoadedMsg:
		if a.stats != nil {
			a.stats.SetUser(msg.User)
		}
		if a.dashboard != nil {
			a.dashboard.SetUser(msg.User)
		}

	case UserStatsLoadedMsg:
		a.isLoading = false
		if a.stats != nil {
			a.stats.SetSubmissions(msg.Submissions)
			a.stats.SetRatingHistory(msg.RatingHistory)
		}
		a.currentView = ViewStats

	case ErrorMsg:
		a.isLoading = false
		a.statusMsg = fmt.Sprintf("Error: %v", msg.Err)
		a.isError = true

	case StatusMsg:
		a.statusMsg = msg.Message
		a.isError = msg.IsError

	case SwitchViewMsg:
		a.currentView = msg.View

	case ProblemSelectedMsg:
		if a.practice == nil {
			a.practice = NewPracticeModel(a.keys)
			a.practice.SetSize(a.width, a.height-4)
		}
		a.practice.SetProblem(msg.Problem)
		a.currentView = ViewPractice
	}

	// Update current view
	switch a.currentView {
	case ViewDashboard:
		if a.dashboard != nil {
			model, cmd := a.dashboard.Update(msg)
			a.dashboard = model.(*DashboardModel)
			cmds = append(cmds, cmd)
		}
	case ViewProblems:
		if a.problems != nil {
			model, cmd := a.problems.Update(msg)
			a.problems = model.(*ProblemsModel)
			cmds = append(cmds, cmd)
		}
	case ViewPractice:
		if a.practice != nil {
			model, cmd := a.practice.Update(msg)
			a.practice = model.(*PracticeModel)
			cmds = append(cmds, cmd)
		}
	case ViewStats:
		if a.stats != nil {
			model, cmd := a.stats.Update(msg)
			a.stats = model.(*StatsModel)
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// View implements tea.Model
func (a *App) View() string {
	var b strings.Builder

	// Header with tabs
	b.WriteString(a.renderHeader())
	b.WriteString("\n")

	// Main content area
	if a.isLoading {
		b.WriteString(a.renderLoading())
	} else {
		b.WriteString(a.renderCurrentView())
	}

	// Footer with help
	b.WriteString("\n")
	b.WriteString(a.renderFooter())

	return b.String()
}

// renderHeader renders the header with navigation tabs
func (a *App) renderHeader() string {
	tabs := []string{"Dashboard", "Problems", "Practice", "Stats"}
	var tabViews []string

	for i, tab := range tabs {
		style := InactiveTabStyle
		if View(i) == a.currentView {
			style = ActiveTabStyle
		}
		tabViews = append(tabViews, style.Render(fmt.Sprintf(" %d %s ", i+1, tab)))
	}

	// Combine tabs
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)

	// Add user handle if configured
	handle := config.GetCFHandle()
	var userInfo string
	if handle != "" {
		userInfo = TextMuted.Render(fmt.Sprintf("@%s", handle))
	}

	// Create header with tabs on left, user on right
	headerWidth := a.width - 4
	if headerWidth < 0 {
		headerWidth = 80
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tabBar,
		lipgloss.NewStyle().Width(headerWidth-lipgloss.Width(tabBar)-lipgloss.Width(userInfo)).Render(""),
		userInfo,
	)

	return BoxStyle.Width(a.width - 2).Render(header)
}

// renderLoading renders the loading view
func (a *App) renderLoading() string {
	msg := a.loadingMsg
	if msg == "" {
		msg = "Loading..."
	}

	content := fmt.Sprintf("\n\n%s %s\n\n", a.spinner.View(), msg)
	return lipgloss.NewStyle().
		Width(a.width - 4).
		Height(a.height - 8).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// renderCurrentView renders the current view's content
func (a *App) renderCurrentView() string {
	var content string

	switch a.currentView {
	case ViewDashboard:
		if a.dashboard != nil {
			content = a.dashboard.View()
		} else {
			content = "Dashboard loading..."
		}
	case ViewProblems:
		if a.problems != nil {
			content = a.problems.View()
		} else {
			content = "Problems loading..."
		}
	case ViewPractice:
		if a.practice != nil {
			content = a.practice.View()
		} else {
			content = "Select a problem to practice"
		}
	case ViewStats:
		if a.stats != nil {
			content = a.stats.View()
		} else {
			content = "Statistics loading..."
		}
	}

	return content
}

// renderFooter renders the footer with status and help
func (a *App) renderFooter() string {
	var status string
	if a.statusMsg != "" {
		style := SuccessStyle
		if a.isError {
			style = ErrorStyle
		}
		status = style.Render(a.statusMsg)
	}

	helpView := a.help.View(a.keys)

	if status != "" {
		return lipgloss.JoinVertical(lipgloss.Left, status, helpView)
	}

	return helpView
}

// initializeViews creates the view models after data is loaded
func (a *App) initializeViews(problems []codeforces.Problem, stats []codeforces.ProblemStatistic) {
	cfg := config.Get()

	// Create dashboard
	a.dashboard = NewDashboardModel(a.keys)
	a.dashboard.SetProblems(problems)
	a.dashboard.SetStatistics(stats)
	a.dashboard.SetSize(a.width, a.height-4)

	// Create problems view
	a.problems = NewProblemsModel(a.keys)
	a.problems.SetProblems(problems)
	a.problems.SetStatistics(stats)
	a.problems.SetDifficultyRange(cfg.Difficulty.Min, cfg.Difficulty.Max)
	a.problems.SetSize(a.width, a.height-4)

	// Create practice view (empty until problem selected)
	a.practice = NewPracticeModel(a.keys)
	a.practice.SetSize(a.width, a.height-4)

	// Create stats view
	a.stats = NewStatsModel(a.keys)
	a.stats.SetSize(a.width, a.height-4)

	// Load user data if handle is configured
	if cfg.CFHandle != "" {
		// Will be loaded async
	}
}

// loadUserStats loads user statistics from Codeforces
func (a *App) loadUserStats() tea.Cmd {
	return func() tea.Msg {
		handle := config.GetCFHandle()
		if handle == "" {
			return StatusMsg{Message: "Set your CF handle with: dsaprep config set cf_handle <handle>", IsError: true}
		}

		a.isLoading = true
		a.loadingMsg = "Loading user stats..."

		ctx := context.Background()

		// Load user info
		users, err := a.client.GetUserInfo(ctx, handle)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to load user info: %w", err)}
		}

		if len(users) == 0 {
			return ErrorMsg{Err: fmt.Errorf("user not found: %s", handle)}
		}

		// Load submissions
		submissions, err := a.client.GetUserStatus(ctx, handle, 100)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to load submissions: %w", err)}
		}

		// Load rating history
		ratingHistory, err := a.client.GetUserRating(ctx, handle)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to load rating history: %w", err)}
		}

		return tea.Batch(
			func() tea.Msg { return UserLoadedMsg{User: users[0]} },
			func() tea.Msg {
				return UserStatsLoadedMsg{
					Submissions:   submissions,
					RatingHistory: ratingHistory,
				}
			},
		)()
	}
}

// refreshData refreshes all data from the API
func (a *App) refreshData() tea.Cmd {
	a.client.ClearCache()
	return a.loadInitialData()
}

// Run starts the TUI application
func Run() error {
	// Initialize config
	if err := config.Init(""); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running app: %w", err)
	}

	return nil
}
