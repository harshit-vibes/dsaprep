package cfapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// ============ Mock HTTP Transport ============

type mockTransport struct {
	statusCode int
	body       string
	err        error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     make(http.Header),
	}, nil
}

// ============ Request Error Path Tests ============

func TestClient_Request_HTTPError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("network connection refused"),
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetProblems(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for HTTP failure")
	}
	if !strings.Contains(err.Error(), "http request") {
		t.Errorf("Expected 'http request' error, got: %v", err)
	}
}

func TestClient_Request_Non200Status(t *testing.T) {
	transport := &mockTransport{
		statusCode: 500,
		body:       "Internal Server Error",
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetProblems(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "api error (status 500)") {
		t.Errorf("Expected 'api error (status 500)' error, got: %v", err)
	}
}

func TestClient_Request_InvalidJSON(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "not valid json",
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetProblems(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse response") {
		t.Errorf("Expected 'parse response' error, got: %v", err)
	}
}

func TestClient_Request_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"API error message"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetProblems(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for API FAILED status")
	}
	if !strings.Contains(err.Error(), "API error message") {
		t.Errorf("Expected 'API error message' error, got: %v", err)
	}
}

func TestClient_Request_RateLimitError(t *testing.T) {
	client := NewClient()

	// Cancel the context immediately to trigger rate limit error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetProblems(ctx, nil)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

// ============ GetUserInfo Error Paths ============

func TestClient_GetUserInfo_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"handles: User with handle nonexistent not found"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetUserInfo(context.Background(), []string{"nonexistent"})
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestClient_GetUserInfo_InvalidJSON(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "invalid json",
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetUserInfo(context.Background(), []string{"test"})
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// ============ GetUserSubmissions Error Paths ============

func TestClient_GetUserSubmissions_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"handle: User with handle nonexistent not found"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetUserSubmissions(context.Background(), "nonexistent", 1, 10)
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

func TestClient_GetUserSubmissions_InvalidJSON(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "invalid json",
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetUserSubmissions(context.Background(), "test", 1, 10)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// ============ GetUserRating Error Paths ============

func TestClient_GetUserRating_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"handle: User with handle nonexistent not found"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetUserRating(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

func TestClient_GetUserRating_InvalidJSON(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "invalid json",
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetUserRating(context.Background(), "test")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// ============ GetContest Error Paths ============

func TestClient_GetContest_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"contestId: Contest with id 999999 not found"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetContest(context.Background(), 999999)
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

func TestClient_GetContest_InvalidJSON(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "invalid json",
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetContest(context.Background(), 1)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// ============ GetContests Error Paths ============

func TestClient_GetContests_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"Some API error"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetContests(context.Background(), false)
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

func TestClient_GetContests_InvalidJSON(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "invalid json",
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetContests(context.Background(), false)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// ============ GetContestStandings Error Paths ============

func TestClient_GetContestStandings_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"contestId: Contest with id 999999 not found"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetContestStandings(context.Background(), 999999, 1, 10, nil, false)
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

func TestClient_GetContestStandings_InvalidJSON(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "invalid json",
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetContestStandings(context.Background(), 1, 1, 10, nil, false)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// ============ GetProblem Error Paths ============

func TestClient_GetProblem_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"Problem not found"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetProblem(context.Background(), 999999, "Z")
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

// ============ GetSolvedProblems Error Paths ============

func TestClient_GetSolvedProblems_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"handle: User with handle nonexistent not found"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetSolvedProblems(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

// ============ FilterProblems Error Paths ============

func TestClient_FilterProblems_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"API error"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.FilterProblems(context.Background(), 800, 1000, nil, false, "")
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

// ============ Cache Tests ============

func TestCache_RemoveExpired_Manual(t *testing.T) {
	cache := NewCache(10 * time.Millisecond)

	// Add items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Wait for expiry
	time.Sleep(20 * time.Millisecond)

	// Manually call removeExpired (for coverage of this internal method)
	cache.removeExpired()

	// Items should be removed
	if cache.Size() != 0 {
		t.Errorf("Expected cache to be empty, got size %d", cache.Size())
	}
}

// ============ Read Body Error ============

func TestClient_Request_ReadBodyError(t *testing.T) {
	transport := &errorBodyTransport{}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetProblems(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for read body failure")
	}
	if !strings.Contains(err.Error(), "read response") {
		t.Errorf("Expected 'read response' error, got: %v", err)
	}
}

type errorBodyTransport struct{}

func (e *errorBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       &errorReadCloser{},
		Header:     make(http.Header),
	}, nil
}

type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (e *errorReadCloser) Close() error {
	return nil
}

// ============ Ping Tests ============

func TestClient_Ping_HTTPError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("connection refused"),
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	err := client.Ping(context.Background())
	if err == nil {
		t.Error("Expected error for HTTP failure")
	}
}

func TestClient_Ping_APIFailed(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"FAILED","comment":"API is down"}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	err := client.Ping(context.Background())
	if err == nil {
		t.Error("Expected error for API FAILED")
	}
}

// ============ Success Path Mock Tests ============

func TestClient_GetProblems_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":{"problems":[{"contestId":1,"index":"A","name":"Test","rating":800}],"problemStatistics":[]}}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	problems, err := client.GetProblems(context.Background(), nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(problems.Problems) != 1 {
		t.Errorf("Expected 1 problem, got %d", len(problems.Problems))
	}
}

func TestClient_GetUserInfo_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"handle":"tourist","rating":3800}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	users, err := client.GetUserInfo(context.Background(), []string{"tourist"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}

func TestClient_GetUserRating_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"contestId":1,"rank":1,"newRating":1500}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	ratings, err := client.GetUserRating(context.Background(), "tourist")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(ratings) != 1 {
		t.Errorf("Expected 1 rating, got %d", len(ratings))
	}
}

func TestClient_GetContests_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"name":"Contest 1","phase":"FINISHED"}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	contests, err := client.GetContests(context.Background(), false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(contests) != 1 {
		t.Errorf("Expected 1 contest, got %d", len(contests))
	}
}

func TestClient_GetContest_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"name":"Contest 1","phase":"FINISHED"}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	contest, err := client.GetContest(context.Background(), 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if contest.ID != 1 {
		t.Errorf("Expected contest ID 1, got %d", contest.ID)
	}
}

