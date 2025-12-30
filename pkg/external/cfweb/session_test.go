package cfweb

import (
	"strings"
	"testing"
)

func TestNewSession(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	if session == nil {
		t.Fatal("NewSession() returned nil")
	}

	if session.client == nil {
		t.Error("session.client should not be nil")
	}

	if session.jar == nil {
		t.Error("session.jar should not be nil")
	}

	if session.Handle() != "" {
		t.Error("new session handle should be empty")
	}

	if session.GetCSRFToken() != "" {
		t.Error("new session CSRF token should be empty")
	}
}

func TestNewSessionWithCookie(t *testing.T) {
	cookieStr := "JSESSIONID=test123; 39ce7=abc456"
	session, err := NewSessionWithCookie(cookieStr)
	if err != nil {
		t.Fatalf("NewSessionWithCookie() failed: %v", err)
	}

	if session == nil {
		t.Fatal("NewSessionWithCookie() returned nil")
	}

	if !session.HasCookies() {
		t.Error("session should have cookies set")
	}
}

func TestSession_SetCookie(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	if session.HasCookies() {
		t.Error("new session should not have cookies")
	}

	// Set a cookie string
	session.SetCookie("JSESSIONID=test123; 39ce7=abc456; cf_clearance=xyz789")

	if !session.HasCookies() {
		t.Error("session should have cookies after SetCookie")
	}
}

func TestSession_SetCookie_Complex(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	// Full browser cookie string
	cookieStr := "JSESSIONID=24FF903C9002F539DCDE4C869C77C1DD; 39ce7=CFtzSSKd; _gid=GA1.2.1125406761.1766998527; cf_clearance=abc123"
	session.SetCookie(cookieStr)

	if !session.HasCookies() {
		t.Error("session should have cookies")
	}
}

func TestSession_SetCookie_Empty(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	session.SetCookie("")

	if session.HasCookies() {
		t.Error("session should not have cookies after empty SetCookie")
	}
}

func TestSession_SetHandle(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	if session.Handle() != "" {
		t.Error("new session handle should be empty")
	}

	session.SetHandle("testuser")
	if session.Handle() != "testuser" {
		t.Errorf("Handle() = %s, want testuser", session.Handle())
	}
}

func TestSession_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		cookie   string
		wantAuth bool
	}{
		{
			name:     "no cookies",
			cookie:   "",
			wantAuth: false,
		},
		{
			name:     "only random cookie",
			cookie:   "random=value",
			wantAuth: false,
		},
		{
			name:     "with JSESSIONID",
			cookie:   "JSESSIONID=test123",
			wantAuth: true,
		},
		{
			name:     "with X-User",
			cookie:   "X-User=test123",
			wantAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, _ := NewSession()
			if tt.cookie != "" {
				session.SetCookie(tt.cookie)
			}

			if session.IsAuthenticated() != tt.wantAuth {
				t.Errorf("IsAuthenticated() = %v, want %v", session.IsAuthenticated(), tt.wantAuth)
			}
		})
	}
}

func TestSession_IsReadyForSubmission(t *testing.T) {
	tests := []struct {
		name      string
		cookie    string
		handle    string
		wantReady bool
	}{
		{
			name:      "nothing set",
			wantReady: false,
		},
		{
			name:      "only cookie",
			cookie:    "JSESSIONID=test123",
			handle:    "",
			wantReady: false,
		},
		{
			name:      "only handle",
			cookie:    "",
			handle:    "testuser",
			wantReady: false,
		},
		{
			name:      "both set",
			cookie:    "JSESSIONID=test123",
			handle:    "testuser",
			wantReady: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, _ := NewSession()
			if tt.cookie != "" {
				session.SetCookie(tt.cookie)
			}
			if tt.handle != "" {
				session.SetHandle(tt.handle)
			}

			if session.IsReadyForSubmission() != tt.wantReady {
				t.Errorf("IsReadyForSubmission() = %v, want %v", session.IsReadyForSubmission(), tt.wantReady)
			}
		})
	}
}

func TestSession_HasCookies(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	if session.HasCookies() {
		t.Error("new session should not have cookies")
	}

	session.SetCookie("test=value")
	if !session.HasCookies() {
		t.Error("session should have cookies after SetCookie")
	}
}

