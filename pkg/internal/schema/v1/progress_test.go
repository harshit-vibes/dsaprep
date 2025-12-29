package v1

import (
	"testing"
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
	"gopkg.in/yaml.v3"
)

func TestNewProgress(t *testing.T) {
	p := NewProgress()

	if p == nil {
		t.Fatal("NewProgress() returned nil")
	}
	if p.Schema.Type != schema.TypeProgress {
		t.Errorf("Schema.Type = %v, want %v", p.Schema.Type, schema.TypeProgress)
	}
	if p.TotalSolved != 0 {
		t.Errorf("TotalSolved = %v, want 0", p.TotalSolved)
	}
	if p.TotalAttempted != 0 {
		t.Errorf("TotalAttempted = %v, want 0", p.TotalAttempted)
	}
	if p.TotalTime != 0 {
		t.Errorf("TotalTime = %v, want 0", p.TotalTime)
	}
	if p.CurrentStreak != 0 {
		t.Errorf("CurrentStreak = %v, want 0", p.CurrentStreak)
	}
	if p.LongestStreak != 0 {
		t.Errorf("LongestStreak = %v, want 0", p.LongestStreak)
	}
	if p.LastActivity != nil {
		t.Error("LastActivity should be nil initially")
	}
	if p.RatingDistribution == nil {
		t.Error("RatingDistribution should be initialized")
	}
	if p.TagDistribution == nil {
		t.Error("TagDistribution should be initialized")
	}
	if p.Daily == nil {
		t.Error("Daily should be initialized")
	}
}

func TestProgress_AddSolved(t *testing.T) {
	p := NewProgress()

	p.AddSolved("1325A", 800, []string{"math", "number theory"}, 300)

	if p.TotalSolved != 1 {
		t.Errorf("TotalSolved = %v, want 1", p.TotalSolved)
	}
	if p.TotalTime != 300 {
		t.Errorf("TotalTime = %v, want 300", p.TotalTime)
	}
	if p.RatingDistribution["800-999"] != 1 {
		t.Errorf("RatingDistribution[800-999] = %v, want 1", p.RatingDistribution["800-999"])
	}
	if p.TagDistribution["math"] != 1 {
		t.Errorf("TagDistribution[math] = %v, want 1", p.TagDistribution["math"])
	}
	if p.TagDistribution["number theory"] != 1 {
		t.Errorf("TagDistribution[number theory] = %v, want 1", p.TagDistribution["number theory"])
	}
	if p.CurrentStreak != 1 {
		t.Errorf("CurrentStreak = %v, want 1", p.CurrentStreak)
	}
	if p.LongestStreak != 1 {
		t.Errorf("LongestStreak = %v, want 1", p.LongestStreak)
	}
	if p.LastActivity == nil {
		t.Error("LastActivity should be set")
	}
}

func TestProgress_AddSolved_MultipleTags(t *testing.T) {
	p := NewProgress()

	p.AddSolved("1A", 1200, []string{"dp", "math"}, 100)
	p.AddSolved("2A", 1200, []string{"dp", "greedy"}, 100)

	if p.TotalSolved != 2 {
		t.Errorf("TotalSolved = %v, want 2", p.TotalSolved)
	}
	if p.TagDistribution["dp"] != 2 {
		t.Errorf("TagDistribution[dp] = %v, want 2", p.TagDistribution["dp"])
	}
	if p.TagDistribution["math"] != 1 {
		t.Errorf("TagDistribution[math] = %v, want 1", p.TagDistribution["math"])
	}
	if p.TagDistribution["greedy"] != 1 {
		t.Errorf("TagDistribution[greedy] = %v, want 1", p.TagDistribution["greedy"])
	}
}

func TestProgress_AddAttempted(t *testing.T) {
	p := NewProgress()

	p.AddAttempted("1325A", 120)

	if p.TotalAttempted != 1 {
		t.Errorf("TotalAttempted = %v, want 1", p.TotalAttempted)
	}
	if p.TotalTime != 120 {
		t.Errorf("TotalTime = %v, want 120", p.TotalTime)
	}
	if p.TotalSolved != 0 {
		t.Errorf("TotalSolved = %v, want 0", p.TotalSolved)
	}
}

func TestProgress_Streak_FirstDay(t *testing.T) {
	p := NewProgress()

	// First solve should set streak to 1
	p.AddSolved("1A", 800, nil, 0)

	if p.CurrentStreak != 1 {
		t.Errorf("CurrentStreak = %v, want 1", p.CurrentStreak)
	}
	if p.LongestStreak != 1 {
		t.Errorf("LongestStreak = %v, want 1", p.LongestStreak)
	}
}

func TestProgress_Streak_SameDay(t *testing.T) {
	p := NewProgress()

	// Multiple solves on the same day shouldn't increment streak
	p.AddSolved("1A", 800, nil, 0)
	initialStreak := p.CurrentStreak

	// Simulate second solve on same day (LastActivity is already today)
	p.AddSolved("2A", 800, nil, 0)

	// Since daysBetween returns 0 for same day, streak stays the same
	if p.CurrentStreak != initialStreak {
		t.Errorf("CurrentStreak = %v, want %v (same day)", p.CurrentStreak, initialStreak)
	}
}

