package v1

import (
	"testing"
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
	"gopkg.in/yaml.v3"
)

func TestNewSubmission(t *testing.T) {
	sub := NewSubmission(123456, "1325A", 1325, "GNU C++17", 54)

	if sub.ID != 123456 {
		t.Errorf("ID = %v, want 123456", sub.ID)
	}
	if sub.ProblemID != "1325A" {
		t.Errorf("ProblemID = %v, want 1325A", sub.ProblemID)
	}
	if sub.ContestID != 1325 {
		t.Errorf("ContestID = %v, want 1325", sub.ContestID)
	}
	if sub.Language != "GNU C++17" {
		t.Errorf("Language = %v, want GNU C++17", sub.Language)
	}
	if sub.LanguageID != 54 {
		t.Errorf("LanguageID = %v, want 54", sub.LanguageID)
	}
	if sub.Verdict != VerdictPending {
		t.Errorf("Verdict = %v, want %v", sub.Verdict, VerdictPending)
	}
	if sub.Schema.Type != schema.TypeSubmission {
		t.Errorf("Schema.Type = %v, want %v", sub.Schema.Type, schema.TypeSubmission)
	}
	if sub.SubmittedAt.IsZero() {
		t.Error("SubmittedAt should not be zero")
	}
}

func TestVerdict_IsAccepted(t *testing.T) {
	tests := []struct {
		verdict Verdict
		want    bool
	}{
		{VerdictOK, true},
		{VerdictWrongAnswer, false},
		{VerdictTimeLimitExceeded, false},
		{VerdictMemoryLimitExceeded, false},
		{VerdictRuntimeError, false},
		{VerdictCompilationError, false},
		{VerdictPending, false},
		{VerdictTesting, false},
		{Verdict("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.verdict), func(t *testing.T) {
			if got := tt.verdict.IsAccepted(); got != tt.want {
				t.Errorf("IsAccepted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerdictConstants(t *testing.T) {
	// Verify all verdict constants have expected string values
	verdicts := map[Verdict]string{
		VerdictOK:                  "OK",
		VerdictWrongAnswer:         "WRONG_ANSWER",
		VerdictTimeLimitExceeded:   "TIME_LIMIT_EXCEEDED",
		VerdictMemoryLimitExceeded: "MEMORY_LIMIT_EXCEEDED",
		VerdictRuntimeError:        "RUNTIME_ERROR",
		VerdictCompilationError:    "COMPILATION_ERROR",
		VerdictPending:             "PENDING",
		VerdictTesting:             "TESTING",
	}

	for verdict, expected := range verdicts {
		if string(verdict) != expected {
			t.Errorf("Verdict %v = %q, want %q", verdict, string(verdict), expected)
		}
	}
}

func TestSubmission_Fields(t *testing.T) {
	now := time.Now()
	sub := &Submission{
		Schema:      schema.NewSchemaHeader(schema.TypeSubmission),
		ID:          999,
		ProblemID:   "1A",
		ContestID:   1,
		SubmittedAt: now,
		Language:    "Python 3",
		LanguageID:  31,
		Verdict:     VerdictOK,
		TimeUsed:    "46 ms",
		MemoryUsed:  "256 KB",
		SourceFile:  "solution.py",
		SourceHash:  "abc123def456",
	}

	if sub.ID != 999 {
		t.Errorf("ID = %v, want 999", sub.ID)
	}
	if sub.TimeUsed != "46 ms" {
		t.Errorf("TimeUsed = %v, want '46 ms'", sub.TimeUsed)
	}
	if sub.MemoryUsed != "256 KB" {
		t.Errorf("MemoryUsed = %v, want '256 KB'", sub.MemoryUsed)
	}
	if sub.SourceFile != "solution.py" {
		t.Errorf("SourceFile = %v, want 'solution.py'", sub.SourceFile)
	}
	if sub.SourceHash != "abc123def456" {
		t.Errorf("SourceHash = %v, want 'abc123def456'", sub.SourceHash)
	}
}

func TestSubmission_YAML_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := &Submission{
		Schema:      schema.NewSchemaHeader(schema.TypeSubmission),
		ID:          12345,
		ProblemID:   "1325A",
		ContestID:   1325,
		SubmittedAt: now,
		Language:    "GNU C++17",
		LanguageID:  54,
		Verdict:     VerdictOK,
		TimeUsed:    "31 ms",
		MemoryUsed:  "0 KB",
		SourceFile:  "main.cpp",
		SourceHash:  "sha256hash",
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	// Unmarshal back
	var decoded Submission
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}

	// Verify fields
	if decoded.ID != original.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, original.ID)
	}
	if decoded.ProblemID != original.ProblemID {
		t.Errorf("ProblemID = %v, want %v", decoded.ProblemID, original.ProblemID)
	}
	if decoded.Verdict != original.Verdict {
		t.Errorf("Verdict = %v, want %v", decoded.Verdict, original.Verdict)
	}
	if decoded.Language != original.Language {
		t.Errorf("Language = %v, want %v", decoded.Language, original.Language)
	}
	if decoded.SourceHash != original.SourceHash {
		t.Errorf("SourceHash = %v, want %v", decoded.SourceHash, original.SourceHash)
	}
}

func TestSubmission_EmptySourceHash(t *testing.T) {
	sub := NewSubmission(1, "1A", 1, "cpp", 1)

	// SourceHash should be empty by default
	if sub.SourceHash != "" {
		t.Errorf("SourceHash should be empty by default, got %q", sub.SourceHash)
	}
}

func TestSubmission_VerdictTransitions(t *testing.T) {
	sub := NewSubmission(1, "1A", 1, "cpp", 1)

	// Initial verdict should be pending
	if sub.Verdict != VerdictPending {
		t.Errorf("Initial verdict = %v, want %v", sub.Verdict, VerdictPending)
	}

	// Simulate verdict update to testing
	sub.Verdict = VerdictTesting
	if sub.Verdict != VerdictTesting {
		t.Errorf("Verdict = %v, want %v", sub.Verdict, VerdictTesting)
	}

	// Simulate verdict update to OK
	sub.Verdict = VerdictOK
	if !sub.Verdict.IsAccepted() {
		t.Error("Verdict should be accepted after setting to OK")
	}
}
