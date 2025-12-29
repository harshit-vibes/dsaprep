// Package config manages application configuration
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Codeforces settings
	CFHandle string `mapstructure:"cf_handle"`

	// Practice settings
	Difficulty DifficultyRange `mapstructure:"difficulty"`
	DailyGoal  int             `mapstructure:"daily_goal"`

	// Paths
	WorkspacePath string `mapstructure:"workspace_path"`
}

// DifficultyRange represents min/max difficulty
type DifficultyRange struct {
	Min int `mapstructure:"min"`
	Max int `mapstructure:"max"`
}

var (
	globalConfig *Config
	configMu     sync.RWMutex
)

// configDir returns the config directory path
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cf"), nil
}

// configFilePath returns the config file path
func configFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// MigrateFromLegacy migrates config from ~/.dsaprep to ~/.cf
// Returns true if migration was performed
func MigrateFromLegacy() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	oldDir := filepath.Join(home, ".dsaprep")
	newDir := filepath.Join(home, ".cf")
	oldEnv := filepath.Join(home, ".dsaprep.env")
	newEnv := filepath.Join(home, ".cf.env")

	migrated := false

	// Check if old config directory exists and new one doesn't
	if _, err := os.Stat(oldDir); err == nil {
		if _, err := os.Stat(newDir); os.IsNotExist(err) {
			// Copy old directory to new location
			if err := copyDir(oldDir, newDir); err != nil {
				return false, fmt.Errorf("failed to migrate config directory: %w", err)
			}
			migrated = true
			fmt.Printf("Migrated config: %s -> %s\n", oldDir, newDir)
		}
	}

	// Check if old env file exists and new one doesn't
	if _, err := os.Stat(oldEnv); err == nil {
		if _, err := os.Stat(newEnv); os.IsNotExist(err) {
			// Copy old env file to new location
			if err := copyFile(oldEnv, newEnv); err != nil {
				return false, fmt.Errorf("failed to migrate env file: %w", err)
			}
			migrated = true
			fmt.Printf("Migrated credentials: %s -> %s\n", oldEnv, newEnv)
		}
	}

	return migrated, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

// copyDir copies a directory from src to dst
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

// Init initializes the configuration
func Init(workspacePath string) error {
	configMu.Lock()
	defer configMu.Unlock()

	// Try to migrate from legacy config
	if _, err := MigrateFromLegacy(); err != nil {
		// Log but don't fail on migration errors
		fmt.Printf("Warning: failed to migrate legacy config: %v\n", err)
	}

	dir, err := configDir()
	if err != nil {
		return fmt.Errorf("failed to get config dir: %w", err)
	}

	// Create config directory if needed
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	// Setup viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	// Set defaults
	viper.SetDefault("cf_handle", "")
	viper.SetDefault("difficulty.min", 800)
	viper.SetDefault("difficulty.max", 1400)
	viper.SetDefault("daily_goal", 3)
	viper.SetDefault("workspace_path", "")

	// Try to read existing config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Create default config
			if err := viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		} else {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Override workspace path if provided
	if workspacePath != "" {
		viper.Set("workspace_path", workspacePath)
	}

	// Unmarshal config
	globalConfig = &Config{}
	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Get returns the global configuration
func Get() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

// GetCFHandle returns the configured CF handle
func GetCFHandle() string {
	cfg := Get()
	if cfg == nil {
		return ""
	}
	return cfg.CFHandle
}

// Set updates a configuration value
func Set(key string, value interface{}) error {
	configMu.Lock()
	defer configMu.Unlock()

	viper.Set(key, value)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Re-unmarshal
	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	return nil
}

// SetCFHandle sets the CF handle
func SetCFHandle(handle string) error {
	return Set("cf_handle", handle)
}

// SetDifficulty sets the difficulty range
func SetDifficulty(min, max int) error {
	if err := Set("difficulty.min", min); err != nil {
		return err
	}
	return Set("difficulty.max", max)
}

// SetDailyGoal sets the daily goal
func SetDailyGoal(goal int) error {
	return Set("daily_goal", goal)
}

// SetWorkspacePath sets the workspace path
func SetWorkspacePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	return Set("workspace_path", absPath)
}

// GetWorkspacePath returns the workspace path
func GetWorkspacePath() string {
	cfg := Get()
	if cfg == nil || cfg.WorkspacePath == "" {
		// Default to current directory
		cwd, _ := os.Getwd()
		return cwd
	}
	return cfg.WorkspacePath
}
