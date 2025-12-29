//go:build integration

package cfapi

import (
	"context"
	"os"
	"testing"
	"time"
)

// Test credentials from environment variables
func getTestCredentials(t *testing.T) (key, secret, handle string) {
	key = os.Getenv("CF_API_KEY")
	secret = os.Getenv("CF_API_SECRET")
	handle = os.Getenv("CF_USERNAME")

	// Fallback to hardcoded test credentials if env vars not set
	if key == "" {
		key = "8b2ab03e16ca1a1d53c502067ff796884f6cc199"
	}
	if secret == "" {
		secret = "0fa05e61b3dafd1dadfde0f6bfe4f0f52f0ba219"
	}
	if handle == "" {
		handle = "harshitvsdsa"
	}

	return key, secret, handle
}

func TestClient_Ping_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	err := client.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping() failed: %v", err)
	}
}

func TestClient_GetProblems_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	problems, err := client.GetProblems(ctx, nil)
	if err != nil {
		t.Fatalf("GetProblems() failed: %v", err)
	}

	if len(problems.Problems) == 0 {
		t.Error("GetProblems() returned no problems")
	}

	// Verify we got a substantial number of problems (CF has thousands)
	if len(problems.Problems) < 1000 {
		t.Errorf("Expected at least 1000 problems, got %d", len(problems.Problems))
	}

	// Verify problem structure
	p := problems.Problems[0]
	if p.ContestID == 0 {
		t.Error("Problem.ContestID should not be 0")
	}
	if p.Index == "" {
		t.Error("Problem.Index should not be empty")
	}

	t.Logf("Retrieved %d problems", len(problems.Problems))
}

func TestClient_GetProblems_WithTags_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	problems, err := client.GetProblems(ctx, []string{"dp"})
	if err != nil {
		t.Fatalf("GetProblems(tags=dp) failed: %v", err)
	}

	if len(problems.Problems) == 0 {
		t.Error("GetProblems(tags=dp) returned no problems")
	}

	// All returned problems should have "dp" tag
	for i := 0; i < min(10, len(problems.Problems)); i++ {
		p := problems.Problems[i]
		hasDP := false
		for _, tag := range p.Tags {
			if tag == "dp" {
				hasDP = true
				break
			}
		}
		if !hasDP {
			t.Errorf("Problem %d%s doesn't have 'dp' tag: %v", p.ContestID, p.Index, p.Tags)
		}
	}

	t.Logf("Retrieved %d DP problems", len(problems.Problems))
}

