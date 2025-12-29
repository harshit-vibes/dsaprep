package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCredentials_HasHandle(t *testing.T) {
	tests := []struct {
		name  string
		creds Credentials
		want  bool
	}{
		{
			name:  "with handle",
			creds: Credentials{CFHandle: "tourist"},
			want:  true,
		},
		{
			name:  "empty handle",
			creds: Credentials{CFHandle: ""},
			want:  false,
		},
		{
			name:  "whitespace handle treated as non-empty",
			creds: Credentials{CFHandle: "  "},
			want:  true, // Implementation doesn't trim whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.creds.HasHandle(); got != tt.want {
				t.Errorf("Credentials.HasHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentials_IsAPIConfigured(t *testing.T) {
	tests := []struct {
		name  string
		creds Credentials
		want  bool
	}{
		{
			name: "fully configured",
			creds: Credentials{
				APIKey:    "key123",
				APISecret: "secret456",
			},
			want: true,
		},
		{
			name: "missing key",
			creds: Credentials{
				APIKey:    "",
				APISecret: "secret456",
			},
			want: false,
		},
		{
			name: "missing secret",
			creds: Credentials{
				APIKey:    "key123",
				APISecret: "",
			},
			want: false,
		},
		{
			name:  "both missing",
			creds: Credentials{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.creds.IsAPIConfigured(); got != tt.want {
				t.Errorf("Credentials.IsAPIConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentials_HasSessionCookies(t *testing.T) {
	tests := []struct {
		name  string
		creds Credentials
		want  bool
	}{
		{
			name: "has JSESSIONID",
			creds: Credentials{
				JSESSIONID: "session123",
			},
			want: true,
		},
		{
			name: "has CE7Cookie",
			creds: Credentials{
				CE7Cookie: "ce7value",
			},
			want: true,
		},
		{
			name: "has both",
			creds: Credentials{
				JSESSIONID: "session123",
				CE7Cookie:  "ce7value",
			},
			want: true,
		},
		{
			name:  "has neither",
			creds: Credentials{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.creds.HasSessionCookies(); got != tt.want {
				t.Errorf("Credentials.HasSessionCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentials_IsReadyForSubmission(t *testing.T) {
	tests := []struct {
		name  string
		creds Credentials
		want  bool
	}{
		{
			name: "fully configured for submission",
			creds: Credentials{
				CFHandle:           "tourist",
				JSESSIONID:         "session123",
				CE7Cookie:          "ce7value",
				CFClearance:        "clearance",
				CFClearanceExpires: 9999999999,
			},
			want: true,
		},
		{
			name: "missing handle",
			creds: Credentials{
				JSESSIONID:         "session123",
				CFClearance:        "clearance",
				CFClearanceExpires: 9999999999,
			},
			want: false,
		},
		{
			name: "missing session cookies",
			creds: Credentials{
				CFHandle:           "tourist",
				CFClearance:        "clearance",
				CFClearanceExpires: 9999999999,
			},
			want: false,
		},
		{
			name: "expired cf_clearance",
			creds: Credentials{
				CFHandle:           "tourist",
				JSESSIONID:         "session123",
				CFClearance:        "clearance",
				CFClearanceExpires: 1, // Expired
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.creds.IsReadyForSubmission(); got != tt.want {
				t.Errorf("Credentials.IsReadyForSubmission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureEnvFile(t *testing.T) {
	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Ensure env file doesn't exist
	envPath := filepath.Join(tmpDir, ".cf.env")
	os.Remove(envPath)

	// Create env file
	err := EnsureEnvFile()
	if err != nil {
		t.Fatalf("EnsureEnvFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Error("EnsureEnvFile() did not create the file")
	}

	// Verify file contents
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("Failed to read env file: %v", err)
	}

	// Check for expected content
	if !strings.Contains(string(content), "CF_HANDLE") {
		t.Error("Env file should contain CF_HANDLE")
	}
	if !strings.Contains(string(content), "CF_API_KEY") {
		t.Error("Env file should contain CF_API_KEY")
	}
	if !strings.Contains(string(content), "CF_CLEARANCE") {
		t.Error("Env file should contain CF_CLEARANCE")
	}
	if !strings.Contains(string(content), "CF_JSESSIONID") {
		t.Error("Env file should contain CF_JSESSIONID")
	}
}

func TestLoadCredentials(t *testing.T) {
	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with test values
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `CF_HANDLE=testuser
CF_API_KEY=testkey
CF_API_SECRET=testsecret
CF_JSESSIONID=session123
CF_39CE7=ce7value
CF_CLEARANCE=clearance_value
CF_CLEARANCE_EXPIRES=9999999999
CF_CLEARANCE_UA=TestUserAgent
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	// Load credentials
	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	// Verify values
	if creds.CFHandle != "testuser" {
		t.Errorf("CFHandle = %v, want %v", creds.CFHandle, "testuser")
	}
	if creds.APIKey != "testkey" {
		t.Errorf("APIKey = %v, want %v", creds.APIKey, "testkey")
	}
	if creds.APISecret != "testsecret" {
		t.Errorf("APISecret = %v, want %v", creds.APISecret, "testsecret")
	}
	if creds.JSESSIONID != "session123" {
		t.Errorf("JSESSIONID = %v, want %v", creds.JSESSIONID, "session123")
	}
	if creds.CE7Cookie != "ce7value" {
		t.Errorf("CE7Cookie = %v, want %v", creds.CE7Cookie, "ce7value")
	}
	if creds.CFClearance != "clearance_value" {
		t.Errorf("CFClearance = %v, want %v", creds.CFClearance, "clearance_value")
	}
	if creds.CFClearanceUA != "TestUserAgent" {
		t.Errorf("CFClearanceUA = %v, want %v", creds.CFClearanceUA, "TestUserAgent")
	}
}

func TestLoadCredentials_FileNotFound(t *testing.T) {
	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home (no env file)
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Load credentials - should return empty struct, not error
	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	if creds == nil {
		t.Error("LoadCredentials() should return empty credentials, not nil")
	}
}

func TestGetEnvFilePath(t *testing.T) {
	path, err := GetEnvFilePath()
	if err != nil {
		t.Fatalf("GetEnvFilePath() error = %v", err)
	}

	if path == "" {
		t.Error("GetEnvFilePath() returned empty path")
	}

	if !strings.HasSuffix(path, ".cf.env") {
		t.Errorf("GetEnvFilePath() = %v, should end with .cf.env", path)
	}
}

func TestSaveCredentials(t *testing.T) {
	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create credentials to save
	creds := &Credentials{
		CFHandle:           "testuser",
		APIKey:             "myapikey",
		APISecret:          "myapisecret",
		JSESSIONID:         "session123",
		CE7Cookie:          "ce7value",
		CFClearance:        "clearance_value",
		CFClearanceExpires: 9999999999,
		CFClearanceUA:      "TestUserAgent",
	}

	// Save credentials
	err := SaveCredentials(creds)
	if err != nil {
		t.Fatalf("SaveCredentials() error = %v", err)
	}

	// Verify file was created
	envPath := filepath.Join(tmpDir, ".cf.env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatal("SaveCredentials() did not create the file")
	}

	// Load and verify credentials
	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	if loaded.CFHandle != creds.CFHandle {
		t.Errorf("CFHandle = %v, want %v", loaded.CFHandle, creds.CFHandle)
	}
	if loaded.APIKey != creds.APIKey {
		t.Errorf("APIKey = %v, want %v", loaded.APIKey, creds.APIKey)
	}
	if loaded.APISecret != creds.APISecret {
		t.Errorf("APISecret = %v, want %v", loaded.APISecret, creds.APISecret)
	}
	if loaded.JSESSIONID != creds.JSESSIONID {
		t.Errorf("JSESSIONID = %v, want %v", loaded.JSESSIONID, creds.JSESSIONID)
	}
	if loaded.CE7Cookie != creds.CE7Cookie {
		t.Errorf("CE7Cookie = %v, want %v", loaded.CE7Cookie, creds.CE7Cookie)
	}
}

func TestCredentials_IsCFClearanceValid(t *testing.T) {
	tests := []struct {
		name  string
		creds Credentials
		want  bool
	}{
		{
			name: "valid clearance - future expiry",
			creds: Credentials{
				CFClearance:        "cf_clearance_value",
				CFClearanceExpires: 9999999999, // Far future
			},
			want: true,
		},
		{
			name: "empty clearance",
			creds: Credentials{
				CFClearance:        "",
				CFClearanceExpires: 9999999999,
			},
			want: false,
		},
		{
			name: "expired clearance",
			creds: Credentials{
				CFClearance:        "cf_clearance_value",
				CFClearanceExpires: 1, // Way in the past
			},
			want: false,
		},
		{
			name: "zero expiry",
			creds: Credentials{
				CFClearance:        "cf_clearance_value",
				CFClearanceExpires: 0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.creds.IsCFClearanceValid(); got != tt.want {
				t.Errorf("Credentials.IsCFClearanceValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentials_CFClearanceExpiresIn(t *testing.T) {
	// Test zero expiry
	creds := Credentials{CFClearanceExpires: 0}
	if creds.CFClearanceExpiresIn() != 0 {
		t.Error("CFClearanceExpiresIn() should return 0 for zero expiry")
	}

	// Test future expiry
	futureTime := int64(9999999999)
	creds = Credentials{CFClearanceExpires: futureTime}
	expiresIn := creds.CFClearanceExpiresIn()
	if expiresIn <= 0 {
		t.Error("CFClearanceExpiresIn() should return positive duration for future expiry")
	}

	// Test past expiry
	pastTime := int64(1)
	creds = Credentials{CFClearanceExpires: pastTime}
	expiresIn = creds.CFClearanceExpiresIn()
	if expiresIn >= 0 {
		t.Error("CFClearanceExpiresIn() should return negative duration for past expiry")
	}
}

func TestCredentials_SetCFClearance(t *testing.T) {
	creds := &Credentials{}

	clearance := "test_cf_clearance"
	userAgent := "Test User Agent"
	expiresAt := time.Now().Add(30 * time.Minute)

	creds.SetCFClearance(clearance, userAgent, expiresAt)

	if creds.CFClearance != clearance {
		t.Errorf("CFClearance = %v, want %v", creds.CFClearance, clearance)
	}
	if creds.CFClearanceUA != userAgent {
		t.Errorf("CFClearanceUA = %v, want %v", creds.CFClearanceUA, userAgent)
	}
	if creds.CFClearanceExpires != expiresAt.Unix() {
		t.Errorf("CFClearanceExpires = %v, want %v", creds.CFClearanceExpires, expiresAt.Unix())
	}
}

func TestCredentials_GetCFClearanceStatus(t *testing.T) {
	tests := []struct {
		name     string
		creds    Credentials
		contains string
	}{
		{
			name:     "not configured",
			creds:    Credentials{},
			contains: "not configured",
		},
		{
			name: "expired",
			creds: Credentials{
				CFClearance:        "test",
				CFClearanceExpires: 1, // Way in past
			},
			contains: "expired",
		},
		{
			name: "valid",
			creds: Credentials{
				CFClearance:        "test",
				CFClearanceExpires: 9999999999, // Far future
			},
			contains: "valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.creds.GetCFClearanceStatus()
			if !strings.Contains(status, tt.contains) {
				t.Errorf("GetCFClearanceStatus() = %v, should contain %v", status, tt.contains)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "negative duration",
			d:    -1 * time.Hour,
			want: "expired",
		},
		{
			name: "seconds",
			d:    30 * time.Second,
			want: "30s",
		},
		{
			name: "minutes",
			d:    5 * time.Minute,
			want: "5m",
		},
		{
			name: "hours and minutes",
			d:    2*time.Hour + 30*time.Minute,
			want: "2h 30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveCredentials_WithCFClearance(t *testing.T) {
	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create credentials with cf_clearance
	creds := &Credentials{
		CFHandle:           "testuser",
		CFClearance:        "test_clearance_value",
		CFClearanceExpires: 9999999999,
		CFClearanceUA:      "Test User Agent",
	}

	// Save credentials
	err := SaveCredentials(creds)
	if err != nil {
		t.Fatalf("SaveCredentials() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	if loaded.CFClearance != creds.CFClearance {
		t.Errorf("CFClearance = %v, want %v", loaded.CFClearance, creds.CFClearance)
	}
	if loaded.CFClearanceExpires != creds.CFClearanceExpires {
		t.Errorf("CFClearanceExpires = %v, want %v", loaded.CFClearanceExpires, creds.CFClearanceExpires)
	}
	if loaded.CFClearanceUA != creds.CFClearanceUA {
		t.Errorf("CFClearanceUA = %v, want %v", loaded.CFClearanceUA, creds.CFClearanceUA)
	}
}

func TestLoadCredentials_WithComments(t *testing.T) {
	// Save original home and restore after test
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Use temp directory as home
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create env file with comments and blank lines
	envPath := filepath.Join(tmpDir, ".cf.env")
	envContent := `# This is a comment
CF_HANDLE=testuser

# Another comment
CF_API_KEY=testkey
CF_API_SECRET=testsecret
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	// Load credentials
	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	// Verify comments and blank lines were skipped
	if creds.CFHandle != "testuser" {
		t.Errorf("CFHandle = %v, want %v", creds.CFHandle, "testuser")
	}
	if creds.APIKey != "testkey" {
		t.Errorf("APIKey = %v, want %v", creds.APIKey, "testkey")
	}
}
