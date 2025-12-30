package health

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/workspace"
)

// ============ Config Tests ============

func TestConfigCheck_Name(t *testing.T) {
	check := &ConfigCheck{}
	if check.Name() != "Configuration" {
		t.Errorf("Name() = %v, want %v", check.Name(), "Configuration")
	}
}

func TestConfigCheck_Category(t *testing.T) {
	check := &ConfigCheck{}
	if check.Category() != "internal" {
		t.Errorf("Category() = %v, want %v", check.Category(), "internal")
	}
}

func TestConfigCheck_Check(t *testing.T) {
	check := &ConfigCheck{}
	result := check.Check(context.Background())

	// Result should have the check name
	if result.Name != check.Name() {
		t.Errorf("Result.Name = %v, want %v", result.Name, check.Name())
	}
	if result.Duration == 0 {
		t.Error("Result.Duration should not be zero")
	}
}

func TestConfigCheck_AutoFix(t *testing.T) {
	check := &ConfigCheck{}

	// AutoFix calls config.Init("")
	err := check.AutoFix(context.Background())
	// May succeed or fail depending on environment
	// Just verify it doesn't panic
	_ = err
}

// ============ Cookie Tests ============

func TestCookieCheck_Name(t *testing.T) {
	check := &CookieCheck{}
	if check.Name() != "Cookie" {
		t.Errorf("Name() = %v, want %v", check.Name(), "Cookie")
	}
}

func TestCookieCheck_Category(t *testing.T) {
	check := &CookieCheck{}
	if check.Category() != "internal" {
		t.Errorf("Category() = %v, want %v", check.Category(), "internal")
	}
}

func TestCookieCheck_IsCritical(t *testing.T) {
	check := &CookieCheck{}
	if check.IsCritical() {
		t.Error("IsCritical() should return false for CookieCheck")
	}
}

func TestCookieCheck_Check_NoCookie(t *testing.T) {
	// Set global config without cookie
	config.SetGlobalConfig(&config.Config{CFHandle: "testuser", Cookie: ""})

	check := &CookieCheck{}
	result := check.Check(context.Background())

	if result.Status != StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, StatusDegraded)
	}
	if result.Message != "Browser cookie not configured" {
		t.Errorf("Message = %v, want 'Browser cookie not configured'", result.Message)
	}
	if result.Action != ActionUserPrompt {
		t.Errorf("Action = %v, want %v", result.Action, ActionUserPrompt)
	}
}

func TestCookieCheck_Check_WithCookie(t *testing.T) {
	// Set global config with cookie
	config.SetGlobalConfig(&config.Config{CFHandle: "testuser", Cookie: "JSESSIONID=test123"})

	check := &CookieCheck{}
	result := check.Check(context.Background())

	if result.Status != StatusHealthy {
		t.Errorf("Status = %v, want %v", result.Status, StatusHealthy)
	}
	if result.Message != "Cookie configured" {
		t.Errorf("Message = %v, want 'Cookie configured'", result.Message)
	}
}

// ============ Workspace Tests ============

func TestWorkspaceCheck_Name(t *testing.T) {
	ws := workspace.New("/tmp/test")
	check := NewWorkspaceCheck(ws)
	if check.Name() != "Workspace" {
		t.Errorf("Name() = %v, want %v", check.Name(), "Workspace")
	}
}

func TestWorkspaceCheck_Category(t *testing.T) {
	ws := workspace.New("/tmp/test")
	check := NewWorkspaceCheck(ws)
	if check.Category() != "internal" {
		t.Errorf("Category() = %v, want %v", check.Category(), "internal")
	}
}

func TestWorkspaceCheck_Check_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(filepath.Join(tmpDir, "nonexistent"))
	check := NewWorkspaceCheck(ws)

	result := check.Check(context.Background())

	if result.Status != StatusCritical {
		t.Errorf("Status = %v, want %v", result.Status, StatusCritical)
	}
	if !result.Recoverable {
		t.Error("Recoverable should be true for missing workspace")
	}
	if result.Action != ActionAutoFix {
		t.Errorf("Action = %v, want %v", result.Action, ActionAutoFix)
	}
}

