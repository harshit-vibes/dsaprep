package v1

import (
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
)

// Problem represents a problem stored locally (problem.yaml)
type Problem struct {
	Schema schema.SchemaHeader `yaml:"_schema" json:"_schema"`

	// Identity
	ID        string `yaml:"id" json:"id"`               // e.g., "1325A"
	Platform  string `yaml:"platform" json:"platform"`   // "codeforces"
	ContestID int    `yaml:"contestId" json:"contestId"` // e.g., 1325
	Index     string `yaml:"index" json:"index"`         // e.g., "A"

	// Basic info
	Name string `yaml:"name" json:"name"`
	URL  string `yaml:"url" json:"url"`

	// Metadata from platform
	Metadata ProblemMetadata `yaml:"metadata" json:"metadata"`

	// Constraints
	Limits ProblemLimits `yaml:"limits" json:"limits"`

	// Sample test cases
	Samples []Sample `yaml:"samples" json:"samples"`

	// User practice data
	Practice PracticeData `yaml:"practice" json:"practice"`

	// User notes
	Notes UserNotes `yaml:"notes,omitempty" json:"notes,omitempty"`

	// Fetch metadata
	FetchedAt   time.Time `yaml:"fetchedAt" json:"fetchedAt"`
	FetchMethod string    `yaml:"fetchMethod" json:"fetchMethod"` // "api", "web", "cache"
}

// ProblemMetadata holds platform-provided metadata
type ProblemMetadata struct {
	Rating      int      `yaml:"rating" json:"rating"`
	Tags        []string `yaml:"tags" json:"tags"`
	SolvedCount int      `yaml:"solvedCount,omitempty" json:"solvedCount,omitempty"`
}

// ProblemLimits holds problem constraints
type ProblemLimits struct {
	TimeLimit   string `yaml:"timeLimit" json:"timeLimit"`
	MemoryLimit string `yaml:"memoryLimit" json:"memoryLimit"`
	InputType   string `yaml:"inputType" json:"inputType"`   // "standard input" or filename
	OutputType  string `yaml:"outputType" json:"outputType"` // "standard output" or filename
}

// Sample represents a sample test case
type Sample struct {
	Index  int    `yaml:"index" json:"index"`
	Input  string `yaml:"input" json:"input"`
	Output string `yaml:"output" json:"output"`
}

// PracticeData holds user's practice progress
type PracticeData struct {
	Status         PracticeStatus `yaml:"status" json:"status"`
	FirstAttempt   *time.Time     `yaml:"firstAttempt,omitempty" json:"firstAttempt,omitempty"`
	SolvedAt       *time.Time     `yaml:"solvedAt,omitempty" json:"solvedAt,omitempty"`
	AttemptCount   int            `yaml:"attemptCount" json:"attemptCount"`
	BestSubmission *int64         `yaml:"bestSubmission,omitempty" json:"bestSubmission,omitempty"`
	TimeSpent      int            `yaml:"timeSpent,omitempty" json:"timeSpent,omitempty"` // seconds
}

// PracticeStatus represents the practice state
type PracticeStatus string

const (
	StatusUnseen    PracticeStatus = "unseen"
	StatusAttempted PracticeStatus = "attempted"
	StatusSolved    PracticeStatus = "solved"
)

// UserNotes holds user's personal notes
type UserNotes struct {
	Difficulty string   `yaml:"difficulty,omitempty" json:"difficulty,omitempty"` // User perception
	CustomTags []string `yaml:"customTags,omitempty" json:"customTags,omitempty"`
	Approach   string   `yaml:"approach,omitempty" json:"approach,omitempty"`
	Reminder   string   `yaml:"reminder,omitempty" json:"reminder,omitempty"`
	Review     bool     `yaml:"review,omitempty" json:"review,omitempty"`
}

// NewProblem creates a new problem with defaults
func NewProblem(contestID int, index, name string) *Problem {
	return &Problem{
		Schema:    schema.NewSchemaHeader(schema.TypeProblem),
		ID:        formatProblemID(contestID, index),
		Platform:  "codeforces",
		ContestID: contestID,
		Index:     index,
		Name:      name,
		Practice: PracticeData{
			Status: StatusUnseen,
		},
		FetchedAt: time.Now(),
	}
}

func formatProblemID(contestID int, index string) string {
	return string(rune('0'+contestID/1000)) + string(rune('0'+(contestID/100)%10)) +
		string(rune('0'+(contestID/10)%10)) + string(rune('0'+contestID%10)) + index
}
