package cfapi

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	BaseURL            = "https://codeforces.com/api"
	DefaultTimeout     = 30 * time.Second
	DefaultTTL         = 5 * time.Minute
	RateLimit          = 5  // requests per second
	MaxResponseSize    = 10 * 1024 * 1024 // 10MB max response size to prevent OOM
)

// Client is the Codeforces API client
type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
	cache      *Cache
	apiKey     string
	apiSecret  string
}

// ClientOption configures the client
type ClientOption func(*Client)

// WithAPICredentials sets API key and secret for authenticated requests
func WithAPICredentials(key, secret string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
		c.apiSecret = secret
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithCacheTTL sets custom cache TTL
func WithCacheTTL(ttl time.Duration) ClientOption {
	return func(c *Client) {
		c.cache = NewCache(ttl)
	}
}

// NewClient creates a new Codeforces API client
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		limiter:    rate.NewLimiter(rate.Limit(RateLimit), 1),
		cache:      NewCache(DefaultTTL),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// request makes an API request with rate limiting
func (c *Client) request(ctx context.Context, method string, params url.Values, useAuth bool) ([]byte, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	// Build URL
	u := fmt.Sprintf("%s/%s", BaseURL, method)

	if params == nil {
		params = url.Values{}
	}

	// Add authentication if required
	if useAuth && c.apiKey != "" && c.apiSecret != "" {
		c.signRequest(method, params)
	}

	fullURL := u + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "cf/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// Use bounded reader to prevent OOM from large responses
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// signRequest adds authentication parameters to the request
func (c *Client) signRequest(method string, params url.Values) {
	rand := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
	params.Set("apiKey", c.apiKey)
	params.Set("time", strconv.FormatInt(time.Now().Unix(), 10))

	// Sort parameters
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build signature string
	var sb strings.Builder
	sb.WriteString(rand)
	sb.WriteString("/")
	sb.WriteString(method)
	sb.WriteString("?")

	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params.Get(k))
	}
	sb.WriteString("#")
	sb.WriteString(c.apiSecret)

	// Calculate SHA512 hash
	hash := sha512.Sum512([]byte(sb.String()))
	sig := hex.EncodeToString(hash[:])

	params.Set("apiSig", rand+sig)
}

// GetProblems retrieves all problems from the problemset
func (c *Client) GetProblems(ctx context.Context, tags []string) (*ProblemsResponse, error) {
	cacheKey := "problems:" + strings.Join(tags, ",")

	// Check cache
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*ProblemsResponse), nil
	}

	params := url.Values{}
	if len(tags) > 0 {
		params.Set("tags", strings.Join(tags, ";"))
	}

	body, err := c.request(ctx, "problemset.problems", params, false)
	if err != nil {
		return nil, err
	}

	var resp Response[ProblemsResponse]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Status != "OK" {
		return nil, fmt.Errorf("api error: %s", resp.Comment)
	}

	c.cache.Set(cacheKey, &resp.Result)
	return &resp.Result, nil
}

// GetUserInfo retrieves information about users
func (c *Client) GetUserInfo(ctx context.Context, handles []string) ([]User, error) {
	if len(handles) == 0 {
		return nil, fmt.Errorf("no handles provided")
	}

	cacheKey := "users:" + strings.Join(handles, ",")

	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.([]User), nil
	}

	params := url.Values{}
	params.Set("handles", strings.Join(handles, ";"))

	body, err := c.request(ctx, "user.info", params, false)
	if err != nil {
		return nil, err
	}

	var resp Response[[]User]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Status != "OK" {
		return nil, fmt.Errorf("api error: %s", resp.Comment)
	}

	c.cache.Set(cacheKey, resp.Result)
	return resp.Result, nil
}

// GetUserSubmissions retrieves submissions for a user
func (c *Client) GetUserSubmissions(ctx context.Context, handle string, from, count int) ([]Submission, error) {
	cacheKey := fmt.Sprintf("submissions:%s:%d:%d", handle, from, count)

	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.([]Submission), nil
	}

	params := url.Values{}
	params.Set("handle", handle)
	if from > 0 {
		params.Set("from", strconv.Itoa(from))
	}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}

	body, err := c.request(ctx, "user.status", params, false)
	if err != nil {
		return nil, err
	}

	var resp Response[[]Submission]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Status != "OK" {
		return nil, fmt.Errorf("api error: %s", resp.Comment)
	}

	c.cache.Set(cacheKey, resp.Result)
	return resp.Result, nil
}

// GetUserRating retrieves rating history for a user
func (c *Client) GetUserRating(ctx context.Context, handle string) ([]RatingChange, error) {
	cacheKey := "rating:" + handle

	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.([]RatingChange), nil
	}

	params := url.Values{}
	params.Set("handle", handle)

	body, err := c.request(ctx, "user.rating", params, false)
	if err != nil {
		return nil, err
	}

	var resp Response[[]RatingChange]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Status != "OK" {
		return nil, fmt.Errorf("api error: %s", resp.Comment)
	}

	c.cache.Set(cacheKey, resp.Result)
	return resp.Result, nil
}