func TestWorkspaceCheck_Check_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(tmpDir)

	// Initialize workspace
	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	check := NewWorkspaceCheck(ws)
	result := check.Check(context.Background())

	if result.Status != StatusHealthy {
		t.Errorf("Status = %v, want %v (message: %s)", result.Status, StatusHealthy, result.Message)
	}
}

func TestWorkspaceCheck_AutoFix(t *testing.T) {
	tmpDir := t.TempDir()
	wsPath := filepath.Join(tmpDir, "new-workspace")
	ws := workspace.New(wsPath)
	check := NewWorkspaceCheck(ws)

	// AutoFix should create the workspace
	err := check.AutoFix(context.Background())
	if err != nil {
		t.Fatalf("AutoFix() error = %v", err)
	}

	// Workspace should now exist
	if !ws.Exists() {
		t.Error("AutoFix() should create workspace")
	}
}

func TestNewWorkspaceCheck(t *testing.T) {
	ws := workspace.New("/tmp/test")
	check := NewWorkspaceCheck(ws)

	if check == nil {
		t.Fatal("NewWorkspaceCheck() returned nil")
	}
	if check.ws != ws {
		t.Error("NewWorkspaceCheck() should set ws field")
	}
}

// WorkspaceCheck test for validation failure
func TestWorkspaceCheck_Check_ValidationFail(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(tmpDir)

	// Initialize workspace
	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Delete a required directory to cause validation failure
	problemsDir := filepath.Join(tmpDir, "problems")
	os.RemoveAll(problemsDir)

	check := NewWorkspaceCheck(ws)
	result := check.Check(context.Background())

	// Should return critical (validation failed)
	if result.Status != StatusCritical {
		t.Errorf("Status = %v, want %v", result.Status, StatusCritical)
	}
	if result.Message != "Workspace validation failed" {
		t.Errorf("Message = %v, want 'Workspace validation failed'", result.Message)
	}
}

// ============ Schema Version Tests ============

func TestSchemaVersionCheck_Name(t *testing.T) {
	ws := workspace.New("/tmp/test")
	check := NewSchemaVersionCheck(ws)
	if check.Name() != "Schema Version" {
		t.Errorf("Name() = %v, want %v", check.Name(), "Schema Version")
	}
}

func TestSchemaVersionCheck_Category(t *testing.T) {
	ws := workspace.New("/tmp/test")
	check := NewSchemaVersionCheck(ws)
	if check.Category() != "internal" {
		t.Errorf("Category() = %v, want %v", check.Category(), "internal")
	}
}

func TestSchemaVersionCheck_Check_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(filepath.Join(tmpDir, "nonexistent"))
	check := NewSchemaVersionCheck(ws)

	result := check.Check(context.Background())

	// No workspace means nothing to check
	if result.Status != StatusHealthy {
		t.Errorf("Status = %v, want %v (no workspace to check)", result.Status, StatusHealthy)
	}
}

func TestSchemaVersionCheck_Check_ValidVersion(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(tmpDir)

	err := ws.Init("Test", "user")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	check := NewSchemaVersionCheck(ws)
	result := check.Check(context.Background())

	if result.Status != StatusHealthy {
		t.Errorf("Status = %v, want %v (message: %s)", result.Status, StatusHealthy, result.Message)
	}
}

func TestNewSchemaVersionCheck(t *testing.T) {
	ws := workspace.New("/tmp/test")
	check := NewSchemaVersionCheck(ws)

	if check == nil {
		t.Fatal("NewSchemaVersionCheck() returned nil")
	}
	if check.ws != ws {
		t.Error("NewSchemaVersionCheck() should set ws field")
	}
}

// SchemaVersionCheck tests for uncovered branches
func TestSchemaVersionCheck_Check_GetVersionError(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(tmpDir)

	// Create a workspace with corrupted manifest
	os.MkdirAll(tmpDir, 0755)
	manifestPath := filepath.Join(tmpDir, "workspace.yaml")
	// Write invalid YAML
	err := os.WriteFile(manifestPath, []byte("this is not valid yaml: [[["), 0644)
	if err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	check := NewSchemaVersionCheck(ws)
	result := check.Check(context.Background())

	// Should return critical (cannot read schema version)
	if result.Status != StatusCritical {
		t.Errorf("Status = %v, want %v", result.Status, StatusCritical)
	}
	if result.Message != "Cannot read schema version" {
		t.Errorf("Message = %v, want 'Cannot read schema version'", result.Message)
	}
}

