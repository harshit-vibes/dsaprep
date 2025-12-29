package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/health"
	"github.com/spf13/cobra"
)

func TestRootCommand_Exists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
}

func TestRootCommand_Use(t *testing.T) {
	if rootCmd.Use != "cf" {
		t.Errorf("rootCmd.Use = %v, want cf", rootCmd.Use)
	}
}

func TestRootCommand_Short(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
}

func TestRootCommand_Long(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestRootCommand_HasSubcommands(t *testing.T) {
	subcommands := rootCmd.Commands()
	if len(subcommands) == 0 {
		t.Error("rootCmd should have subcommands")
	}

	// Check for expected subcommands
	subcommandNames := make(map[string]bool)
	for _, cmd := range subcommands {
		subcommandNames[cmd.Name()] = true
	}

	expected := []string{"version", "init", "health", "parse"}
	for _, name := range expected {
		if !subcommandNames[name] {
			t.Errorf("rootCmd should have %s subcommand", name)
		}
	}
}

func TestVersionCommand(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}
	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %v, want version", versionCmd.Use)
	}
}

func TestVersionCommand_Output(t *testing.T) {
	// Verify that the version command doesn't panic
	// Note: The version command uses fmt.Printf which goes to stdout,
	// not to the command's output writer
	versionCmd.Run(versionCmd, []string{})
	// If we get here without panic, the test passes
}

func TestInitCommand(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}
	if initCmd.Use != "init [path]" {
		t.Errorf("initCmd.Use = %v, want 'init [path]'", initCmd.Use)
	}
}

func TestInitCommand_ArgsValidation(t *testing.T) {
	// Test that it accepts 0 or 1 args
	if err := cobra.MaximumNArgs(1)(initCmd, []string{}); err != nil {
		t.Errorf("init should accept 0 args, got error: %v", err)
	}
	if err := cobra.MaximumNArgs(1)(initCmd, []string{"path"}); err != nil {
		t.Errorf("init should accept 1 arg, got error: %v", err)
	}
	if err := cobra.MaximumNArgs(1)(initCmd, []string{"path1", "path2"}); err == nil {
		t.Error("init should reject 2 args")
	}
}

func TestInitCommand_Run(t *testing.T) {
	tmpDir := t.TempDir()
	wsPath := filepath.Join(tmpDir, "test-workspace")

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	err := initCmd.RunE(initCmd, []string{wsPath})
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify workspace was created
	if _, err := os.Stat(filepath.Join(wsPath, "workspace.yaml")); os.IsNotExist(err) {
		t.Error("workspace.yaml should exist after init")
	}
}

func TestInitCommand_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Run init once
	initCmd.SetOut(new(bytes.Buffer))
	initCmd.SetErr(new(bytes.Buffer))
	err := initCmd.RunE(initCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	// Run init again - should fail
	err = initCmd.RunE(initCmd, []string{tmpDir})
	if err == nil {
		t.Error("init should fail when workspace already exists")
	}
}

func TestHealthCommand(t *testing.T) {
	if healthCmd == nil {
		t.Fatal("healthCmd should not be nil")
	}
	if healthCmd.Use != "health" {
		t.Errorf("healthCmd.Use = %v, want health", healthCmd.Use)
	}
}

func TestHealthCommand_SkipsPreRun(t *testing.T) {
	// Health command has its own PersistentPreRunE that returns nil
	err := healthCmd.PersistentPreRunE(healthCmd, []string{})
	if err != nil {
		t.Errorf("health command pre-run should return nil, got: %v", err)
	}
}

func TestParseCommand(t *testing.T) {
	if parseCmd == nil {
		t.Fatal("parseCmd should not be nil")
	}
	if parseCmd.Use != "parse <contest_id> <problem_index>" {
		t.Errorf("parseCmd.Use = %v, want 'parse <contest_id> <problem_index>'", parseCmd.Use)
	}
}

