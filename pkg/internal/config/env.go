package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const envFileName = ".cf.env"

// Credentials holds CF API and cookie-based web credentials
type Credentials struct {
	// API authentication
	APIKey    string
	APISecret string
	CFHandle  string

	// Session cookies (extracted from browser)
	JSESSIONID string // JSESSIONID cookie for CF session
	CE7Cookie  string // 39ce7 cookie for CF session

	// Cloudflare bypass
	CFClearance        string // cf_clearance cookie value
	CFClearanceExpires int64  // Unix timestamp when cf_clearance expires
	CFClearanceUA      string // User-Agent tied to cf_clearance (MUST match)
}

// GetEnvFilePath returns the path to the .env file
func GetEnvFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, envFileName), nil
}

// EnsureEnvFile creates the .env file if it doesn't exist
func EnsureEnvFile() error {
	envPath, err := GetEnvFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		content := `# cf - Codeforces CLI Configuration
# Get your API key and secret from: https://codeforces.com/settings/api

# === API Authentication (for read operations) ===
CF_HANDLE=
CF_API_KEY=
CF_API_SECRET=

# === Session Cookies (extracted from browser - for submissions) ===
# 1. Open https://codeforces.com and log in
# 2. Open DevTools (F12) > Application > Cookies > codeforces.com
# 3. Copy the values for JSESSIONID and 39ce7 cookies
CF_JSESSIONID=
CF_39CE7=

# === Cloudflare Bypass (for automated access) ===
# Extract from browser: DevTools > Application > Cookies > cf_clearance
# Also copy your User-Agent from browser console: navigator.userAgent
# IMPORTANT: User-Agent MUST match exactly when using cf_clearance!
CF_CLEARANCE=
CF_CLEARANCE_EXPIRES=
CF_CLEARANCE_UA=
`
		if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
			return fmt.Errorf("failed to create env file: %w", err)
		}
	}

	return nil
}

// LoadCredentials loads credentials from the .env file
func LoadCredentials() (*Credentials, error) {
	envPath, err := GetEnvFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
		return nil, err
	}
	defer file.Close()

	creds := &Credentials{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "CF_HANDLE":
			creds.CFHandle = value
		case "CF_API_KEY":
			creds.APIKey = value
		case "CF_API_SECRET":
			creds.APISecret = value
		case "CF_JSESSIONID":
			creds.JSESSIONID = value
		case "CF_39CE7":
			creds.CE7Cookie = value
		case "CF_CLEARANCE":
			creds.CFClearance = value
		case "CF_CLEARANCE_EXPIRES":
			if value != "" {
				_, _ = fmt.Sscanf(value, "%d", &creds.CFClearanceExpires)
			}
		case "CF_CLEARANCE_UA":
			creds.CFClearanceUA = value
		}
	}

	return creds, scanner.Err()
}

// SaveCredentials saves credentials to the .env file
func SaveCredentials(creds *Credentials) error {
	envPath, err := GetEnvFilePath()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`# cf - Codeforces CLI Configuration
# Get your API key and secret from: https://codeforces.com/settings/api

# === API Authentication (for read operations) ===
CF_HANDLE=%s
CF_API_KEY=%s
CF_API_SECRET=%s

# === Session Cookies (extracted from browser - for submissions) ===
CF_JSESSIONID=%s
CF_39CE7=%s

# === Cloudflare Bypass ===
CF_CLEARANCE=%s
CF_CLEARANCE_EXPIRES=%d
CF_CLEARANCE_UA=%s
`, creds.CFHandle, creds.APIKey, creds.APISecret,
		creds.JSESSIONID, creds.CE7Cookie,
		creds.CFClearance, creds.CFClearanceExpires, creds.CFClearanceUA)

	return os.WriteFile(envPath, []byte(content), 0600)
}

// IsAPIConfigured returns true if API credentials are set
func (c *Credentials) IsAPIConfigured() bool {
	return c.APIKey != "" && c.APISecret != ""
}

// HasHandle returns true if CF handle is set
func (c *Credentials) HasHandle() bool {
	return c.CFHandle != ""
}

// HasSessionCookies returns true if session cookies are set
func (c *Credentials) HasSessionCookies() bool {
	return c.JSESSIONID != "" || c.CE7Cookie != ""
}

// IsCFClearanceValid returns true if cf_clearance cookie is set and not expired
func (c *Credentials) IsCFClearanceValid() bool {
	if c.CFClearance == "" || c.CFClearanceExpires == 0 {
		return false
	}
	return time.Now().Unix() < c.CFClearanceExpires
}

// IsReadyForSubmission returns true if we have everything needed to submit
// Requires: cf_clearance (valid) + session cookies + handle
func (c *Credentials) IsReadyForSubmission() bool {
	return c.IsCFClearanceValid() && c.HasSessionCookies() && c.CFHandle != ""
}

// CFClearanceExpiresIn returns time until cf_clearance expires (negative if expired)
func (c *Credentials) CFClearanceExpiresIn() time.Duration {
	if c.CFClearanceExpires == 0 {
		return 0
	}
	return time.Until(time.Unix(c.CFClearanceExpires, 0))
}

// SetCFClearance sets the cf_clearance cookie with expiration and User-Agent
func (c *Credentials) SetCFClearance(clearance, userAgent string, expiresAt time.Time) {
	c.CFClearance = clearance
	c.CFClearanceExpires = expiresAt.Unix()
	c.CFClearanceUA = userAgent
}

// GetCFClearanceStatus returns a human-readable status string
func (c *Credentials) GetCFClearanceStatus() string {
	if c.CFClearance == "" {
		return "not configured"
	}
	if !c.IsCFClearanceValid() {
		return "expired"
	}
	remaining := c.CFClearanceExpiresIn()
	if remaining < 5*time.Minute {
		return fmt.Sprintf("expiring soon (%s)", formatDuration(remaining))
	}
	return fmt.Sprintf("valid (%s remaining)", formatDuration(remaining))
}

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}
