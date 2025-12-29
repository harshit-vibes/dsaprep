package v1

import (
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
)

// Progress represents overall progress tracking
type Progress struct {
	Schema schema.SchemaHeader `yaml:"_schema" json:"_schema"`

	// Lifetime stats
	TotalSolved    int `yaml:"totalSolved" json:"totalSolved"`
	TotalAttempted int `yaml:"totalAttempted" json:"totalAttempted"`
	TotalTime      int `yaml:"totalTime" json:"totalTime"` // seconds

	// Rating distribution of solved problems
	RatingDistribution map[string]int `yaml:"ratingDistribution" json:"ratingDistribution"`

	// Tag distribution
	TagDistribution map[string]int `yaml:"tagDistribution" json:"tagDistribution"`

	// Streak tracking
	CurrentStreak int        `yaml:"currentStreak" json:"currentStreak"`
	LongestStreak int        `yaml:"longestStreak" json:"longestStreak"`
	LastActivity  *time.Time `yaml:"lastActivity,omitempty" json:"lastActivity,omitempty"`

	// Daily entries
	Daily []DailyProgress `yaml:"daily,omitempty" json:"daily,omitempty"`
}

// DailyProgress represents a single day's progress
type DailyProgress struct {
	Date     string   `yaml:"date" json:"date"` // YYYY-MM-DD
	Solved   int      `yaml:"solved" json:"solved"`
	Attempted int     `yaml:"attempted" json:"attempted"`
	Problems []string `yaml:"problems" json:"problems"` // Problem IDs
	TimeSpent int     `yaml:"timeSpent" json:"timeSpent"` // seconds
}

// NewProgress creates a new progress tracker
func NewProgress() *Progress {
	return &Progress{
		Schema:             schema.NewSchemaHeader(schema.TypeProgress),
		RatingDistribution: make(map[string]int),
		TagDistribution:    make(map[string]int),
		Daily:              []DailyProgress{},
	}
}

// AddSolved records a solved problem
func (p *Progress) AddSolved(problemID string, rating int, tags []string, timeSpent int) {
	p.TotalSolved++
	p.TotalTime += timeSpent

	// Update rating distribution
	ratingBucket := getRatingBucket(rating)
	p.RatingDistribution[ratingBucket]++

	// Update tag distribution
	for _, tag := range tags {
		p.TagDistribution[tag]++
	}

	// Update streak
	now := time.Now()
	if p.LastActivity != nil {
		daysDiff := daysBetween(*p.LastActivity, now)
		if daysDiff == 1 {
			p.CurrentStreak++
		} else if daysDiff > 1 {
			p.CurrentStreak = 1
		}
	} else {
		p.CurrentStreak = 1
	}

	if p.CurrentStreak > p.LongestStreak {
		p.LongestStreak = p.CurrentStreak
	}

	p.LastActivity = &now

	// Update daily
	p.updateDaily(problemID, true, timeSpent)
}

// AddAttempted records an attempted problem
func (p *Progress) AddAttempted(problemID string, timeSpent int) {
	p.TotalAttempted++
	p.TotalTime += timeSpent
	p.updateDaily(problemID, false, timeSpent)
}

func (p *Progress) updateDaily(problemID string, solved bool, timeSpent int) {
	today := time.Now().Format("2006-01-02")

	// Find or create today's entry
	var todayEntry *DailyProgress
	for i := range p.Daily {
		if p.Daily[i].Date == today {
			todayEntry = &p.Daily[i]
			break
		}
	}

	if todayEntry == nil {
		p.Daily = append(p.Daily, DailyProgress{Date: today})
		todayEntry = &p.Daily[len(p.Daily)-1]
	}

	if solved {
		todayEntry.Solved++
	} else {
		todayEntry.Attempted++
	}
	todayEntry.Problems = append(todayEntry.Problems, problemID)
	todayEntry.TimeSpent += timeSpent
}

func getRatingBucket(rating int) string {
	switch {
	case rating < 1000:
		return "800-999"
	case rating < 1200:
		return "1000-1199"
	case rating < 1400:
		return "1200-1399"
	case rating < 1600:
		return "1400-1599"
	case rating < 1900:
		return "1600-1899"
	case rating < 2100:
		return "1900-2099"
	case rating < 2400:
		return "2100-2399"
	default:
		return "2400+"
	}
}

func daysBetween(a, b time.Time) int {
	aDate := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
	bDate := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
	return int(bDate.Sub(aDate).Hours() / 24)
}