func TestClient_GetContestStandings_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":{"contest":{"id":1},"problems":[],"rows":[]}}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	standings, err := client.GetContestStandings(context.Background(), 1, 1, 10, nil, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if standings.Contest.ID != 1 {
		t.Errorf("Expected contest ID 1, got %d", standings.Contest.ID)
	}
}

// ============ Additional Coverage Tests ============

func TestClient_GetUserInfo_EmptyHandles(t *testing.T) {
	client := NewClient()
	_, err := client.GetUserInfo(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty handles")
	}
	if !strings.Contains(err.Error(), "no handles provided") {
		t.Errorf("Expected 'no handles provided' error, got: %v", err)
	}
}

func TestClient_GetUserInfo_CacheHit(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"handle":"tourist","rating":3800}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// First call - cache miss
	users, err := client.GetUserInfo(context.Background(), []string{"tourist"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	// Second call - cache hit (transport won't be called)
	transport.err = fmt.Errorf("should not be called")
	users2, err := client.GetUserInfo(context.Background(), []string{"tourist"})
	if err != nil {
		t.Errorf("Unexpected error on cache hit: %v", err)
	}
	if len(users2) != 1 {
		t.Errorf("Expected 1 user from cache, got %d", len(users2))
	}
}

func TestClient_GetUserSubmissions_CacheHit(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"verdict":"OK"}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// First call - cache miss
	subs, err := client.GetUserSubmissions(context.Background(), "tourist", 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Second call - cache hit
	transport.err = fmt.Errorf("should not be called")
	subs2, err := client.GetUserSubmissions(context.Background(), "tourist", 1, 10)
	if err != nil {
		t.Errorf("Unexpected error on cache hit: %v", err)
	}
	if len(subs2) != len(subs) {
		t.Errorf("Expected same submission count from cache")
	}
}

func TestClient_GetUserSubmissions_NoFromCount(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// Call with 0 from and 0 count (should not set these params)
	_, err := client.GetUserSubmissions(context.Background(), "tourist", 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestClient_GetUserRating_CacheHit(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"contestId":1,"newRating":1500}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// First call
	ratings, err := client.GetUserRating(context.Background(), "tourist")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Second call - cache hit
	transport.err = fmt.Errorf("should not be called")
	ratings2, err := client.GetUserRating(context.Background(), "tourist")
	if err != nil {
		t.Errorf("Unexpected error on cache hit: %v", err)
	}
	if len(ratings2) != len(ratings) {
		t.Errorf("Expected same rating count from cache")
	}
}

func TestClient_GetContest_CacheHit(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"name":"Contest 1"}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// First call
	contest, err := client.GetContest(context.Background(), 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Second call - cache hit
	transport.err = fmt.Errorf("should not be called")
	contest2, err := client.GetContest(context.Background(), 1)
	if err != nil {
		t.Errorf("Unexpected error on cache hit: %v", err)
	}
	if contest2.ID != contest.ID {
		t.Errorf("Expected same contest from cache")
	}
}

func TestClient_GetContest_NotFound_Mock(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"name":"Contest 1"}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetContest(context.Background(), 999999)
	if err == nil {
		t.Error("Expected error for contest not found")
	}
	if !strings.Contains(err.Error(), "contest 999999 not found") {
		t.Errorf("Expected 'contest 999999 not found' error, got: %v", err)
	}
}

func TestClient_GetContests_CacheHit(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"name":"Contest 1"}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// First call
	contests, err := client.GetContests(context.Background(), false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Second call - cache hit
	transport.err = fmt.Errorf("should not be called")
	contests2, err := client.GetContests(context.Background(), false)
	if err != nil {
		t.Errorf("Unexpected error on cache hit: %v", err)
	}
	if len(contests2) != len(contests) {
		t.Errorf("Expected same contest count from cache")
	}
}

func TestClient_GetContestStandings_WithHandles_Mock(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":{"contest":{"id":1},"problems":[],"rows":[]}}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	standings, err := client.GetContestStandings(context.Background(), 1, 1, 10, []string{"tourist", "jiangly"}, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if standings == nil {
		t.Error("Expected standings, got nil")
	}
}

