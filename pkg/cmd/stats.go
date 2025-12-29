package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
)

var statsCmd = &cobra.Command{
	Use:   "stats [handle]",
	Short: "Show practice statistics",
	Long: `Display practice statistics and problem-solving progress.

Shows solved problems by rating, tags, and recent activity.

Examples:
  cf stats           # Your statistics
  cf stats tourist   # tourist's statistics`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStats,
}

func init() {
	// Stats is a top-level command, no subcommands
}

func runStats(cmd *cobra.Command, args []string) error {
	handle, err := getHandle(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := getAPIClient()

	// Get user info
	users, err := client.GetUserInfo(ctx, []string{handle})
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	if len(users) == 0 {
		return fmt.Errorf("user %s not found", handle)
	}
	user := users[0]

	// Get submissions
	submissions, err := client.GetUserSubmissions(ctx, handle, 1, 10000)
	if err != nil {
		return fmt.Errorf("failed to get submissions: %w", err)
	}

	// Calculate stats
	stats := calculateStats(submissions)

	// Display
	fmt.Printf("\n📊 Statistics for %s\n", handle)
	fmt.Println(strings.Repeat("═", 60))

	// User summary
	rankColor := getRankColor(user.Rating)
	fmt.Printf("\n%s%s\033[0m (Rating: %s%d\033[0m)\n",
		rankColor, user.Rank, rankColor, user.Rating)

	// Overall stats
	fmt.Printf("\n📈 Overall:\n")
	fmt.Printf("   Total Solved:     %d unique problems\n", stats.TotalSolved)
	fmt.Printf("   Total Submissions: %d\n", stats.TotalSubmissions)
	fmt.Printf("   Acceptance Rate:  %.1f%%\n", stats.AcceptanceRate)

	// Problems by rating
	fmt.Printf("\n⭐ By Rating:\n")
	ratings := make([]int, 0, len(stats.ByRating))
	for r := range stats.ByRating {
		ratings = append(ratings, r)
	}
	sort.Ints(ratings)

	for _, r := range ratings {
		count := stats.ByRating[r]
		bar := strings.Repeat("█", min(count/2, 30))
		if r == 0 {
			fmt.Printf("   Unrated: %3d %s\n", count, bar)
		} else {
			fmt.Printf("   %4d:    %3d %s\n", r, count, bar)
		}
	}

	// Top tags
	fmt.Printf("\n🏷️  Top Tags:\n")
	type tagCount struct {
		tag   string
		count int
	}
	tags := make([]tagCount, 0, len(stats.ByTag))
	for t, c := range stats.ByTag {
		tags = append(tags, tagCount{t, c})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].count > tags[j].count
	})

	for i, tc := range tags {
		if i >= 10 {
			break
		}
		bar := strings.Repeat("█", min(tc.count/2, 20))
		fmt.Printf("   %-20s %3d %s\n", tc.tag, tc.count, bar)
	}

	// Recent activity
	fmt.Printf("\n📅 Recent Activity (last 30 days):\n")
	recentSolved := 0
	recentSubmissions := 0
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	for _, s := range submissions {
		if s.SubmissionTime().After(thirtyDaysAgo) {
			recentSubmissions++
			if s.IsAccepted() {
				recentSolved++
			}
		}
	}

	fmt.Printf("   Submissions: %d\n", recentSubmissions)
	fmt.Printf("   Solved:      %d unique problems\n", recentSolved)

	// Streak info
	if len(submissions) > 0 {
		streak := calculateStreak(submissions)
		if streak > 0 {
			fmt.Printf("   Current Streak: 🔥 %d days\n", streak)
		}
	}

	fmt.Println()
	return nil
}

type Stats struct {
	TotalSolved      int
	TotalSubmissions int
	AcceptanceRate   float64
	ByRating         map[int]int
	ByTag            map[string]int
}

func calculateStats(submissions []cfapi.Submission) Stats {
	stats := Stats{
		ByRating: make(map[int]int),
		ByTag:    make(map[string]int),
	}

	seen := make(map[string]bool)
	accepted := 0

	for _, s := range submissions {
		stats.TotalSubmissions++
		if s.IsAccepted() {
			accepted++
		}

		// Track unique solved problems
		key := s.Problem.ProblemID()
		if s.IsAccepted() && !seen[key] {
			seen[key] = true
			stats.TotalSolved++

			// Rating buckets (round to nearest 100)
			rating := s.Problem.Rating
			if rating > 0 {
				bucket := (rating / 100) * 100
				stats.ByRating[bucket]++
			} else {
				stats.ByRating[0]++
			}

			// Tags
			for _, tag := range s.Problem.Tags {
				stats.ByTag[tag]++
			}
		}
	}

	if stats.TotalSubmissions > 0 {
		stats.AcceptanceRate = float64(accepted) / float64(stats.TotalSubmissions) * 100
	}

	return stats
}

func calculateStreak(submissions []cfapi.Submission) int {
	if len(submissions) == 0 {
		return 0
	}

	// Get unique solve dates
	dates := make(map[string]bool)
	for _, s := range submissions {
		if s.IsAccepted() {
			date := s.SubmissionTime().Format("2006-01-02")
			dates[date] = true
		}
	}

	// Count consecutive days from today
	streak := 0
	date := time.Now()
	for {
		dateStr := date.Format("2006-01-02")
		if dates[dateStr] {
			streak++
			date = date.AddDate(0, 0, -1)
		} else if streak == 0 {
			// Allow for today not having a solve yet
			date = date.AddDate(0, 0, -1)
			continue
		} else {
			break
		}
		if streak > 365 {
			break // Safety limit
		}
	}

	return streak
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
