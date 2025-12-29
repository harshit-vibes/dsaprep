package v1

import (
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
)

// Submission represents a submission record
type Submission struct {
	Schema schema.SchemaHeader `yaml:"_schema" json:"_schema"`

	// Identity
	ID        int64  `yaml:"id" json:"id"`
	ProblemID string `yaml:"problemId" json:"problemId"`
	ContestID int    `yaml:"contestId" json:"contestId"`

	// Timestamp
	SubmittedAt time.Time `yaml:"submittedAt" json:"submittedAt"`

	// Submission details
	Language   string `yaml:"language" json:"language"`
	LanguageID int    `yaml:"languageId" json:"languageId"`

	// Result
	Verdict    Verdict `yaml:"verdict" json:"verdict"`
	TimeUsed   string  `yaml:"timeUsed" json:"timeUsed"`
	MemoryUsed string  `yaml:"memoryUsed" json:"memoryUsed"`

	// Source reference
	SourceFile string `yaml:"sourceFile" json:"sourceFile"`
	SourceHash string `yaml:"sourceHash,omitempty" json:"sourceHash,omitempty"`
}

// Verdict represents submission verdict
type Verdict string

const (
	VerdictOK                  Verdict = "OK"
	VerdictWrongAnswer         Verdict = "WRONG_ANSWER"
	VerdictTimeLimitExceeded   Verdict = "TIME_LIMIT_EXCEEDED"
	VerdictMemoryLimitExceeded Verdict = "MEMORY_LIMIT_EXCEEDED"
	VerdictRuntimeError        Verdict = "RUNTIME_ERROR"
	VerdictCompilationError    Verdict = "COMPILATION_ERROR"
	VerdictPending             Verdict = "PENDING"
	VerdictTesting             Verdict = "TESTING"
)

// IsAccepted returns true if verdict is OK
func (v Verdict) IsAccepted() bool {
	return v == VerdictOK
}

// NewSubmission creates a new submission record
func NewSubmission(id int64, problemID string, contestID int, language string, langID int) *Submission {
	return &Submission{
		Schema:      schema.NewSchemaHeader(schema.TypeSubmission),
		ID:          id,
		ProblemID:   problemID,
		ContestID:   contestID,
		SubmittedAt: time.Now(),
		Language:    language,
		LanguageID:  langID,
		Verdict:     VerdictPending,
	}
}
