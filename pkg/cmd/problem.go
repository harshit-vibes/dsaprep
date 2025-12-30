package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/harshit-vibes/cf/pkg/external/cfweb"
	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/workspace"
)

var (
	// problem list flags
	problemTags      []string
	problemMinRating int
	problemMaxRating int
	problemLimit     int
	excludeSolved    bool
)

var problemCmd = &cobra.Command{
	Use:     "problem",
	Aliases: []string{"p", "prob"},
	Short:   "Manage Codeforces problems",
	Long: `Commands for working with Codeforces problems.

Parse problems from contests, list available problems, and fetch them to your workspace.`,
}

var problemParseCmd = &cobra.Command{
	Use:   "parse <contest_id> <problem_index>",
	Short: "Parse a problem from Codeforces",
	Long: `Parse a problem from Codeforces and display its details.

The problem will be saved to your workspace if one is configured.

Examples:
  cf problem parse 1 A       # Parse problem A from contest 1
  cf problem parse 1234 B    # Parse problem B from contest 1234`,
	Args: cobra.ExactArgs(2),
	RunE: runProblemParse,
}

var problemListCmd = &cobra.Command{
	Use:   "list",
	Short: "List problems from Codeforces",
	Long: `List problems from the Codeforces problemset.

Filter by tags, rating range, and exclude already solved problems.

Examples:
  cf problem list                          # List all problems
  cf problem list --tag dp --tag graphs    # Filter by tags
  cf problem list --rating 800-1200        # Filter by rating range
  cf problem list --limit 20               # Limit results`,
	RunE: runProblemList,
}

var problemFetchCmd = &cobra.Command{
	Use:   "fetch <contest_id> [problem_index]",
	Short: "Fetch problem(s) to workspace",
	Long: `Fetch a problem or all problems from a contest to your workspace.

If problem_index is provided, fetches only that problem.
Otherwise, fetches all problems from the contest.

Examples:
  cf problem fetch 1 A      # Fetch problem A from contest 1
  cf problem fetch 1234     # Fetch all problems from contest 1234`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runProblemFetch,
}

func init() {
	// Add problem subcommands
	problemCmd.AddCommand(problemParseCmd)
	problemCmd.AddCommand(problemListCmd)
	problemCmd.AddCommand(problemFetchCmd)

	// problem list flags
	problemListCmd.Flags().StringArrayVar(&problemTags, "tag", nil, "Filter by tag (can be specified multiple times)")
	problemListCmd.Flags().IntVar(&problemMinRating, "min-rating", 0, "Minimum problem rating")
	problemListCmd.Flags().IntVar(&problemMaxRating, "max-rating", 0, "Maximum problem rating")
	problemListCmd.Flags().IntVar(&problemLimit, "limit", 25, "Maximum number of problems to display")
	problemListCmd.Flags().BoolVar(&excludeSolved, "unsolved", false, "Exclude already solved problems")
}

func runProblemParse(cmd *cobra.Command, args []string) error {
	var contestID int
	if _, err := fmt.Sscanf(args[0], "%d", &contestID); err != nil {
		return fmt.Errorf("invalid contest ID: %s", args[0])
	}
	problemIndex := strings.ToUpper(args[1])

	parser := cfweb.NewParserWithClient(nil)
	problem, err := parser.ParseProblem(contestID, problemIndex)
	if err != nil {
		return fmt.Errorf("failed to parse problem: %w", err)
	}

	fmt.Printf("✓ Parsed: %s. %s\n", problem.Index, problem.Name)
	fmt.Printf("  Rating: %d | Time: %s | Memory: %s\n",
		problem.Rating, problem.TimeLimit, problem.MemoryLimit)
	fmt.Printf("  Tags: %v\n", problem.Tags)
	fmt.Printf("  Samples: %d\n", len(problem.Samples))

	// Save to workspace if available
	cfg := config.Get()
	if cfg != nil && cfg.WorkspacePath != "" {
		ws := workspace.New(cfg.WorkspacePath)
		if ws.Exists() {
			schemaProblem := problem.ToSchemaProblem()
			if err := ws.SaveProblem(schemaProblem); err != nil {
				return fmt.Errorf("failed to save problem: %w", err)
			}
			fmt.Printf("✓ Saved to workspace\n")
		}
	}

	return nil
}

