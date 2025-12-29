package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/internal/config"
)

var (
	// contest list flags
	contestShowGym    bool
	contestLimit      int
	contestPhase      string
)

var contestCmd = &cobra.Command{
	Use:     "contest",
	Aliases: []string{"c", "contests"},
	Short:   "Browse Codeforces contests",
	Long: `Commands for viewing Codeforces contest information.

List upcoming and past contests, view contest problems.`,
}

var contestListCmd = &cobra.Command{
	Use:   "list",
	Short: "List contests",
	Long: `List Codeforces contests.

Shows upcoming and recent contests by default.

Examples:
  cf contest list              # List recent contests
  cf contest list --gym        # List gym contests
  cf contest list --limit 50   # Show more contests`,
	RunE: runContestList,
}

var contestProblemsCmd = &cobra.Command{
	Use:   "problems <contest_id>",
	Short: "Show contest problems",
	Long: `Display problems from a specific contest.

Examples:
  cf contest problems 1234    # Show problems from contest 1234`,
	Args: cobra.ExactArgs(1),
	RunE: runContestProblems,
}

func init() {
	// Add contest subcommands
	contestCmd.AddCommand(contestListCmd)
	contestCmd.AddCommand(contestProblemsCmd)

	// contest list flags
	contestListCmd.Flags().BoolVar(&contestShowGym, "gym", false, "Show gym contests instead of regular contests")
	contestListCmd.Flags().IntVar(&contestLimit, "limit", 20, "Maximum number of contests to display")
	contestListCmd.Flags().StringVar(&contestPhase, "phase", "", "Filter by phase (BEFORE, CODING, FINISHED)")
}

func runContestList(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := getAPIClient()
	contests, err := client.GetContests(ctx, contestShowGym)
	if err != nil {
		return fmt.Errorf("failed to fetch contests: %w", err)
	}

	// Filter by phase if specified
	if contestPhase != "" {
		phase := strings.ToUpper(contestPhase)
		var filtered []cfapi.Contest
		for _, c := range contests {
			if c.Phase == phase {
				filtered = append(filtered, c)
			}
		}
		contests = filtered
	}

	// Limit results
	if contestLimit > 0 && len(contests) > contestLimit {
		contests = contests[:contestLimit]
	}

	if len(contests) == 0 {
		fmt.Println("No contests found.")
		return nil
	}

	contestType := "Contests"
	if contestShowGym {
		contestType = "Gym Contests"
	}

	fmt.Printf("\n%s:\n\n", contestType)
	fmt.Printf("%-8s %-50s %-12s %s\n", "ID", "Name", "Phase", "Start Time")
	fmt.Println(strings.Repeat("─", 100))

	for _, c := range contests {
		name := c.Name
		if len(name) > 48 {
			name = name[:45] + "..."
		}

		phaseColor := getPhaseColor(c.Phase)
		startTime := "-"
		if c.StartTimeSeconds > 0 {
			startTime = c.StartTime().Format("Jan 02, 2006 15:04")
		}

		fmt.Printf("%-8d %-50s %s%-12s\033[0m %s\n",
			c.ID,
			name,
			phaseColor,
			c.Phase,
			startTime,
		)
	}

	fmt.Println()
	return nil
}

func runContestProblems(cmd *cobra.Command, args []string) error {
	var contestID int
	if _, err := fmt.Sscanf(args[0], "%d", &contestID); err != nil {
		return fmt.Errorf("invalid contest ID: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	creds, _ := config.LoadCredentials()
	var client *cfapi.Client
	if creds != nil && creds.IsAPIConfigured() {
		client = cfapi.NewClient(cfapi.WithAPICredentials(creds.APIKey, creds.APISecret))
	} else {
		client = cfapi.NewClient()
	}

	standings, err := client.GetContestStandings(ctx, contestID, 1, 1, nil, false)
	if err != nil {
		return fmt.Errorf("failed to get contest: %w", err)
	}

	contest := standings.Contest
	problems := standings.Problems

	fmt.Printf("\n%s\n", contest.Name)
	fmt.Printf("Contest #%d | %s | Duration: %s\n",
		contest.ID,
		contest.Phase,
		formatDuration(contest.Duration()),
	)
	fmt.Println(strings.Repeat("─", 80))

	if len(problems) == 0 {
		fmt.Println("No problems available.")
		return nil
	}

	fmt.Printf("\n%-6s %-50s %8s  %s\n", "Index", "Name", "Rating", "Tags")
	fmt.Println(strings.Repeat("─", 80))

	for _, p := range problems {
		name := p.Name
		if len(name) > 48 {
			name = name[:45] + "..."
		}

		ratingStr := "-"
		if p.Rating > 0 {
			ratingStr = fmt.Sprintf("%d", p.Rating)
		}

		tags := strings.Join(p.Tags, ", ")
		if len(tags) > 20 {
			tags = tags[:17] + "..."
		}

		fmt.Printf("%-6s %-50s %8s  %s\n",
			p.Index,
			name,
			ratingStr,
			tags,
		)
	}

	fmt.Printf("\nContest URL: https://codeforces.com/contest/%d\n\n", contestID)

	return nil
}

// getPhaseColor returns ANSI color code for contest phase
func getPhaseColor(phase string) string {
	switch phase {
	case cfapi.PhaseCoding:
		return "\033[32m" // green - active
	case cfapi.PhaseBefore:
		return "\033[33m" // yellow - upcoming
	case cfapi.PhaseFinished:
		return "\033[90m" // gray - finished
	case cfapi.PhasePendingTest, cfapi.PhaseSystemTest:
		return "\033[36m" // cyan - testing
	default:
		return ""
	}
}

// formatDuration formats a duration as hours and minutes
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", minutes)
}
