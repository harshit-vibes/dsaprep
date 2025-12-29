package v1

import (
	"testing"
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
)

func TestNewProblem(t *testing.T) {
	problem := NewProblem(1325, "A", "EhAb AnD gCd")

	if problem.ContestID != 1325 {
		t.Errorf("NewProblem().ContestID = %v, want %v", problem.ContestID, 1325)
	}
	if problem.Index != "A" {
		t.Errorf("NewProblem().Index = %v, want %v", problem.Index, "A")
	}
	if problem.Name != "EhAb AnD gCd" {
		t.Errorf("NewProblem().Name = %v, want %v", problem.Name, "EhAb AnD gCd")
	}
	if problem.Platform != "codeforces" {
		t.Errorf("NewProblem().Platform = %v, want %v", problem.Platform, "codeforces")
	}
	if problem.Practice.Status != StatusUnseen {
		t.Errorf("NewProblem().Practice.Status = %v, want %v", problem.Practice.Status, StatusUnseen)
	}
	if problem.Schema.Type != schema.TypeProblem {
		t.Errorf("NewProblem().Schema.Type = %v, want %v", problem.Schema.Type, schema.TypeProblem)
	}
	if problem.FetchedAt.IsZero() {
		t.Error("NewProblem().FetchedAt should not be zero")
	}
}

func TestProblem_ID(t *testing.T) {
	tests := []struct {
		name      string
		contestID int
		index     string
		wantID    string
	}{
		{
			name:      "simple problem",
			contestID: 1,
			index:     "A",
			wantID:    "0001A",
		},
		{
			name:      "four digit contest",
			contestID: 1325,
			index:     "B",
			wantID:    "1325B",
		},
		{
			name:      "problem with number suffix",
			contestID: 1500,
			index:     "B2",
			wantID:    "1500B2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problem := NewProblem(tt.contestID, tt.index, "Test")
			if problem.ID != tt.wantID {
				t.Errorf("Problem.ID = %v, want %v", problem.ID, tt.wantID)
			}
		})
	}
}

func TestPracticeStatus(t *testing.T) {
	// Verify all statuses are defined
	statuses := []PracticeStatus{StatusUnseen, StatusAttempted, StatusSolved}
	for _, s := range statuses {
		if s == "" {
			t.Error("PracticeStatus should not be empty")
		}
	}
}

func TestProblemLimits(t *testing.T) {
	problem := NewProblem(1, "A", "Test")
	problem.Limits = ProblemLimits{
		TimeLimit:   "1 second",
		MemoryLimit: "256 megabytes",
		InputType:   "standard input",
		OutputType:  "standard output",
	}

	if problem.Limits.TimeLimit != "1 second" {
		t.Errorf("Limits.TimeLimit = %v, want %v", problem.Limits.TimeLimit, "1 second")
	}
	if problem.Limits.MemoryLimit != "256 megabytes" {
		t.Errorf("Limits.MemoryLimit = %v, want %v", problem.Limits.MemoryLimit, "256 megabytes")
	}
}

func TestProblemMetadata(t *testing.T) {
	problem := NewProblem(1, "A", "Test")
	problem.Metadata = ProblemMetadata{
		Rating:      1400,
		Tags:        []string{"math", "dp"},
		SolvedCount: 10000,
	}

	if problem.Metadata.Rating != 1400 {
		t.Errorf("Metadata.Rating = %v, want %v", problem.Metadata.Rating, 1400)
	}
	if len(problem.Metadata.Tags) != 2 {
		t.Errorf("len(Metadata.Tags) = %v, want %v", len(problem.Metadata.Tags), 2)
	}
}

func TestSample(t *testing.T) {
	sample := Sample{
		Index:  1,
		Input:  "3\n",
		Output: "4\n",
	}

	if sample.Index != 1 {
		t.Errorf("Sample.Index = %v, want %v", sample.Index, 1)
	}
	if sample.Input != "3\n" {
		t.Errorf("Sample.Input = %v, want %v", sample.Input, "3\n")
	}
	if sample.Output != "4\n" {
		t.Errorf("Sample.Output = %v, want %v", sample.Output, "4\n")
	}
}

func TestPracticeData(t *testing.T) {
	now := time.Now()
	submissionID := int64(123456)

	practice := PracticeData{
		Status:         StatusSolved,
		FirstAttempt:   &now,
		SolvedAt:       &now,
		AttemptCount:   3,
		BestSubmission: &submissionID,
		TimeSpent:      600, // 10 minutes
	}

	if practice.Status != StatusSolved {
		t.Errorf("PracticeData.Status = %v, want %v", practice.Status, StatusSolved)
	}
	if practice.AttemptCount != 3 {
		t.Errorf("PracticeData.AttemptCount = %v, want %v", practice.AttemptCount, 3)
	}
	if *practice.BestSubmission != submissionID {
		t.Errorf("PracticeData.BestSubmission = %v, want %v", *practice.BestSubmission, submissionID)
	}
}