func runProblemList(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := getAPIClient()

	// Get handle for exclude-solved
	handle := ""
	if excludeSolved {
		handle = config.GetCFHandle()
		if handle == "" {
			return fmt.Errorf("--unsolved requires CF handle to be configured")
		}
	}

	// Filter problems
	problems, err := client.FilterProblems(ctx, problemMinRating, problemMaxRating, problemTags, excludeSolved, handle)
	if err != nil {
		return fmt.Errorf("failed to fetch problems: %w", err)
	}

	if len(problems) == 0 {
		fmt.Println("No problems found matching the criteria.")
		return nil
	}

	// Limit results
	if problemLimit > 0 && len(problems) > problemLimit {
		problems = problems[:problemLimit]
	}

	// Display problems
	fmt.Printf("Found %d problems:\n\n", len(problems))
	fmt.Printf("%-10s %-50s %6s  %s\n", "ID", "Name", "Rating", "Tags")
	fmt.Println(strings.Repeat("─", 100))

	for _, p := range problems {
		name := p.Name
		if len(name) > 48 {
			name = name[:45] + "..."
		}
		tags := strings.Join(p.Tags, ", ")
		if len(tags) > 30 {
			tags = tags[:27] + "..."
		}

		ratingStr := "-"
		if p.Rating > 0 {
			ratingStr = fmt.Sprintf("%d", p.Rating)
		}

		fmt.Printf("%-10s %-50s %6s  %s\n", p.ProblemID(), name, ratingStr, tags)
	}

	return nil
}

func runProblemFetch(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var contestID int
	if _, err := fmt.Sscanf(args[0], "%d", &contestID); err != nil {
		return fmt.Errorf("invalid contest ID: %s", args[0])
	}

	// Check workspace
	cfg := config.Get()
	if cfg == nil || cfg.WorkspacePath == "" {
		return fmt.Errorf("no workspace configured. Run 'cf init' first")
	}
	ws := workspace.New(cfg.WorkspacePath)
	if !ws.Exists() {
		return fmt.Errorf("workspace not found at %s. Run 'cf init' first", cfg.WorkspacePath)
	}

	parser := cfweb.NewParserWithClient(nil)

	if len(args) == 2 {
		// Fetch single problem
		problemIndex := strings.ToUpper(args[1])
		problem, err := parser.ParseProblem(contestID, problemIndex)
		if err != nil {
			return fmt.Errorf("failed to parse problem: %w", err)
		}

		schemaProblem := problem.ToSchemaProblem()
		if err := ws.SaveProblem(schemaProblem); err != nil {
			return fmt.Errorf("failed to save problem: %w", err)
		}

		fmt.Printf("✓ Fetched %s. %s to workspace\n", problem.Index, problem.Name)
	} else {
		// Fetch all problems from contest
		client := getAPIClient()
		standings, err := client.GetContestStandings(ctx, contestID, 1, 1, nil, false)
		if err != nil {
			return fmt.Errorf("failed to get contest problems: %w", err)
		}

		fmt.Printf("Fetching %d problems from contest %d...\n", len(standings.Problems), contestID)

		for _, p := range standings.Problems {
			problem, err := parser.ParseProblem(contestID, p.Index)
			if err != nil {
				fmt.Printf("  ✗ Failed to fetch %s: %v\n", p.Index, err)
				continue
			}

			schemaProblem := problem.ToSchemaProblem()
			if err := ws.SaveProblem(schemaProblem); err != nil {
				fmt.Printf("  ✗ Failed to save %s: %v\n", p.Index, err)
				continue
			}

			fmt.Printf("  ✓ %s. %s\n", problem.Index, problem.Name)
		}

		fmt.Printf("✓ Fetched contest %d to workspace\n", contestID)
	}

	return nil
}
