// Package tui provides a terminal user interface for cf
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/tui/styles"
	"github.com/harshit-vibes/cf/pkg/tui/views"
)

// App is the main application model
type App struct {
	// State
	currentView View
	loading     bool
	statusMsg   string
	err         error

	// Dimensions
	width  int
	height int

	// Components
	keys    KeyMap
	help    help.Model
	spinner spinner.Model

	// Views
	dashboard   views.DashboardModel
	problems    views.ProblemsModel
	submissions views.SubmissionsModel
	profile     views.ProfileModel
	settings    views.SettingsModel

	// Data
	client *cfapi.Client
	handle string
	user   *cfapi.User
}

// New creates a new App instance
func New() *App {
	// Load credentials
	creds, _ := config.LoadCredentials()
	handle := ""
	if creds != nil {
		handle = creds.CFHandle
	}

	// Create API client
	var client *cfapi.Client
	if creds != nil && creds.IsAPIConfigured() {
		client = cfapi.NewClient(cfapi.WithAPICredentials(creds.APIKey, creds.APISecret))
	} else {
		client = cfapi.NewClient()
	}

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	return &App{
		currentView: ViewDashboard,
		keys:        DefaultKeyMap(),
		help:        help.New(),
		spinner:     s,
		client:      client,
		handle:      handle,
		width:       styles.DefaultWidth,
		height:      styles.DefaultHeight,
		dashboard:   views.NewDashboardModel(),
		problems:    views.NewProblemsModel(),
		submissions: views.NewSubmissionsModel(),
		profile:     views.NewProfileModel(),
		settings:    views.NewSettingsModel(),
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		a.loadInitialData(),
	)
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = msg.Width

		// Update all views with new size
		a.dashboard.SetSize(msg.Width, msg.Height-styles.HeaderHeight-styles.FooterHeight-styles.TabHeight)
		a.problems.SetSize(msg.Width, msg.Height-styles.HeaderHeight-styles.FooterHeight-styles.TabHeight)
		a.submissions.SetSize(msg.Width, msg.Height-styles.HeaderHeight-styles.FooterHeight-styles.TabHeight)
		a.profile.SetSize(msg.Width, msg.Height-styles.HeaderHeight-styles.FooterHeight-styles.TabHeight)
		a.settings.SetSize(msg.Width, msg.Height-styles.HeaderHeight-styles.FooterHeight-styles.TabHeight)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit

		case key.Matches(msg, a.keys.Tab1):
			a.currentView = ViewDashboard
			cmds = append(cmds, a.refreshCurrentView())

		case key.Matches(msg, a.keys.Tab2):
			a.currentView = ViewProblems
			cmds = append(cmds, a.refreshCurrentView())

		case key.Matches(msg, a.keys.Tab3):
			a.currentView = ViewSubmissions
			cmds = append(cmds, a.refreshCurrentView())

		case key.Matches(msg, a.keys.Tab4):
			a.currentView = ViewProfile
			cmds = append(cmds, a.refreshCurrentView())

		case key.Matches(msg, a.keys.Tab5):
			a.currentView = ViewSettings

		case key.Matches(msg, a.keys.NextTab):
			a.currentView = (a.currentView + 1) % 5
			cmds = append(cmds, a.refreshCurrentView())

		case key.Matches(msg, a.keys.PrevTab):
			a.currentView = (a.currentView + 4) % 5
			cmds = append(cmds, a.refreshCurrentView())

		case key.Matches(msg, a.keys.Refresh):
			cmds = append(cmds, a.refreshCurrentView())

		case key.Matches(msg, a.keys.Help):
			a.help.ShowAll = !a.help.ShowAll
		}

	case SwitchViewMsg:
		a.currentView = msg.View
		cmds = append(cmds, a.refreshCurrentView())

	case LoadingMsg:
		a.loading = msg.Loading
		a.statusMsg = msg.Message
		if a.loading {
			cmds = append(cmds, a.spinner.Tick)
		}

	case StatusMsg:
		a.statusMsg = msg.Message

	case ErrorMsg:
		a.err = msg.Err
		a.loading = false

	case UserLoadedMsg:
		a.user = &msg.User
		a.loading = false
		a.profile.SetUser(&msg.User)
		a.dashboard.SetUser(&msg.User)

	case ProblemsLoadedMsg:
		a.problems.SetProblems(msg.Problems)
		a.loading = false

	case SubmissionsLoadedMsg:
		a.submissions.SetSubmissions(msg.Submissions)
		a.dashboard.SetSubmissions(msg.Submissions)
		a.loading = false

	case RatingLoadedMsg:
		a.profile.SetRatingHistory(msg.RatingChanges)
		a.loading = false

	case StatsLoadedMsg:
		a.dashboard.SetStats(msg.TotalSolved, msg.RecentSolved, msg.Streak)
		a.loading = false

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update current view
	switch a.currentView {
	case ViewDashboard:
		var cmd tea.Cmd
		a.dashboard, cmd = a.dashboard.Update(msg)
		cmds = append(cmds, cmd)
	case ViewProblems:
		var cmd tea.Cmd
		a.problems, cmd = a.problems.Update(msg)
		cmds = append(cmds, cmd)
	case ViewSubmissions:
		var cmd tea.Cmd
		a.submissions, cmd = a.submissions.Update(msg)
		cmds = append(cmds, cmd)
	case ViewProfile:
		var cmd tea.Cmd
		a.profile, cmd = a.profile.Update(msg)
		cmds = append(cmds, cmd)
	case ViewSettings:
		var cmd tea.Cmd
		a.settings, cmd = a.settings.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	var b strings.Builder

	// Header
	b.WriteString(a.renderHeader())
	b.WriteString("\n")

	// Tab bar
	b.WriteString(a.renderTabBar())
	b.WriteString("\n")

	// Content
	contentHeight := a.height - styles.HeaderHeight - styles.FooterHeight - styles.TabHeight - 2
	content := a.renderContent()
	b.WriteString(lipgloss.NewStyle().Height(contentHeight).Render(content))
	b.WriteString("\n")

	// Footer
	b.WriteString(a.renderFooter())

	return styles.AppStyle.Render(b.String())
}

