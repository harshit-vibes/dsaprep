package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/harshit-vibes/cf/pkg/internal/schema/v1"
)

func TestWorkspace_SaveProblem(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	problem := v1.NewProblem(1325, "A", "EhAb AnD gCd")
	problem.Samples = []v1.Sample{
		{Index: 1, Input: "1\n", Output: "2\n"},
		{Index: 2, Input: "2\n", Output: "3 1\n"},
	}

	err = ws.SaveProblem(problem)
	if err != nil {
		t.Fatalf("SaveProblem() error = %v", err)
	}

	// Check problem.yaml exists
	problemPath := filepath.Join(ws.ProblemPath("codeforces", 1325, "A"), "problem.yaml")
	if _, err := os.Stat(problemPath); os.IsNotExist(err) {
		t.Error("SaveProblem() did not create problem.yaml")
	}

	// Check tests directory exists with samples
	testsDir := filepath.Join(ws.ProblemPath("codeforces", 1325, "A"), "tests")
	if _, err := os.Stat(testsDir); os.IsNotExist(err) {
		t.Error("SaveProblem() did not create tests directory")
	}

	// Check sample files
	for i := 1; i <= 2; i++ {
		inputPath := filepath.Join(testsDir, "sample_1.in")
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Errorf("SaveProblem() did not create sample_%d.in", i)
		}
		outputPath := filepath.Join(testsDir, "sample_1.out")
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("SaveProblem() did not create sample_%d.out", i)
		}
	}

	// Check solutions directory exists
	solutionsDir := filepath.Join(ws.ProblemPath("codeforces", 1325, "A"), "solutions")
	if _, err := os.Stat(solutionsDir); os.IsNotExist(err) {
		t.Error("SaveProblem() did not create solutions directory")
	}
}

func TestWorkspace_LoadProblem(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Save problem first
	problem := v1.NewProblem(1325, "A", "EhAb AnD gCd")
	problem.Metadata.Rating = 800
	problem.Metadata.Tags = []string{"math", "number theory"}

	err = ws.SaveProblem(problem)
	if err != nil {
		t.Fatalf("SaveProblem() error = %v", err)
	}

	// Load problem
	loaded, err := ws.LoadProblem("codeforces", 1325, "A")
	if err != nil {
		t.Fatalf("LoadProblem() error = %v", err)
	}

	if loaded.ContestID != 1325 {
		t.Errorf("LoadProblem().ContestID = %v, want %v", loaded.ContestID, 1325)
	}
	if loaded.Index != "A" {
		t.Errorf("LoadProblem().Index = %v, want %v", loaded.Index, "A")
	}
	if loaded.Name != "EhAb AnD gCd" {
		t.Errorf("LoadProblem().Name = %v, want %v", loaded.Name, "EhAb AnD gCd")
	}
	if loaded.Metadata.Rating != 800 {
		t.Errorf("LoadProblem().Metadata.Rating = %v, want %v", loaded.Metadata.Rating, 800)
	}
}

func TestWorkspace_LoadProblem_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	_, err = ws.LoadProblem("codeforces", 9999, "Z")
	if err == nil {
		t.Error("LoadProblem() should return error for non-existent problem")
	}
}

func TestWorkspace_ProblemExists(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Should not exist before saving
	if ws.ProblemExists("codeforces", 1325, "A") {
		t.Error("ProblemExists() should return false before saving")
	}

	// Save problem
	problem := v1.NewProblem(1325, "A", "Test")
	err = ws.SaveProblem(problem)
	if err != nil {
		t.Fatalf("SaveProblem() error = %v", err)
	}

	// Should exist after saving
	if !ws.ProblemExists("codeforces", 1325, "A") {
		t.Error("ProblemExists() should return true after saving")
	}
}

