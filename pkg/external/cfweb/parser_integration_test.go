//go:build integration

package cfweb

import (
	"testing"
	"time"
)

func TestParser_ParseProblem_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Parse problem 1A (Theatre Square)
	problem, err := parser.ParseProblem(1, "A")
	if err != nil {
		t.Fatalf("ParseProblem(1, A) failed: %v", err)
	}

	if problem.ContestID != 1 {
		t.Errorf("ContestID = %d, want 1", problem.ContestID)
	}

	if problem.Index != "A" {
		t.Errorf("Index = %s, want A", problem.Index)
	}

	if problem.Name == "" {
		t.Error("Name should not be empty")
	}

	if problem.TimeLimit == "" {
		t.Error("TimeLimit should not be empty")
	}

	if problem.MemoryLimit == "" {
		t.Error("MemoryLimit should not be empty")
	}

	if len(problem.Samples) == 0 {
		t.Error("Samples should not be empty")
	}

	t.Logf("Problem 1A: Name=%s, TimeLimit=%s, MemoryLimit=%s, Samples=%d",
		problem.Name, problem.TimeLimit, problem.MemoryLimit, len(problem.Samples))
}

func TestParser_ParseProblemset_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Parse problem 1A from problemset URL
	problem, err := parser.ParseProblemset(1, "A")
	if err != nil {
		t.Fatalf("ParseProblemset(1, A) failed: %v", err)
	}

	if problem.ContestID != 1 {
		t.Errorf("ContestID = %d, want 1", problem.ContestID)
	}

	if problem.Index != "A" {
		t.Errorf("Index = %s, want A", problem.Index)
	}

	if problem.URL == "" {
		t.Error("URL should not be empty")
	}

	t.Logf("Problemset 1/A: Name=%s, URL=%s", problem.Name, problem.URL)
}

func TestParser_ParseContestProblems_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Parse all problems from contest 1
	problems, err := parser.ParseContestProblems(1)
	if err != nil {
		t.Fatalf("ParseContestProblems(1) failed: %v", err)
	}

	// Contest 1 should have at least 3 problems
	if len(problems) < 3 {
		t.Errorf("expected at least 3 problems, got %d", len(problems))
	}

	// Verify problem A exists
	var foundA bool
	for _, p := range problems {
		if p.Index == "A" {
			foundA = true
			if p.ContestID != 1 {
				t.Errorf("Problem A has ContestID = %d, want 1", p.ContestID)
			}
		}
	}

	if !foundA {
		t.Error("Problem A not found in contest 1")
	}

	t.Logf("Contest 1 has %d problems", len(problems))
	for _, p := range problems {
		t.Logf("  - %s: %s", p.Index, p.Name)
	}
}

func TestParser_ParseProblem_ExtractSamples_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Parse problem 1A which has known sample tests
	problem, err := parser.ParseProblem(1, "A")
	if err != nil {
		t.Fatalf("ParseProblem(1, A) failed: %v", err)
	}

	if len(problem.Samples) == 0 {
		t.Fatal("expected samples but got none")
	}

	// Check first sample
	sample := problem.Samples[0]
	if sample.Index != 1 {
		t.Errorf("First sample index = %d, want 1", sample.Index)
	}

	if sample.Input == "" {
		t.Error("First sample input should not be empty")
	}

	if sample.Output == "" {
		t.Error("First sample output should not be empty")
	}

	t.Logf("Sample 1 - Input: %q, Output: %q", sample.Input, sample.Output)
}

func TestParser_ParseProblem_ExtractTags_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Parse problem 1A
	problem, err := parser.ParseProblem(1, "A")
	if err != nil {
		t.Fatalf("ParseProblem(1, A) failed: %v", err)
	}

	// Problem 1A should have tags (math is common for this problem)
	if len(problem.Tags) == 0 {
		t.Log("Warning: no tags found for problem 1A")
	} else {
		t.Logf("Problem 1A tags: %v", problem.Tags)
	}
}

func TestParser_ParseProblem_ExtractRating_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Parse problem 1A (should have rating around 1000)
	problem, err := parser.ParseProblem(1, "A")
	if err != nil {
		t.Fatalf("ParseProblem(1, A) failed: %v", err)
	}

	// Rating might or might not be available
	if problem.Rating > 0 {
		if problem.Rating < 800 || problem.Rating > 3500 {
			t.Errorf("Rating %d seems out of normal range (800-3500)", problem.Rating)
		}
		t.Logf("Problem 1A rating: %d", problem.Rating)
	} else {
		t.Log("Warning: rating not found or is 0 for problem 1A")
	}
}

func TestParser_VerifyPageStructure_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	err = parser.VerifyPageStructure()
	if err != nil {
		t.Fatalf("VerifyPageStructure() failed: %v", err)
	}

	t.Log("Page structure verification passed")
}

func TestParser_ToSchemaProblem_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Parse and convert problem
	parsedProblem, err := parser.ParseProblem(1, "A")
	if err != nil {
		t.Fatalf("ParseProblem(1, A) failed: %v", err)
	}

	schemaProblem := parsedProblem.ToSchemaProblem()

	if schemaProblem.ID != "1A" {
		t.Errorf("ID = %s, want 1A", schemaProblem.ID)
	}

	if schemaProblem.Platform != "codeforces" {
		t.Errorf("Platform = %s, want codeforces", schemaProblem.Platform)
	}

	if schemaProblem.ContestID != 1 {
		t.Errorf("ContestID = %d, want 1", schemaProblem.ContestID)
	}

	if schemaProblem.Index != "A" {
		t.Errorf("Index = %s, want A", schemaProblem.Index)
	}

	if schemaProblem.URL == "" {
		t.Error("URL should not be empty")
	}

	if len(schemaProblem.Samples) == 0 {
		t.Error("Samples should not be empty")
	}

	t.Logf("Schema problem: ID=%s, Name=%s, Rating=%d",
		schemaProblem.ID, schemaProblem.Name, schemaProblem.Metadata.Rating)
}

func TestParser_ParseProblem_NotFound_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Try to parse a non-existent problem
	problem, err := parser.ParseProblem(1, "Z")
	if err != nil {
		t.Logf("Correctly got error for non-existent problem: %v", err)
		return
	}

	// If no error, check that the problem data is empty/invalid
	// (CF might return 200 with an error page)
	if problem.Name == "" {
		t.Log("ParseProblem returned empty problem for non-existent problem index")
		return
	}

	// If we got a valid problem, that's unexpected for 1Z
	t.Errorf("ParseProblem(1, Z) should return error or empty problem, got: %+v", problem)
}

func TestParser_ParseRecentProblem_Real(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	parser := NewParser(session)

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Parse a more recent problem (contest 1800+)
	problem, err := parser.ParseProblem(1800, "A")
	if err != nil {
		t.Fatalf("ParseProblem(1800, A) failed: %v", err)
	}

	if problem.ContestID != 1800 {
		t.Errorf("ContestID = %d, want 1800", problem.ContestID)
	}

	if problem.Name == "" {
		t.Error("Name should not be empty")
	}

	t.Logf("Problem 1800A: Name=%s, Rating=%d", problem.Name, problem.Rating)
}
