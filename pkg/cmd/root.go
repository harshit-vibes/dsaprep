// Package cmd provides CLI commands for cf
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/external/cfweb"
	exthealth "github.com/harshit-vibes/cf/pkg/external/health"
	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/health"
	"github.com/harshit-vibes/cf/pkg/internal/workspace"
	"github.com/harshit-vibes/cf/pkg/tui"
)

var (
	// Version information
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"

	// Command line flags
	skipChecks bool
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "cf",
	Short: "Codeforces CLI - Your competitive programming companion",
	Long: `cf is a command-line tool for competitive programming with Codeforces.

Features:
  • Problem parsing and workspace management
  • Solution submission and verdict tracking
  • Practice progress tracking
  • Beautiful TUI dashboard

Run 'cf' without arguments to launch the interactive TUI.`,
	PersistentPreRunE: runPreChecks,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Launch TUI
		return tui.Run()
	},
}

func runPreChecks(cmd *cobra.Command, args []string) error {
	// Skip health checks for version and help commands
	if cmd.Name() == "version" || cmd.Name() == "help" {
		return nil
	}
	return runStartupChecks()
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Initialize configuration
	cobra.OnInitialize(initConfig)

	// Add flags
	rootCmd.PersistentFlags().BoolVar(&skipChecks, "skip-checks", false, "Skip startup health checks")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")

	// Core commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(tuiCmd)

	// Feature commands
	rootCmd.AddCommand(problemCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(contestCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(configCmd)

	// Legacy parse command (deprecated, redirects to problem parse)
	rootCmd.AddCommand(parseCmd)
}

// tuiCmd launches the TUI explicitly
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI",
	Long:  `Launch the interactive terminal user interface for browsing problems, viewing stats, and more.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

func initConfig() {
	// Load configuration
	if err := config.Init(""); err != nil {
		// Config init failure is handled by health checks
		return
	}
}

func runStartupChecks() error {
	if skipChecks {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	checker := health.NewChecker()

	// Get workspace path
	cfg := config.Get()
	wsPath := "."
	if cfg != nil && cfg.WorkspacePath != "" {
		wsPath = cfg.WorkspacePath
	}
	ws := workspace.New(wsPath)

	// Internal checks
	checker.AddCheck(&health.ConfigCheck{})
	checker.AddCheck(&health.CookieCheck{})
	checker.AddCheck(health.NewWorkspaceCheck(ws))
	checker.AddCheck(health.NewSchemaVersionCheck(ws))

	// External checks
	apiClient := cfapi.NewClient()
	parser := cfweb.NewParserWithClient(nil)

	checker.AddCheck(exthealth.NewCFAPICheck(apiClient))
	checker.AddCheck(exthealth.NewCFWebCheck(parser))
	checker.AddCheck(exthealth.NewCFHandleCheck(apiClient))

	// Run checks
	report := checker.Run(ctx)

	// Display results
	if verbose || report.OverallStatus != health.StatusHealthy {
		displayHealthReport(report)
	}

	if !report.CanProceed {
		fmt.Println("\n❌ Cannot proceed due to critical errors. Please fix the issues above.")
		return fmt.Errorf("startup checks failed")
	}

	if report.OverallStatus == health.StatusDegraded {
		fmt.Println("\n⚠️  Some features may be unavailable. See warnings above.")
	}

	return nil
}

func displayHealthReport(report *health.Report) {
	fmt.Printf("\n🔍 Health Check Report (took %s)\n", report.Duration.Round(time.Millisecond))
	fmt.Println("─────────────────────────────────")

	for _, result := range report.Results {
		var icon string
		switch result.Status {
		case health.StatusHealthy:
			icon = "✓"
		case health.StatusDegraded:
			icon = "⚠"
		case health.StatusCritical:
			icon = "✗"
		}

		fmt.Printf("%s %-20s %s\n", icon, result.Name, result.Message)
		if result.Details != "" && result.Status != health.StatusHealthy {
			fmt.Printf("  └─ %s\n", result.Details)
		}
	}

	fmt.Println("─────────────────────────────────")
	fmt.Printf("Status: %s | Schema: %s\n", report.OverallStatus, report.CurrentSchemaVersion)
}

// versionCmd shows version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cf %s\n", Version)
		fmt.Printf("  Commit:     %s\n", Commit)
		fmt.Printf("  Build Date: %s\n", BuildDate)
	},
}

// initCmd initializes the workspace
var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new cf workspace",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		ws := workspace.New(path)

		if ws.Exists() {
			return fmt.Errorf("workspace already exists at %s", path)
		}

		handle := config.GetCFHandle()

		if err := ws.Init("DSA Practice", handle); err != nil {
			return fmt.Errorf("failed to initialize workspace: %w", err)
		}

		fmt.Printf("✓ Initialized workspace at %s\n", path)
		return nil
	},
}

// healthCmd shows health status
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check system health",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip the regular pre-run checks for health command
		// since we'll run them ourselves with verbose=true
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Force verbose output
		verbose = true
		return runStartupChecks()
	},
}

// parseCmd is deprecated, kept for backward compatibility
var parseCmd = &cobra.Command{
	Use:        "parse <contest_id> <problem_index>",
	Short:      "[Deprecated] Use 'cf problem parse' instead",
	Args:       cobra.ExactArgs(2),
	Deprecated: "use 'cf problem parse' instead",
	Hidden:     true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("⚠️  'cf parse' is deprecated. Use 'cf problem parse' instead.")
		fmt.Println()
		return runProblemParse(cmd, args)
	},
}
