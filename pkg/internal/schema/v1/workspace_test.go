package v1

import (
	"testing"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
)

func TestNewWorkspace(t *testing.T) {
	ws := NewWorkspace("My Practice", "tourist")

	if ws.Name != "My Practice" {
		t.Errorf("NewWorkspace().Name = %v, want %v", ws.Name, "My Practice")
	}
	if ws.Codeforces.Handle != "tourist" {
		t.Errorf("NewWorkspace().Codeforces.Handle = %v, want %v", ws.Codeforces.Handle, "tourist")
	}
	if ws.Schema.Type != schema.TypeWorkspace {
		t.Errorf("NewWorkspace().Schema.Type = %v, want %v", ws.Schema.Type, schema.TypeWorkspace)
	}
	if ws.CreatedAt.IsZero() {
		t.Error("NewWorkspace().CreatedAt should not be zero")
	}
	if ws.UpdatedAt.IsZero() {
		t.Error("NewWorkspace().UpdatedAt should not be zero")
	}
}

func TestWorkspace_DefaultValues(t *testing.T) {
	ws := NewWorkspace("Test", "")

	// Check default language
	if ws.Codeforces.DefaultLanguage != "cpp" {
		t.Errorf("Default language = %v, want %v", ws.Codeforces.DefaultLanguage, "cpp")
	}

	// Check default difficulty range
	if ws.Practice.DifficultyMin != 800 {
		t.Errorf("DifficultyMin = %v, want %v", ws.Practice.DifficultyMin, 800)
	}
	if ws.Practice.DifficultyMax != 1400 {
		t.Errorf("DifficultyMax = %v, want %v", ws.Practice.DifficultyMax, 1400)
	}

	// Check default daily goal
	if ws.Practice.DailyGoal != 3 {
		t.Errorf("DailyGoal = %v, want %v", ws.Practice.DailyGoal, 3)
	}

	// Check default paths
	if ws.Paths.Problems != "problems" {
		t.Errorf("Paths.Problems = %v, want %v", ws.Paths.Problems, "problems")
	}
	if ws.Paths.Templates != "templates" {
		t.Errorf("Paths.Templates = %v, want %v", ws.Paths.Templates, "templates")
	}
	if ws.Paths.Submissions != "submissions" {
		t.Errorf("Paths.Submissions = %v, want %v", ws.Paths.Submissions, "submissions")
	}
	if ws.Paths.Stats != "stats" {
		t.Errorf("Paths.Stats = %v, want %v", ws.Paths.Stats, "stats")
	}
}

func TestCFConfig(t *testing.T) {
	cfg := CFConfig{
		Handle:          "tourist",
		DefaultLanguage: "cpp17",
		PreferredTags:   []string{"dp", "math"},
	}

	if cfg.Handle != "tourist" {
		t.Errorf("CFConfig.Handle = %v, want %v", cfg.Handle, "tourist")
	}
	if cfg.DefaultLanguage != "cpp17" {
		t.Errorf("CFConfig.DefaultLanguage = %v, want %v", cfg.DefaultLanguage, "cpp17")
	}
	if len(cfg.PreferredTags) != 2 {
		t.Errorf("len(CFConfig.PreferredTags) = %v, want %v", len(cfg.PreferredTags), 2)
	}
	if cfg.PreferredTags[0] != "dp" {
		t.Errorf("CFConfig.PreferredTags[0] = %v, want %v", cfg.PreferredTags[0], "dp")
	}
}

func TestPracticeConfig(t *testing.T) {
	cfg := PracticeConfig{
		DifficultyMin: 1000,
		DifficultyMax: 1800,
		DailyGoal:     5,
		WeeklyGoal:    20,
	}

	if cfg.DifficultyMin != 1000 {
		t.Errorf("PracticeConfig.DifficultyMin = %v, want %v", cfg.DifficultyMin, 1000)
	}
	if cfg.DifficultyMax != 1800 {
		t.Errorf("PracticeConfig.DifficultyMax = %v, want %v", cfg.DifficultyMax, 1800)
	}
	if cfg.DailyGoal != 5 {
		t.Errorf("PracticeConfig.DailyGoal = %v, want %v", cfg.DailyGoal, 5)
	}
	if cfg.WeeklyGoal != 20 {
		t.Errorf("PracticeConfig.WeeklyGoal = %v, want %v", cfg.WeeklyGoal, 20)
	}
}

func TestPathConfig(t *testing.T) {
	cfg := PathConfig{
		Problems:    "my-problems",
		Templates:   "my-templates",
		Submissions: "my-submissions",
		Stats:       "my-stats",
	}

	if cfg.Problems != "my-problems" {
		t.Errorf("PathConfig.Problems = %v, want %v", cfg.Problems, "my-problems")
	}
	if cfg.Templates != "my-templates" {
		t.Errorf("PathConfig.Templates = %v, want %v", cfg.Templates, "my-templates")
	}
}

func TestWorkspace_EmptyHandle(t *testing.T) {
	ws := NewWorkspace("Test", "")

	if ws.Codeforces.Handle != "" {
		t.Errorf("Empty handle should remain empty, got %v", ws.Codeforces.Handle)
	}
}

func TestDefaultWorkspace(t *testing.T) {
	ws := DefaultWorkspace()

	if ws == nil {
		t.Fatal("DefaultWorkspace() returned nil")
	}
	if ws.Name != "DSA Practice" {
		t.Errorf("DefaultWorkspace().Name = %v, want %v", ws.Name, "DSA Practice")
	}
	if ws.Codeforces.Handle != "" {
		t.Errorf("DefaultWorkspace().Codeforces.Handle = %v, want empty", ws.Codeforces.Handle)
	}
	if ws.Codeforces.DefaultLanguage != "cpp" {
		t.Errorf("DefaultWorkspace().Codeforces.DefaultLanguage = %v, want cpp", ws.Codeforces.DefaultLanguage)
	}
}

func TestWorkspace_SchemaHeader(t *testing.T) {
	ws := NewWorkspace("Test", "user")

	if ws.Schema.Version != "1.0.0" {
		t.Errorf("Schema.Version = %v, want 1.0.0", ws.Schema.Version)
	}
	if ws.Schema.Type != schema.TypeWorkspace {
		t.Errorf("Schema.Type = %v, want %v", ws.Schema.Type, schema.TypeWorkspace)
	}
}

func TestWorkspace_Timestamps(t *testing.T) {
	ws := NewWorkspace("Test", "user")

	if ws.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if ws.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
	if !ws.CreatedAt.Equal(ws.UpdatedAt) {
		t.Error("CreatedAt and UpdatedAt should be equal on creation")
	}
}

func TestCFConfig_EmptyTags(t *testing.T) {
	cfg := CFConfig{
		Handle:          "user",
		DefaultLanguage: "cpp",
	}

	if len(cfg.PreferredTags) != 0 {
		t.Errorf("PreferredTags should be empty, got %v", cfg.PreferredTags)
	}
}

func TestPracticeConfig_ZeroValues(t *testing.T) {
	cfg := PracticeConfig{}

	if cfg.DifficultyMin != 0 {
		t.Errorf("DifficultyMin should be 0, got %v", cfg.DifficultyMin)
	}
	if cfg.WeeklyGoal != 0 {
		t.Errorf("WeeklyGoal should be 0, got %v", cfg.WeeklyGoal)
	}
}
