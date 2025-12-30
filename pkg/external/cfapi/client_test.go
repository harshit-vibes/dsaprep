package cfapi

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.limiter == nil {
		t.Error("limiter should not be nil")
	}
	if client.cache == nil {
		t.Error("cache should not be nil")
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 60 * time.Second}
	client := NewClient(
		WithHTTPClient(customClient),
	)

	if client.httpClient != customClient {
		t.Error("httpClient should be the custom client")
	}
}

func TestNewClient_WithCacheTTL(t *testing.T) {
	client := NewClient(
		WithCacheTTL(10 * time.Minute),
	)

	if client.cache.ttl != 10*time.Minute {
		t.Errorf("cache.ttl = %v, want %v", client.cache.ttl, 10*time.Minute)
	}
}

func TestNewClient_MultipleOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 45 * time.Second}
	client := NewClient(
		WithHTTPClient(customClient),
		WithCacheTTL(15 * time.Minute),
	)

	if client.httpClient != customClient {
		t.Error("httpClient should be custom")
	}
	if client.cache.ttl != 15*time.Minute {
		t.Errorf("cache.ttl = %v, want %v", client.cache.ttl, 15*time.Minute)
	}
}

func TestClient_ClearCache(t *testing.T) {
	client := NewClient()

	// Add some items to cache
	client.cache.Set("key1", "value1")
	client.cache.Set("key2", "value2")

	if client.cache.Size() != 2 {
		t.Errorf("cache.Size() = %v, want 2", client.cache.Size())
	}

	client.ClearCache()

	if client.cache.Size() != 0 {
		t.Errorf("cache.Size() after ClearCache() = %v, want 0", client.cache.Size())
	}
}

func TestClient_GetUserInfo_NoHandles(t *testing.T) {
	client := NewClient()

	_, err := client.GetUserInfo(context.Background(), []string{})
	if err == nil {
		t.Error("GetUserInfo() should return error for empty handles")
	}
}

func TestConstants(t *testing.T) {
	if BaseURL != "https://codeforces.com/api" {
		t.Errorf("BaseURL = %v, want https://codeforces.com/api", BaseURL)
	}
	if DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultTimeout = %v, want 30s", DefaultTimeout)
	}
	if DefaultTTL != 5*time.Minute {
		t.Errorf("DefaultTTL = %v, want 5m", DefaultTTL)
	}
	if RateLimit != 5 {
		t.Errorf("RateLimit = %v, want 5", RateLimit)
	}
}

func TestClient_CacheInteraction(t *testing.T) {
	client := NewClient(WithCacheTTL(1 * time.Hour))

	// Simulate caching some problems
	problems := &ProblemsResponse{
		Problems: []Problem{
			{ContestID: 1, Index: "A", Name: "Test"},
		},
	}

	client.cache.Set("problems:", problems)

	// Verify cache works
	cached, ok := client.cache.Get("problems:")
	if !ok {
		t.Error("Cache should have problems")
	}

	cachedProblems := cached.(*ProblemsResponse)
	if len(cachedProblems.Problems) != 1 {
		t.Errorf("len(Problems) = %v, want 1", len(cachedProblems.Problems))
	}
}

func TestWithHTTPClient_Option(t *testing.T) {
	httpClient := &http.Client{Timeout: 1 * time.Minute}
	opt := WithHTTPClient(httpClient)

	client := &Client{}
	opt(client)

	if client.httpClient != httpClient {
		t.Error("httpClient not set correctly")
	}
}

func TestWithCacheTTL_Option(t *testing.T) {
	opt := WithCacheTTL(20 * time.Minute)

	client := &Client{}
	opt(client)

	if client.cache == nil {
		t.Fatal("cache should be created")
	}
	if client.cache.ttl != 20*time.Minute {
		t.Errorf("cache.ttl = %v, want 20m", client.cache.ttl)
	}
}

func TestResponse_Generic(t *testing.T) {
	// Test the generic Response type
	resp := Response[[]User]{
		Status: "OK",
		Result: []User{
			{Handle: "tourist", Rating: 3800},
		},
	}

	if resp.Status != "OK" {
		t.Errorf("Status = %v, want OK", resp.Status)
	}
	if len(resp.Result) != 1 {
		t.Errorf("len(Result) = %v, want 1", len(resp.Result))
	}
	if resp.Result[0].Handle != "tourist" {
		t.Errorf("Handle = %v, want tourist", resp.Result[0].Handle)
	}
}

func TestResponse_WithError(t *testing.T) {
	resp := Response[[]User]{
		Status:  "FAILED",
		Comment: "handle: User with handle tourist not found",
	}

	if resp.Status != "FAILED" {
		t.Errorf("Status = %v, want FAILED", resp.Status)
	}
	if resp.Comment == "" {
		t.Error("Comment should contain error message")
	}
}

func TestProblemsResponse_Fields(t *testing.T) {
	resp := ProblemsResponse{
		Problems: []Problem{
			{ContestID: 1, Index: "A"},
			{ContestID: 1, Index: "B"},
		},
		ProblemStatistics: []ProblemStatistics{
			{ContestID: 1, Index: "A", SolvedCount: 10000},
		},
	}

	if len(resp.Problems) != 2 {
		t.Errorf("len(Problems) = %v, want 2", len(resp.Problems))
	}
	if len(resp.ProblemStatistics) != 1 {
		t.Errorf("len(ProblemStatistics) = %v, want 1", len(resp.ProblemStatistics))
	}
}

