package health

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harshit-vibes/dsaprep/pkg/external/cfapi"
	"github.com/harshit-vibes/dsaprep/pkg/external/cfweb"
	"github.com/harshit-vibes/dsaprep/pkg/internal/health"
)

// ============ Mock Transport for HTTP Mocking ============

type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

// ============ CFAPICheck Mock Tests ============

func TestCFAPICheck_Check_PingError(t *testing.T) {
	// Create a client with a mock transport that returns an error
	httpClient := &http.Client{
		Transport: &mockTransport{
			err: &mockError{msg: "connection refused"},
		},
		Timeout: 100 * time.Millisecond,
	}
	client := cfapi.NewClient(cfapi.WithHTTPClient(httpClient))
	check := NewCFAPICheck(client)

	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "CF API unreachable" {
		t.Errorf("Message = %v, want 'CF API unreachable'", result.Message)
	}
	if result.Action != health.ActionRetry {
		t.Errorf("Action = %v, want %v", result.Action, health.ActionRetry)
	}
	if result.Details == "" {
		t.Error("Details should contain error message")
	}
}

// ============ CFWebCheck Mock Tests ============

func TestCFWebCheck_Check_VerifyStructureError(t *testing.T) {
	// Create parser with a mock transport that returns error
	httpClient := &http.Client{
		Transport: &mockTransport{
			err: &mockError{msg: "network error"},
		},
		Timeout: 100 * time.Millisecond,
	}
	parser := cfweb.NewParserWithClient(httpClient)
	check := NewCFWebCheck(parser)

	result := check.Check(context.Background())

	// VerifyPageStructure will fail due to network error
	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "CF page structure changed" {
		t.Errorf("Message = %v, want 'CF page structure changed'", result.Message)
	}
	if result.Action != health.ActionManualFix {
		t.Errorf("Action = %v, want %v", result.Action, health.ActionManualFix)
	}
}

// ============ CFHandleCheck Mock Tests ============

func TestCFHandleCheck_Check_NoHandle(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp dir with no .dsaprep.env file
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	client := cfapi.NewClient()
	check := NewCFHandleCheck(client)

	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "CF handle not configured" {
		t.Errorf("Message = %v, want 'CF handle not configured'", result.Message)
	}
	if result.Action != health.ActionUserPrompt {
		t.Errorf("Action = %v, want %v", result.Action, health.ActionUserPrompt)
	}
}