func TestClient_GetContestStandings_NoFromCount(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":{"contest":{"id":1},"problems":[],"rows":[]}}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	standings, err := client.GetContestStandings(context.Background(), 1, 0, 0, nil, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if standings == nil {
		t.Error("Expected standings, got nil")
	}
}

func TestClient_GetProblem_CacheHit(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":{"problems":[{"contestId":1,"index":"A","name":"Test"}],"problemStatistics":[]}}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// First call
	problem, err := client.GetProblem(context.Background(), 1, "A")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Second call - cache hit
	transport.err = fmt.Errorf("should not be called")
	problem2, err := client.GetProblem(context.Background(), 1, "A")
	if err != nil {
		t.Errorf("Unexpected error on cache hit: %v", err)
	}
	if problem2.ContestID != problem.ContestID {
		t.Errorf("Expected same problem from cache")
	}
}

func TestClient_GetProblem_NotFound_Mock(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":{"problems":[{"contestId":1,"index":"A","name":"Test"}],"problemStatistics":[]}}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	_, err := client.GetProblem(context.Background(), 999999, "Z")
	if err == nil {
		t.Error("Expected error for problem not found")
	}
	if !strings.Contains(err.Error(), "problem 999999Z not found") {
		t.Errorf("Expected 'problem 999999Z not found' error, got: %v", err)
	}
}

func TestClient_GetSolvedProblems_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"verdict":"OK","problem":{"contestId":1,"index":"A","name":"Test"}}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	solved, err := client.GetSolvedProblems(context.Background(), "tourist")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(solved) != 1 {
		t.Errorf("Expected 1 solved problem, got %d", len(solved))
	}
}

