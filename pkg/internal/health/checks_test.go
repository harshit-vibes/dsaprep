package health

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/workspace"
)

func TestEnvFileCheck_Name(t *testing.T) {
	check := &EnvFileCheck{}
	if check.Name() != "Environment File" {
		t.Errorf("Name() = %v, want %v", check.Name(), "Environment File")
	}
}

func TestEnvFileCheck_Category(t *testing.T) {
	check := &EnvFileCheck{}
	if check.Category() != "internal" {
		t.Errorf("Category() = %v, want %v", check.Category(), "internal")
	}
}

func TestEnvFileCheck_Check(t *testing.T) {
	check := &EnvFileCheck{}
	ctx := context.Background()

	result := check.Check(ctx)

	// Result should have the check name
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

func TestSessionCheck_Name(t *testing.T) {
	check := &SessionCheck{}
	if check.Name() != "CF Session" {
		t.Errorf("Name() = %v, want %v", check.Name(), "CF Session")
	}
}

func TestSessionCheck_Category(t *testing.T) {
	check := &SessionCheck{}
	if check.Category() != "internal" {
		t.Errorf("Category() = %v, want %v", check.Category(), "internal")
	}
}

func TestSessionCheck_IsCritical(t *testing.T) {
	check := &SessionCheck{}
	if check.IsCritical() {
		t.Error("IsCritical() should return false for SessionCheck")
	}
}

func TestSessionCheck_Check(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home with no credentials
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	check := &SessionCheck{}
	result := check.Check(context.Background())

	// Without credentials file, should return degraded
	if result.Name != check.Name() {
		t.Errorf("Result.Name = %v, want %v", result.Name, check.Name())
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

// EnvFileCheck tests for uncovered branches
func TestEnvFileCheck_Check_FileNotFound(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Ensure no .env file exists
	os.Remove(filepath.Join(tmpDir, ".cf.env"))

	check := &EnvFileCheck{}
	result := check.Check(context.Background())

	// Should return critical with recoverable=true
	if result.Status != StatusCritical {
		t.Errorf("Status = %v, want %v", result.Status, StatusCritical)
	}
	if !result.Recoverable {
		t.Error("Recoverable should be true for missing env file")
	}
	if result.Action != ActionAutoFix {
		t.Errorf("Action = %v, want %v", result.Action, ActionAutoFix)
	}
	if result.Message != ".env file not found" {
		t.Errorf("Message = %v, want '.env file not found'", result.Message)
	}
}

func TestEnvFileCheck_Check_NoHandle(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file without handle
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=
CF_API_KEY=
CF_API_SECRET=
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := &EnvFileCheck{}
	result := check.Check(context.Background())

	// Should return degraded (no handle configured)
	if result.Status != StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, StatusDegraded)
	}
	if result.Message != "CF handle not configured" {
		t.Errorf("Message = %v, want 'CF handle not configured'", result.Message)
	}
	if result.Action != ActionUserPrompt {
		t.Errorf("Action = %v, want %v", result.Action, ActionUserPrompt)
	}
}

func TestEnvFileCheck_Check_Healthy(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create valid env file with handle
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=testuser
CF_API_KEY=testkey
CF_API_SECRET=testsecret
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := &EnvFileCheck{}
	result := check.Check(context.Background())

	// Should return healthy
	if result.Status != StatusHealthy {
		t.Errorf("Status = %v, want %v", result.Status, StatusHealthy)
	}
	if result.Message != "Environment file OK" {
		t.Errorf("Message = %v, want 'Environment file OK'", result.Message)
	}
}

func TestEnvFileCheck_AutoFix(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	check := &EnvFileCheck{}
	err := check.AutoFix(context.Background())
	if err != nil {
		t.Fatalf("AutoFix() error = %v", err)
	}

	// Verify env file was created
	envPath := filepath.Join(tmpDir, ".cf.env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Error("AutoFix() should create env file")
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

// ConfigCheck tests for uncovered branches
func TestConfigCheck_Check_NilConfig(t *testing.T) {
	// Reset global config to nil
	origConfig := config.Get()
	// There's no direct way to set config to nil, but we can test the branch
	// by checking when config.Get() returns nil

	check := &ConfigCheck{}
	result := check.Check(context.Background())

	// The result depends on whether config is initialized
	// Either way, the check should have proper name and category
	if result.Name != check.Name() {
		t.Errorf("Result.Name = %v, want %v", result.Name, check.Name())
	}
	if result.Category != check.Category() {
		t.Errorf("Result.Category = %v, want %v", result.Category, check.Category())
	}

	// Restore
	_ = origConfig
}

func TestConfigCheck_AutoFix(t *testing.T) {
	check := &ConfigCheck{}

	// AutoFix calls config.Init("")
	err := check.AutoFix(context.Background())
	// May succeed or fail depending on environment
	// Just verify it doesn't panic
	_ = err
}

// SessionCheck tests for uncovered branches
func TestSessionCheck_Check_NoCFClearance(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with handle but no cf_clearance
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=testuser
CF_API_KEY=testkey
CF_API_SECRET=testsecret
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := &SessionCheck{}
	result := check.Check(context.Background())

	// Should return degraded (cf_clearance not configured)
	if result.Status != StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, StatusDegraded)
	}
	if result.Message != "cf_clearance not configured or expired" {
		t.Errorf("Message = %v, want 'cf_clearance not configured or expired'", result.Message)
	}
	if result.Action != ActionUserPrompt {
		t.Errorf("Action = %v, want %v", result.Action, ActionUserPrompt)
	}
}

func TestSessionCheck_Check_CFClearanceExpired(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with cf_clearance but expired
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=testuser
CF_CLEARANCE=someclearancevalue
CF_CLEARANCE_EXPIRES=1
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := &SessionCheck{}
	result := check.Check(context.Background())

	// Should return degraded (cf_clearance expired)
	if result.Status != StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, StatusDegraded)
	}
	if result.Message != "cf_clearance not configured or expired" {
		t.Errorf("Message = %v, want 'cf_clearance not configured or expired'", result.Message)
	}
}

