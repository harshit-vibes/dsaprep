package workspace

import (
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/harshit-vibes/dsaprep/pkg/internal/schema/v1"
)

// ============ Additional Coverage Tests ============

func TestWorkspace_SaveManifest_NilManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// manifest is nil
	err := ws.SaveManifest()
	if err == nil {
		t.Error("SaveManifest() should error when manifest is nil")
	}
}

func TestWorkspace_LoadManifest_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create invalid YAML file - tabs in wrong places cause actual YAML errors
	manifestPath := filepath.Join(tmpDir, ManifestFile)
	err := os.WriteFile(manifestPath, []byte(":\t:\t:\ninvalid\n\t- bad indent"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = ws.LoadManifest()
	if err == nil {
		t.Error("LoadManifest() should error on invalid YAML")
	}
}

func TestWorkspace_Load_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create invalid manifest - tabs cause YAML errors
	manifestPath := filepath.Join(tmpDir, ManifestFile)
	err := os.WriteFile(manifestPath, []byte(":\t:\t:\ninvalid\n\t- bad indent"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	err = ws.Load()
	if err == nil {
		t.Error("Load() should error on invalid manifest")
	}
}

func TestWorkspace_GetSchemaVersion_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create invalid manifest
	manifestPath := filepath.Join(tmpDir, ManifestFile)
	err := os.WriteFile(manifestPath, []byte("not: valid: yaml: {{}}"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = ws.GetSchemaVersion()
	if err == nil {
		t.Error("GetSchemaVersion() should error on invalid manifest")
	}
}

func TestWorkspace_Validate_NotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(filepath.Join(tmpDir, "nonexistent"))

	err := ws.Validate()
	if err == nil {
		t.Error("Validate() should error when workspace not initialized")
	}
}

func TestWorkspace_Validate_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create invalid manifest
	manifestPath := filepath.Join(tmpDir, ManifestFile)
	err := os.WriteFile(manifestPath, []byte("invalid yaml {{}}"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	err = ws.Validate()
	if err == nil {
		t.Error("Validate() should error on invalid manifest")
	}
}

func TestWorkspace_Validate_InvalidSchemaVersion(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create manifest with invalid schema version
	manifestPath := filepath.Join(tmpDir, ManifestFile)
	content := `schema:
  version: "invalid.version.format"
name: test
handle: testuser
paths:
  problems: problems
  templates: templates
  submissions: submissions
  stats: stats
`
	err := os.WriteFile(manifestPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	err = ws.Validate()
	if err == nil {
		t.Error("Validate() should error on invalid schema version")
	}
}

func TestWorkspace_Validate_MissingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create valid manifest but don't create directories
	manifestPath := filepath.Join(tmpDir, ManifestFile)
	content := `schema:
  version: "1.0"
name: test
handle: testuser
paths:
  problems: problems
  templates: templates
  submissions: submissions
  stats: stats
`
	err := os.WriteFile(manifestPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	err = ws.Validate()
	if err == nil {
		t.Error("Validate() should error when directories are missing")
	}
}

func TestWorkspace_LoadProblem_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create problem directory with invalid YAML - tabs cause errors
	problemDir := ws.ProblemPath("codeforces", 1, "A")
	err := os.MkdirAll(problemDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	problemPath := filepath.Join(problemDir, "problem.yaml")
	err = os.WriteFile(problemPath, []byte(":\t:\t:\ninvalid\n\t- bad indent"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = ws.LoadProblem("codeforces", 1, "A")
	if err == nil {
		t.Error("LoadProblem() should error on invalid YAML")
	}
}

func TestWorkspace_LoadProblem_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	_, err := ws.LoadProblem("codeforces", 999999, "Z")
	if err == nil {
		t.Error("LoadProblem() should error when problem not found")
	}
}

func TestWorkspace_ListProblems_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create empty problems directory
	problemsDir := ws.ProblemsPath()
	err := os.MkdirAll(problemsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	problems, err := ws.ListProblems()
	if err != nil {
		t.Fatalf("ListProblems() error: %v", err)
	}

	if len(problems) != 0 {
		t.Errorf("Expected 0 problems, got %d", len(problems))
	}
}

func TestWorkspace_ListProblems_WithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Create problem directory with invalid YAML (should be skipped)
	problemDir := ws.ProblemPath("codeforces", 1, "A")
	err := os.MkdirAll(problemDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	problemPath := filepath.Join(problemDir, "problem.yaml")
	// Use invalid YAML that will cause parse error
	err = os.WriteFile(problemPath, []byte(":\t:\t:\ninvalid\n\t- bad indent"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Should not error - just skip invalid files
	problems, err := ws.ListProblems()
	if err != nil {
		t.Fatalf("ListProblems() error: %v", err)
	}

	// Invalid YAML should be skipped
	if len(problems) != 0 {
		t.Errorf("Expected 0 problems (invalid one skipped), got %d", len(problems))
	}
}

func TestWorkspace_SaveStatement_Success(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	problem := &v1.Problem{
		Platform:  "codeforces",
		ContestID: 1,
		Index:     "A",
		Name:      "Theatre Square",
		URL:       "https://codeforces.com/contest/1/problem/A",
		Limits: v1.ProblemLimits{
			TimeLimit:   "1s",
			MemoryLimit: "256MB",
		},
		Metadata: v1.ProblemMetadata{
			Rating: 1000,
			Tags:   []string{"math", "geometry"},
		},
		Samples: []v1.Sample{
			{Index: 1, Input: "1 1 1", Output: "1"},
		},
	}

	// Save problem first
	err := ws.SaveProblem(problem)
	if err != nil {
		t.Fatalf("SaveProblem() error: %v", err)
	}

	// Save statement
	statement := "This is the problem statement."
	err = ws.SaveStatement(problem, statement)
	if err != nil {
		t.Fatalf("SaveStatement() error: %v", err)
	}

	// Load statement
	loaded, err := ws.LoadStatement("codeforces", 1, "A")
	if err != nil {
		t.Fatalf("LoadStatement() error: %v", err)
	}

	if loaded == "" {
		t.Error("LoadStatement() returned empty string")
	}
}

func TestWorkspace_SaveNotes_ProblemNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	notes := &v1.UserNotes{
		Approach: "test",
	}

	err := ws.SaveNotes("codeforces", 999999, "Z", notes)
	if err == nil {
		t.Error("SaveNotes() should error when problem not found")
	}
}

func TestWorkspace_UpdatePractice_ProblemNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	practice := &v1.PracticeData{
		AttemptCount: 1,
	}

	err := ws.UpdatePractice("codeforces", 999999, "Z", practice)
	if err == nil {
		t.Error("UpdatePractice() should error when problem not found")
	}
}

func TestWorkspace_ProblemExists_False(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	exists := ws.ProblemExists("codeforces", 999999, "Z")
	if exists {
		t.Error("ProblemExists() should return false for non-existent problem")
	}
}

func TestWorkspace_ProblemsPath_WithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	// Set manifest with custom path
	ws.manifest = &v1.Workspace{
		Paths: v1.PathConfig{
			Problems: "custom_problems",
		},
	}

	path := ws.ProblemsPath()
	expected := filepath.Join(tmpDir, "custom_problems")
	if path != expected {
		t.Errorf("ProblemsPath() = %v, want %v", path, expected)
	}
}

func TestWorkspace_TemplatesPath_WithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	ws.manifest = &v1.Workspace{
		Paths: v1.PathConfig{
			Templates: "custom_templates",
		},
	}

	path := ws.TemplatesPath()
	expected := filepath.Join(tmpDir, "custom_templates")
	if path != expected {
		t.Errorf("TemplatesPath() = %v, want %v", path, expected)
	}
}

func TestWorkspace_SubmissionsPath_WithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	ws.manifest = &v1.Workspace{
		Paths: v1.PathConfig{
			Submissions: "custom_submissions",
		},
	}

	path := ws.SubmissionsPath()
	expected := filepath.Join(tmpDir, "custom_submissions")
	if path != expected {
		t.Errorf("SubmissionsPath() = %v, want %v", path, expected)
	}
}

func TestWorkspace_StatsPath_WithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	ws.manifest = &v1.Workspace{
		Paths: v1.PathConfig{
			Stats: "custom_stats",
		},
	}

	path := ws.StatsPath()
	expected := filepath.Join(tmpDir, "custom_stats")
	if path != expected {
		t.Errorf("StatsPath() = %v, want %v", path, expected)
	}
}

func TestWorkspace_LoadStatement_Missing(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	_, err := ws.LoadStatement("codeforces", 999999, "Z")
	if err == nil {
		t.Error("LoadStatement() should error when statement not found")
	}
}

func TestFormatStatement_WithTags(t *testing.T) {
	problem := &v1.Problem{
		Index: "A",
		Name:  "Test Problem",
		URL:   "https://codeforces.com/test",
		Metadata: v1.ProblemMetadata{
			Rating: 1000,
			Tags:   []string{"math", "dp"},
		},
		Limits: v1.ProblemLimits{
			TimeLimit:   "1s",
			MemoryLimit: "256MB",
		},
		Samples: []v1.Sample{
			{Index: 1, Input: "1", Output: "2"},
		},
	}

	result := formatStatement(problem, "Statement text")

	if result == "" {
		t.Error("formatStatement() returned empty string")
	}

	// Check content contains expected parts
	if !contains(result, "A. Test Problem") {
		t.Error("Missing problem title")
	}
	if !contains(result, "math, dp") {
		t.Error("Missing tags")
	}
	if !contains(result, "Example 1") {
		t.Error("Missing example")
	}
}

func TestFormatStatement_NoTags(t *testing.T) {
	problem := &v1.Problem{
		Index: "B",
		Name:  "No Tags Problem",
		URL:   "https://codeforces.com/test",
		Metadata: v1.ProblemMetadata{
			Rating: 800,
			Tags:   []string{}, // Empty tags
		},
		Limits: v1.ProblemLimits{
			TimeLimit:   "2s",
			MemoryLimit: "512MB",
		},
	}

	result := formatStatement(problem, "Statement text")

	if result == "" {
		t.Error("formatStatement() returned empty string")
	}
}

func TestFormatStatement_NoSamples(t *testing.T) {
	problem := &v1.Problem{
		Index: "C",
		Name:  "No Samples Problem",
		URL:   "https://codeforces.com/test",
		Metadata: v1.ProblemMetadata{
			Rating: 1200,
		},
		Limits: v1.ProblemLimits{
			TimeLimit:   "1s",
			MemoryLimit: "256MB",
		},
		Samples: []v1.Sample{}, // Empty samples
	}

	result := formatStatement(problem, "Statement text")

	if result == "" {
		t.Error("formatStatement() returned empty string")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