// ============================================================================
// Real API Integration Tests
// These tests use the real Codeforces API
// ============================================================================

func TestClient_GetProblems_WithTags(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	problems, err := client.GetProblems(ctx, []string{"dp"})
	if err != nil {
		t.Skipf("Skipping: API error: %v", err)
	}

	if problems == nil {
		t.Fatal("GetProblems() returned nil")
	}

	// All problems should have dp tag
	for _, p := range problems.Problems[:min(10, len(problems.Problems))] {
		hasDPTag := false
		for _, tag := range p.Tags {
			if tag == "dp" {
				hasDPTag = true
				break
			}
		}
		if !hasDPTag {
			t.Errorf("Problem %s should have dp tag, got %v", p.Name, p.Tags)
		}
	}
}

func TestClient_GetProblems_CacheHit(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// First call
	problems1, err := client.GetProblems(ctx, nil)
	if err != nil {
		t.Skipf("Skipping: API error: %v", err)
	}

	// Second call should use cache
	problems2, err := client.GetProblems(ctx, nil)
	if err != nil {
		t.Fatalf("Second GetProblems() failed: %v", err)
	}

	// Should return same data (from cache)
	if len(problems1.Problems) != len(problems2.Problems) {
		t.Error("Cached response should match original")
	}
}

func TestClient_GetUserInfo_MultipleHandles(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Test with multiple well-known handles
	handles := []string{"tourist", "jiangly"}
	users, err := client.GetUserInfo(ctx, handles)
	if err != nil {
		t.Skipf("Skipping: API error: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	for _, user := range users {
		if user.Handle != "tourist" && user.Handle != "jiangly" {
			t.Errorf("Unexpected handle: %s", user.Handle)
		}
	}
}

func TestClient_GetUserInfo_InvalidHandle(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.GetUserInfo(ctx, []string{"nonexistent_user_12345678901234567890"})
	if err == nil {
		t.Error("Expected error for invalid handle")
	}
}

func TestClient_GetUserSubmissions_Tourist(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Tourist always has submissions
	submissions, err := client.GetUserSubmissions(ctx, "tourist", 1, 5)
	if err != nil {
		t.Skipf("Skipping: API error: %v", err)
	}

	if len(submissions) == 0 {
		t.Error("Tourist should have submissions")
	}

	// Check submission structure
	sub := submissions[0]
	if sub.Verdict == "" {
		t.Error("Submission should have verdict")
	}
	if sub.Problem.ContestID == 0 {
		t.Error("Submission problem should have contest ID")
	}

	t.Logf("Tourist's latest submission: problem=%s, verdict=%s",
		sub.Problem.Name, sub.Verdict)
}

func TestClient_GetContest_NotFound(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.GetContest(ctx, 9999999)
	if err == nil {
		t.Error("Expected error for nonexistent contest")
	}
}

func TestClient_GetContestStandings_WithHandles(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Find tourist in a contest
	standings, err := client.GetContestStandings(ctx, 1, 1, 100, []string{"tourist"}, true)
	if err != nil {
		t.Skipf("Skipping: API error: %v", err)
	}

	// May or may not find tourist in contest 1
	t.Logf("Standings with tourist filter: %d rows", len(standings.Rows))
}

func TestClient_GetProblem_NotFound(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.GetProblem(ctx, 9999999, "Z")
	if err == nil {
		t.Error("Expected error for nonexistent problem")
	}
}

func TestClient_FilterProblems_Real(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Filter by rating range
	filtered, err := client.FilterProblems(ctx, 800, 1000, nil, false, "")
	if err != nil {
		t.Skipf("Skipping: API error: %v", err)
	}

	if len(filtered) == 0 {
		t.Error("Expected some problems in rating range")
	}

	// Verify all are in range
	for _, p := range filtered {
		if p.Rating > 0 && (p.Rating < 800 || p.Rating > 1000) {
			t.Errorf("Problem %s has rating %d, outside range [800, 1000]",
				p.Name, p.Rating)
		}
	}

	t.Logf("Found %d problems in rating range 800-1000", len(filtered))
}

func TestClient_FilterProblems_WithTags(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	filtered, err := client.FilterProblems(ctx, 1200, 1400, []string{"greedy"}, false, "")
	if err != nil {
		t.Skipf("Skipping: API error: %v", err)
	}

	t.Logf("Found %d greedy problems in rating range 1200-1400", len(filtered))
}

func TestClient_FilterProblems_ExcludeSolved(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Get problems excluding tourist's solved
	filtered, err := client.FilterProblems(ctx, 800, 1000, nil, true, "tourist")
	if err != nil {
		t.Skipf("Skipping: API error: %v", err)
	}

	t.Logf("Found %d unsolved problems for tourist in rating range 800-1000", len(filtered))
}

func TestClient_Request_RateLimiting(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	start := time.Now()

	// Make several requests quickly - should be rate limited
	for i := 0; i < 3; i++ {
		err := client.Ping(ctx)
		if err != nil {
			t.Skipf("Skipping: API error: %v", err)
		}
	}

	elapsed := time.Since(start)
	// With rate limit of 5/sec, 3 requests should take at least 400ms
	if elapsed < 200*time.Millisecond {
		t.Logf("Rate limiting may not be effective, elapsed: %v", elapsed)
	}
}
