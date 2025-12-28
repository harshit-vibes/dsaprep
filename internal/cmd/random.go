package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
	"github.com/harshit-vibes/dsaprep/internal/config"
)

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Get a random problem",
	Long: `Get a random Codeforces problem based on your difficulty settings.

By default, uses the difficulty range from your configuration.
You can override with --min and --max flags.

Examples:
  dsaprep random               # Use config difficulty range
  dsaprep random --min 1200    # Problems rated 1200+
  dsaprep random --max 1600    # Problems rated up to 1600
  dsaprep random --tag dp      # Only dynamic programming problems
  dsaprep random --open        # Open directly in browser`,
	RunE: func(cmd *cobra.Command, args []string) error {
		minRating, _ := cmd.Flags().GetInt("min")
		maxRating, _ := cmd.Flags().GetInt("max")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		openInBrowser, _ := cmd.Flags().GetBool("open")

		// Use config defaults if not specified
		cfg := config.Get()
		if minRating == 0 {
			minRating = cfg.Difficulty.Min
		}
		if maxRating == 0 {
			maxRating = cfg.Difficulty.Max
		}

		// Fetch problems
		client := codeforces.NewClient()
		ctx := context.Background()

		var result *codeforces.ProblemsResult
		var err error

		if len(tags) > 0 {
			result, err = client.GetProblemsWithTags(ctx, tags)
		} else {
			result, err = client.GetProblems(ctx)
		}

		if err != nil {
			return fmt.Errorf("failed to fetch problems: %w", err)
		}

		// Filter problems by rating
		var eligible []codeforces.Problem
		for _, p := range result.Problems {
			if p.Rating > 0 && p.Rating >= minRating && p.Rating <= maxRating {
				eligible = append(eligible, p)
			}
		}

		if len(eligible) == 0 {
			return fmt.Errorf("no problems found matching criteria (rating %d-%d)", minRating, maxRating)
		}

		// Pick a random problem
		rand.Seed(time.Now().UnixNano())
		problem := eligible[rand.Intn(len(eligible))]

		// Open in browser if requested
		if openInBrowser {
			return openURL(problem.URL())
		}

		// Display problem
		fmt.Printf("🎲 Random Problem\n\n")
		fmt.Printf("Name:    %s\n", problem.Name)
		fmt.Printf("ID:      %s\n", problem.ID())
		fmt.Printf("Rating:  %d (%s)\n", problem.Rating, problem.RankName())

		if len(problem.Tags) > 0 {
			fmt.Printf("Tags:    %s\n", strings.Join(problem.Tags, ", "))
		}

		fmt.Printf("\nURL: %s\n", problem.URL())
		fmt.Printf("\nUse --open flag to open in browser\n")
		fmt.Printf("Run 'dsaprep random' again for another problem\n")

		return nil
	},
}

func init() {
	randomCmd.Flags().Int("min", 0, "Minimum rating (default: from config)")
	randomCmd.Flags().Int("max", 0, "Maximum rating (default: from config)")
	randomCmd.Flags().StringSlice("tag", nil, "Filter by tags (e.g., --tag dp --tag graphs)")
	randomCmd.Flags().BoolP("open", "o", false, "Open problem in browser")
}