func TestProgress_updateDaily(t *testing.T) {
	p := NewProgress()

	p.AddSolved("1A", 800, nil, 100)

	if len(p.Daily) != 1 {
		t.Fatalf("len(Daily) = %v, want 1", len(p.Daily))
	}

	today := time.Now().Format("2006-01-02")
	if p.Daily[0].Date != today {
		t.Errorf("Daily[0].Date = %v, want %v", p.Daily[0].Date, today)
	}
	if p.Daily[0].Solved != 1 {
		t.Errorf("Daily[0].Solved = %v, want 1", p.Daily[0].Solved)
	}
	if p.Daily[0].Attempted != 0 {
		t.Errorf("Daily[0].Attempted = %v, want 0", p.Daily[0].Attempted)
	}
	if len(p.Daily[0].Problems) != 1 {
		t.Errorf("len(Daily[0].Problems) = %v, want 1", len(p.Daily[0].Problems))
	}
	if p.Daily[0].Problems[0] != "1A" {
		t.Errorf("Daily[0].Problems[0] = %v, want 1A", p.Daily[0].Problems[0])
	}
	if p.Daily[0].TimeSpent != 100 {
		t.Errorf("Daily[0].TimeSpent = %v, want 100", p.Daily[0].TimeSpent)
	}
}

func TestProgress_updateDaily_MultipleSolves(t *testing.T) {
	p := NewProgress()

	p.AddSolved("1A", 800, nil, 100)
	p.AddSolved("2A", 900, nil, 150)
	p.AddAttempted("3A", 50)

	if len(p.Daily) != 1 {
		t.Fatalf("len(Daily) = %v, want 1 (same day)", len(p.Daily))
	}

	if p.Daily[0].Solved != 2 {
		t.Errorf("Daily[0].Solved = %v, want 2", p.Daily[0].Solved)
	}
	if p.Daily[0].Attempted != 1 {
		t.Errorf("Daily[0].Attempted = %v, want 1", p.Daily[0].Attempted)
	}
	if len(p.Daily[0].Problems) != 3 {
		t.Errorf("len(Daily[0].Problems) = %v, want 3", len(p.Daily[0].Problems))
	}
	if p.Daily[0].TimeSpent != 300 {
		t.Errorf("Daily[0].TimeSpent = %v, want 300", p.Daily[0].TimeSpent)
	}
}

func TestGetRatingBucket(t *testing.T) {
	tests := []struct {
		rating int
		want   string
	}{
		{0, "800-999"},
		{500, "800-999"},
		{800, "800-999"},
		{999, "800-999"},
		{1000, "1000-1199"},
		{1100, "1000-1199"},
		{1199, "1000-1199"},
		{1200, "1200-1399"},
		{1300, "1200-1399"},
		{1399, "1200-1399"},
		{1400, "1400-1599"},
		{1500, "1400-1599"},
		{1599, "1400-1599"},
		{1600, "1600-1899"},
		{1800, "1600-1899"},
		{1899, "1600-1899"},
		{1900, "1900-2099"},
		{2000, "1900-2099"},
		{2099, "1900-2099"},
		{2100, "2100-2399"},
		{2200, "2100-2399"},
		{2399, "2100-2399"},
		{2400, "2400+"},
		{2500, "2400+"},
		{3000, "2400+"},
		{3500, "2400+"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.rating)), func(t *testing.T) {
			if got := getRatingBucket(tt.rating); got != tt.want {
				t.Errorf("getRatingBucket(%d) = %v, want %v", tt.rating, got, tt.want)
			}
		})
	}
}

