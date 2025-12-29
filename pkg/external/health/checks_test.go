package health

import (
	"context"
	"testing"
	"time"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/external/cfweb"
	"github.com/harshit-vibes/cf/pkg/internal/health"
)

func TestCFAPICheck_Name(t *testing.T) {
	check := NewCFAPICheck(nil)
	if check.Name() != "CF API" {
		t.Errorf("Name() = %v, want 'CF API'", check.Name())
	}
}

func TestCFAPICheck_Category(t *testing.T) {
	check := NewCFAPICheck(nil)
	if check.Category() != "external" {
		t.Errorf("Category() = %v, want 'external'", check.Category())
	}
}

func TestCFAPICheck_IsCritical(t *testing.T) {
	check := NewCFAPICheck(nil)
	if check.IsCritical() {
		t.Error("IsCritical() should return false")
	}
}

func TestCFAPICheck_Check_NilClient(t *testing.T) {
	check := NewCFAPICheck(nil)
	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Name != check.Name() {
		t.Errorf("Result.Name = %v, want %v", result.Name, check.Name())
	}
}

func TestCFAPICheck_Check_WithClient(t *testing.T) {
	client := cfapi.NewClient()
	check := NewCFAPICheck(client)

	result := check.Check(context.Background())

	// Should have Name and Category set
	if result.Name != check.Name() {
		t.Errorf("Result.Name = %v, want %v", result.Name, check.Name())
	}
	if result.Category != check.Category() {
		t.Errorf("Result.Category = %v, want %v", result.Category, check.Category())
	}
	// Duration should be set
	if result.Duration == 0 {
		t.Error("Result.Duration should not be zero")
	}
}

func TestNewCFAPICheck(t *testing.T) {
	client := cfapi.NewClient()
	check := NewCFAPICheck(client)

	if check == nil {
		t.Fatal("NewCFAPICheck() returned nil")
	}
	if check.client != client {
		t.Error("client not set correctly")
	}
}

func TestCFWebCheck_Name(t *testing.T) {
	check := NewCFWebCheck(nil)
	if check.Name() != "CF Web Structure" {
		t.Errorf("Name() = %v, want 'CF Web Structure'", check.Name())
	}
}

func TestCFWebCheck_Category(t *testing.T) {
	check := NewCFWebCheck(nil)
	if check.Category() != "external" {
		t.Errorf("Category() = %v, want 'external'", check.Category())
	}
}

func TestCFWebCheck_IsCritical(t *testing.T) {
	check := NewCFWebCheck(nil)
	if check.IsCritical() {
		t.Error("IsCritical() should return false")
	}
}

func TestCFWebCheck_Check_NilParser(t *testing.T) {
	check := NewCFWebCheck(nil)
	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "Parser not initialized" {
		t.Errorf("Message = %v", result.Message)
	}
}

func TestNewCFWebCheck(t *testing.T) {
	parser := cfweb.NewParser(nil)
	check := NewCFWebCheck(parser)

	if check == nil {
		t.Fatal("NewCFWebCheck() returned nil")
	}
	if check.parser != parser {
		t.Error("parser not set correctly")
	}
}

func TestCFSessionCheck_Name(t *testing.T) {
	check := NewCFSessionCheck(nil)
	if check.Name() != "CF Session" {
		t.Errorf("Name() = %v, want 'CF Session'", check.Name())
	}
}

func TestCFSessionCheck_Category(t *testing.T) {
	check := NewCFSessionCheck(nil)
	if check.Category() != "external" {
		t.Errorf("Category() = %v, want 'external'", check.Category())
	}
}

func TestCFSessionCheck_IsCritical(t *testing.T) {
	check := NewCFSessionCheck(nil)
	if check.IsCritical() {
		t.Error("IsCritical() should return false")
	}
}

func TestCFSessionCheck_Check_NilSession(t *testing.T) {
	check := NewCFSessionCheck(nil)
	result := check.Check(context.Background())

	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	if result.Message != "No active session" {
		t.Errorf("Message = %v", result.Message)
	}
	if result.Action != health.ActionUserPrompt {
		t.Errorf("Action = %v, want %v", result.Action, health.ActionUserPrompt)
	}
}

func TestNewCFSessionCheck(t *testing.T) {
	check := NewCFSessionCheck(nil)

	if check == nil {
		t.Fatal("NewCFSessionCheck() returned nil")
	}
}

func TestCFHandleCheck_Name(t *testing.T) {
	check := NewCFHandleCheck(nil)
	if check.Name() != "CF Handle" {
		t.Errorf("Name() = %v, want 'CF Handle'", check.Name())
	}
}