func TestClient_GetSolvedProblems_Deduplication(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body: `{"status":"OK","result":[
			{"id":1,"verdict":"OK","problem":{"contestId":1,"index":"A","name":"Test"}},
			{"id":2,"verdict":"OK","problem":{"contestId":1,"index":"A","name":"Test"}},
			{"id":3,"verdict":"WRONG_ANSWER","problem":{"contestId":1,"index":"B","name":"Test2"}}
		]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	solved, err := client.GetSolvedProblems(context.Background(), "tourist")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Should deduplicate: only 1A is accepted (B is WA)
	if len(solved) != 1 {
		t.Errorf("Expected 1 unique solved problem, got %d", len(solved))
	}
}

func TestClient_FilterProblems_ExcludeSolved_Mock(t *testing.T) {
	// Sequential mock for multiple calls
	callCount := 0
	transport := &sequentialTransport{
		responses: []mockResponse{
			// GetProblems response
			{statusCode: 200, body: `{"status":"OK","result":{"problems":[{"contestId":1,"index":"A","name":"Test1","rating":800},{"contestId":1,"index":"B","name":"Test2","rating":900}],"problemStatistics":[]}}`},
			// GetSolvedProblems -> GetUserSubmissions
			{statusCode: 200, body: `{"status":"OK","result":[{"id":1,"verdict":"OK","problem":{"contestId":1,"index":"A","name":"Test1"}}]}`},
		},
		callCount: &callCount,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	filtered, err := client.FilterProblems(context.Background(), 0, 0, nil, true, "tourist")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// 1A should be excluded (solved), 1B should remain
	if len(filtered) != 1 {
		t.Errorf("Expected 1 problem (1B), got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].Index != "B" {
		t.Errorf("Expected problem B, got %s", filtered[0].Index)
	}
}

func TestClient_FilterProblems_RatingFilter(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body: `{"status":"OK","result":{"problems":[
			{"contestId":1,"index":"A","name":"Test1","rating":800},
			{"contestId":1,"index":"B","name":"Test2","rating":1200},
			{"contestId":1,"index":"C","name":"Test3","rating":1600},
			{"contestId":1,"index":"D","name":"Test4","rating":0}
		],"problemStatistics":[]}}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// Filter for rating 1000-1500
	filtered, err := client.FilterProblems(context.Background(), 1000, 1500, nil, false, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Only B (1200) should match. D (0) doesn't have rating so it's not filtered by rating.
	expectedCount := 2 // B and D (D has rating 0 which bypasses the rating filter)
	if len(filtered) != expectedCount {
		t.Errorf("Expected %d problems, got %d", expectedCount, len(filtered))
	}
}

func TestClient_WithAPICredentials_Mock(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"name":"Contest 1"}]}`,
	}
	client := NewClient(
		WithAPICredentials("test-api-key", "test-api-secret"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)

	if !client.HasCredentials() {
		t.Error("Client should have credentials")
	}

	// Make a request that uses auth (contest.standings uses auth)
	_, err := client.GetContests(context.Background(), false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestClient_ClearCache_Mock(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `{"status":"OK","result":[{"id":1,"name":"Contest 1"}]}`,
	}
	client := NewClient(WithHTTPClient(&http.Client{Transport: transport}))

	// Fill cache
	_, _ = client.GetContests(context.Background(), false)

	// Clear cache
	client.ClearCache()

	// Should fetch again (not from cache)
	transport.err = nil
	transport.body = `{"status":"OK","result":[{"id":2,"name":"Contest 2"}]}`
	contests, err := client.GetContests(context.Background(), false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(contests) > 0 && contests[0].ID != 2 {
		t.Errorf("Expected contest ID 2 after cache clear, got %d", contests[0].ID)
	}
}

func TestClient_HasCredentials_False(t *testing.T) {
	client := NewClient()
	if client.HasCredentials() {
		t.Error("Client should not have credentials by default")
	}
}

// Sequential transport for tests that need multiple different responses
type sequentialTransport struct {
	responses []mockResponse
	callCount *int
}

type mockResponse struct {
	statusCode int
	body       string
	err        error
}

func (s *sequentialTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	idx := *s.callCount
	*s.callCount++
	if idx >= len(s.responses) {
		return nil, fmt.Errorf("unexpected request #%d", idx)
	}
	resp := s.responses[idx]
	if resp.err != nil {
		return nil, resp.err
	}
	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(strings.NewReader(resp.body)),
		Header:     make(http.Header),
	}, nil
}
