package health

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/external/cfweb"
	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/health"
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
	// Set global config without handle
	config.SetGlobalConfig(&config.Config{CFHandle: "", Cookie: ""})

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
	// Set global config with handle
	config.SetGlobalConfig(&config.Config{CFHandle: "testuser", Cookie: ""})

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
	// Set global config with handle
	config.SetGlobalConfig(&config.Config{CFHandle: "testuser", Cookie: ""})

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
	// Set global config with a handle that doesn't exist
	config.SetGlobalConfig(&config.Config{CFHandle: "nonexistent_user_12345678", Cookie: ""})

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
	// Set global config with a known valid handle
	config.SetGlobalConfig(&config.Config{CFHandle: "tourist", Cookie: ""})

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

// ============ Helper Types and Functions ============

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// ============ Additional Coverage Tests ============

func TestFormatRating_Negative(t *testing.T) {
	// Edge case: negative rating (shouldn't happen but let's cover it)
	result := formatRating(-100)
	if result != "-100" {
		t.Errorf("formatRating(-100) = %v, want '-100'", result)
	}
}