func TestSchemaVersionCheck_Check_IncompatibleVersion(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(tmpDir)

	// Create workspace with incompatible version (major version 2)
	os.MkdirAll(tmpDir, 0755)
	manifestPath := filepath.Join(tmpDir, "workspace.yaml")
	// Use correct _schema field name
	manifest := `_schema:
  version: "2.0.0"
  type: workspace
name: Test
codeforces:
  handle: user
`
	err := os.WriteFile(manifestPath, []byte(manifest), 0644)
	if err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	check := NewSchemaVersionCheck(ws)
	result := check.Check(context.Background())

	// Should return critical (incompatible version)
	if result.Status != StatusCritical {
		t.Errorf("Status = %v, want %v", result.Status, StatusCritical)
	}
	if result.Message != "Incompatible schema version" {
		t.Errorf("Message = %v, want 'Incompatible schema version'", result.Message)
	}
	if result.Action != ActionManualFix {
		t.Errorf("Action = %v, want %v", result.Action, ActionManualFix)
	}
}

func TestSchemaVersionCheck_Check_NeedsMigration(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(tmpDir)

	// Create workspace with older minor version (needs migration)
	os.MkdirAll(tmpDir, 0755)
	manifestPath := filepath.Join(tmpDir, "workspace.yaml")
	// Version 1.1.0 is same major but different minor - needs migration
	// Use correct _schema field name
	manifest := `_schema:
  version: "1.1.0"
  type: workspace
name: Test
codeforces:
  handle: user
`
	err := os.WriteFile(manifestPath, []byte(manifest), 0644)
	if err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	check := NewSchemaVersionCheck(ws)
	result := check.Check(context.Background())

	// Should return degraded (migration available)
	if result.Status != StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, StatusDegraded)
	}
	if result.Message != "Schema migration available" {
		t.Errorf("Message = %v, want 'Schema migration available'", result.Message)
	}
	if !result.Recoverable {
		t.Error("Recoverable should be true for migration")
	}
	if result.Action != ActionUserPrompt {
		t.Errorf("Action = %v, want %v", result.Action, ActionUserPrompt)
	}
}

// Integration test for full check flow
func TestCheckerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(tmpDir)

	checker := NewChecker()
	checker.AddCheck(NewWorkspaceCheck(ws))
	checker.AddCheck(NewSchemaVersionCheck(ws))

	// Run checks on uninitialized workspace
	// Note: WorkspaceCheck has AutoFix which will initialize the workspace
	report := checker.Run(context.Background())

	// The auto-fix should have worked, so report may be healthy
	// Verify that results contain the workspace check
	if len(report.Results) == 0 {
		t.Error("Report should have results")
	}

	// The workspace should now exist after auto-fix
	if !ws.Exists() {
		t.Error("Workspace should exist after auto-fix")
	}

	// Run checks again
	report = checker.Run(context.Background())

	// Both checks should pass on initialized workspace
	if report.OverallStatus != StatusHealthy {
		t.Errorf("Report.OverallStatus = %v, want %v after init", report.OverallStatus, StatusHealthy)
	}
	if !report.CanProceed {
		t.Error("Report.CanProceed should be true")
	}
}

// Test checker without auto-fix
func TestCheckerWithoutAutoFix(t *testing.T) {
	tmpDir := t.TempDir()
	ws := workspace.New(tmpDir)

	checker := NewChecker()
	// Only add SchemaVersionCheck which returns healthy for non-existent workspace
	checker.AddCheck(NewSchemaVersionCheck(ws))

	report := checker.Run(context.Background())

	// SchemaVersionCheck returns healthy when workspace doesn't exist
	if report.OverallStatus != StatusHealthy {
		t.Errorf("Report.OverallStatus = %v, want %v", report.OverallStatus, StatusHealthy)
	}
}