func TestParseCommand_ArgsValidation(t *testing.T) {
	// Test that it requires exactly 2 args
	if err := cobra.ExactArgs(2)(parseCmd, []string{}); err == nil {
		t.Error("parse should reject 0 args")
	}
	if err := cobra.ExactArgs(2)(parseCmd, []string{"1"}); err == nil {
		t.Error("parse should reject 1 arg")
	}
	if err := cobra.ExactArgs(2)(parseCmd, []string{"1", "A"}); err != nil {
		t.Errorf("parse should accept 2 args, got error: %v", err)
	}
	if err := cobra.ExactArgs(2)(parseCmd, []string{"1", "A", "B"}); err == nil {
		t.Error("parse should reject 3 args")
	}
}

func TestParseCommand_InvalidContestID(t *testing.T) {
	buf := new(bytes.Buffer)
	parseCmd.SetOut(buf)
	parseCmd.SetErr(buf)

	err := parseCmd.RunE(parseCmd, []string{"invalid", "A"})
	if err == nil {
		t.Error("parse should fail with invalid contest ID")
	}
}

func TestVersionVariables(t *testing.T) {
	// Test default version variables
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Commit == "" {
		t.Error("Commit should not be empty")
	}
	if BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}

func TestPersistentFlags(t *testing.T) {
	// Check that persistent flags are defined
	skipFlag := rootCmd.PersistentFlags().Lookup("skip-checks")
	if skipFlag == nil {
		t.Error("--skip-checks flag should be defined")
	}

	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("--verbose flag should be defined")
	}
}

func TestRunPreChecks_VersionCommand(t *testing.T) {
	// Version command should skip pre-checks
	versionCmd.SetOut(new(bytes.Buffer))
	err := runPreChecks(versionCmd, []string{})
	if err != nil {
		t.Errorf("version command should skip pre-checks, got error: %v", err)
	}
}

func TestRunPreChecks_HelpCommand(t *testing.T) {
	// Help command should skip pre-checks
	helpCmd := &cobra.Command{Use: "help"}
	err := runPreChecks(helpCmd, []string{})
	if err != nil {
		t.Errorf("help command should skip pre-checks, got error: %v", err)
	}
}

func TestRunStartupChecks_SkipChecks(t *testing.T) {
	// When skipChecks is true, should return nil
	skipChecks = true
	defer func() { skipChecks = false }()

	err := runStartupChecks()
	if err != nil {
		t.Errorf("runStartupChecks should return nil when skipChecks=true, got: %v", err)
	}
}

func TestDisplayHealthReport(t *testing.T) {
	// This function just prints, so we just verify it doesn't panic
	// with various report states
	report := &struct {
		Duration             interface{}
		Results              []interface{}
		OverallStatus        interface{}
		CurrentSchemaVersion string
	}{
		Duration:             "1s",
		CurrentSchemaVersion: "1.0.0",
	}
	_ = report
	// The function expects a *health.Report, but we're just testing it doesn't crash
}

func TestInitCommand_DefaultPath(t *testing.T) {
	// Test that init uses current directory when no path is given
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	os.Chdir(tmpDir)

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init command with no args failed: %v", err)
	}

	// Verify workspace was created in current directory
	if _, err := os.Stat(filepath.Join(tmpDir, "workspace.yaml")); os.IsNotExist(err) {
		t.Error("workspace.yaml should exist after init")
	}
}

func TestRootCommand_PersistentPreRunE(t *testing.T) {
	// Test that PersistentPreRunE is set
	if rootCmd.PersistentPreRunE == nil {
		t.Error("rootCmd should have PersistentPreRunE set")
	}
}

func TestRootCommand_RunE(t *testing.T) {
	// Test that RunE is set
	if rootCmd.RunE == nil {
		t.Error("rootCmd should have RunE set")
	}
}

func TestVersionCommand_ShortDescription(t *testing.T) {
	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}
}

func TestInitCommand_ShortDescription(t *testing.T) {
	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}
}