func TestCFHandleCheck_Check_NilClientWithHandle(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with handle
	envPath := filepath.Join(tmpDir, ".dsaprep.env")
	err := os.WriteFile(envPath, []byte("CF_HANDLE=testuser\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := NewCFHandleCheck(nil)
	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "Cannot verify handle" {
		t.Errorf("Message = %v, want 'Cannot verify handle'", result.Message)
	}
	if result.Details != "API client not initialized" {
		t.Errorf("Details = %v, want 'API client not initialized'", result.Details)
	}
}

func TestCFHandleCheck_Check_GetUserInfoError(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with handle
	envPath := filepath.Join(tmpDir, ".dsaprep.env")
	err := os.WriteFile(envPath, []byte("CF_HANDLE=testuser\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	// Create client with mock transport that returns error
	httpClient := &http.Client{
		Transport: &mockTransport{
			err: &mockError{msg: "API error"},
		},
		Timeout: 100 * time.Millisecond,
	}
	client := cfapi.NewClient(cfapi.WithHTTPClient(httpClient))
	check := NewCFHandleCheck(client)

	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "Cannot verify handle" {
		t.Errorf("Message = %v, want 'Cannot verify handle'", result.Message)
	}
	if result.Action != health.ActionRetry {
		t.Errorf("Action = %v, want %v", result.Action, health.ActionRetry)
	}
}

func TestCFHandleCheck_Check_HandleNotFound(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with handle
	envPath := filepath.Join(tmpDir, ".dsaprep.env")
	err := os.WriteFile(envPath, []byte("CF_HANDLE=nonexistent_user_12345678\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	// Use real client but with a handle that doesn't exist
	// The API returns empty result for non-existent handles
	client := cfapi.NewClient()
	check := NewCFHandleCheck(client)

	result := check.Check(context.Background())

	// Should be Critical when handle not found
	if result.Status != health.StatusCritical {
		// API might return error instead of empty result
		if result.Status != health.StatusDegraded {
			t.Errorf("Status = %v, want Critical or Degraded", result.Status)
		}
	}
}

func TestCFHandleCheck_Check_ValidHandle(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with a known valid handle
	envPath := filepath.Join(tmpDir, ".dsaprep.env")
	err := os.WriteFile(envPath, []byte("CF_HANDLE=tourist\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	client := cfapi.NewClient()
	check := NewCFHandleCheck(client)

	result := check.Check(context.Background())

	// Should be Healthy when handle is found
	if result.Status == health.StatusCritical {
		t.Errorf("Status = %v for valid handle", result.Status)
	}
	// Duration should be set
	if result.Duration == 0 {
		t.Error("Duration should not be zero")
	}
}

// ============ CFClearanceCheck Mock Tests ============

func TestCFClearanceCheck_Check_EmptyClearance(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file without cf_clearance
	envPath := filepath.Join(tmpDir, ".dsaprep.env")
	err := os.WriteFile(envPath, []byte("CF_HANDLE=testuser\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := NewCFClearanceCheck()
	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "cf_clearance not configured" {
		t.Errorf("Message = %v, want 'cf_clearance not configured'", result.Message)
	}
	if result.Action != health.ActionUserPrompt {
		t.Errorf("Action = %v, want %v", result.Action, health.ActionUserPrompt)
	}
}

func TestCFClearanceCheck_Check_ExpiredClearance(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with expired cf_clearance (timestamp in the past)
	envPath := filepath.Join(tmpDir, ".dsaprep.env")
	content := `CF_HANDLE=testuser
CF_CLEARANCE=test_clearance
CF_CLEARANCE_EXPIRES=1000000000
CF_CLEARANCE_UA=TestUA
`
	err := os.WriteFile(envPath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := NewCFClearanceCheck()
	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "cf_clearance expired" {
		t.Errorf("Message = %v, want 'cf_clearance expired'", result.Message)
	}
	if !result.Recoverable {
		t.Error("Result should be recoverable")
	}
}

func TestCFClearanceCheck_Check_ExpiringSoon(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with cf_clearance expiring in 2 minutes
	expiresAt := time.Now().Add(2 * time.Minute).Unix()
	envPath := filepath.Join(tmpDir, ".dsaprep.env")
	err := os.WriteFile(envPath, []byte(
		"CF_HANDLE=testuser\nCF_CLEARANCE=test_clearance\nCF_CLEARANCE_EXPIRES="+
			formatTimestamp(expiresAt)+"\nCF_CLEARANCE_UA=TestUA\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := NewCFClearanceCheck()
	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "cf_clearance expiring soon" {
		t.Errorf("Message = %v, want 'cf_clearance expiring soon'", result.Message)
	}
	if !result.Recoverable {
		t.Error("Result should be recoverable")
	}
}

func TestCFClearanceCheck_Check_ValidClearance(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with cf_clearance valid for 1 hour
	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	envPath := filepath.Join(tmpDir, ".dsaprep.env")
	err := os.WriteFile(envPath, []byte(
		"CF_HANDLE=testuser\nCF_CLEARANCE=test_clearance\nCF_CLEARANCE_EXPIRES="+
			formatTimestamp(expiresAt)+"\nCF_CLEARANCE_UA=TestUA\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := NewCFClearanceCheck()
	result := check.Check(context.Background())

	if result.Status != health.StatusHealthy {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusHealthy)
	}
	if result.Duration == 0 {
		t.Error("Duration should not be zero")
	}
}

// ============ Helper Types and Functions ============

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func formatTimestamp(ts int64) string {
	return fmt.Sprintf("%d", ts)
}

// ============ Additional Coverage Tests ============

func TestFormatDuration_ZeroSeconds(t *testing.T) {
	result := formatDuration(0)
	if result != "0s" {
		t.Errorf("formatDuration(0) = %v, want '0s'", result)
	}
}

func TestFormatDuration_OneMinute(t *testing.T) {
	result := formatDuration(60 * time.Second)
	if result != "1m" {
		t.Errorf("formatDuration(60s) = %v, want '1m'", result)
	}
}

func TestFormatDuration_OneHour(t *testing.T) {
	result := formatDuration(60 * time.Minute)
	if result != "1h 0m" {
		t.Errorf("formatDuration(60m) = %v, want '1h 0m'", result)
	}
}

func TestFormatRating_Negative(t *testing.T) {
	// Edge case: negative rating (shouldn't happen but let's cover it)
	result := formatRating(-100)
	if result != "-100" {
		t.Errorf("formatRating(-100) = %v, want '-100'", result)
	}
}

func TestNetworkCheck_Check_Error(t *testing.T) {
	// NetworkCheck creates its own parser, so we can't mock it directly
	// But we can verify it handles errors gracefully
	check := &NetworkCheck{}

	// With actual network, this should succeed or degrade gracefully
	result := check.Check(context.Background())

	// Either healthy or degraded, never critical
	if result.Status == health.StatusCritical {
		t.Error("NetworkCheck should never be critical")
	}
	if result.Name != "Network" {
		t.Errorf("Name = %v, want 'Network'", result.Name)
	}
	if result.Category != "external" {
		t.Errorf("Category = %v, want 'external'", result.Category)
	}
}

func TestCFSessionCheck_Check_NotAuthenticated(t *testing.T) {
	// Create a new session (not logged in)
	session, err := cfweb.NewSession()
	if err != nil {
		t.Skipf("NewSession() error: %v", err)
	}

	check := NewCFSessionCheck(session)
	result := check.Check(context.Background())

	// New session is not authenticated
	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	// Should have action to prompt user
	if result.Action != health.ActionUserPrompt {
		t.Errorf("Action = %v, want %v", result.Action, health.ActionUserPrompt)
	}
	// Should be recoverable
	if !result.Recoverable {
		t.Error("Session not authenticated should be recoverable")
	}
}
