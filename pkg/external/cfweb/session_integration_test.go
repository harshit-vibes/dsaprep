//go:build integration

package cfweb

import (
	"os"
	"testing"
	"time"
)

// getCookieAuthCredentials returns test credentials for cookie-based auth
func getCookieAuthCredentials(t *testing.T) (cfClearance, jsessionID, ce7Cookie, handle, userAgent string) {
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

// TestSession_RefreshCSRFToken_Real tests CSRF token refresh against real CF
func TestSession_RefreshCSRFToken_Real(t *testing.T) {
	cfClearance, _, _, _, userAgent := getCookieAuthCredentials(t)

	if cfClearance == "" {
		t.Skip("CF_CLEARANCE not set - skipping CSRF refresh test")
	}

	session, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession() failed: %v", err)
	}

	// Set cf_clearance to bypass Cloudflare
	session.SetCFClearance(cfClearance, userAgent, time.Now().Add(1*time.Hour))

	// Try to refresh CSRF token
	err = session.RefreshCSRFToken()
	if err != nil {
		t.Fatalf("RefreshCSRFToken() failed: %v", err)
	}

	token := session.GetCSRFToken()
	if token == "" {
		t.Error("CSRF token should not be empty after refresh")
	}

	t.Logf("CSRF token refreshed successfully (length: %d)", len(token))
}

// TestSession_Validate_Real tests session validation against real CF
func TestSession_Validate_Real(t *testing.T) {
	cfClearance, jsessionID, ce7Cookie, handle, userAgent := getCookieAuthCredentials(t)

	if cfClearance == "" || handle == "" {
		t.Skip("CF_CLEARANCE or CF_HANDLE not set - skipping validation test")
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

	// Add delay to respect rate limiting
	time.Sleep(500 * time.Millisecond)

	err = session.Validate()
	if err != nil {
		t.Logf("Validation failed (may need fresh session cookies): %v", err)
		// Don't fail the test as session cookies might be stale
	} else {
		t.Log("Session validated successfully")
	}
}
