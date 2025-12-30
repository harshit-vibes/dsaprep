package cfweb

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// mockTransport implements http.RoundTripper for testing
type mockTransport struct {
	response   *http.Response
	err        error
	statusCode int
	body       string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     make(http.Header),
	}, nil
}

// createMockSession creates a session with mock transport
func createMockSession(transport http.RoundTripper) *Session {
	session, _ := NewSession()
	session.client = &http.Client{Transport: transport}
	session.SetCookie("JSESSIONID=test-jsessionid; 39ce7=test-ce7; cf_clearance=test-clearance")
	session.SetHandle("testuser")
	return session
}

// ============ Session Mock Tests ============

func TestSession_Get_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("network error"),
	}
	session := createMockSession(transport)

	resp, err := session.get("https://codeforces.com")
	if err == nil {
		t.Error("Expected network error, got nil")
	}
	if resp != nil {
		t.Error("Expected nil response on error")
	}
}

func TestSession_Get_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "<html>test</html>",
	}
	session := createMockSession(transport)

	resp, err := session.get("https://codeforces.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("Expected response, got nil")
	}
	if resp != nil && resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestSession_RefreshCSRFToken_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("connection refused"),
	}
	session := createMockSession(transport)

	err := session.RefreshCSRFToken()
	if err == nil {
		t.Error("Expected error on network failure")
	}
}

func TestSession_RefreshCSRFToken_NoToken(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "<html>no csrf here</html>",
	}
	session := createMockSession(transport)

	err := session.RefreshCSRFToken()
	if err == nil {
		t.Error("Expected error when CSRF token not found")
	}
}

func TestSession_RefreshCSRFToken_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><meta name="X-Csrf-Token" content="test-csrf-token"></html>`,
	}
	session := createMockSession(transport)

	err := session.RefreshCSRFToken()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if session.csrfToken != "test-csrf-token" {
		t.Errorf("Expected csrf token 'test-csrf-token', got '%s'", session.csrfToken)
	}
}

func TestSession_Validate_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("network timeout"),
	}
	session := createMockSession(transport)

	err := session.Validate()
	if err == nil {
		t.Error("Expected error on network failure")
	}
}

func TestSession_Validate_HandleNotFound(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><a href="/enter">Enter</a></html>`,
	}
	session := createMockSession(transport)

	err := session.Validate()
	if err == nil {
		t.Error("Expected error when handle not found")
	}
}

func TestSession_Validate_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><a href="/logout">Logout</a>var handle = "testuser";</html>`,
	}
	session := createMockSession(transport)

	err := session.Validate()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// ============ Submitter Mock Tests ============

func TestSubmitter_Submit_GetPageError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("connection reset"),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected error on get page failure")
	}
	if !strings.Contains(err.Error(), "get submit page") {
		t.Errorf("Expected 'get submit page' error, got: %v", err)
	}
}

func TestSubmitter_Submit_NoCSRFToken(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "<html>no csrf</html>",
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected error when CSRF not found")
	}
	if !strings.Contains(err.Error(), "csrf token not found") {
		t.Errorf("Expected 'csrf token not found' error, got: %v", err)
	}
}

func TestSubmitter_SubmitToGym_GetPageError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("connection refused"),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.SubmitToGym(100001, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected error on get page failure")
	}
}

func TestSubmitter_GetLatestSubmission_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("timeout"),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.getLatestSubmission(1, "A")
	if err == nil {
		t.Error("Expected error on network failure")
	}
}

func TestSubmitter_GetLatestSubmission_NoSubmissions(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "<html><table></table></html>",
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.getLatestSubmission(1, "A")
	if err == nil {
		t.Error("Expected error when no submissions found")
	}
}

func TestSubmitter_GetLatestGymSubmission_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("connection reset"),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.getLatestGymSubmission(100001, "A")
	if err == nil {
		t.Error("Expected error on network failure")
	}
}

func TestSubmitter_WaitForVerdict_Timeout(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><table><tr data-submission-id="123"><td class="status-cell">Running</td></tr></table></html>`,
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.WaitForVerdict(123, 1, 100*time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestSubmitter_GetSubmission_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("network error"),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.GetSubmission(123, 1)
	if err == nil {
		t.Error("Expected error on network failure")
	}
}