func TestDaysBetween(t *testing.T) {
	tests := []struct {
		name     string
		a        time.Time
		b        time.Time
		expected int
	}{
		{
			name:     "same day",
			a:        time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			b:        time.Date(2024, 1, 15, 23, 0, 0, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "consecutive days",
			a:        time.Date(2024, 1, 15, 23, 0, 0, 0, time.UTC),
			b:        time.Date(2024, 1, 16, 1, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "one day apart",
			a:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			b:        time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "two days apart",
			a:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			b:        time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
			expected: 2,
		},
		{
			name:     "week apart",
			a:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			b:        time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
			expected: 7,
		},
		{
			name:     "across months",
			a:        time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			b:        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "across years",
			a:        time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			b:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := daysBetween(tt.a, tt.b); got != tt.expected {
				t.Errorf("daysBetween() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProgress_RatingDistribution(t *testing.T) {
	p := NewProgress()

	// Add problems with different ratings
	p.AddSolved("1A", 800, nil, 0)
	p.AddSolved("2A", 1000, nil, 0)
	p.AddSolved("3A", 1200, nil, 0)
	p.AddSolved("4A", 1200, nil, 0)
	p.AddSolved("5A", 2500, nil, 0)

	if p.RatingDistribution["800-999"] != 1 {
		t.Errorf("RatingDistribution[800-999] = %v, want 1", p.RatingDistribution["800-999"])
	}
	if p.RatingDistribution["1000-1199"] != 1 {
		t.Errorf("RatingDistribution[1000-1199] = %v, want 1", p.RatingDistribution["1000-1199"])
	}
	if p.RatingDistribution["1200-1399"] != 2 {
		t.Errorf("RatingDistribution[1200-1399] = %v, want 2", p.RatingDistribution["1200-1399"])
	}
	if p.RatingDistribution["2400+"] != 1 {
		t.Errorf("RatingDistribution[2400+] = %v, want 1", p.RatingDistribution["2400+"])
	}
}

func TestProgress_YAML_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := &Progress{
		Schema:             schema.NewSchemaHeader(schema.TypeProgress),
		TotalSolved:        10,
		TotalAttempted:     5,
		TotalTime:          3600,
		RatingDistribution: map[string]int{"800-999": 5, "1000-1199": 5},
		TagDistribution:    map[string]int{"math": 3, "dp": 7},
		CurrentStreak:      5,
		LongestStreak:      10,
		LastActivity:       &now,
		Daily: []DailyProgress{
			{Date: "2024-01-15", Solved: 2, Attempted: 1, Problems: []string{"1A", "2A"}, TimeSpent: 600},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	// Unmarshal back
	var decoded Progress
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}

	// Verify fields
	if decoded.TotalSolved != original.TotalSolved {
		t.Errorf("TotalSolved = %v, want %v", decoded.TotalSolved, original.TotalSolved)
	}
	if decoded.CurrentStreak != original.CurrentStreak {
		t.Errorf("CurrentStreak = %v, want %v", decoded.CurrentStreak, original.CurrentStreak)
	}
	if decoded.RatingDistribution["800-999"] != 5 {
		t.Errorf("RatingDistribution[800-999] = %v, want 5", decoded.RatingDistribution["800-999"])
	}
	if len(decoded.Daily) != 1 {
		t.Errorf("len(Daily) = %v, want 1", len(decoded.Daily))
	}
}

func TestDailyProgress_Fields(t *testing.T) {
	dp := DailyProgress{
		Date:      "2024-01-15",
		Solved:    3,
		Attempted: 2,
		Problems:  []string{"1A", "2A", "3A", "4A", "5A"},
		TimeSpent: 1800,
	}

	if dp.Date != "2024-01-15" {
		t.Errorf("Date = %v, want 2024-01-15", dp.Date)
	}
	if dp.Solved != 3 {
		t.Errorf("Solved = %v, want 3", dp.Solved)
	}
	if dp.Attempted != 2 {
		t.Errorf("Attempted = %v, want 2", dp.Attempted)
	}
	if len(dp.Problems) != 5 {
		t.Errorf("len(Problems) = %v, want 5", len(dp.Problems))
	}
	if dp.TimeSpent != 1800 {
		t.Errorf("TimeSpent = %v, want 1800", dp.TimeSpent)
	}
}

func TestProgress_EmptyTags(t *testing.T) {
	p := NewProgress()

	// Add solved with nil tags
	p.AddSolved("1A", 800, nil, 100)

	if len(p.TagDistribution) != 0 {
		t.Errorf("TagDistribution should be empty with nil tags, got %v", p.TagDistribution)
	}

	// Add solved with empty tags slice
	p.AddSolved("2A", 800, []string{}, 100)

	if len(p.TagDistribution) != 0 {
		t.Errorf("TagDistribution should be empty with empty tags, got %v", p.TagDistribution)
	}
}

func TestProgress_LongestStreakUpdate(t *testing.T) {
	p := NewProgress()

	// Simulate building up a streak
	p.CurrentStreak = 5
	p.LongestStreak = 5

	// Manually set LastActivity to yesterday to trigger streak increment
	yesterday := time.Now().AddDate(0, 0, -1)
	p.LastActivity = &yesterday

	p.AddSolved("1A", 800, nil, 0)

	if p.CurrentStreak != 6 {
		t.Errorf("CurrentStreak = %v, want 6", p.CurrentStreak)
	}
	if p.LongestStreak != 6 {
		t.Errorf("LongestStreak = %v, want 6", p.LongestStreak)
	}
}

func TestProgress_StreakReset(t *testing.T) {
	p := NewProgress()

	// Simulate having a streak
	p.CurrentStreak = 5
	p.LongestStreak = 10

	// Set LastActivity to 3 days ago to trigger streak reset
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	p.LastActivity = &threeDaysAgo

	p.AddSolved("1A", 800, nil, 0)

	if p.CurrentStreak != 1 {
		t.Errorf("CurrentStreak = %v, want 1 (reset)", p.CurrentStreak)
	}
	// LongestStreak should not change
	if p.LongestStreak != 10 {
		t.Errorf("LongestStreak = %v, want 10 (unchanged)", p.LongestStreak)
	}
}