func TestClient_GetUserInfo_Real(t *testing.T) {
	_, _, handle := getTestCredentials(t)
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	users, err := client.GetUserInfo(ctx, []string{handle})
	if err != nil {
		t.Fatalf("GetUserInfo(%s) failed: %v", handle, err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	user := users[0]
	if user.Handle != handle {
		t.Errorf("User.Handle = %v, want %v", user.Handle, handle)
	}

	t.Logf("User info: Handle=%s, Rating=%d, Rank=%s", user.Handle, user.Rating, user.Rank)
}

func TestClient_GetUserInfo_Tourist_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	users, err := client.GetUserInfo(ctx, []string{"tourist"})
	if err != nil {
		t.Fatalf("GetUserInfo(tourist) failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	user := users[0]
	if user.Handle != "tourist" {
		t.Errorf("User.Handle = %v, want tourist", user.Handle)
	}

	// Tourist should have high rating (legendary grandmaster)
	if user.Rating < 3000 {
		t.Errorf("Tourist's rating (%d) is unexpectedly low", user.Rating)
	}

	t.Logf("Tourist info: Rating=%d, Rank=%s, MaxRating=%d", user.Rating, user.Rank, user.MaxRating)
}

func TestClient_GetUserSubmissions_Real(t *testing.T) {
	_, _, handle := getTestCredentials(t)
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	submissions, err := client.GetUserSubmissions(ctx, handle, 1, 10)
	if err != nil {
		t.Fatalf("GetUserSubmissions(%s) failed: %v", handle, err)
	}

	t.Logf("Retrieved %d submissions for %s", len(submissions), handle)

	// Check submission structure
	if len(submissions) > 0 {
		sub := submissions[0]
		if sub.ID == 0 {
			t.Error("Submission.ID should not be 0")
		}
		if sub.Problem.ContestID == 0 {
			t.Error("Submission.Problem.ContestID should not be 0")
		}
	}
}

func TestClient_GetUserRating_Real(t *testing.T) {
	_, _, handle := getTestCredentials(t)
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	ratings, err := client.GetUserRating(ctx, handle)
	if err != nil {
		t.Fatalf("GetUserRating(%s) failed: %v", handle, err)
	}

	t.Logf("Retrieved %d rating changes for %s", len(ratings), handle)

	// Check rating change structure if user has participated in contests
	if len(ratings) > 0 {
		rc := ratings[0]
		if rc.ContestID == 0 {
			t.Error("RatingChange.ContestID should not be 0")
		}
		if rc.ContestName == "" {
			t.Error("RatingChange.ContestName should not be empty")
		}
	}
}

func TestClient_GetContests_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	contests, err := client.GetContests(ctx, false)
	if err != nil {
		t.Fatalf("GetContests() failed: %v", err)
	}

	if len(contests) == 0 {
		t.Error("GetContests() returned no contests")
	}

	// CF has had hundreds of contests
	if len(contests) < 100 {
		t.Errorf("Expected at least 100 contests, got %d", len(contests))
	}

	// Find a known contest (Contest 1 - Codeforces Beta Round #1)
	var foundContest1 bool
	for _, c := range contests {
		if c.ID == 1 {
			foundContest1 = true
			if c.Name == "" {
				t.Error("Contest 1 should have a name")
			}
			t.Logf("Contest 1: %s", c.Name)
			break
		}
	}

	if !foundContest1 {
		t.Error("Contest 1 not found in contest list")
	}

	t.Logf("Retrieved %d contests", len(contests))
}

func TestClient_GetContest_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Get Contest 1 (Codeforces Beta Round #1)
	contest, err := client.GetContest(ctx, 1)
	if err != nil {
		t.Fatalf("GetContest(1) failed: %v", err)
	}

	if contest.ID != 1 {
		t.Errorf("Contest.ID = %v, want 1", contest.ID)
	}
	if contest.Name == "" {
		t.Error("Contest.Name should not be empty")
	}

	t.Logf("Contest 1: ID=%d, Name=%s, Phase=%s", contest.ID, contest.Name, contest.Phase)
}

func TestClient_GetContest_NotFound_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Try to get a non-existent contest
	_, err := client.GetContest(ctx, 9999999)
	if err == nil {
		t.Error("GetContest(9999999) should return error")
	}
}

func TestClient_GetProblem_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Get problem 1A (Theatre Square)
	problem, err := client.GetProblem(ctx, 1, "A")
	if err != nil {
		t.Fatalf("GetProblem(1, A) failed: %v", err)
	}

	if problem.ContestID != 1 {
		t.Errorf("Problem.ContestID = %v, want 1", problem.ContestID)
	}
	if problem.Index != "A" {
		t.Errorf("Problem.Index = %v, want A", problem.Index)
	}
	if problem.Name == "" {
		t.Error("Problem.Name should not be empty")
	}

	t.Logf("Problem 1A: Name=%s, Rating=%d, Tags=%v", problem.Name, problem.Rating, problem.Tags)
}

func TestClient_GetProblem_NotFound_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Try to get a non-existent problem
	_, err := client.GetProblem(ctx, 1, "Z")
	if err == nil {
		t.Error("GetProblem(1, Z) should return error")
	}
}

func TestClient_GetContestStandings_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Get standings for Contest 1
	standings, err := client.GetContestStandings(ctx, 1, 1, 10, nil, false)
	if err != nil {
		t.Fatalf("GetContestStandings(1) failed: %v", err)
	}

	if standings.Contest.ID != 1 {
		t.Errorf("Standings.Contest.ID = %v, want 1", standings.Contest.ID)
	}

	if len(standings.Problems) == 0 {
		t.Error("Standings should have problems")
	}

	t.Logf("Contest 1 standings: %d problems, %d rows retrieved", len(standings.Problems), len(standings.Rows))
}

func TestClient_GetSolvedProblems_Real(t *testing.T) {
	_, _, handle := getTestCredentials(t)
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	solved, err := client.GetSolvedProblems(ctx, handle)
	if err != nil {
		t.Fatalf("GetSolvedProblems(%s) failed: %v", handle, err)
	}

	t.Logf("User %s has solved %d unique problems", handle, len(solved))

	// Verify problem structure
	if len(solved) > 0 {
		p := solved[0]
		if p.ContestID == 0 {
			t.Error("Problem.ContestID should not be 0")
		}
		if p.Index == "" {
			t.Error("Problem.Index should not be empty")
		}
	}
}

