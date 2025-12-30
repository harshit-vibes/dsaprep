package cfweb

import (
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Note: time is still used for time.Millisecond in tests

func TestParseTime(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Duration
	}{
		{"normal ms", "46 ms", 46 * time.Millisecond},
		{"large ms", "1000 ms", 1000 * time.Millisecond},
		{"no space", "15ms", 15 * time.Millisecond},
		{"with text before", "Time: 100 ms", 100 * time.Millisecond},
		{"zero", "0 ms", 0},
		{"empty", "", 0},
		{"no ms unit", "100", 0},
		{"no time", "no time here", 0},
		{"seconds not ms", "5 s", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTime(tt.input)
			if got != tt.want {
				t.Errorf("parseTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseMemory(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{"KB normal", "256 KB", 256 * 1024},
		{"KB large", "1024 KB", 1024 * 1024},
		{"KB no space", "100KB", 100 * 1024},
		{"MB normal", "1 MB", 1 * 1024 * 1024},
		{"MB large", "10 MB", 10 * 1024 * 1024},
		{"MB no space", "5MB", 5 * 1024 * 1024},
		{"with text", "Memory: 512 KB used", 512 * 1024},
		{"empty", "", 0},
		{"no memory", "no memory here", 0},
		{"bytes not KB", "1024 bytes", 0},
		{"zero KB", "0 KB", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMemory(tt.input)
			if got != tt.want {
				t.Errorf("parseMemory(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeVerdict(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"accepted", "Accepted", "OK"},
		{"accepted with whitespace", "  Accepted  ", "OK"},
		{"wrong answer", "Wrong answer on test 5", "WRONG_ANSWER"},
		{"time limit", "Time limit exceeded on test 3", "TIME_LIMIT_EXCEEDED"},
		{"memory limit", "Memory limit exceeded", "MEMORY_LIMIT_EXCEEDED"},
		{"runtime error", "Runtime error on test 1", "RUNTIME_ERROR"},
		{"compilation error", "Compilation error", "COMPILATION_ERROR"},
		{"presentation error", "Presentation error", "PRESENTATION_ERROR"},
		{"idleness", "Idleness limit exceeded", "IDLENESS_LIMIT_EXCEEDED"},
		{"hacked", "Hacked", "CHALLENGED"},
		{"unknown", "Unknown verdict", "Unknown verdict"},
		{"empty", "", ""},
		{"partial accepted", "Accepted (partial)", "OK"},
		{"wrong answer short", "Wrong answer", "WRONG_ANSWER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeVerdict(tt.input)
			if got != tt.want {
				t.Errorf("normalizeVerdict(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSubmissionResult_AllFields(t *testing.T) {
	now := time.Now()
	result := &SubmissionResult{
		SubmissionID: 123456789,
		ContestID:    1,
		ProblemIndex: "A",
		Verdict:      "OK",
		Time:         46 * time.Millisecond,
		Memory:       256 * 1024,
		PassedTests:  10,
		SubmittedAt:  now,
		Status:       "Judged",
	}

	if result.SubmissionID != 123456789 {
		t.Errorf("SubmissionID = %d, want 123456789", result.SubmissionID)
	}
	if result.ContestID != 1 {
		t.Errorf("ContestID = %d, want 1", result.ContestID)
	}
	if result.ProblemIndex != "A" {
		t.Errorf("ProblemIndex = %s, want A", result.ProblemIndex)
	}
	if result.Verdict != "OK" {
		t.Errorf("Verdict = %s, want OK", result.Verdict)
	}
	if result.Time != 46*time.Millisecond {
		t.Errorf("Time = %v, want 46ms", result.Time)
	}
	if result.Memory != 256*1024 {
		t.Errorf("Memory = %d, want %d", result.Memory, 256*1024)
	}
	if result.PassedTests != 10 {
		t.Errorf("PassedTests = %d, want 10", result.PassedTests)
	}
	if result.Status != "Judged" {
		t.Errorf("Status = %s, want Judged", result.Status)
	}
}

func TestNewSubmitter_NotAuthenticated(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	_, err = NewSubmitter(session)
	if err == nil {
		t.Error("NewSubmitter() should fail with unauthenticated session")
	}
}

func TestNewSubmitter_MissingHandle(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	// Set cookies but no handle
	session.SetCookie("JSESSIONID=test-jsessionid; 39ce7=test-39ce7; cf_clearance=test-clearance")
	// Don't set handle

	_, err = NewSubmitter(session)
	if err == nil {
		t.Error("NewSubmitter() should fail without handle")
	}
}

func TestNewSubmitter_Authenticated(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	// Set full authentication using new API
	session.SetCookie("JSESSIONID=test-jsessionid; 39ce7=test-39ce7; cf_clearance=test-clearance")
	session.SetHandle("testuser")

	submitter, err := NewSubmitter(session)
	if err != nil {
		t.Fatalf("NewSubmitter() failed: %v", err)
	}

	if submitter == nil {
		t.Fatal("NewSubmitter() returned nil")
	}

	if submitter.session != session {
		t.Error("submitter.session should match input session")
	}
}

func TestSubmitter_Struct(t *testing.T) {
	session, _ := NewSession()
	session.SetCookie("JSESSIONID=test-jsessionid; 39ce7=test-39ce7; cf_clearance=test-clearance")
	session.SetHandle("testuser")

	submitter := &Submitter{session: session}

	if submitter.session == nil {
		t.Error("submitter.session should not be nil")
	}
}

func TestParseSubmissionRow(t *testing.T) {
	html := `<table>
<tr data-submission-id="123456789">
<td class="id-cell">A</td>
<td class="status-cell"><span class="verdict-accepted">Accepted</span></td>
<td class="time-consumed-cell">46 ms</td>
<td class="memory-consumed-cell">256 KB</td>
</tr>
</table>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	row := doc.Find("tr[data-submission-id]").First()
	result, err := parseSubmissionRow(row, 1)
	if err != nil {
		t.Fatalf("parseSubmissionRow() error = %v", err)
	}

	if result.SubmissionID != 123456789 {
		t.Errorf("SubmissionID = %d, want 123456789", result.SubmissionID)
	}
	if result.ContestID != 1 {
		t.Errorf("ContestID = %d, want 1", result.ContestID)
	}
	if result.Verdict != "OK" {
		t.Errorf("Verdict = %s, want OK", result.Verdict)
	}
	if result.Time != 46*time.Millisecond {
		t.Errorf("Time = %v, want 46ms", result.Time)
	}
	if result.Memory != 256*1024 {
		t.Errorf("Memory = %d, want %d", result.Memory, 256*1024)
	}
}

func TestParseSubmissionRow_NoSubmissionID(t *testing.T) {
	html := `<table><tr><td>No ID</td></tr></table>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	row := doc.Find("tr").First()
	_, err = parseSubmissionRow(row, 1)
	if err == nil {
		t.Error("parseSubmissionRow() should return error when no submission ID")
	}
}

func TestParseSubmissionRow_InvalidSubmissionID(t *testing.T) {
	html := `<table><tr data-submission-id="invalid"><td>Invalid</td></tr></table>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	row := doc.Find("tr[data-submission-id]").First()
	_, err = parseSubmissionRow(row, 1)
	if err == nil {
		t.Error("parseSubmissionRow() should return error for invalid submission ID")
	}
}

func TestParseSubmissionRow_InQueue(t *testing.T) {
	html := `<table>
<tr data-submission-id="123456789">
<td class="id-cell">A</td>
<td class="status-cell">In queue</td>
<td class="time-consumed-cell"></td>
<td class="memory-consumed-cell"></td>
</tr>
</table>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	row := doc.Find("tr[data-submission-id]").First()
	result, err := parseSubmissionRow(row, 1)
	if err != nil {
		t.Fatalf("parseSubmissionRow() error = %v", err)
	}

	if result.Status != "In queue" {
		t.Errorf("Status = %s, want 'In queue'", result.Status)
	}
}

func TestParseSubmissionRow_Running(t *testing.T) {
	html := `<table>
<tr data-submission-id="123456789">
<td class="id-cell">A</td>
<td class="status-cell">Running on test 5</td>
<td class="time-consumed-cell"></td>
<td class="memory-consumed-cell"></td>
</tr>
</table>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	row := doc.Find("tr[data-submission-id]").First()
	result, err := parseSubmissionRow(row, 1)
	if err != nil {
		t.Fatalf("parseSubmissionRow() error = %v", err)
	}

	if result.Status != "Running" {
		t.Errorf("Status = %s, want 'Running'", result.Status)
	}
}