func TestSession_Client(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	client := session.Client()
	if client == nil {
		t.Error("Client() should not return nil")
	}

	if client != session.client {
		t.Error("Client() should return the internal client")
	}
}

func TestSession_GetCSRFToken(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	if session.GetCSRFToken() != "" {
		t.Error("new session CSRF token should be empty")
	}

	session.csrfToken = "test-token-12345"
	if session.GetCSRFToken() != "test-token-12345" {
		t.Errorf("GetCSRFToken() = %s, want test-token-12345", session.GetCSRFToken())
	}
}

func TestExtractCSRFToken(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "input field name first",
			html: `<input type="hidden" name="csrf_token" value="abc123def456"/>`,
			want: "abc123def456",
		},
		{
			name: "input field value first",
			html: `<input type="hidden" value="xyz789abc123" name="csrf_token"/>`,
			want: "xyz789abc123",
		},
		{
			name: "meta tag",
			html: `<meta name="X-Csrf-Token" content="meta123token"/>`,
			want: "meta123token",
		},
		{
			name: "javascript variable",
			html: `<script>Codeforces.getCsrfToken = function() { return "js_token_123"; }</script>`,
			want: "js_token_123",
		},
		{
			name: "no token",
			html: `<html><body>No token here</body></html>`,
			want: "",
		},
		{
			name: "empty input",
			html: "",
			want: "",
		},
		{
			name: "real CF-like HTML",
			html: `<!DOCTYPE html>
<html>
<head><title>Codeforces</title></head>
<body>
<form>
<input type="hidden" name="csrf_token" value="8a9b0c1d2e3f4g5h6i7j8k9l0m"/>
<input type="text" name="username"/>
</form>
</body>
</html>`,
			want: "8a9b0c1d2e3f4g5h6i7j8k9l0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCSRFToken(tt.html)
			if got != tt.want {
				t.Errorf("extractCSRFToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractHiddenInput(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		field string
		want  string
	}{
		{
			name:  "ftaa name first",
			html:  `<input type="hidden" name="ftaa" value="ftaa_value_123"/>`,
			field: "ftaa",
			want:  "ftaa_value_123",
		},
		{
			name:  "bfaa value first",
			html:  `<input type="hidden" value="bfaa_value_456" name="bfaa"/>`,
			field: "bfaa",
			want:  "bfaa_value_456",
		},
		{
			name:  "field not found",
			html:  `<input type="hidden" name="other" value="something"/>`,
			field: "ftaa",
			want:  "",
		},
		{
			name:  "empty value",
			html:  `<input type="hidden" name="ftaa" value=""/>`,
			field: "ftaa",
			want:  "",
		},
		{
			name:  "empty html",
			html:  "",
			field: "ftaa",
			want:  "",
		},
		{
			name:  "multiple inputs",
			html:  `<input name="a" value="1"/><input name="ftaa" value="found"/><input name="b" value="2"/>`,
			field: "ftaa",
			want:  "found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHiddenInput(tt.html, tt.field)
			if got != tt.want {
				t.Errorf("extractHiddenInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetHTMLDocument(t *testing.T) {
	html := `<!DOCTYPE html><html><head><title>Test</title></head><body><p>Hello</p></body></html>`
	reader := strings.NewReader(html)

	doc, err := GetHTMLDocument(reader)
	if err != nil {
		t.Fatalf("GetHTMLDocument() failed: %v", err)
	}

	if doc == nil {
		t.Error("GetHTMLDocument() returned nil")
	}
}

func TestGetHTMLDocument_Empty(t *testing.T) {
	reader := strings.NewReader("")

	doc, err := GetHTMLDocument(reader)
	// Empty HTML should still parse (to an empty document)
	if err != nil {
		t.Fatalf("GetHTMLDocument() failed: %v", err)
	}

	if doc == nil {
		t.Error("GetHTMLDocument() returned nil")
	}
}

func TestConstants(t *testing.T) {
	if BaseURL != "https://codeforces.com" {
		t.Errorf("BaseURL = %s, want https://codeforces.com", BaseURL)
	}
	if UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
	if MaxPageSize != 5*1024*1024 {
		t.Errorf("MaxPageSize = %d, want 5MB", MaxPageSize)
	}
}

func TestSession_Validate_NoCookies(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	err = session.Validate()
	if err == nil {
		t.Error("Validate() should return error when no cookies set")
	}
}