func (a *App) renderHeader() string {
	logo := styles.LogoStyle.Render("◆ cf")
	title := " - Codeforces CLI"

	var status string
	if a.loading {
		status = a.spinner.View() + " " + a.statusMsg
	} else if a.err != nil {
		status = styles.ErrorStyle.Render("Error: " + a.err.Error())
	} else if a.handle != "" {
		status = styles.SubtitleStyle.Render("@" + a.handle)
	}

	left := logo + title
	right := status

	gap := a.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 0 {
		gap = 0
	}

	return styles.HeaderStyle.Width(a.width - 2).Render(
		left + strings.Repeat(" ", gap) + right,
	)
}

func (a *App) renderTabBar() string {
	tabs := []View{ViewDashboard, ViewProblems, ViewSubmissions, ViewProfile, ViewSettings}
	var renderedTabs []string

	for _, tab := range tabs {
		label := fmt.Sprintf("%s %s", tab.Icon(), tab.String())
		if tab == a.currentView {
			renderedTabs = append(renderedTabs, styles.ActiveTabStyle.Render(label))
		} else {
			renderedTabs = append(renderedTabs, styles.InactiveTabStyle.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (a *App) renderContent() string {
	switch a.currentView {
	case ViewDashboard:
		return a.dashboard.View()
	case ViewProblems:
		return a.problems.View()
	case ViewSubmissions:
		return a.submissions.View()
	case ViewProfile:
		return a.profile.View()
	case ViewSettings:
		return a.settings.View()
	default:
		return "Unknown view"
	}
}

func (a *App) renderFooter() string {
	helpView := a.help.View(a.keys)
	return styles.FooterStyle.Render(helpView)
}

// Commands

func (a *App) loadInitialData() tea.Cmd {
	return tea.Batch(
		a.loadUser(),
		a.loadSubmissions(),
	)
}

func (a *App) loadUser() tea.Cmd {
	return func() tea.Msg {
		if a.handle == "" {
			return ErrorMsg{Err: fmt.Errorf("no CF handle configured")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		users, err := a.client.GetUserInfo(ctx, []string{a.handle})
		if err != nil {
			return ErrorMsg{Err: err}
		}

		if len(users) == 0 {
			return ErrorMsg{Err: fmt.Errorf("user not found: %s", a.handle)}
		}

		return UserLoadedMsg{User: users[0]}
	}
}

func (a *App) loadSubmissions() tea.Cmd {
	return func() tea.Msg {
		if a.handle == "" {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		subs, err := a.client.GetUserSubmissions(ctx, a.handle, 1, 100)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return SubmissionsLoadedMsg{Submissions: subs}
	}
}

func (a *App) loadProblems() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cfg := config.Get()
		minRating := 800
		maxRating := 1400
		if cfg != nil {
			minRating = cfg.Difficulty.Min
			maxRating = cfg.Difficulty.Max
		}

		problems, err := a.client.FilterProblems(ctx, minRating, maxRating, nil, false, "")
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Limit to 100 problems for performance
		if len(problems) > 100 {
			problems = problems[:100]
		}

		return ProblemsLoadedMsg{Problems: problems}
	}
}

func (a *App) loadRating() tea.Cmd {
	return func() tea.Msg {
		if a.handle == "" {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		changes, err := a.client.GetUserRating(ctx, a.handle)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return RatingLoadedMsg{RatingChanges: changes}
	}
}

func (a *App) refreshCurrentView() tea.Cmd {
	switch a.currentView {
	case ViewDashboard:
		return tea.Batch(a.loadUser(), a.loadSubmissions())
	case ViewProblems:
		return a.loadProblems()
	case ViewSubmissions:
		return a.loadSubmissions()
	case ViewProfile:
		return tea.Batch(a.loadUser(), a.loadRating())
	default:
		return nil
	}
}

// Run starts the TUI application
func Run() error {
	app := New()
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