func TestCFHandleCheck_Category(t *testing.T) {
	check := NewCFHandleCheck(nil)
	if check.Category() != "external" {
		t.Errorf("Category() = %v, want 'external'", check.Category())
	}
}

func TestCFHandleCheck_IsCritical(t *testing.T) {
	check := NewCFHandleCheck(nil)
	if check.IsCritical() {
		t.Error("IsCritical() should return false")
	}
}

func TestNewCFHandleCheck(t *testing.T) {
	client := cfapi.NewClient()
	check := NewCFHandleCheck(client)

	if check == nil {
		t.Fatal("NewCFHandleCheck() returned nil")
	}
	if check.client != client {
		t.Error("client not set correctly")
	}
}

func TestNetworkCheck_Name(t *testing.T) {
	check := &NetworkCheck{}
	if check.Name() != "Network" {
		t.Errorf("Name() = %v, want 'Network'", check.Name())
	}
}

func TestNetworkCheck_Category(t *testing.T) {
	check := &NetworkCheck{}
	if check.Category() != "external" {
		t.Errorf("Category() = %v, want 'external'", check.Category())
	}
}

func TestNetworkCheck_IsCritical(t *testing.T) {
	check := &NetworkCheck{}
	if check.IsCritical() {
		t.Error("IsCritical() should return false")
	}
}

func TestFormatRating(t *testing.T) {
	tests := []struct {
		rating int
		want   string
	}{
		{0, "unrated"},
		{800, "800"},
		{1400, "1400"},
		{2100, "2100"},
		{3500, "3500"},
	}

	for _, tt := range tests {
		got := formatRating(tt.rating)
		if got != tt.want {
			t.Errorf("formatRating(%d) = %v, want %v", tt.rating, got, tt.want)
		}
	}
}

func TestAllChecksImplementInterface(t *testing.T) {
	// Verify all checks implement health.Check interface
	var _ health.Check = &CFAPICheck{}
	var _ health.Check = &CFWebCheck{}
	var _ health.Check = &CFSessionCheck{}
	var _ health.Check = &CFHandleCheck{}
	var _ health.Check = &NetworkCheck{}
}

func TestAllChecksImplementCritical(t *testing.T) {
	// Verify all checks implement health.Critical interface
	var _ health.Critical = &CFAPICheck{}
	var _ health.Critical = &CFWebCheck{}
	var _ health.Critical = &CFSessionCheck{}
	var _ health.Critical = &CFHandleCheck{}
	var _ health.Critical = &NetworkCheck{}
}

func TestCFAPICheck_CheckReturnsResult(t *testing.T) {
	check := NewCFAPICheck(nil)
	result := check.Check(context.Background())

	// Verify result fields are populated
	if result.Name == "" {
		t.Error("Result.Name should not be empty")
	}
	if result.Category == "" {
		t.Error("Result.Category should not be empty")
	}
	if result.Message == "" {
		t.Error("Result.Message should not be empty")
	}
}

func TestCFWebCheck_CheckReturnsResult(t *testing.T) {
	check := NewCFWebCheck(nil)
	result := check.Check(context.Background())

	if result.Name == "" {
		t.Error("Result.Name should not be empty")
	}
	if result.Category == "" {
		t.Error("Result.Category should not be empty")
	}
	if result.Message == "" {
		t.Error("Result.Message should not be empty")
	}
}

func TestCFSessionCheck_CheckReturnsResult(t *testing.T) {
	check := NewCFSessionCheck(nil)
	result := check.Check(context.Background())

	if result.Name == "" {
		t.Error("Result.Name should not be empty")
	}
	if result.Category == "" {
		t.Error("Result.Category should not be empty")
	}
	if result.Message == "" {
		t.Error("Result.Message should not be empty")
	}
}

func TestExternalChecksAreDegraded(t *testing.T) {
	// External checks should be degraded, not critical, when unavailable
	checks := []health.Check{
		NewCFAPICheck(nil),
		NewCFWebCheck(nil),
		NewCFSessionCheck(nil),
		NewCFHandleCheck(nil),
	}

	for _, check := range checks {
		result := check.Check(context.Background())
		if result.Status == health.StatusCritical {
			t.Errorf("%s should not be critical when unavailable", check.Name())
		}
	}
}

func TestCFClearanceCheck_Name(t *testing.T) {
	check := NewCFClearanceCheck()
	if check.Name() != "CF Clearance" {
		t.Errorf("Name() = %v, want 'CF Clearance'", check.Name())
	}
}

