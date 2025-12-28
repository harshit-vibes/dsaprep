package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/harshit-vibes/dsaprep/internal/config"
	"github.com/harshit-vibes/dsaprep/internal/tui"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dsaprep",
	Short: "DSA Prep - A Codeforces practice companion",
	Long: `DSA Prep is a beautiful terminal UI application for practicing
competitive programming problems from Codeforces.

Features:
  • Browse 8000+ problems with filtering and search
  • Track your practice sessions with built-in timer
  • View your Codeforces statistics and rating history
  • Get random problems matching your skill level

Run without arguments to launch the interactive TUI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dsaprep/config.yaml)")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(problemCmd)
	rootCmd.AddCommand(randomCmd)
	rootCmd.AddCommand(statsCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := config.Init(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
	}
}
