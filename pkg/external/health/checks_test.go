package health

import (
	"context"
	"testing"

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
	var _ health.Check = &CFHandleCheck{}
}

func TestAllChecksImplementCritical(t *testing.T) {
	// Verify all checks implement health.Critical interface
	var _ health.Critical = &CFAPICheck{}
	var _ health.Critical = &CFWebCheck{}
	var _ health.Critical = &CFHandleCheck{}
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

func TestExternalChecksAreDegraded(t *testing.T) {
	// External checks should be degraded, not critical, when unavailable
	checks := []health.Check{
		NewCFAPICheck(nil),
		NewCFWebCheck(nil),
		NewCFHandleCheck(nil),
	}

	for _, check := range checks {
		result := check.Check(context.Background())
		if result.Status == health.StatusCritical {
			t.Errorf("%s should not be critical when unavailable", check.Name())
		}
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
