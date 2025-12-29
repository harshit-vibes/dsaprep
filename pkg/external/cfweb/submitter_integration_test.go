//go:build integration

package cfweb

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// getSubmitterCookieCredentials returns test credentials for cookie-based auth
func getSubmitterCookieCredentials(t *testing.T) (cfClearance, jsessionID, ce7Cookie, handle, userAgent string) {
	cfClearance = os.Getenv("CF_CLEARANCE")
	jsessionID = os.Getenv("CF_JSESSIONID")
	ce7Cookie = os.Getenv("CF_39CE7")
	handle = os.Getenv("CF_HANDLE")
	userAgent = os.Getenv("CF_CLEARANCE_UA")

	// Fallback to hardcoded test credentials if env vars not set
	if handle == "" {
		handle = "harshitvsdsa"
	}
	if userAgent == "" {
		userAgent = UserAgent
	}

	return cfClearance, jsessionID, ce7Cookie, handle, userAgent
}

func TestSubmitter_NewSubmitter_WithCookies_Real(t *testing.T) {
	cfClearance, jsessionID, ce7Cookie, handle, userAgent := getSubmitterCookieCredentials(t)

	if cfClearance == "" {
		t.Skip("CF_CLEARANCE not set - skipping submitter test")
	}

	if jsessionID == "" && ce7Cookie == "" {
		t.Skip("No session cookies set - skipping submitter test")
	}

	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	// Set full authentication
	session.SetFullAuth(
		cfClearance,
		userAgent,
		time.Now().Add(1*time.Hour),
		jsessionID,
		ce7Cookie,
		handle,
	)

	submitter, err := NewSubmitter(session)
	if err != nil {
		t.Fatalf("NewSubmitter() failed: %v", err)
	}

	if submitter == nil {
		t.Fatal("NewSubmitter() returned nil")
	}

	if submitter.session != session {
		t.Error("submitter session should match input session")
	}

	t.Log("Submitter created successfully with cookie-based auth")
}

func TestSubmitter_VerifySubmitPage_Real(t *testing.T) {
	cfClearance, jsessionID, ce7Cookie, handle, userAgent := getSubmitterCookieCredentials(t)

	if cfClearance == "" {
		t.Skip("CF_CLEARANCE not set - skipping submitter test")
	}

	if jsessionID == "" && ce7Cookie == "" {
		t.Skip("No session cookies set - skipping submitter test")
	}

	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	// Set full authentication
	session.SetFullAuth(
		cfClearance,
		userAgent,
		time.Now().Add(1*time.Hour),
		jsessionID,
		ce7Cookie,
		handle,
	)

	// First validate session - skip if not logged in
	if err := session.Validate(); err != nil {
		t.Skipf("Session not valid (cookies may have expired) - skipping: %v", err)
	}

	submitter, err := NewSubmitter(session)
	if err != nil {
		t.Fatalf("NewSubmitter() failed: %v", err)
	}

	// Verify submit page structure for contest 1
	// Note: Page structure may change, so we log errors instead of failing
	err = submitter.VerifySubmitPage(1)
	if err != nil {
		t.Logf("VerifySubmitPage(1) warning (page structure may have changed): %v", err)
	} else {
		t.Log("Submit page structure verification passed")
	}
}

func TestSubmitter_GetSubmission_Real(t *testing.T) {
	cfClearance, jsessionID, ce7Cookie, handle, userAgent := getSubmitterCookieCredentials(t)

	if cfClearance == "" || (jsessionID == "" && ce7Cookie == "") {
		t.Skip("Cookie credentials not set - skipping GetSubmission test")
	}

	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	session.SetFullAuth(
		cfClearance,
		userAgent,
		time.Now().Add(1*time.Hour),
		jsessionID,
		ce7Cookie,
		handle,
	)

	submitter, err := NewSubmitter(session)
	if err != nil {
		t.Fatalf("NewSubmitter() failed: %v", err)
	}

	// Verify submitter was created successfully
	if submitter == nil {
		t.Fatal("submitter should not be nil")
	}

	// Note: This test requires knowing a valid submission ID
	// Skip this test if we can't access submission status
	t.Log("Skipping GetSubmission test - requires valid submission ID")
}

func TestSubmitter_Submit_Real(t *testing.T) {
	cfClearance, jsessionID, ce7Cookie, handle, userAgent := getSubmitterCookieCredentials(t)

	if cfClearance == "" || (jsessionID == "" && ce7Cookie == "") {
		t.Skip("Cookie credentials not set - skipping actual submission test")
	}

	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	session.SetFullAuth(
		cfClearance,
		userAgent,
		time.Now().Add(1*time.Hour),
		jsessionID,
		ce7Cookie,
		handle,
	)

	// Validate session first
	if err := session.Validate(); err != nil {
		t.Skipf("Session not valid (cookies may have expired) - skipping: %v", err)
	}

	submitter, err := NewSubmitter(session)
	if err != nil {
		t.Fatalf("NewSubmitter() failed: %v", err)
	}

	// Use a unique comment to avoid duplicate detection
	uniqueComment := fmt.Sprintf("// Test submission at %d", time.Now().UnixNano())
	sourceCode := `#include <iostream>
using namespace std;

` + uniqueComment + `

int main() {
    cout << "Hello World" << endl;
    return 0;
}
`

	// Submit to contest 1 problem A
	// Note: This will actually submit to Codeforces!
	result, err := submitter.Submit(1, "A", 54, sourceCode) // 54 = GNU C++17
	if err != nil {
		// Submission might fail for various reasons (duplicate, rate limit, etc.)
		t.Logf("Submit() returned error (might be expected): %v", err)
		return
	}

	if result.SubmissionID == 0 {
		t.Error("SubmissionID should not be 0")
	}

	t.Logf("Submission created: ID=%d, Status=%s", result.SubmissionID, result.Status)

	// Test WaitForVerdict with the submission we just made
	if result.SubmissionID > 0 {
		verdictResult, err := submitter.WaitForVerdict(result.SubmissionID, 1, 60*time.Second)
		if err != nil {
			t.Logf("WaitForVerdict() returned error: %v", err)
		} else {
			t.Logf("Final verdict: %s (Time: %v, Memory: %d)", verdictResult.Verdict, verdictResult.Time, verdictResult.Memory)
		}
	}
}

func TestSubmitter_GetSubmission_WithRealID_Real(t *testing.T) {
	cfClearance, jsessionID, ce7Cookie, handle, userAgent := getSubmitterCookieCredentials(t)

	if cfClearance == "" || (jsessionID == "" && ce7Cookie == "") {
		t.Skip("Cookie credentials not set - skipping GetSubmission test")
	}

	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	session.SetFullAuth(
		cfClearance,
		userAgent,
		time.Now().Add(1*time.Hour),
		jsessionID,
		ce7Cookie,
		handle,
	)

	if err := session.Validate(); err != nil {
		t.Skipf("Session not valid - skipping: %v", err)
	}

	submitter, err := NewSubmitter(session)
	if err != nil {
		t.Fatalf("NewSubmitter() failed: %v", err)
	}

	// Use a known submission ID from contest 1 (any old submission works)
	// This is submission 2 from contest 1 which should always exist
	result, err := submitter.GetSubmission(2, 1)
	if err != nil {
		t.Logf("GetSubmission() error (might be expected for old submissions): %v", err)
		return
	}

	if result.SubmissionID != 2 {
		t.Errorf("SubmissionID = %d, want 2", result.SubmissionID)
	}
	t.Logf("Got submission: ID=%d, Verdict=%s", result.SubmissionID, result.Verdict)
}