func TestUserNotes(t *testing.T) {
	notes := UserNotes{
		Difficulty: "medium",
		CustomTags: []string{"review", "tricky"},
		Approach:   "Use binary search",
		Reminder:   "Watch for edge cases",
		Review:     true,
	}

	if notes.Difficulty != "medium" {
		t.Errorf("UserNotes.Difficulty = %v, want %v", notes.Difficulty, "medium")
	}
	if !notes.Review {
		t.Error("UserNotes.Review should be true")
	}
	if len(notes.CustomTags) != 2 {
		t.Errorf("len(UserNotes.CustomTags) = %v, want %v", len(notes.CustomTags), 2)
	}
}

func TestFormatProblemID(t *testing.T) {
	tests := []struct {
		contestID int
		index     string
		want      string
	}{
		{1, "A", "0001A"},
		{12, "B", "0012B"},
		{123, "C", "0123C"},
		{1234, "D", "1234D"},
		{1, "B1", "0001B1"},
		{9999, "E", "9999E"},
	}

	for _, tt := range tests {
		got := formatProblemID(tt.contestID, tt.index)
		if got != tt.want {
			t.Errorf("formatProblemID(%d, %s) = %v, want %v", tt.contestID, tt.index, got, tt.want)
		}
	}
}

func TestProblem_FullStruct(t *testing.T) {
	now := time.Now()
	submissionID := int64(123456)

	problem := &Problem{
		Schema:    schema.NewSchemaHeader(schema.TypeProblem),
		ID:        "1325A",
		Platform:  "codeforces",
		ContestID: 1325,
		Index:     "A",
		Name:      "EhAb AnD gCd",
		URL:       "https://codeforces.com/problemset/problem/1325/A",
		Metadata: ProblemMetadata{
			Rating:      800,
			Tags:        []string{"constructive algorithms", "greedy", "number theory"},
			SolvedCount: 50000,
		},
		Limits: ProblemLimits{
			TimeLimit:   "1 second",
			MemoryLimit: "256 megabytes",
			InputType:   "standard input",
			OutputType:  "standard output",
		},
		Samples: []Sample{
			{Index: 1, Input: "1\n", Output: "1 0\n"},
			{Index: 2, Input: "2\n", Output: "1 1\n"},
		},
		Practice: PracticeData{
			Status:         StatusSolved,
			FirstAttempt:   &now,
			SolvedAt:       &now,
			AttemptCount:   2,
			BestSubmission: &submissionID,
			TimeSpent:      300,
		},
		Notes: UserNotes{
			Difficulty: "easy",
			CustomTags: []string{"math"},
			Approach:   "GCD approach",
		},
		FetchedAt:   now,
		FetchMethod: "web",
	}

	// Verify all fields
	if problem.Schema.Type != schema.TypeProblem {
		t.Errorf("Schema.Type = %v, want %v", problem.Schema.Type, schema.TypeProblem)
	}
	if problem.ID != "1325A" {
		t.Errorf("ID = %v, want 1325A", problem.ID)
	}
	if problem.URL == "" {
		t.Error("URL should not be empty")
	}
	if len(problem.Samples) != 2 {
		t.Errorf("len(Samples) = %v, want 2", len(problem.Samples))
	}
	if problem.FetchMethod != "web" {
		t.Errorf("FetchMethod = %v, want web", problem.FetchMethod)
	}
}

func TestPracticeStatus_Values(t *testing.T) {
	// Test status string values
	if string(StatusUnseen) != "unseen" {
		t.Errorf("StatusUnseen = %v, want unseen", StatusUnseen)
	}
	if string(StatusAttempted) != "attempted" {
		t.Errorf("StatusAttempted = %v, want attempted", StatusAttempted)
	}
	if string(StatusSolved) != "solved" {
		t.Errorf("StatusSolved = %v, want solved", StatusSolved)
	}
}

func TestProblem_WithSamples(t *testing.T) {
	problem := NewProblem(1, "A", "Theatre Square")
	problem.Samples = []Sample{
		{Index: 1, Input: "6 6 4\n", Output: "4\n"},
	}

	if len(problem.Samples) != 1 {
		t.Errorf("len(Samples) = %v, want 1", len(problem.Samples))
	}
	if problem.Samples[0].Input != "6 6 4\n" {
		t.Errorf("Samples[0].Input = %v, want '6 6 4\\n'", problem.Samples[0].Input)
	}
}

func TestProblem_EmptyNotes(t *testing.T) {
	problem := NewProblem(1, "A", "Test")

	// Notes should be empty by default
	if problem.Notes.Difficulty != "" {
		t.Errorf("Notes.Difficulty should be empty by default")
	}
	if problem.Notes.Review {
		t.Error("Notes.Review should be false by default")
	}
	if len(problem.Notes.CustomTags) != 0 {
		t.Errorf("Notes.CustomTags should be empty by default")
	}
}