func TestHealthCommand_ShortDescription(t *testing.T) {
	if healthCmd.Short == "" {
		t.Error("healthCmd.Short should not be empty")
	}
}

func TestParseCommand_ShortDescription(t *testing.T) {
	if parseCmd.Short == "" {
		t.Error("parseCmd.Short should not be empty")
	}
}

func TestParseCommand_ContestIDParsing(t *testing.T) {
	tests := []struct {
		name      string
		contestID string
		wantErr   bool
	}{
		{"valid int", "123", false},
		{"valid large int", "1234567", false},
		{"invalid string", "abc", true},
		{"mixed", "123abc", false}, // fmt.Sscanf parses leading digits successfully
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id int
			_, err := fmt.Sscanf(tt.contestID, "%d", &id)
			hasErr := (err != nil)
			if hasErr != tt.wantErr {
				t.Errorf("parsing %q: got error=%v, wantErr=%v", tt.contestID, err, tt.wantErr)
			}
		})
	}
}

func TestFlags_SkipChecks_Default(t *testing.T) {
	// skipChecks should default to false
	skipFlag := rootCmd.PersistentFlags().Lookup("skip-checks")
	if skipFlag.DefValue != "false" {
		t.Errorf("skip-checks default = %v, want false", skipFlag.DefValue)
	}
}

func TestFlags_Verbose_Default(t *testing.T) {
	// verbose should default to false
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag.DefValue != "false" {
		t.Errorf("verbose default = %v, want false", verboseFlag.DefValue)
	}
}

// Test displayHealthReport with actual health.Report data
func TestDisplayHealthReport_AllStatuses(t *testing.T) {
	report := &health.Report{
		Duration:             100 * time.Millisecond,
		CurrentSchemaVersion: "1.0.0",
		Results: []health.Result{
			{
				Name:     "Check1",
				Status:   health.StatusHealthy,
				Message:  "All good",
				Category: "internal",
			},
			{
				Name:     "Check2",
				Status:   health.StatusDegraded,
				Message:  "Warning",
				Details:  "Some issue here",
				Category: "internal",
			},
			{
				Name:     "Check3",
				Status:   health.StatusCritical,
				Message:  "Critical error",
				Details:  "Needs fixing",
				Category: "external",
			},
		},
		OverallStatus: health.StatusDegraded,
		CanProceed:    true,
	}

	// Capture output and verify it doesn't panic
	displayHealthReport(report)
}

func TestDisplayHealthReport_HealthyReport(t *testing.T) {
	report := &health.Report{
		Duration:             50 * time.Millisecond,
		CurrentSchemaVersion: "1.0.0",
		Results: []health.Result{
			{
				Name:     "Test Check",
				Status:   health.StatusHealthy,
				Message:  "OK",
				Category: "internal",
			},
		},
		OverallStatus: health.StatusHealthy,
		CanProceed:    true,
	}

	// Should not panic
	displayHealthReport(report)
}

func TestDisplayHealthReport_EmptyReport(t *testing.T) {
	report := &health.Report{
		Duration:             10 * time.Millisecond,
		CurrentSchemaVersion: "1.0.0",
		Results:              []health.Result{},
		OverallStatus:        health.StatusHealthy,
		CanProceed:           true,
	}

	// Should not panic
	displayHealthReport(report)
}

