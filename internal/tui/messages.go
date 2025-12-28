package tui

import (
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
)

// View represents the current view/screen
type View int

const (
	ViewDashboard View = iota
	ViewProblems
	ViewPractice
	ViewStats
)

// String returns the view name
func (v View) String() string {
	switch v {
	case ViewDashboard:
		return "Dashboard"
	case ViewProblems:
		return "Problems"
	case ViewPractice:
		return "Practice"
	case ViewStats:
		return "Statistics"
	default:
		return "Unknown"
	}
}

// SwitchViewMsg is sent to switch between views
type SwitchViewMsg struct {
	View View
}

// ErrorMsg represents an error message
type ErrorMsg struct {
	Err error
}

func (e ErrorMsg) Error() string {
	return e.Err.Error()
}

// StatusMsg represents a status message to display
type StatusMsg struct {
	Message string
	IsError bool
}

// LoadingMsg indicates a loading state change
type LoadingMsg struct {
	IsLoading bool
	Message   string
}

// ProblemsLoadedMsg is sent when problems are loaded
type ProblemsLoadedMsg struct {
	Problems   []codeforces.Problem
	Statistics []codeforces.ProblemStatistic
}

// UserLoadedMsg is sent when user info is loaded
type UserLoadedMsg struct {
	User codeforces.User
}

// UserStatsLoadedMsg is sent when user statistics are loaded
type UserStatsLoadedMsg struct {
	Submissions  []codeforces.Submission
	RatingHistory []codeforces.RatingChange
}

// ContestsLoadedMsg is sent when contests are loaded
type ContestsLoadedMsg struct {
	Contests []codeforces.Contest
}

// ProblemSelectedMsg is sent when a problem is selected
type ProblemSelectedMsg struct {
	Problem codeforces.Problem
}

// ProblemMarkedSolvedMsg is sent when a problem is marked as solved
type ProblemMarkedSolvedMsg struct {
	ProblemID string
}

// RefreshMsg triggers a data refresh
type RefreshMsg struct{}

// TickMsg is sent periodically for timer updates
type TickMsg struct{}

// WindowSizeMsg is sent when the terminal window is resized
type WindowSizeMsg struct {
	Width  int
	Height int
}

// FilterMsg represents filter criteria for problems
type FilterMsg struct {
	MinRating int
	MaxRating int
	Tags      []string
	Search    string
	Solved    *bool // nil = all, true = solved only, false = unsolved only
}

// SortMsg represents sorting options
type SortMsg struct {
	Field     string // "rating", "name", "solved_count", "contest_id"
	Ascending bool
}

// PracticeSessionMsg manages practice session state
type PracticeSessionMsg struct {
	Action      string // "start", "pause", "resume", "end"
	Problem     *codeforces.Problem
	ElapsedTime int // seconds
}

// OpenURLMsg requests opening a URL in browser
type OpenURLMsg struct {
	URL string
}

// CopyToClipboardMsg requests copying text to clipboard
type CopyToClipboardMsg struct {
	Text string
}

// ClearCacheMsg clears the API cache
type ClearCacheMsg struct{}

// ConfigUpdatedMsg is sent when configuration is updated
type ConfigUpdatedMsg struct {
	Key   string
	Value interface{}
}

// HelpToggleMsg toggles help visibility
type HelpToggleMsg struct{}

// QuitMsg signals application quit
type QuitMsg struct{}