func TestCFClearanceCheck_Category(t *testing.T) {
	check := NewCFClearanceCheck()
	if check.Category() != "external" {
		t.Errorf("Category() = %v, want 'external'", check.Category())
	}
}

func TestCFClearanceCheck_IsCritical(t *testing.T) {
	check := NewCFClearanceCheck()
	if check.IsCritical() {
		t.Error("IsCritical() should return false")
	}
}

func TestNewCFClearanceCheck(t *testing.T) {
	check := NewCFClearanceCheck()
	if check == nil {
		t.Fatal("NewCFClearanceCheck() returned nil")
	}
}

func TestCFClearanceCheck_CheckReturnsResult(t *testing.T) {
	check := NewCFClearanceCheck()
	result := check.Check(context.Background())

	if result.Name == "" {
		t.Error("Result.Name should not be empty")
	}
	if result.Category == "" {
		t.Error("Result.Category should not be empty")
	}
	if result.Message == "" {
		t.Error("Result.Message should not be empty")
	}
}

func TestCFClearanceCheck_ImplementsInterface(t *testing.T) {
	var _ health.Check = &CFClearanceCheck{}
	var _ health.Critical = &CFClearanceCheck{}
}

func TestFormatDuration_External(t *testing.T) {
	// Test the formatDuration helper function
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"negative", -1 * time.Hour, "expired"},
		{"seconds", 30 * time.Second, "30s"},
		{"minutes", 5 * time.Minute, "5m"},
		{"hours", 2*time.Hour + 30*time.Minute, "2h 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Tests with real clients (integration tests - not tagged as they are quick)

func TestCFAPICheck_Check_APISuccess(t *testing.T) {
	// This test makes a real API call
	client := cfapi.NewClient()
	check := NewCFAPICheck(client)

	result := check.Check(context.Background())

	// Should either succeed or fail gracefully
	if result.Status == health.StatusHealthy {
		if result.Message != "CF API OK" {
			t.Errorf("Message = %v, want 'CF API OK'", result.Message)
		}
	} else {
		// Network issues - check it's degraded not critical
		if result.Status == health.StatusCritical {
			t.Error("Should be degraded, not critical")
		}
	}
}

func TestCFWebCheck_Check_WithParser(t *testing.T) {
	// This test makes a real web request
	parser := cfweb.NewParser(nil)
	check := NewCFWebCheck(parser)

	result := check.Check(context.Background())

	// Either succeeds or degrades gracefully
	if result.Status == health.StatusHealthy {
		if result.Message == "" {
			t.Error("Message should not be empty on success")
		}
	} else {
		// Should be degraded, not critical
		if result.Status == health.StatusCritical {
			t.Error("Should be degraded, not critical")
		}
	}
}

func TestCFHandleCheck_Check_NilClient(t *testing.T) {
	// Test with nil client but handle configured
	check := NewCFHandleCheck(nil)
	result := check.Check(context.Background())

	// Should be degraded (either no handle or can't verify)
	if result.Status == health.StatusHealthy {
		// If it's healthy, handle must exist
		t.Log("Handle check passed (handle configured locally)")
	}
	if result.Name != "CF Handle" {
		t.Errorf("Name = %v, want 'CF Handle'", result.Name)
	}
}

func TestNetworkCheck_Check(t *testing.T) {
	check := &NetworkCheck{}
	result := check.Check(context.Background())

	// Either succeeds or degrades gracefully
	if result.Status == health.StatusHealthy {
		if result.Message != "Network OK" {
			t.Errorf("Message = %v, want 'Network OK'", result.Message)
		}
	} else {
		// Network issues - check it's degraded not critical
		if result.Status == health.StatusCritical {
			t.Error("Should be degraded, not critical for network issues")
		}
	}
	// Duration should always be set
	if result.Duration == 0 {
		t.Error("Duration should not be zero")
	}
}

// Test CFSessionCheck with a session that exists but is not logged in
func TestCFSessionCheck_Check_SessionNotLoggedIn(t *testing.T) {
	// Create a session that's not logged in
	session, err := cfweb.NewSession()
	if err != nil {
		t.Skipf("Skipping: NewSession() failed: %v", err)
	}

	check := NewCFSessionCheck(session)
	result := check.Check(context.Background())

	// Session exists but not logged in - should be degraded
	if result.Status != health.StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, health.StatusDegraded)
	}
	// Should have meaningful message
	if result.Message == "" {
		t.Error("Message should not be empty")
	}
}