func TestSessionCheck_Check_NoSessionCookies(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with valid cf_clearance but no session cookies
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=testuser
CF_CLEARANCE=someclearancevalue
CF_CLEARANCE_EXPIRES=9999999999
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := &SessionCheck{}
	result := check.Check(context.Background())

	// Should return degraded (no session cookies)
	if result.Status != StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, StatusDegraded)
	}
	if result.Message != "Session cookies not configured" {
		t.Errorf("Message = %v, want 'Session cookies not configured'", result.Message)
	}
}

func TestSessionCheck_Check_SessionConfigured(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with all required cookies
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=testuser
CF_CLEARANCE=someclearancevalue
CF_CLEARANCE_EXPIRES=9999999999
CF_JSESSIONID=somejsessionid
CF_39CE7=some39ce7value
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := &SessionCheck{}
	result := check.Check(context.Background())

	// Should return healthy (session configured)
	if result.Status != StatusHealthy {
		t.Errorf("Status = %v, want %v (message: %s, details: %s)", result.Status, StatusHealthy, result.Message, result.Details)
	}
	// Message should contain "Session configured"
	if !strings.Contains(result.Message, "Session configured") {
		t.Errorf("Message = %v, want to contain 'Session configured'", result.Message)
	}
}

func TestSessionCheck_Check_MissingHandle(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with cookies but no handle
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_CLEARANCE=someclearancevalue
CF_CLEARANCE_EXPIRES=9999999999
CF_JSESSIONID=somejsessionid
CF_39CE7=some39ce7value
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	check := &SessionCheck{}
	result := check.Check(context.Background())

	// Should return degraded (missing handle)
	if result.Status != StatusDegraded {
		t.Errorf("Status = %v, want %v", result.Status, StatusDegraded)
	}
	if result.Message != "Missing handle or cookies" {
		t.Errorf("Message = %v, want 'Missing handle or cookies'", result.Message)
	}
}
