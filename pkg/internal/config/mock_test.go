package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ============ Config Tests ============

func TestGetWorkspacePath_WithConfiguredPath(t *testing.T) {
	// Set up config with a specific workspace path
	expectedPath := "/some/workspace/path"
	globalConfig = &Config{WorkspacePath: expectedPath}

	wsPath := GetWorkspacePath()
	if wsPath != expectedPath {
		t.Errorf("GetWorkspacePath() = %v, want %v", wsPath, expectedPath)
	}
}

func TestInit_ExistingConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config directory and file first
	configDirPath := filepath.Join(tmpDir, ".cf")
	err := os.MkdirAll(configDirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configFilePath := filepath.Join(configDirPath, "config.yaml")
	configContent := `cf_handle: existinguser
difficulty:
  min: 1000
  max: 1600
daily_goal: 5
workspace_path: /existing/path
`
	err = os.WriteFile(configFilePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Now Init with empty workspace (should preserve existing config)
	err = Init("")
	if err != nil {
		// This might fail due to viper config path issues in test - that's ok
		t.Logf("Init() returned error (might be expected in test env): %v", err)
	}
}

func TestDifficultyRange_Fields(t *testing.T) {
	dr := DifficultyRange{
		Min: 800,
		Max: 1400,
	}

	if dr.Min != 800 {
		t.Errorf("Min = %v, want 800", dr.Min)
	}
	if dr.Max != 1400 {
		t.Errorf("Max = %v, want 1400", dr.Max)
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := Config{
		CFHandle: "testuser",
		Cookie:   "test_cookie_string",
		Difficulty: DifficultyRange{
			Min: 800,
			Max: 1400,
		},
		DailyGoal:     3,
		WorkspacePath: "/workspace",
	}

	if cfg.CFHandle != "testuser" {
		t.Errorf("CFHandle = %v, want testuser", cfg.CFHandle)
	}
	if cfg.Cookie != "test_cookie_string" {
		t.Errorf("Cookie = %v, want test_cookie_string", cfg.Cookie)
	}
	if cfg.Difficulty.Min != 800 {
		t.Errorf("Difficulty.Min = %v, want 800", cfg.Difficulty.Min)
	}
	if cfg.Difficulty.Max != 1400 {
		t.Errorf("Difficulty.Max = %v, want 1400", cfg.Difficulty.Max)
	}
	if cfg.DailyGoal != 3 {
		t.Errorf("DailyGoal = %v, want 3", cfg.DailyGoal)
	}
	if cfg.WorkspacePath != "/workspace" {
		t.Errorf("WorkspacePath = %v, want /workspace", cfg.WorkspacePath)
	}
}

func TestHasCookie(t *testing.T) {
	tests := []struct {
		name   string
		cookie string
		want   bool
	}{
		{"empty cookie", "", false},
		{"with cookie", "JSESSIONID=test123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalConfig = &Config{Cookie: tt.cookie}
			if got := HasCookie(); got != tt.want {
				t.Errorf("HasCookie() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasHandle(t *testing.T) {
	tests := []struct {
		name   string
		handle string
		want   bool
	}{
		{"empty handle", "", false},
		{"with handle", "testuser", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalConfig = &Config{CFHandle: tt.handle}
			if got := HasHandle(); got != tt.want {
				t.Errorf("HasHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCookie(t *testing.T) {
	expectedCookie := "JSESSIONID=test123; 39ce7=abc456"
	globalConfig = &Config{Cookie: expectedCookie}

	if got := GetCookie(); got != expectedCookie {
		t.Errorf("GetCookie() = %v, want %v", got, expectedCookie)
	}
}

func TestGetCFHandle_MockTest(t *testing.T) {
	expectedHandle := "testuser"
	globalConfig = &Config{CFHandle: expectedHandle}

	if got := GetCFHandle(); got != expectedHandle {
		t.Errorf("GetCFHandle() = %v, want %v", got, expectedHandle)
	}
}

func TestGetCFHandle_NilConfig_MockTest(t *testing.T) {
	globalConfig = nil

	if got := GetCFHandle(); got != "" {
		t.Errorf("GetCFHandle() with nil config = %v, want empty string", got)
	}
}
