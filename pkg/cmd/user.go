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
	// user submissions flags
	submissionsLimit  int
	submissionsVerdict string
)

var userCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"u"},
	Short:   "User profile and statistics",
	Long: `Commands for viewing Codeforces user information.

View user profiles, submission history, and rating changes.`,
}

var userInfoCmd = &cobra.Command{
	Use:   "info [handle]",
	Short: "Show user profile information",
	Long: `Display Codeforces user profile information.

If no handle is provided, uses the configured CF handle.

Examples:
  cf user info           # Show your profile
  cf user info tourist   # Show tourist's profile`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUserInfo,
}

var userSubmissionsCmd = &cobra.Command{
	Use:     "submissions [handle]",
	Aliases: []string{"sub", "subs"},
	Short:   "Show user submissions",
	Long: `Display recent submissions for a user.

If no handle is provided, uses the configured CF handle.

Examples:
  cf user submissions                 # Your recent submissions
  cf user submissions --limit 50      # Last 50 submissions
  cf user submissions --verdict AC    # Only accepted submissions`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUserSubmissions,
}

var userRatingCmd = &cobra.Command{
	Use:   "rating [handle]",
	Short: "Show rating history",
	Long: `Display rating change history for a user.

If no handle is provided, uses the configured CF handle.

Examples:
  cf user rating           # Your rating history
  cf user rating tourist   # tourist's rating history`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUserRating,
}

func init() {
	// Add user subcommands
	userCmd.AddCommand(userInfoCmd)
	userCmd.AddCommand(userSubmissionsCmd)
	userCmd.AddCommand(userRatingCmd)

	// user submissions flags
	userSubmissionsCmd.Flags().IntVar(&submissionsLimit, "limit", 10, "Number of submissions to show")
	userSubmissionsCmd.Flags().StringVar(&submissionsVerdict, "verdict", "", "Filter by verdict (AC, WA, TLE, etc.)")
}

func getHandle(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	creds, err := config.LoadCredentials()
	if err != nil || creds == nil || creds.CFHandle == "" {
		return "", fmt.Errorf("no handle provided and no CF handle configured. Set with 'cf config set cf_handle <handle>'")
	}

	return creds.CFHandle, nil
}

func getAPIClient() *cfapi.Client {
	creds, _ := config.LoadCredentials()
	if creds != nil && creds.IsAPIConfigured() {
		return cfapi.NewClient(cfapi.WithAPICredentials(creds.APIKey, creds.APISecret))
	}
	return cfapi.NewClient()
}

func runUserInfo(cmd *cobra.Command, args []string) error {
	handle, err := getHandle(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := getAPIClient()
	users, err := client.GetUserInfo(ctx, []string{handle})
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("user %s not found", handle)
	}

	u := users[0]

	fmt.Printf("\n%s\n", u.Handle)
	fmt.Println(strings.Repeat("─", 40))

	// Rank with color indicator
	rankColor := getRankColor(u.Rating)
	fmt.Printf("  Rank:         %s%s\033[0m\n", rankColor, u.Rank)
	fmt.Printf("  Rating:       %s%d\033[0m (max: %d)\n", rankColor, u.Rating, u.MaxRating)

	if u.Country != "" {
		location := u.Country
		if u.City != "" {
			location = u.City + ", " + u.Country
		}
		fmt.Printf("  Location:     %s\n", location)
	}

	if u.Organization != "" {
		fmt.Printf("  Organization: %s\n", u.Organization)
	}

	fmt.Printf("  Contribution: %d\n", u.Contribution)
	fmt.Printf("  Friends:      %d\n", u.FriendOfCount)
	fmt.Printf("  Registered:   %s\n", u.RegistrationTime().Format("Jan 2006"))
	fmt.Printf("  Last Online:  %s\n", formatTimeAgo(u.LastOnline()))

	fmt.Println()
	return nil
}