// GetContest retrieves contest information
func (c *Client) GetContest(ctx context.Context, contestID int) (*Contest, error) {
	cacheKey := fmt.Sprintf("contest:%d", contestID)

	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*Contest), nil
	}

	// Get all contests and filter
	contests, err := c.GetContests(ctx, false)
	if err != nil {
		return nil, err
	}

	for i := range contests {
		if contests[i].ID == contestID {
			c.cache.Set(cacheKey, &contests[i])
			return &contests[i], nil
		}
	}

	return nil, fmt.Errorf("contest %d not found", contestID)
}

// GetContests retrieves list of contests
func (c *Client) GetContests(ctx context.Context, gym bool) ([]Contest, error) {
	cacheKey := fmt.Sprintf("contests:%v", gym)

	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.([]Contest), nil
	}

	params := url.Values{}
	params.Set("gym", strconv.FormatBool(gym))

	body, err := c.request(ctx, "contest.list", params, false)
	if err != nil {
		return nil, err
	}

	var resp Response[[]Contest]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Status != "OK" {
		return nil, fmt.Errorf("api error: %s", resp.Comment)
	}

	c.cache.Set(cacheKey, resp.Result)
	return resp.Result, nil
}

// GetContestStandings retrieves contest standings
func (c *Client) GetContestStandings(ctx context.Context, contestID int, from, count int, handles []string, showUnofficial bool) (*ContestStandings, error) {
	params := url.Values{}
	params.Set("contestId", strconv.Itoa(contestID))
	if from > 0 {
		params.Set("from", strconv.Itoa(from))
	}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	if len(handles) > 0 {
		params.Set("handles", strings.Join(handles, ";"))
	}
	params.Set("showUnofficial", strconv.FormatBool(showUnofficial))

	body, err := c.request(ctx, "contest.standings", params, false)
	if err != nil {
		return nil, err
	}

	var resp Response[ContestStandings]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if resp.Status != "OK" {
		return nil, fmt.Errorf("api error: %s", resp.Comment)
	}

	return &resp.Result, nil
}

// GetProblem retrieves a single problem by contest ID and index
func (c *Client) GetProblem(ctx context.Context, contestID int, index string) (*Problem, error) {
	cacheKey := fmt.Sprintf("problem:%d:%s", contestID, index)

	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*Problem), nil
	}

	problems, err := c.GetProblems(ctx, nil)
	if err != nil {
		return nil, err
	}

	for i := range problems.Problems {
		p := &problems.Problems[i]
		if p.ContestID == contestID && p.Index == index {
			c.cache.Set(cacheKey, p)
			return p, nil
		}
	}

	return nil, fmt.Errorf("problem %d%s not found", contestID, index)
}

// GetSolvedProblems returns all problems solved by a user
func (c *Client) GetSolvedProblems(ctx context.Context, handle string) ([]Problem, error) {
	submissions, err := c.GetUserSubmissions(ctx, handle, 1, 10000)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var solved []Problem

	for _, sub := range submissions {
		if sub.IsAccepted() {
			key := sub.Problem.ProblemID()
			if !seen[key] {
				seen[key] = true
				solved = append(solved, sub.Problem)
			}
		}
	}

	return solved, nil
}

// FilterProblems filters problems by criteria
func (c *Client) FilterProblems(ctx context.Context, minRating, maxRating int, tags []string, excludeSolved bool, handle string) ([]Problem, error) {
	problems, err := c.GetProblems(ctx, tags)
	if err != nil {
		return nil, err
	}

	var solvedSet map[string]bool
	if excludeSolved && handle != "" {
		solved, err := c.GetSolvedProblems(ctx, handle)
		if err != nil {
			return nil, err
		}
		solvedSet = make(map[string]bool)
		for _, p := range solved {
			solvedSet[p.ProblemID()] = true
		}
	}

	var filtered []Problem
	for _, p := range problems.Problems {
		// Rating filter
		if p.Rating > 0 {
			if minRating > 0 && p.Rating < minRating {
				continue
			}
			if maxRating > 0 && p.Rating > maxRating {
				continue
			}
		}

		// Exclude solved
		if excludeSolved && solvedSet != nil {
			if solvedSet[p.ProblemID()] {
				continue
			}
		}

		filtered = append(filtered, p)
	}

	return filtered, nil
}

// Ping checks if the API is accessible
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.GetContests(ctx, false)
	return err
}

// ClearCache clears the API cache
func (c *Client) ClearCache() {
	c.cache.Clear()
}

// HasCredentials returns true if API credentials are configured
func (c *Client) HasCredentials() bool {
	return c.apiKey != "" && c.apiSecret != ""
}
