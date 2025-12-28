package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
)

var problemCmd = &cobra.Command{
	Use:   "problem <id>",
	Short: "View or open a problem",
	Long: `View details of a Codeforces problem or open it in browser.

Problem ID format: <contest_id><problem_index>
Examples: 1A, 1234B, 500C

Examples:
  dsaprep problem 1A          # Show problem details
  dsaprep problem 1234B --open # Open in browser`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		problemID := strings.ToUpper(args[0])
		openInBrowser, _ := cmd.Flags().GetBool("open")

		// Parse problem ID
		var contestID int
		var index string
		for i, c := range problemID {
			if c >= 'A' && c <= 'Z' {
				fmt.Sscanf(problemID[:i], "%d", &contestID)
				index = problemID[i:]
				break
			}
		}

		if contestID == 0 || index == "" {
			return fmt.Errorf("invalid problem ID format: %s (expected format: 1234A)", problemID)
		}

		// Get problem from API
		client := codeforces.NewClient()
		ctx := context.Background()

		result, err := client.GetProblems(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch problems: %w", err)
		}

		// Find the problem
		var problem *codeforces.Problem
		for _, p := range result.Problems {
			if p.ContestID == contestID && p.Index == index {
				problem = &p
				break
			}
		}

		if problem == nil {
			return fmt.Errorf("problem not found: %s", problemID)
		}

		// Open in browser if requested
		if openInBrowser {
			url := problem.URL()
			return openURL(url)
		}

		// Display problem info
		fmt.Printf("Problem: %s\n", problem.Name)
		fmt.Printf("ID:      %s\n", problem.ID())
		fmt.Printf("Contest: %d\n", problem.ContestID)
		fmt.Printf("Index:   %s\n", problem.Index)

		if problem.Rating > 0 {
			fmt.Printf("Rating:  %d (%s)\n", problem.Rating, problem.RankName())
		} else {
			fmt.Printf("Rating:  Unrated\n")
		}

		if len(problem.Tags) > 0 {
			fmt.Printf("Tags:    %s\n", strings.Join(problem.Tags, ", "))
		}

		fmt.Printf("\nURL: %s\n", problem.URL())
		fmt.Printf("\nUse --open flag to open in browser\n")

		return nil
	},
}

func init() {
	problemCmd.Flags().BoolP("open", "o", false, "Open problem in browser")
}

// openURL opens the specified URL in the default browser
func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	fmt.Printf("Opening %s in browser...\n", url)
	return cmd.Start()
}
