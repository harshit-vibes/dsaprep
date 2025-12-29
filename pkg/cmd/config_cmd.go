package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harshit-vibes/cf/pkg/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `View and modify cf configuration settings.

Configuration is stored in ~/.cf/config.yaml`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value(s)",
	Long: `Display configuration values.

If no key is provided, shows all configuration.

Available keys:
  cf_handle       - Your Codeforces handle
  difficulty.min  - Minimum problem difficulty
  difficulty.max  - Maximum problem difficulty
  daily_goal      - Daily problem solving goal
  workspace_path  - Path to workspace directory

Examples:
  cf config get              # Show all config
  cf config get cf_handle    # Show CF handle`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long: `Set a configuration value.

Available keys:
  cf_handle       - Your Codeforces handle
  difficulty.min  - Minimum problem difficulty (e.g., 800)
  difficulty.max  - Maximum problem difficulty (e.g., 1400)
  daily_goal      - Daily problem solving goal (e.g., 3)
  workspace_path  - Path to workspace directory

Examples:
  cf config set cf_handle tourist
  cf config set difficulty.min 1000
  cf config set daily_goal 5`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file paths",
	Long:  `Display the paths to configuration files.`,
	RunE:  runConfigPath,
}

func init() {
	// Add config subcommands
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configPathCmd)
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if len(args) == 0 {
		// Show all config
		fmt.Println("\n📋 Configuration:")
		fmt.Println(strings.Repeat("─", 40))
		fmt.Printf("  cf_handle:       %s\n", valueOrEmpty(cfg.CFHandle))
		fmt.Printf("  difficulty.min:  %d\n", cfg.Difficulty.Min)
		fmt.Printf("  difficulty.max:  %d\n", cfg.Difficulty.Max)
		fmt.Printf("  daily_goal:      %d\n", cfg.DailyGoal)
		fmt.Printf("  workspace_path:  %s\n", valueOrEmpty(cfg.WorkspacePath))
		fmt.Println()

		// Also show credentials status
		creds, _ := config.LoadCredentials()
		if creds != nil {
			fmt.Println("🔑 Credentials (~/.cf.env):")
			fmt.Println(strings.Repeat("─", 40))
			fmt.Printf("  CF_HANDLE:       %s\n", valueOrEmpty(creds.CFHandle))
			fmt.Printf("  CF_API_KEY:      %s\n", maskValue(creds.APIKey))
			fmt.Printf("  CF_API_SECRET:   %s\n", maskValue(creds.APISecret))
			fmt.Printf("  CF_JSESSIONID:   %s\n", maskValue(creds.JSESSIONID))
			fmt.Printf("  CF_CLEARANCE:    %s\n", creds.GetCFClearanceStatus())
			fmt.Println()
		}

		return nil
	}

	// Show specific key
	key := strings.ToLower(args[0])
	switch key {
	case "cf_handle":
		fmt.Println(valueOrEmpty(cfg.CFHandle))
	case "difficulty.min":
		fmt.Println(cfg.Difficulty.Min)
	case "difficulty.max":
		fmt.Println(cfg.Difficulty.Max)
	case "daily_goal":
		fmt.Println(cfg.DailyGoal)
	case "workspace_path":
		fmt.Println(valueOrEmpty(cfg.WorkspacePath))
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := strings.ToLower(args[0])
	value := args[1]

	var err error
	switch key {
	case "cf_handle":
		err = config.SetCFHandle(value)
	case "difficulty.min":
		var min int
		if _, e := fmt.Sscanf(value, "%d", &min); e != nil {
			return fmt.Errorf("invalid value for difficulty.min: %s", value)
		}
		cfg := config.Get()
		err = config.SetDifficulty(min, cfg.Difficulty.Max)
	case "difficulty.max":
		var max int
		if _, e := fmt.Sscanf(value, "%d", &max); e != nil {
			return fmt.Errorf("invalid value for difficulty.max: %s", value)
		}
		cfg := config.Get()
		err = config.SetDifficulty(cfg.Difficulty.Min, max)
	case "daily_goal":
		var goal int
		if _, e := fmt.Sscanf(value, "%d", &goal); e != nil {
			return fmt.Errorf("invalid value for daily_goal: %s", value)
		}
		err = config.SetDailyGoal(goal)
	case "workspace_path":
		err = config.SetWorkspacePath(value)
	default:
		return fmt.Errorf("unknown config key: %s\n\nAvailable keys: cf_handle, difficulty.min, difficulty.max, daily_goal, workspace_path", key)
	}

	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	fmt.Printf("✓ Set %s = %s\n", key, value)
	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	envPath, err := config.GetEnvFilePath()
	if err != nil {
		return err
	}

	fmt.Println("\n📁 Configuration Files:")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("  Config:      ~/.cf/config.yaml\n")
	fmt.Printf("  Credentials: %s\n", envPath)
	fmt.Println()

	return nil
}

func valueOrEmpty(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

func maskValue(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