func TestWorkspace_ListProblems(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Save multiple problems
	problems := []*v1.Problem{
		v1.NewProblem(1325, "A", "Problem A"),
		v1.NewProblem(1325, "B", "Problem B"),
		v1.NewProblem(1400, "A", "Problem C"),
	}

	for _, p := range problems {
		if err := ws.SaveProblem(p); err != nil {
			t.Fatalf("SaveProblem() error = %v", err)
		}
	}

	// List problems
	listed, err := ws.ListProblems()
	if err != nil {
		t.Fatalf("ListProblems() error = %v", err)
	}

	if len(listed) != 3 {
		t.Errorf("ListProblems() returned %d problems, want %d", len(listed), 3)
	}
}

func TestWorkspace_ListProblems_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	listed, err := ws.ListProblems()
	if err != nil {
		t.Fatalf("ListProblems() error = %v", err)
	}

	if len(listed) != 0 {
		t.Errorf("ListProblems() should return empty for new workspace, got %d", len(listed))
	}
}

func TestWorkspace_SaveStatement(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	problem := v1.NewProblem(1325, "A", "EhAb AnD gCd")
	problem.Metadata.Rating = 800
	problem.Metadata.Tags = []string{"math"}
	problem.Limits.TimeLimit = "1 second"
	problem.Limits.MemoryLimit = "256 megabytes"
	problem.Samples = []v1.Sample{
		{Index: 1, Input: "1\n", Output: "2\n"},
	}
	problem.URL = "https://codeforces.com/contest/1325/problem/A"

	err = ws.SaveProblem(problem)
	if err != nil {
		t.Fatalf("SaveProblem() error = %v", err)
	}

	statement := "This is the problem statement."
	err = ws.SaveStatement(problem, statement)
	if err != nil {
		t.Fatalf("SaveStatement() error = %v", err)
	}

	// Verify statement.md exists
	statementPath := filepath.Join(ws.ProblemPath("codeforces", 1325, "A"), "statement.md")
	if _, err := os.Stat(statementPath); os.IsNotExist(err) {
		t.Error("SaveStatement() did not create statement.md")
	}

	// Read and verify content
	content, err := os.ReadFile(statementPath)
	if err != nil {
		t.Fatalf("Failed to read statement: %v", err)
	}

	md := string(content)
	if !strings.Contains(md, "# A. EhAb AnD gCd") {
		t.Error("Statement should contain problem title")
	}
	if !strings.Contains(md, "**Rating:** 800") {
		t.Error("Statement should contain rating")
	}
	if !strings.Contains(md, "**Tags:** math") {
		t.Error("Statement should contain tags")
	}
	if !strings.Contains(md, "This is the problem statement.") {
		t.Error("Statement should contain the statement text")
	}
	if !strings.Contains(md, "## Examples") {
		t.Error("Statement should contain examples section")
	}
}

func TestWorkspace_LoadStatement(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	problem := v1.NewProblem(1325, "A", "Test")
	err = ws.SaveProblem(problem)
	if err != nil {
		t.Fatalf("SaveProblem() error = %v", err)
	}

	statement := "Test statement content"
	err = ws.SaveStatement(problem, statement)
	if err != nil {
		t.Fatalf("SaveStatement() error = %v", err)
	}

	loaded, err := ws.LoadStatement("codeforces", 1325, "A")
	if err != nil {
		t.Fatalf("LoadStatement() error = %v", err)
	}

	if !strings.Contains(loaded, statement) {
		t.Errorf("LoadStatement() should contain original statement")
	}
}

func TestWorkspace_LoadStatement_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	_, err = ws.LoadStatement("codeforces", 9999, "Z")
	if err == nil {
		t.Error("LoadStatement() should return error for non-existent statement")
	}
}