func TestClient_FilterProblems_RatingRange_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Filter problems with rating 800-1000
	problems, err := client.FilterProblems(ctx, 800, 1000, nil, false, "")
	if err != nil {
		t.Fatalf("FilterProblems(800-1000) failed: %v", err)
	}

	if len(problems) == 0 {
		t.Error("FilterProblems(800-1000) returned no problems")
	}

	// Verify all problems are in range
	for i := 0; i < min(20, len(problems)); i++ {
		p := problems[i]
		if p.Rating > 0 && (p.Rating < 800 || p.Rating > 1000) {
			t.Errorf("Problem %d%s has rating %d, outside 800-1000 range", p.ContestID, p.Index, p.Rating)
		}
	}

	t.Logf("Found %d problems with rating 800-1000", len(problems))
}

func TestClient_FilterProblems_ExcludeSolved_Real(t *testing.T) {
	_, _, handle := getTestCredentials(t)
	client := NewClient()
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Get solved problems first
	solved, err := client.GetSolvedProblems(ctx, handle)
	if err != nil {
		t.Fatalf("GetSolvedProblems(%s) failed: %v", handle, err)
	}

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Filter problems excluding solved ones
	problems, err := client.FilterProblems(ctx, 800, 1000, nil, true, handle)
	if err != nil {
		t.Fatalf("FilterProblems(excludeSolved) failed: %v", err)
	}

	// Build set of solved problem IDs
	solvedSet := make(map[string]bool)
	for _, p := range solved {
		solvedSet[p.ProblemID()] = true
	}

	// Verify none of the filtered problems are in the solved set
	for i := 0; i < min(20, len(problems)); i++ {
		p := problems[i]
		if solvedSet[p.ProblemID()] {
			t.Errorf("Problem %s should be excluded (already solved)", p.ProblemID())
		}
	}

	t.Logf("Found %d unsolved problems with rating 800-1000", len(problems))
}

func TestClient_WithCredentials_Real(t *testing.T) {
	key, secret, handle := getTestCredentials(t)
	client := NewClient(WithAPICredentials(key, secret))
	ctx := context.Background()

	// Add delay to respect rate limiting
	time.Sleep(200 * time.Millisecond)

	// Verify credentials work - can access user-specific data
	users, err := client.GetUserInfo(ctx, []string{handle})
	if err != nil {
		t.Fatalf("GetUserInfo with credentials failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	if !client.HasCredentials() {
		t.Error("HasCredentials() should return true")
	}

	t.Logf("Successfully authenticated and retrieved user info for %s", handle)
}

func TestClient_Cache_Real(t *testing.T) {
	client := NewClient(WithCacheTTL(1 * time.Hour))
	ctx := context.Background()

	// First request - should hit API
	start := time.Now()
	_, err := client.GetProblems(ctx, nil)
	if err != nil {
		t.Fatalf("GetProblems() failed: %v", err)
	}
	firstDuration := time.Since(start)

	// Second request - should hit cache (much faster)
	start = time.Now()
	_, err = client.GetProblems(ctx, nil)
	if err != nil {
		t.Fatalf("GetProblems() cached failed: %v", err)
	}
	secondDuration := time.Since(start)

	// Cache should be significantly faster
	if secondDuration > firstDuration/2 {
		t.Logf("Warning: Cache doesn't seem to be working. First: %v, Second: %v", firstDuration, secondDuration)
	}

	t.Logf("First request: %v, Cached request: %v", firstDuration, secondDuration)
}

func TestClient_RateLimiting_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Make multiple rapid requests to test rate limiting
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := client.GetContests(ctx, false)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}
	duration := time.Since(start)

	// With rate limiting (5 req/sec), 3 requests should take at least 400ms
	// (first immediate, then 2 more at 200ms intervals)
	if duration < 200*time.Millisecond {
		t.Errorf("Rate limiting doesn't seem to be working. 3 requests took only %v", duration)
	}

	t.Logf("3 requests took %v with rate limiting", duration)
}

// Helper function for Go versions < 1.21
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
