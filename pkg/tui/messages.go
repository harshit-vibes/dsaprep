package tui

import (
	"github.com/harshit-vibes/cf/pkg/external/cfapi"
)

// View represents different views/tabs in the application
type View int

const (
	ViewDashboard View = iota
	ViewProblems
	ViewSubmissions
	ViewProfile
	ViewSettings
)

// String returns the view name
func (v View) String() string {
	switch v {
	case ViewDashboard:
		return "Dashboard"
	case ViewProblems:
		return "Problems"
	case ViewSubmissions:
		return "Submissions"
	case ViewProfile:
		return "Profile"
	case ViewSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}

// Icon returns an icon for the view
func (v View) Icon() string {
	switch v {
	case ViewDashboard:
		return "📊"
	case ViewProblems:
		return "📝"
	case ViewSubmissions:
		return "📤"
	case ViewProfile:
		return "👤"
	case ViewSettings:
		return "⚙️"
	default:
		return "?"
	}
}

// Messages

// SwitchViewMsg switches to a different view
type SwitchViewMsg struct {
	View View
}

// ErrorMsg represents an error
type ErrorMsg struct {
	Err error
}

// StatusMsg represents a status update
type StatusMsg struct {
	Message string
}

// LoadingMsg indicates loading state
type LoadingMsg struct {
	Loading bool
	Message string
}

// Data messages

// UserLoadedMsg is sent when user data is loaded
type UserLoadedMsg struct {
	User cfapi.User
}

// ProblemsLoadedMsg is sent when problems are loaded
type ProblemsLoadedMsg struct {
	Problems []cfapi.Problem
}

// SubmissionsLoadedMsg is sent when submissions are loaded
type SubmissionsLoadedMsg struct {
	Submissions []cfapi.Submission
}

// RatingLoadedMsg is sent when rating history is loaded
type RatingLoadedMsg struct {
	RatingChanges []cfapi.RatingChange
}

// ContestsLoadedMsg is sent when contests are loaded
type ContestsLoadedMsg struct {
	Contests []cfapi.Contest
}

// StatsLoadedMsg is sent when statistics are loaded
type StatsLoadedMsg struct {
	TotalSolved      int
	TotalSubmissions int
	AcceptanceRate   float64
	ByRating         map[int]int
	ByTag            map[string]int
	RecentSolved     int
	Streak           int
}

// WindowSizeMsg is sent when the window is resized
type WindowSizeMsg struct {
	Width  int
	Height int
}

// RefreshMsg triggers a data refresh
type RefreshMsg struct{}

// TickMsg is sent periodically for animations
type TickMsg struct{}