func TestWorkspace_SaveNotes(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	problem := v1.NewProblem(1325, "A", "Test")
	err = ws.SaveProblem(problem)
	if err != nil {
		t.Fatalf("SaveProblem() error = %v", err)
	}

	notes := &v1.UserNotes{
		Difficulty: "easy",
		CustomTags: []string{"review"},
		Approach:   "Use GCD",
		Review:     true,
	}

	err = ws.SaveNotes("codeforces", 1325, "A", notes)
	if err != nil {
		t.Fatalf("SaveNotes() error = %v", err)
	}

	// Load and verify
	loaded, err := ws.LoadProblem("codeforces", 1325, "A")
	if err != nil {
		t.Fatalf("LoadProblem() error = %v", err)
	}

	if loaded.Notes.Difficulty != "easy" {
		t.Errorf("Notes.Difficulty = %v, want %v", loaded.Notes.Difficulty, "easy")
	}
	if loaded.Notes.Approach != "Use GCD" {
		t.Errorf("Notes.Approach = %v, want %v", loaded.Notes.Approach, "Use GCD")
	}
	if !loaded.Notes.Review {
		t.Error("Notes.Review should be true")
	}
}

func TestWorkspace_UpdatePractice(t *testing.T) {
	tmpDir := t.TempDir()
	ws := New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	problem := v1.NewProblem(1325, "A", "Test")
	err = ws.SaveProblem(problem)
	if err != nil {
		t.Fatalf("SaveProblem() error = %v", err)
	}

	practice := &v1.PracticeData{
		Status:       v1.StatusSolved,
		AttemptCount: 2,
		TimeSpent:    300,
	}

	err = ws.UpdatePractice("codeforces", 1325, "A", practice)
	if err != nil {
		t.Fatalf("UpdatePractice() error = %v", err)
	}

	// Load and verify
	loaded, err := ws.LoadProblem("codeforces", 1325, "A")
	if err != nil {
		t.Fatalf("LoadProblem() error = %v", err)
	}

	if loaded.Practice.Status != v1.StatusSolved {
		t.Errorf("Practice.Status = %v, want %v", loaded.Practice.Status, v1.StatusSolved)
	}
	if loaded.Practice.AttemptCount != 2 {
		t.Errorf("Practice.AttemptCount = %v, want %v", loaded.Practice.AttemptCount, 2)
	}
	if loaded.Practice.TimeSpent != 300 {
		t.Errorf("Practice.TimeSpent = %v, want %v", loaded.Practice.TimeSpent, 300)
	}
}

func TestFormatStatement(t *testing.T) {
	problem := v1.NewProblem(1325, "A", "EhAb AnD gCd")
	problem.Metadata.Rating = 800
	problem.Metadata.Tags = []string{"math", "number theory"}
	problem.Limits.TimeLimit = "1 second"
	problem.Limits.MemoryLimit = "256 MB"
	problem.Samples = []v1.Sample{
		{Index: 1, Input: "1\n", Output: "2\n"},
	}
	problem.URL = "https://codeforces.com/contest/1325/problem/A"

	result := formatStatement(problem, "Find x and y such that x + y = n and gcd(x, y) is maximum.")

	// Check title
	if !strings.Contains(result, "# A. EhAb AnD gCd") {
		t.Error("formatStatement() should include title")
	}

	// Check metadata
	if !strings.Contains(result, "**Rating:** 800") {
		t.Error("formatStatement() should include rating")
	}
	if !strings.Contains(result, "**Time:** 1 second") {
		t.Error("formatStatement() should include time limit")
	}
	if !strings.Contains(result, "**Memory:** 256 MB") {
		t.Error("formatStatement() should include memory limit")
	}

	// Check tags
	if !strings.Contains(result, "**Tags:** math, number theory") {
		t.Error("formatStatement() should include tags")
	}

	// Check statement
	if !strings.Contains(result, "## Statement") {
		t.Error("formatStatement() should include statement section")
	}
	if !strings.Contains(result, "Find x and y") {
		t.Error("formatStatement() should include statement content")
	}

	// Check examples
	if !strings.Contains(result, "## Examples") {
		t.Error("formatStatement() should include examples section")
	}
	if !strings.Contains(result, "### Example 1") {
		t.Error("formatStatement() should include example headers")
	}

	// Check link
	if !strings.Contains(result, "[Open on Codeforces]") {
		t.Error("formatStatement() should include Codeforces link")
	}
}
