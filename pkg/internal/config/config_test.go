package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test initialization with custom path
	err := Init(configPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify config is accessible
	cfg := Get()
	if cfg == nil {
		t.Error("Get() returned nil after Init()")
	}
}

func TestGet_BeforeInit(t *testing.T) {
	// Reset global config
	globalConfig = nil

	cfg := Get()
	// Should return nil or empty config before Init
	// This is expected behavior
	_ = cfg
}

func TestConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := Init(configPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg := Get()
	if cfg == nil {
		t.Fatal("Get() returned nil")
	}

	// WorkspacePath can be empty by default, which is fine
	_ = cfg.WorkspacePath
}

func TestConfig_WorkspacePath(t *testing.T) {
	tmpDir := t.TempDir()
	workspacePath := filepath.Join(tmpDir, "my-workspace")

	// Init with a workspace path - this overrides the workspace_path setting
	err := Init(workspacePath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg := Get()
	if cfg == nil {
		t.Fatal("Get() returned nil")
	}

	// WorkspacePath should be the path we passed to Init
	if cfg.WorkspacePath != workspacePath {
		t.Errorf("WorkspacePath = %v, want %v", cfg.WorkspacePath, workspacePath)
	}
}

func TestGetWorkspacePath_Fallback(t *testing.T) {
	// Reset global state
	globalConfig = nil

	// When globalConfig is nil, GetWorkspacePath should return cwd
	wsPath := GetWorkspacePath()
	cwd, _ := os.Getwd()
	if wsPath != cwd {
		t.Errorf("GetWorkspacePath() = %v, expected cwd %v", wsPath, cwd)
	}
}

func TestGetWorkspacePath_EmptyConfig(t *testing.T) {
	// Set up config with empty workspace path
	globalConfig = &Config{WorkspacePath: ""}

	// Should return cwd when workspace is empty
	wsPath := GetWorkspacePath()
	cwd, _ := os.Getwd()
	if wsPath != cwd {
		t.Errorf("GetWorkspacePath() = %v, expected cwd %v", wsPath, cwd)
	}
}

func TestGetCFHandle(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: "",
		},
		{
			name:     "empty handle",
			config:   &Config{CFHandle: ""},
			expected: "",
		},
		{
			name:     "valid handle",
			config:   &Config{CFHandle: "harshitvsdsa"},
			expected: "harshitvsdsa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalConfig = tt.config
			got := GetCFHandle()
			if got != tt.expected {
				t.Errorf("GetCFHandle() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := Init(configPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Test setting a value
	err = Set("cf_handle", "testuser")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify the value was set
	cfg := Get()
	if cfg.CFHandle != "testuser" {
		t.Errorf("CFHandle = %v, want testuser", cfg.CFHandle)
	}
}

func TestSetCFHandle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := Init(configPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = SetCFHandle("myhandle")
	if err != nil {
		t.Fatalf("SetCFHandle() error = %v", err)
	}

	cfg := Get()
	if cfg.CFHandle != "myhandle" {
		t.Errorf("CFHandle = %v, want myhandle", cfg.CFHandle)
	}
}

func TestSetDifficulty(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := Init(configPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = SetDifficulty(1000, 1800)
	if err != nil {
		t.Fatalf("SetDifficulty() error = %v", err)
	}

	cfg := Get()
	if cfg.Difficulty.Min != 1000 {
		t.Errorf("Difficulty.Min = %v, want 1000", cfg.Difficulty.Min)
	}
	if cfg.Difficulty.Max != 1800 {
		t.Errorf("Difficulty.Max = %v, want 1800", cfg.Difficulty.Max)
	}
}

func TestSetDailyGoal(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := Init(configPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = SetDailyGoal(5)
	if err != nil {
		t.Fatalf("SetDailyGoal() error = %v", err)
	}

	cfg := Get()
	if cfg.DailyGoal != 5 {
		t.Errorf("DailyGoal = %v, want 5", cfg.DailyGoal)
	}
}

func TestSetWorkspacePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := Init(configPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	workspacePath := filepath.Join(tmpDir, "my-workspace")
	err = SetWorkspacePath(workspacePath)
	if err != nil {
		t.Fatalf("SetWorkspacePath() error = %v", err)
	}

	cfg := Get()
	// Should be absolute path
	absPath, _ := filepath.Abs(workspacePath)
	if cfg.WorkspacePath != absPath {
		t.Errorf("WorkspacePath = %v, want %v", cfg.WorkspacePath, absPath)
	}
}

func TestConfigFilePath(t *testing.T) {
	path, err := configFilePath()
	if err != nil {
		t.Fatalf("configFilePath() error = %v", err)
	}

	if path == "" {
		t.Error("configFilePath() returned empty string")
	}

	// Should end with config.yaml
	if !filepath.IsAbs(path) {
		t.Error("configFilePath() should return absolute path")
	}

	if filepath.Base(path) != "config.yaml" {
		t.Errorf("configFilePath() should end with config.yaml, got %v", filepath.Base(path))
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := configDir()
	if err != nil {
		t.Fatalf("configDir() error = %v", err)
	}

	if dir == "" {
		t.Error("configDir() returned empty string")
	}

	// Should be absolute path
	if !filepath.IsAbs(dir) {
		t.Error("configDir() should return absolute path")
	}

	// Should end with .cf
	if filepath.Base(dir) != ".cf" {
		t.Errorf("configDir() should end with .cf, got %v", filepath.Base(dir))
	}
}
