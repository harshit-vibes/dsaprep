//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/external/cfweb"
	"github.com/harshit-vibes/cf/pkg/internal/config"
)

// These tests require real credentials and network access
// Run with: go test -tags=integration ./pkg/integration/...

func TestRealAPI_GetUserInfo(t *testing.T) {
	creds, err := config.LoadCredentials()
	if err != nil {
		t.Skipf("Skipping: cannot load credentials: %v", err)
	}
	if !creds.IsAPIConfigured() {
		t.Skip("Skipping: API credentials not configured")
	}

	client := cfapi.NewClient(cfapi.WithAPICredentials(creds.APIKey, creds.APISecret))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users, err := client.GetUserInfo(ctx, []string{creds.CFHandle})
	if err != nil {
		t.Fatalf("GetUserInfo failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	if users[0].Handle != creds.CFHandle {
		t.Errorf("Handle = %v, want %v", users[0].Handle, creds.CFHandle)
	}

	t.Logf("User: %s, Rating: %d, Rank: %s", users[0].Handle, users[0].Rating, users[0].Rank)
}

func TestRealAPI_GetProblems(t *testing.T) {
	creds, err := config.LoadCredentials()
	if err != nil {
		t.Skipf("Skipping: cannot load credentials: %v", err)
	}

	client := cfapi.NewClient()
	if creds != nil && creds.IsAPIConfigured() {
		client = cfapi.NewClient(cfapi.WithAPICredentials(creds.APIKey, creds.APISecret))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	problems, err := client.GetProblems(ctx, nil)
	if err != nil {
		t.Fatalf("GetProblems failed: %v", err)
	}

	if len(problems.Problems) == 0 {
		t.Error("Expected problems, got none")
	}

	t.Logf("Fetched %d problems", len(problems.Problems))

	// Verify some problems have expected fields
	for i, p := range problems.Problems[:5] {
		if p.ContestID == 0 {
			t.Errorf("Problem %d: ContestID is 0", i)
		}
		if p.Index == "" {
			t.Errorf("Problem %d: Index is empty", i)
		}
		if p.Name == "" {
			t.Errorf("Problem %d: Name is empty", i)
		}
	}
}

func TestRealWeb_ParseProblem(t *testing.T) {
	parser := cfweb.NewParserWithClient(nil)

	// Parse a classic problem
	problem, err := parser.ParseProblem(1, "A")
	if err != nil {
		t.Fatalf("ParseProblem failed: %v", err)
	}

	if problem.Name == "" {
		t.Error("Problem name is empty")
	}
	if problem.ContestID != 1 {
		t.Errorf("ContestID = %d, want 1", problem.ContestID)
	}
	if problem.Index != "A" {
		t.Errorf("Index = %s, want A", problem.Index)
	}
	if len(problem.Samples) == 0 {
		t.Error("Expected samples, got none")
	}

	t.Logf("Parsed: %s. %s", problem.Index, problem.Name)
	t.Logf("  Time: %s, Memory: %s", problem.TimeLimit, problem.MemoryLimit)
	t.Logf("  Tags: %v", problem.Tags)
	t.Logf("  Samples: %d", len(problem.Samples))
}

func TestRealWeb_ParseProblem_WithRating(t *testing.T) {
	parser := cfweb.NewParserWithClient(nil)

	// Parse a problem with known rating
	problem, err := parser.ParseProblem(1325, "A")
	if err != nil {
		t.Fatalf("ParseProblem failed: %v", err)
	}

	if problem.Name != "EhAb AnD gCd" {
		t.Errorf("Name = %s, want 'EhAb AnD gCd'", problem.Name)
	}

	t.Logf("Problem: %s. %s (Rating: %d)", problem.Index, problem.Name, problem.Rating)
}

func TestRealAPI_GetContestStandings(t *testing.T) {
	creds, err := config.LoadCredentials()
	if err != nil {
		t.Skipf("Skipping: cannot load credentials: %v", err)
	}

	client := cfapi.NewClient()
	if creds != nil && creds.IsAPIConfigured() {
		client = cfapi.NewClient(cfapi.WithAPICredentials(creds.APIKey, creds.APISecret))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	standings, err := client.GetContestStandings(ctx, 1, 1, 10, nil, false)
	if err != nil {
		t.Fatalf("GetContestStandings failed: %v", err)
	}

	if standings.Contest.ID != 1 {
		t.Errorf("Contest ID = %d, want 1", standings.Contest.ID)
	}

	t.Logf("Contest: %s", standings.Contest.Name)
	t.Logf("Problems: %d", len(standings.Problems))
	t.Logf("Rows: %d", len(standings.Rows))
}

func TestCredentials_Loaded(t *testing.T) {
	creds, err := config.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if creds.CFHandle == "" {
		t.Error("CF_HANDLE is not set")
	}
	if creds.APIKey == "" {
		t.Error("CF_API_KEY is not set")
	}
	if creds.APISecret == "" {
		t.Error("CF_API_SECRET is not set")
	}

	t.Logf("Handle: %s", creds.CFHandle)
	t.Logf("API Key: %s...", creds.APIKey[:8])
	t.Logf("Has API: %v", creds.IsAPIConfigured())
}

