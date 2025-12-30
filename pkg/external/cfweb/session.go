package cfweb

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
)

// Pre-compiled regexes for CSRF token extraction
var (
	reCSRFInput1 = regexp.MustCompile(`<input[^>]+name="csrf_token"[^>]+value="([^"]+)"`)
	reCSRFInput2 = regexp.MustCompile(`<input[^>]+value="([^"]+)"[^>]+name="csrf_token"`)
	reCSRFMeta   = regexp.MustCompile(`<meta[^>]+name="X-Csrf-Token"[^>]+content="([^"]+)"`)
	reCSRFJS     = regexp.MustCompile(`Codeforces\.getCsrfToken[^"]*"([^"]+)"`)
)

const (
	BaseURL     = "https://codeforces.com"
	UserAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	MaxPageSize = 5 * 1024 * 1024 // 5MB max page size to prevent OOM
)

// Session manages CF web authentication using browser cookies
type Session struct {
	client    *http.Client
	jar       *cookiejar.Jar
	csrfToken string
	handle    string
}

// NewSession creates a new CF session
func NewSession() (*Session, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	return &Session{
		client: client,
		jar:    jar,
	}, nil
}

// NewSessionWithCookie creates a session with the provided cookie string
// Cookie format: "JSESSIONID=xxx; 39ce7=xxx; cf_clearance=xxx; ..."
func NewSessionWithCookie(cookieStr string) (*Session, error) {
	session, err := NewSession()
	if err != nil {
		return nil, err
	}

	if cookieStr != "" {
		session.SetCookie(cookieStr)
	}

	return session, nil
}

// SetCookie parses and sets cookies from a browser cookie string
// Cookie format: "JSESSIONID=xxx; 39ce7=xxx; cf_clearance=xxx; ..."
func (s *Session) SetCookie(cookieStr string) {
	cfURL, _ := url.Parse(BaseURL)

	var cookies []*http.Cookie
	pairs := strings.Split(cookieStr, ";")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		cookie := &http.Cookie{
			Name:   name,
			Value:  value,
			Path:   "/",
			Domain: "codeforces.com",
		}

		// Handle specific cookies
		switch name {
		case "cf_clearance", "39ce7":
			cookie.Domain = ".codeforces.com"
			cookie.Secure = true
			cookie.HttpOnly = true
		case "JSESSIONID":
			cookie.HttpOnly = true
		}

		cookies = append(cookies, cookie)
	}

	if len(cookies) > 0 {
		s.jar.SetCookies(cfURL, cookies)
	}
}

// SetHandle sets the user handle
func (s *Session) SetHandle(handle string) {
	s.handle = handle
}

// HasCookies returns true if any cookies are set
func (s *Session) HasCookies() bool {
	cfURL, _ := url.Parse(BaseURL)
	return len(s.jar.Cookies(cfURL)) > 0
}

// IsAuthenticated returns true if session has cookies that indicate login
func (s *Session) IsAuthenticated() bool {
	cfURL, _ := url.Parse(BaseURL)
	cookies := s.jar.Cookies(cfURL)

	hasSession := false
	for _, c := range cookies {
		if c.Name == "JSESSIONID" || c.Name == "X-User" {
			hasSession = true
			break
		}
	}
	return hasSession
}

// get makes a GET request with proper headers
func (s *Session) get(urlStr string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	return s.client.Do(req)
}

// GetCSRFToken returns the current CSRF token
func (s *Session) GetCSRFToken() string {
	return s.csrfToken
}

// RefreshCSRFToken fetches a fresh CSRF token from any CF page
func (s *Session) RefreshCSRFToken() error {
	resp, err := s.get(BaseURL)
	if err != nil {
		return fmt.Errorf("get page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxPageSize))
	if err != nil {
		return fmt.Errorf("read page: %w", err)
	}

	csrfToken := extractCSRFToken(string(body))
	if csrfToken == "" {
		return fmt.Errorf("csrf token not found")
	}

	s.csrfToken = csrfToken
	return nil
}

// Validate checks if the session is still valid
func (s *Session) Validate() error {
	if !s.HasCookies() {
		return fmt.Errorf("no cookies set")
	}

	resp, err := s.get(BaseURL)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxPageSize))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Check if we're logged in by looking for logout link or handle
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "/logout") {
		return fmt.Errorf("session invalid - not logged in")
	}

	// Refresh CSRF token while we're at it
	if csrfToken := extractCSRFToken(bodyStr); csrfToken != "" {
		s.csrfToken = csrfToken
	}

	return nil
}

// Handle returns the configured handle
func (s *Session) Handle() string {
	return s.handle
}

// IsReadyForSubmission returns true if session is ready for submitting solutions
// Requires authentication cookies and a handle to be set
func (s *Session) IsReadyForSubmission() bool {
	return s.IsAuthenticated() && s.handle != ""
}

// Client returns the underlying HTTP client
func (s *Session) Client() *http.Client {
	return s.client
}

// Helper functions

// extractCSRFToken extracts CSRF token from HTML
func extractCSRFToken(htmlStr string) string {
	if matches := reCSRFInput1.FindStringSubmatch(htmlStr); len(matches) > 1 {
		return matches[1]
	}
	if matches := reCSRFInput2.FindStringSubmatch(htmlStr); len(matches) > 1 {
		return matches[1]
	}
	if matches := reCSRFMeta.FindStringSubmatch(htmlStr); len(matches) > 1 {
		return matches[1]
	}
	if matches := reCSRFJS.FindStringSubmatch(htmlStr); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractHiddenInput extracts value of a hidden input field
func extractHiddenInput(htmlStr, name string) string {
	// Pattern: <input ... name="NAME" ... value="VALUE" ...>
	pattern := regexp.MustCompile(`<input[^>]+name="` + regexp.QuoteMeta(name) + `"[^>]+value="([^"]*)"`)
	if matches := pattern.FindStringSubmatch(htmlStr); len(matches) > 1 {
		return matches[1]
	}
	// Try alternate order: value before name
	pattern2 := regexp.MustCompile(`<input[^>]+value="([^"]*)"[^>]+name="` + regexp.QuoteMeta(name) + `"`)
	if matches := pattern2.FindStringSubmatch(htmlStr); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GetHTMLDocument parses HTML into a document
func GetHTMLDocument(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}
