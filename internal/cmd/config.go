package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/harshit-vibes/dsaprep/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `View and modify dsaprep configuration.

Configuration is stored in ~/.dsaprep/config.yaml`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long: `Get the value of a configuration key.

Available keys:
  cf_handle       - Your Codeforces handle
  difficulty.min  - Minimum problem rating (default: 800)
  difficulty.max  - Maximum problem rating (default: 1600)
  daily_goal      - Daily problem solving goal (default: 5)
  theme           - UI theme (dark/light)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Show all config
			cfg := config.Get()
			fmt.Printf("cf_handle:      %s\n", cfg.CFHandle)
			fmt.Printf("difficulty.min: %d\n", cfg.Difficulty.Min)
			fmt.Printf("difficulty.max: %d\n", cfg.Difficulty.Max)
			fmt.Printf("daily_goal:     %d\n", cfg.DailyGoal)
			fmt.Printf("theme:          %s\n", cfg.Theme)
			fmt.Printf("\nConfig file: %s\n", config.GetConfigPath())
			return
		}

		key := args[0]
		cfg := config.Get()

		switch key {
		case "cf_handle":
			fmt.Println(cfg.CFHandle)
		case "difficulty.min":
			fmt.Println(cfg.Difficulty.Min)
		case "difficulty.max":
			fmt.Println(cfg.Difficulty.Max)
		case "daily_goal":
			fmt.Println(cfg.DailyGoal)
		case "theme":
			fmt.Println(cfg.Theme)
		default:
			fmt.Printf("Unknown config key: %s\n", key)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  cf_handle       - Your Codeforces handle
  difficulty.min  - Minimum problem rating
  difficulty.max  - Maximum problem rating
  daily_goal      - Daily problem solving goal
  theme           - UI theme (dark/light)

Examples:
  dsaprep config set cf_handle tourist
  dsaprep config set difficulty.min 1200
  dsaprep config set daily_goal 10`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		switch key {
		case "cf_handle":
			config.Set("cf_handle", value)
		case "difficulty.min":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid value for difficulty.min: %s", value)
			}
			if v < 0 || v > 3500 {
				return fmt.Errorf("difficulty.min must be between 0 and 3500")
			}
			config.Set("difficulty.min", v)
		case "difficulty.max":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid value for difficulty.max: %s", value)
			}
			if v < 0 || v > 3500 {
				return fmt.Errorf("difficulty.max must be between 0 and 3500")
			}
			config.Set("difficulty.max", v)
		case "daily_goal":
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid value for daily_goal: %s", value)
			}
			if v < 1 || v > 100 {
				return fmt.Errorf("daily_goal must be between 1 and 100")
			}
			config.Set("daily_goal", v)
		case "theme":
			if value != "dark" && value != "light" {
				return fmt.Errorf("theme must be 'dark' or 'light'")
			}
			config.Set("theme", value)
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		if err := config.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	Run: func(cmd *cobra.Command, args []string) {
		path := config.GetConfigPath()
		if path == "" {
			dataDir, _ := config.DataDir()
			path = dataDir + "/config.yaml"
		}
		fmt.Println(path)
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configPathCmd)
}
