package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/harshit-vibes/dsaprep/internal/codeforces"
	"github.com/harshit-vibes/dsaprep/internal/config"
)

var statsCmd = &cobra.Command{
	Use:   "stats [handle]",
	Short: "View Codeforces statistics",
	Long: `View Codeforces statistics for a user.

If no handle is provided, uses the configured handle.

Examples:
  dsaprep stats           # Use configured handle
  dsaprep stats tourist   # View tourist's stats
  dsaprep stats --rating  # Show rating history`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var handle string
		if len(args) > 0 {
			handle = args[0]
		} else {
			handle = config.GetCFHandle()
		}

		if handle == "" {
			return fmt.Errorf("no handle provided. Use 'dsaprep config set cf_handle <handle>' or provide one as argument")
		}

		showRating, _ := cmd.Flags().GetBool("rating")
		showSubmissions, _ := cmd.Flags().GetBool("submissions")

		client := codeforces.NewClient()
		ctx := context.Background()

		// Get user info
		users, err := client.GetUserInfo(ctx, handle)
		if err != nil {
			return fmt.Errorf("failed to fetch user info: %w", err)
		}

		if len(users) == 0 {
			return fmt.Errorf("user not found: %s", handle)
		}

		user := users[0]

		// Display user info
		fmt.Printf("📊 Codeforces Stats for @%s\n\n", user.Handle)

		// Basic info
		fmt.Printf("Rank:       %s\n", strings.ToUpper(user.Rank))
		fmt.Printf("Rating:     %d\n", user.Rating)
		fmt.Printf("Max Rating: %d (%s)\n", user.MaxRating, user.MaxRank)

		if user.Country != "" {
			loc := user.Country
			if user.City != "" {
				loc = user.City + ", " + user.Country
			}
			fmt.Printf("Location:   %s\n", loc)
		}

		if user.Organization != "" {
			fmt.Printf("Org:        %s\n", user.Organization)
		}

		fmt.Printf("Friends:    %d\n", user.FriendOfCount)

		// Rating history
		if showRating {
			fmt.Printf("\n📈 Rating History (Last 10 contests)\n")
			fmt.Println(strings.Repeat("-", 60))

			ratingHistory, err := client.GetUserRating(ctx, handle)
			if err != nil {
				return fmt.Errorf("failed to fetch rating history: %w", err)
			}

			if len(ratingHistory) == 0 {
				fmt.Println("No contest history")
			} else {
				start := len(ratingHistory) - 10
				if start < 0 {
					start = 0
				}

				for i := len(ratingHistory) - 1; i >= start; i-- {
					rc := ratingHistory[i]
					delta := rc.NewRating - rc.OldRating
					sign := "+"
					if delta < 0 {
						sign = ""
					}

					contestName := rc.ContestName
					if len(contestName) > 35 {
						contestName = contestName[:32] + "..."
					}

					fmt.Printf("%-35s  Rank: %-5d  %d → %d (%s%d)\n",
						contestName, rc.Rank, rc.OldRating, rc.NewRating, sign, delta)
				}
			}
		}

		// Recent submissions
		if showSubmissions {
			fmt.Printf("\n📝 Recent Submissions (Last 10)\n")
			fmt.Println(strings.Repeat("-", 60))

			submissions, err := client.GetUserStatus(ctx, handle, 10)
			if err != nil {
				return fmt.Errorf("failed to fetch submissions: %w", err)
			}

			if len(submissions) == 0 {
				fmt.Println("No submissions")
			} else {
				for _, sub := range submissions {
					problemName := sub.Problem.Name
					if len(problemName) > 30 {
						problemName = problemName[:27] + "..."
					}

					timeAgo := formatTimeAgo(time.Unix(sub.CreationTimeSeconds, 0))

					verdict := sub.Verdict
					if verdict == "" {
						verdict = "TESTING"
					}

					fmt.Printf("%-10s  %-30s  %-15s  %s\n",
						sub.Problem.ID(), problemName, verdict, timeAgo)
				}
			}
		}

		if !showRating && !showSubmissions {
			fmt.Printf("\nUse --rating to see rating history\n")
			fmt.Printf("Use --submissions to see recent submissions\n")
		}

		return nil
	},
}

func init() {
	statsCmd.Flags().BoolP("rating", "r", false, "Show rating history")
	statsCmd.Flags().BoolP("submissions", "s", false, "Show recent submissions")
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	}
	if diff < 24*time.Hour {
		h := int(diff.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}
	if diff < 30*24*time.Hour {
		d := int(diff.Hours() / 24)
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	}
	if diff < 365*24*time.Hour {
		m := int(diff.Hours() / (24 * 30))
		if m == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", m)
	}
	y := int(diff.Hours() / (24 * 365))
	if y == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", y)
}
