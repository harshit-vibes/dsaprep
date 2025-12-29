// Package v1 contains schema definitions for version 1.x
package v1

import (
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
)

// Workspace represents the workspace manifest (workspace.yaml)
type Workspace struct {
	Schema schema.SchemaHeader `yaml:"_schema" json:"_schema"`

	// Workspace identity
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Timestamps
	CreatedAt time.Time `yaml:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `yaml:"updatedAt" json:"updatedAt"`

	// Platform configuration
	Codeforces CFConfig `yaml:"codeforces" json:"codeforces"`

	// Practice settings
	Practice PracticeConfig `yaml:"practice" json:"practice"`

	// Path configuration
	Paths PathConfig `yaml:"paths" json:"paths"`
}

// CFConfig holds Codeforces-specific settings
type CFConfig struct {
	Handle          string   `yaml:"handle" json:"handle"`
	DefaultLanguage string   `yaml:"defaultLanguage" json:"defaultLanguage"`
	PreferredTags   []string `yaml:"preferredTags,omitempty" json:"preferredTags,omitempty"`
}

// PracticeConfig holds practice session settings
type PracticeConfig struct {
	DifficultyMin int `yaml:"difficultyMin" json:"difficultyMin"`
	DifficultyMax int `yaml:"difficultyMax" json:"difficultyMax"`
	DailyGoal     int `yaml:"dailyGoal" json:"dailyGoal"`
	WeeklyGoal    int `yaml:"weeklyGoal,omitempty" json:"weeklyGoal,omitempty"`
}

// PathConfig holds relative paths within workspace
type PathConfig struct {
	Problems    string `yaml:"problems" json:"problems"`
	Templates   string `yaml:"templates" json:"templates"`
	Submissions string `yaml:"submissions" json:"submissions"`
	Stats       string `yaml:"stats" json:"stats"`
}

// NewWorkspace creates a new workspace with defaults
func NewWorkspace(name, handle string) *Workspace {
	now := time.Now()
	return &Workspace{
		Schema:      schema.NewSchemaHeader(schema.TypeWorkspace),
		Name:        name,
		CreatedAt:   now,
		UpdatedAt:   now,
		Codeforces: CFConfig{
			Handle:          handle,
			DefaultLanguage: "cpp",
		},
		Practice: PracticeConfig{
			DifficultyMin: 800,
			DifficultyMax: 1400,
			DailyGoal:     3,
		},
		Paths: PathConfig{
			Problems:    "problems",
			Templates:   "templates",
			Submissions: "submissions",
			Stats:       "stats",
		},
	}
}

// DefaultWorkspace returns a workspace with empty handle
func DefaultWorkspace() *Workspace {
	return NewWorkspace("DSA Practice", "")
}