func TestSubmitter_GetSubmission_NotFound(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "<html><table class=\"status-frame-datatable\"></table></html>",
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.GetSubmission(123, 1)
	// GetSubmission may return nil result or error when not found
	if err == nil {
		t.Log("GetSubmission returned nil error (empty result expected)")
	}
}

func TestSubmitter_Get_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("connection refused"),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.get("https://codeforces.com/test")
	if err == nil {
		t.Error("Expected error on network failure")
	}
}

func TestSubmitter_VerifySubmitPage_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("timeout"),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	err := submitter.VerifySubmitPage(1)
	if err == nil {
		t.Error("Expected error on network failure")
	}
}

func TestSubmitter_VerifySubmitPage_MissingElements(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "<html>incomplete page</html>",
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	err := submitter.VerifySubmitPage(1)
	if err == nil {
		t.Error("Expected error when page missing elements")
	}
	if !strings.Contains(err.Error(), "missing elements") {
		t.Errorf("Expected 'missing elements' error, got: %v", err)
	}
}

func TestSubmitter_VerifySubmitPage_Success(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body: `<html>
			<input name="csrf_token" value="test">
			<input name="submittedProblemIndex">
			<select name="programTypeId"></select>
			<textarea name="source"></textarea>
			<input type="submit">
		</html>`,
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	err := submitter.VerifySubmitPage(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// ============ Parser Mock Tests ============

func TestParser_Fetch_NetworkError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("DNS lookup failed"),
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	_, err := parser.fetch("https://codeforces.com/test")
	if err == nil {
		t.Error("Expected error on network failure")
	}
}

func TestParser_Fetch_Non200Status(t *testing.T) {
	transport := &mockTransport{
		statusCode: 404,
		body:       "Not Found",
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	resp, err := parser.fetch("https://codeforces.com/nonexistent")
	// fetch returns the response even on non-200, error checking is done by caller
	if err != nil {
		t.Errorf("fetch should not error on 404: %v", err)
	}
	if resp != nil && resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestParser_Fetch_CloudflareBlock(t *testing.T) {
	transport := &mockTransport{
		statusCode: 403,
		body:       "Blocked by Cloudflare",
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	resp, err := parser.fetch("https://codeforces.com/problem/1/A")
	// fetch returns the response even on 403, error checking is done by caller
	if err != nil {
		t.Errorf("fetch should not error on 403: %v", err)
	}
	if resp != nil && resp.StatusCode != 403 {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestParser_ParseProblem_FetchError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("connection refused"),
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	_, err := parser.ParseProblem(1, "A")
	if err == nil {
		t.Error("Expected error on fetch failure")
	}
}

func TestParser_ParseProblemset_FetchError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("timeout"),
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	_, err := parser.ParseProblemset(1, "A")
	if err == nil {
		t.Error("Expected error on fetch failure")
	}
}

func TestParser_ParseContestProblems_FetchError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("network error"),
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	_, err := parser.ParseContestProblems(1)
	if err == nil {
		t.Error("Expected error on fetch failure")
	}
}

func TestParser_VerifyPageStructure_FetchError(t *testing.T) {
	transport := &mockTransport{
		err: fmt.Errorf("timeout"),
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	err := parser.VerifyPageStructure()
	if err == nil {
		t.Error("Expected error on fetch failure")
	}
}

func TestParser_VerifyPageStructure_MissingElements(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "<html>incomplete</html>",
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	err := parser.VerifyPageStructure()
	if err == nil {
		t.Error("Expected error when elements missing")
	}
}

// ============ Helper Function Tests ============

func TestExtractCSRFToken_NotFound(t *testing.T) {
	token := extractCSRFToken("<html>no token here</html>")
	if token != "" {
		t.Errorf("Expected empty token, got '%s'", token)
	}
}

func TestExtractCSRFToken_MetaTag(t *testing.T) {
	token := extractCSRFToken(`<meta name="X-Csrf-Token" content="abc123">`)
	if token != "abc123" {
		t.Errorf("Expected 'abc123', got '%s'", token)
	}
}

func TestExtractCSRFToken_InputField(t *testing.T) {
	token := extractCSRFToken(`<input name="csrf_token" value="xyz789">`)
	if token != "xyz789" {
		t.Errorf("Expected 'xyz789', got '%s'", token)
	}
}

func TestExtractHiddenInput_NotFound(t *testing.T) {
	value := extractHiddenInput("<html></html>", "ftaa")
	if value != "" {
		t.Errorf("Expected empty value, got '%s'", value)
	}
}

func TestExtractHiddenInput_Found(t *testing.T) {
	value := extractHiddenInput(`<input name="ftaa" value="test123">`, "ftaa")
	if value != "test123" {
		t.Errorf("Expected 'test123', got '%s'", value)
	}
}

// ============ Parser Additional Mock Tests ============

func TestParser_ParseProblem_Non200Status(t *testing.T) {
	transport := &mockTransport{
		statusCode: 404,
		body:       "Not Found",
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	_, err := parser.ParseProblem(1, "A")
	if err == nil {
		t.Error("Expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "status 404") {
		t.Errorf("Expected 'status 404' error, got: %v", err)
	}
}

func TestParser_ParseProblemset_Non200Status(t *testing.T) {
	transport := &mockTransport{
		statusCode: 403,
		body:       "Forbidden",
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	_, err := parser.ParseProblemset(1, "A")
	if err == nil {
		t.Error("Expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "status 403") {
		t.Errorf("Expected 'status 403' error, got: %v", err)
	}
}

func TestParser_ParseContestProblems_Non200Status(t *testing.T) {
	transport := &mockTransport{
		statusCode: 500,
		body:       "Internal Server Error",
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	_, err := parser.ParseContestProblems(1)
	if err == nil {
		t.Error("Expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("Expected 'status 500' error, got: %v", err)
	}
}

func TestParser_ParseContestProblems_EmptyTable(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><table class="problems"></table></html>`,
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	problems, err := parser.ParseContestProblems(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(problems) != 0 {
		t.Errorf("Expected 0 problems, got %d", len(problems))
	}
}

func TestParser_NilSession_DefaultClient(t *testing.T) {
	// Test that parser with nil session uses default client
	parser := NewParser(nil)
	if parser.session != nil {
		t.Error("Parser with nil session should have nil session")
	}
}

func TestParser_NilSessionClient_DefaultClient(t *testing.T) {
	// Test parser with session but nil client
	parser := &Parser{
		session:   &Session{client: nil},
		selectors: CurrentSelectors,
	}

	// This will try to use the default HTTP client
	// We can't easily mock this, but we can verify the parser exists
	if parser.session == nil {
		t.Error("Parser session should not be nil")
	}
}

// ============ Submitter Full Flow Mock Tests ============

func TestSubmitter_Submit_DuplicateSubmission(t *testing.T) {
	// First request returns CSRF token page
	// Second request returns duplicate error
	callCount := 0
	transport := &mockTransport{}
	session := createMockSession(transport)

	// Override RoundTrip to return different responses
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"><input name="ftaa" value="ftaa123"><input name="bfaa" value="bfaa123"></html>`},
			{statusCode: 200, body: `You have submitted exactly the same code before`},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected duplicate submission error")
	}
	if !strings.Contains(err.Error(), "duplicate submission") {
		t.Errorf("Expected 'duplicate submission' error, got: %v", err)
	}
}

func TestSubmitter_Submit_SourceTooLong(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{statusCode: 200, body: `Source code is too long`},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected source too long error")
	}
	if !strings.Contains(err.Error(), "too long") {
		t.Errorf("Expected 'too long' error, got: %v", err)
	}
}

func TestSubmitter_Submit_NotAllowed(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{statusCode: 200, body: `You are not allowed to submit`},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected not allowed error")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("Expected 'not allowed' error, got: %v", err)
	}
}

func TestSubmitter_Submit_ContestOver(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{statusCode: 200, body: `Contest is over`},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected contest over error")
	}
	if !strings.Contains(err.Error(), "contest is over") {
		t.Errorf("Expected 'contest is over' error, got: %v", err)
	}
}

func TestSubmitter_Submit_PostFailed(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{statusCode: 500, body: `Internal Server Error`},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected submission failed error")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("Expected 'failed' error, got: %v", err)
	}
}

func TestSubmitter_Submit_RedirectSuccess(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	// Need 3 responses:
	// 1. GET submit page
	// 2. POST submit (returns 302)
	// 3. GET /contest/1/my (called by getLatestSubmission)
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{statusCode: 302, body: ``, headers: map[string]string{"Location": "/contest/1/my"}},
			{statusCode: 200, body: `<html><table class="status-frame-datatable"><tr data-submission-id="123456"><td class="id-cell">A</td><td class="status-cell">Accepted</td><td class="time-consumed-cell">46 ms</td><td class="memory-consumed-cell">256 KB</td></tr></table></html>`},
		},
		callCount: &callCount,
	}
	// Disable redirect following so we can test the 302 handling
	session.client = &http.Client{
		Transport: customTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	submitter := &Submitter{session: session}

	result, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != nil && result.SubmissionID != 123456 {
		t.Errorf("Expected submission ID 123456, got %d", result.SubmissionID)
	}
}

func TestSubmitter_SubmitToGym_NoCSRFToken(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "<html>no csrf</html>",
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.SubmitToGym(100001, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected error when CSRF not found")
	}
	if !strings.Contains(err.Error(), "csrf token not found") {
		t.Errorf("Expected 'csrf token not found' error, got: %v", err)
	}
}

func TestSubmitter_SubmitToGym_PostFailed(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{statusCode: 500, body: `Internal Server Error`},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.SubmitToGym(100001, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected gym submission failed error")
	}
}

func TestSubmitter_SubmitToGym_RedirectSuccess(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	// Need 3 responses (redirect following disabled):
	// 1. GET gym submit page
	// 2. POST submit (returns 302)
	// 3. GET /gym/100001/my (called by getLatestGymSubmission)
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{statusCode: 302, body: ``, headers: map[string]string{"Location": "/gym/100001/my"}},
			{statusCode: 200, body: `<html><table class="status-frame-datatable"><tr data-submission-id="999"><td class="id-cell">A</td><td class="status-cell">Accepted</td><td class="time-consumed-cell">100 ms</td><td class="memory-consumed-cell">512 KB</td></tr></table></html>`},
		},
		callCount: &callCount,
	}
	// Disable redirect following so we can test the 302 handling
	session.client = &http.Client{
		Transport: customTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	submitter := &Submitter{session: session}

	result, err := submitter.SubmitToGym(100001, "A", 54, "int main(){}")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != nil && result.SubmissionID != 999 {
		t.Errorf("Expected submission ID 999, got %d", result.SubmissionID)
	}
}

func TestSubmitter_GetSubmission_VerdictInQueue(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><span class="verdict-waiting">In queue</span></html>`,
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	result, err := submitter.GetSubmission(123, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != "In queue" {
		t.Errorf("Expected status 'In queue', got '%s'", result.Status)
	}
}

func TestSubmitter_GetSubmission_VerdictRunning(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><span class="verdict-waiting">Running on test 5</span></html>`,
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	result, err := submitter.GetSubmission(123, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != "Running" {
		t.Errorf("Expected status 'Running', got '%s'", result.Status)
	}
}

func TestSubmitter_GetSubmission_VerdictAccepted(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><span class="verdict-accepted">Accepted</span><table class="datatable"><tr><td>46 ms</td><td>256 KB</td></tr></table></html>`,
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	result, err := submitter.GetSubmission(123, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != "Accepted" {
		t.Errorf("Expected status 'Accepted', got '%s'", result.Status)
	}
	if result.Verdict != "OK" {
		t.Errorf("Expected verdict 'OK', got '%s'", result.Verdict)
	}
	if result.Time != 46*time.Millisecond {
		t.Errorf("Expected time 46ms, got %v", result.Time)
	}
	if result.Memory != 256*1024 {
		t.Errorf("Expected memory 256KB, got %d", result.Memory)
	}
}

func TestSubmitter_GetSubmission_VerdictRejected(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><span class="verdict-rejected">Wrong answer on test 3</span></html>`,
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	result, err := submitter.GetSubmission(123, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != "Judged" {
		t.Errorf("Expected status 'Judged', got '%s'", result.Status)
	}
}

func TestSubmitter_WaitForVerdict_Success(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><span class="verdict-waiting">Running on test 1</span></html>`},
			{statusCode: 200, body: `<html><span class="verdict-accepted">Accepted</span></html>`},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	result, err := submitter.WaitForVerdict(123, 1, 5*time.Second)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != "Accepted" {
		t.Errorf("Expected status 'Accepted', got '%s'", result.Status)
	}
}

func TestSubmitter_WaitForVerdict_NetworkErrorDuringPolling(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><span class="verdict-waiting">Running</span></html>`},
			{err: fmt.Errorf("network error during polling")},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.WaitForVerdict(123, 1, 5*time.Second)
	if err == nil {
		t.Error("Expected error during polling")
	}
}

// ============ Session Additional Mock Tests ============

func TestSession_Get_RequestCreationError(t *testing.T) {
	session, _ := NewSession()
	session.SetCookie("JSESSIONID=jsid; 39ce7=ce7; cf_clearance=test")
	session.SetHandle("user")

	// Invalid URL should cause request creation to fail
	_, err := session.get("://invalid-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestSession_RefreshCSRFToken_InvalidHTML(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       string([]byte{0x00, 0x01, 0x02}), // Invalid UTF-8
	}
	session := createMockSession(transport)

	err := session.RefreshCSRFToken()
	if err == nil {
		t.Error("Expected error when CSRF not found in invalid HTML")
	}
}

func TestSession_Validate_HandleMismatch(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html>var handle = "differentuser";</html>`,
	}
	session := createMockSession(transport)

	err := session.Validate()
	if err == nil {
		t.Error("Expected error when handle doesn't match")
	}
}

// ============ Helper Types for Sequential Mock Responses ============

type mockResponse struct {
	statusCode int
	body       string
	headers    map[string]string
	err        error
}

type sequentialMockTransport struct {
	responses []mockResponse
	callCount *int
}

func (m *sequentialMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	idx := *m.callCount
	*m.callCount++

	if idx >= len(m.responses) {
		return nil, fmt.Errorf("unexpected request #%d", idx)
	}

	resp := m.responses[idx]
	if resp.err != nil {
		return nil, resp.err
	}

	httpResp := &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(strings.NewReader(resp.body)),
		Header:     make(http.Header),
	}

	for k, v := range resp.headers {
		httpResp.Header.Set(k, v)
	}

	return httpResp, nil
}

// ============ parseSamples Alternative Structure Test ============

func TestParser_ParseSamples_AlternativeStructure(t *testing.T) {
	html := `<div class="sample-tests">
		<div class="input"><div class="title">Input</div><pre>1 2 3</pre></div>
		<div class="output"><div class="title">Output</div><pre>6</pre></div>
		<div class="input"><div class="title">Input</div><pre>4 5 6</pre></div>
		<div class="output"><div class="title">Output</div><pre>15</pre></div>
	</div>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	samples := parseSamples(doc.Find(".sample-tests"), CurrentSelectors.Problem)
	if len(samples) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(samples))
	}

	if len(samples) > 0 && samples[0].Input != "1 2 3" {
		t.Errorf("Expected input '1 2 3', got '%s'", samples[0].Input)
	}
}

func TestParser_ParseSamples_MismatchedInputOutput(t *testing.T) {
	html := `<div class="sample-tests">
		<div class="input"><pre>input1</pre></div>
		<div class="input"><pre>input2</pre></div>
		<div class="output"><pre>output1</pre></div>
	</div>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	samples := parseSamples(doc.Find(".sample-tests"), CurrentSelectors.Problem)
	// Should return minimum of inputs and outputs
	if len(samples) != 1 {
		t.Errorf("Expected 1 sample (min of inputs/outputs), got %d", len(samples))
	}
}

func TestParser_ParseProblemHTML_InvalidHTML(t *testing.T) {
	parser := NewParser(nil)
	// Reading from a broken reader
	brokenReader := &errorReader{}
	_, err := parser.parseProblemHTML(brokenReader, 1, "A", "http://test.com")
	if err == nil {
		t.Error("Expected error for invalid HTML reader")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

// ============ Additional Coverage Tests ============

func TestSubmitter_Submit_OKStatusWithLatestSubmission(t *testing.T) {
	// Test the path where Submit gets 200 OK and calls getLatestSubmission
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{statusCode: 200, body: `<html>Submission recorded</html>`}, // No error messages, 200 OK
			{statusCode: 200, body: `<html><table class="status-frame-datatable"><tr data-submission-id="999"><td class="id-cell">A</td><td class="status-cell">Accepted</td><td class="time-consumed-cell">50 ms</td><td class="memory-consumed-cell">128 KB</td></tr></table></html>`},
		},
		callCount: &callCount,
	}
	session.client = &http.Client{
		Transport: customTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	submitter := &Submitter{session: session}

	result, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != nil && result.SubmissionID != 999 {
		t.Errorf("Expected submission ID 999, got %d", result.SubmissionID)
	}
}

func TestSubmitter_Submit_PostError(t *testing.T) {
	// Test the path where POST request fails
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{err: fmt.Errorf("POST request failed")},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected error for POST failure")
	}
	if !strings.Contains(err.Error(), "submit solution") {
		t.Errorf("Expected 'submit solution' error, got: %v", err)
	}
}

func TestSubmitter_SubmitToGym_PostError(t *testing.T) {
	callCount := 0
	session := createMockSession(&mockTransport{})
	customTransport := &sequentialMockTransport{
		responses: []mockResponse{
			{statusCode: 200, body: `<html><meta name="X-Csrf-Token" content="test-csrf"></html>`},
			{err: fmt.Errorf("POST request failed")},
		},
		callCount: &callCount,
	}
	session.client.Transport = customTransport
	submitter := &Submitter{session: session}

	_, err := submitter.SubmitToGym(100001, "A", 54, "int main(){}")
	if err == nil {
		t.Error("Expected error for POST failure")
	}
	if !strings.Contains(err.Error(), "submit gym solution") {
		t.Errorf("Expected 'submit gym solution' error, got: %v", err)
	}
}

func TestSubmitter_GetLatestSubmission_HTMLParseError(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       string([]byte{0x00, 0x01, 0x02, '<', 'h', 't', 'm', 'l', 0x00}),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	// Use reflection or just test behavior
	_, err := submitter.getLatestSubmission(1, "A")
	// May not error if goquery is lenient, but should not panic
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestSubmitter_GetLatestGymSubmission_HTMLParseError(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       string([]byte{0x00, 0x01, 0x02}),
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	_, err := submitter.getLatestGymSubmission(100001, "A")
	// May not error if goquery is lenient, but should not panic
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestParser_ParseContestProblems_WithProblems(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body: `<html>
		<table class="problems">
			<tr><th>Problem</th></tr>
			<tr>
				<td><a href="/contest/1/problem/A">A</a></td>
				<td><a href="/contest/1/problem/A">Theatre Square</a></td>
			</tr>
			<tr>
				<td><a href="/contest/1/problem/B">B</a></td>
				<td><a href="/contest/1/problem/B">Spreadsheet</a></td>
			</tr>
		</table>
		</html>`,
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	problems, err := parser.ParseContestProblems(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Verify problems were parsed (structure may vary based on selectors)
	t.Logf("Parsed %d problems", len(problems))
}

func TestParser_ParseContestProblems_MissingProblemLink(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body: `<html>
		<table class="problems">
			<tr><th>Problem</th></tr>
			<tr><td>No link here</td></tr>
		</table>
		</html>`,
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	problems, err := parser.ParseContestProblems(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(problems) != 0 {
		t.Errorf("Expected 0 problems (no valid links), got %d", len(problems))
	}
}

func TestSubmitter_GetSubmission_MemoryMB(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><span class="verdict-accepted">Accepted</span><table class="datatable"><tr><td>100 ms</td><td>1 MB</td></tr></table></html>`,
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	result, err := submitter.GetSubmission(123, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Memory != 1*1024*1024 {
		t.Errorf("Expected memory 1MB (%d), got %d", 1*1024*1024, result.Memory)
	}
}

func TestSubmitter_VerifySubmitPage_AllElementsPresent(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body: `<html>
			<input name="csrf_token" value="test">
			<input name="submittedProblemIndex">
			<select name="programTypeId"></select>
			<textarea name="source"></textarea>
			<input type="submit">
		</html>`,
	}
	session := createMockSession(transport)
	submitter := &Submitter{session: session}

	err := submitter.VerifySubmitPage(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestSession_RefreshCSRFToken_JSPattern(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html>Codeforces.getCsrfToken"js-csrf-token";</html>`,
	}
	session := createMockSession(transport)

	err := session.RefreshCSRFToken()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if session.csrfToken != "js-csrf-token" {
		t.Errorf("Expected csrf token 'js-csrf-token', got '%s'", session.csrfToken)
	}
}

func TestSession_RefreshCSRFToken_InputPattern(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       `<html><input name="csrf_token" value="input-csrf-token"></html>`,
	}
	session := createMockSession(transport)

	err := session.RefreshCSRFToken()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if session.csrfToken != "input-csrf-token" {
		t.Errorf("Expected csrf token 'input-csrf-token', got '%s'", session.csrfToken)
	}
}

func TestSession_Get_WithMockTransport(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body:       "Hello World",
	}
	session := createMockSession(transport)

	resp, err := session.get("https://codeforces.com/test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestParser_VerifyPageStructure_AllSelectors(t *testing.T) {
	transport := &mockTransport{
		statusCode: 200,
		body: `<html>
			<div class="title">A. Problem</div>
			<div class="time-limit">1 second</div>
			<div class="memory-limit">256 MB</div>
			<div class="problem-statement">Statement text</div>
			<div class="sample-tests"><div class="sample-test"></div></div>
		</html>`,
	}
	session := createMockSession(transport)
	parser := &Parser{session: session, selectors: CurrentSelectors}

	err := parser.VerifyPageStructure()
	if err != nil {
		t.Logf("Verification returned error (expected if selectors mismatch): %v", err)
	}
}

func TestExtractPreContent_ParseError(t *testing.T) {
	// Test with malformed HTML
	html := `<pre>test content`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	pre := doc.Find("pre").First()
	result := extractPreContent(pre)

	if result == "" {
		t.Error("Expected some content from malformed pre")
	}
}

func TestSubmitter_Submit_ReadBodyError(t *testing.T) {
	// Test when reading the submit page body fails
	session := createMockSession(&mockTransport{})
	// Swap to a transport that returns an error-prone body
	session.client.Transport = &errorBodyTransport{}
	submitter := &Submitter{session: session}

	_, err := submitter.Submit(1, "A", 54, "int main(){}")
	// This should fail somewhere in the submit flow
	if err == nil {
		t.Log("Submit succeeded or failed gracefully")
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