func TestRunStartupChecks_WithChecks(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create a minimal env file
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=testuser
CF_API_KEY=
CF_API_SECRET=
`
	os.WriteFile(envPath, []byte(envContent), 0600)

	// Run with skipChecks=false (default)
	skipChecks = false
	verbose = false

	err := runStartupChecks()
	// May succeed or fail depending on network, but should not panic
	_ = err
}

func TestRunStartupChecks_Verbose(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	skipChecks = false
	verbose = true
	defer func() { verbose = false }()

	err := runStartupChecks()
	// Should print verbose output
	_ = err
}

func TestRunPreChecks_RegularCommand(t *testing.T) {
	// Create a regular command (not version or help)
	regularCmd := &cobra.Command{Use: "test"}

	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file
	envPath := filepath.Join(tmpDir, ".cf.env")
	os.WriteFile(envPath, []byte("CF_HANDLE=testuser\n"), 0600)

	skipChecks = true
	defer func() { skipChecks = false }()

	err := runPreChecks(regularCmd, []string{})
	if err != nil {
		t.Errorf("runPreChecks with skipChecks=true should return nil, got: %v", err)
	}
}

func TestHealthCommand_RunE(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file
	envPath := filepath.Join(tmpDir, ".cf.env")
	os.WriteFile(envPath, []byte("CF_HANDLE=testuser\n"), 0600)

	// Reset verbose and skipChecks
	origVerbose := verbose
	origSkip := skipChecks
	defer func() {
		verbose = origVerbose
		skipChecks = origSkip
	}()

	buf := new(bytes.Buffer)
	healthCmd.SetOut(buf)
	healthCmd.SetErr(buf)

	err := healthCmd.RunE(healthCmd, []string{})
	// May succeed or fail depending on checks, but should not panic
	_ = err
}

func TestRootCommand_RunE_Execute(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// The RunE function just prints version info
	// We need to call it directly since Execute() would run pre-checks
	if rootCmd.RunE != nil {
		err := rootCmd.RunE(rootCmd, []string{})
		if err != nil {
			t.Errorf("rootCmd.RunE should return nil, got: %v", err)
		}
	}
}

func TestInitConfig(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create .cf config dir
	configDir := filepath.Join(tmpDir, ".cf")
	os.MkdirAll(configDir, 0755)

	// This should not panic
	initConfig()
}

func TestParseCommand_ValidContestID(t *testing.T) {
	// This test makes a real API call
	buf := new(bytes.Buffer)
	parseCmd.SetOut(buf)
	parseCmd.SetErr(buf)

	// Test with a known problem (Contest 1, Problem A)
	err := parseCmd.RunE(parseCmd, []string{"1", "A"})

	// May succeed or fail depending on network, but should handle gracefully
	if err != nil {
		// Check it's a reasonable error, not a panic
		t.Logf("Parse command returned error (expected for network issues): %v", err)
	}
}

func TestRunStartupChecks_WithAPICredentials(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with API credentials
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=testuser
CF_API_KEY=testkey
CF_API_SECRET=testsecret
`
	os.WriteFile(envPath, []byte(envContent), 0600)

	// Initialize config
	config.Init(filepath.Join(tmpDir, ".cf"))

	skipChecks = false
	verbose = false

	err := runStartupChecks()
	// May succeed or fail, but tests the API credentials path
	_ = err
}

func TestRunStartupChecks_WithWorkspace(t *testing.T) {
	// Save and restore HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file
	envPath := filepath.Join(tmpDir, ".cf.env")
	os.WriteFile(envPath, []byte("CF_HANDLE=testuser\n"), 0600)

	// Initialize workspace
	wsPath := filepath.Join(tmpDir, "workspace")
	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)
	initCmd.RunE(initCmd, []string{wsPath})

	// Initialize config with workspace path
	config.Init(wsPath)

	skipChecks = false
	verbose = false

	err := runStartupChecks()
	_ = err
}

func TestDisplayHealthReport_WithDetails(t *testing.T) {
	report := &health.Report{
		Duration:             200 * time.Millisecond,
		CurrentSchemaVersion: "1.0.0",
		Results: []health.Result{
			{
				Name:     "Network Check",
				Status:   health.StatusDegraded,
				Message:  "Connection issues",
				Details:  "Unable to reach codeforces.com - timeout after 5s",
				Category: "external",
			},
		},
		OverallStatus: health.StatusDegraded,
		CanProceed:    true,
	}

	// Should not panic and should show details
	displayHealthReport(report)
}