func runUserSubmissions(cmd *cobra.Command, args []string) error {
	handle, err := getHandle(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := getAPIClient()

	// Fetch more if filtering
	fetchCount := submissionsLimit
	if submissionsVerdict != "" {
		fetchCount = submissionsLimit * 5 // Fetch extra to ensure we get enough after filtering
	}

	submissions, err := client.GetUserSubmissions(ctx, handle, 1, fetchCount)
	if err != nil {
		return fmt.Errorf("failed to get submissions: %w", err)
	}

	// Filter by verdict if specified
	if submissionsVerdict != "" {
		verdict := strings.ToUpper(submissionsVerdict)
		var filtered []cfapi.Submission
		for _, s := range submissions {
			if s.Verdict == verdict {
				filtered = append(filtered, s)
			}
		}
		submissions = filtered
	}

	// Limit results
	if len(submissions) > submissionsLimit {
		submissions = submissions[:submissionsLimit]
	}

	if len(submissions) == 0 {
		fmt.Println("No submissions found.")
		return nil
	}

	fmt.Printf("\nRecent submissions for %s:\n\n", handle)
	fmt.Printf("%-12s %-10s %-40s %8s  %s\n", "Time", "Problem", "Name", "Verdict", "Language")
	fmt.Println(strings.Repeat("─", 100))

	for _, s := range submissions {
		name := s.Problem.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}

		verdictColor := getVerdictColor(s.Verdict)
		timeStr := s.SubmissionTime().Format("Jan 02 15:04")

		fmt.Printf("%-12s %-10s %-40s %s%-8s\033[0m  %s\n",
			timeStr,
			s.Problem.ProblemID(),
			name,
			verdictColor,
			s.Verdict,
			s.ProgrammingLanguage,
		)
	}

	fmt.Println()
	return nil
}

func runUserRating(cmd *cobra.Command, args []string) error {
	handle, err := getHandle(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := getAPIClient()
	changes, err := client.GetUserRating(ctx, handle)
	if err != nil {
		return fmt.Errorf("failed to get rating history: %w", err)
	}

	if len(changes) == 0 {
		fmt.Printf("%s has not participated in any rated contests.\n", handle)
		return nil
	}

	fmt.Printf("\nRating history for %s (%d contests):\n\n", handle, len(changes))
	fmt.Printf("%-12s %-50s %5s → %5s  %s\n", "Date", "Contest", "Old", "New", "Delta")
	fmt.Println(strings.Repeat("─", 100))

	// Show last 15 contests (most recent)
	start := 0
	if len(changes) > 15 {
		start = len(changes) - 15
		fmt.Printf("  ... %d earlier contests ...\n", start)
	}

	for i := start; i < len(changes); i++ {
		rc := changes[i]
		contestName := rc.ContestName
		if len(contestName) > 48 {
			contestName = contestName[:45] + "..."
		}

		delta := rc.RatingDelta()
		deltaStr := fmt.Sprintf("%+d", delta)
		deltaColor := "\033[32m" // green
		if delta < 0 {
			deltaColor = "\033[31m" // red
		}

		date := time.Unix(rc.RatingUpdateTimeSeconds, 0).Format("Jan 02 2006")

		fmt.Printf("%-12s %-50s %5d → %5d  %s%s\033[0m\n",
			date,
			contestName,
			rc.OldRating,
			rc.NewRating,
			deltaColor,
			deltaStr,
		)
	}

	// Summary
	first := changes[0]
	last := changes[len(changes)-1]
	totalDelta := last.NewRating - first.OldRating

	fmt.Println(strings.Repeat("─", 100))
	deltaColor := "\033[32m"
	if totalDelta < 0 {
		deltaColor = "\033[31m"
	}
	fmt.Printf("Total change: %s%+d\033[0m over %d contests\n", deltaColor, totalDelta, len(changes))
	fmt.Println()

	return nil
}

// getRankColor returns ANSI color code for CF rank
func getRankColor(rating int) string {
	switch {
	case rating >= 3000:
		return "\033[31m" // red - legendary grandmaster
	case rating >= 2600:
		return "\033[31m" // red - international grandmaster
	case rating >= 2400:
		return "\033[31m" // red - grandmaster
	case rating >= 2300:
		return "\033[33m" // orange - international master
	case rating >= 2100:
		return "\033[33m" // orange - master
	case rating >= 1900:
		return "\033[35m" // violet - candidate master
	case rating >= 1600:
		return "\033[34m" // blue - expert
	case rating >= 1400:
		return "\033[36m" // cyan - specialist
	case rating >= 1200:
		return "\033[32m" // green - pupil
	default:
		return "\033[90m" // gray - newbie
	}
}

// getVerdictColor returns ANSI color code for verdict
func getVerdictColor(verdict string) string {
	switch verdict {
	case cfapi.VerdictOK:
		return "\033[32m" // green
	case cfapi.VerdictWrongAnswer, cfapi.VerdictRuntimeError:
		return "\033[31m" // red
	case cfapi.VerdictTimeLimitExceeded, cfapi.VerdictMemoryLimitExceeded:
		return "\033[33m" // yellow
	case cfapi.VerdictCompilationError:
		return "\033[35m" // magenta
	default:
		return "\033[90m" // gray
	}
}

// formatTimeAgo formats a time as "X ago"
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d < 30*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}
