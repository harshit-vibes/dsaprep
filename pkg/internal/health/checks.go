package health

import (
	"context"
	"os"
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/schema"
	"github.com/harshit-vibes/cf/pkg/internal/workspace"
)

// EnvFileCheck checks the .env file
type EnvFileCheck struct{}

func (c *EnvFileCheck) Name() string     { return "Environment File" }
func (c *EnvFileCheck) Category() string { return "internal" }

func (c *EnvFileCheck) Check(ctx context.Context) Result {
	start := time.Now()

	envPath, err := config.GetEnvFilePath()
	if err != nil {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusCritical,
			Message:  "Cannot determine .env path",
			Details:  err.Error(),
			Duration: time.Since(start),
		}
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      StatusCritical,
			Message:     ".env file not found",
			Details:     "Expected at: " + envPath,
			Recoverable: true,
			Action:      ActionAutoFix,
			Duration:    time.Since(start),
		}
	}

	creds, err := config.LoadCredentials()
	if err != nil {
		return Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      StatusCritical,
			Message:     ".env file corrupted",
			Details:     err.Error(),
			Recoverable: true,
			Action:      ActionAutoFix,
			Duration:    time.Since(start),
		}
	}

	if !creds.HasHandle() {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusDegraded,
			Message:  "CF handle not configured",
			Details:  "Set CF_HANDLE in ~/.cf.env",
			Action:   ActionUserPrompt,
			Duration: time.Since(start),
		}
	}

	return Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusHealthy,
		Message:  "Environment file OK",
		Duration: time.Since(start),
	}
}

func (c *EnvFileCheck) AutoFix(ctx context.Context) error {
	return config.EnsureEnvFile()
}

// WorkspaceCheck checks the workspace
type WorkspaceCheck struct {
	ws *workspace.Workspace
}

func NewWorkspaceCheck(ws *workspace.Workspace) *WorkspaceCheck {
	return &WorkspaceCheck{ws: ws}
}

func (c *WorkspaceCheck) Name() string     { return "Workspace" }
func (c *WorkspaceCheck) Category() string { return "internal" }

func (c *WorkspaceCheck) Check(ctx context.Context) Result {
	start := time.Now()

	if !c.ws.Exists() {
		return Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      StatusCritical,
			Message:     "Workspace not initialized",
			Details:     "Run 'cf init' to create workspace",
			Recoverable: true,
			Action:      ActionAutoFix,
			Duration:    time.Since(start),
		}
	}

	if err := c.ws.Validate(); err != nil {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusCritical,
			Message:  "Workspace validation failed",
			Details:  err.Error(),
			Duration: time.Since(start),
		}
	}

	return Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusHealthy,
		Message:  "Workspace OK",
		Duration: time.Since(start),
	}
}

func (c *WorkspaceCheck) AutoFix(ctx context.Context) error {
	creds, _ := config.LoadCredentials()
	handle := ""
	if creds != nil {
		handle = creds.CFHandle
	}
	return c.ws.Init("DSA Practice", handle)
}

// SchemaVersionCheck checks schema compatibility
type SchemaVersionCheck struct {
	ws *workspace.Workspace
}

func NewSchemaVersionCheck(ws *workspace.Workspace) *SchemaVersionCheck {
	return &SchemaVersionCheck{ws: ws}
}

func (c *SchemaVersionCheck) Name() string     { return "Schema Version" }
func (c *SchemaVersionCheck) Category() string { return "internal" }

func (c *SchemaVersionCheck) Check(ctx context.Context) Result {
	start := time.Now()

	if !c.ws.Exists() {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusHealthy,
			Message:  "No workspace to check",
			Duration: time.Since(start),
		}
	}

	version, err := c.ws.GetSchemaVersion()
	if err != nil {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusCritical,
			Message:  "Cannot read schema version",
			Details:  err.Error(),
			Duration: time.Since(start),
		}
	}

	current := schema.CurrentVersion

	if !current.IsCompatible(version) {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusCritical,
			Message:  "Incompatible schema version",
			Details:  version.String() + " (current: " + current.String() + ")",
			Action:   ActionManualFix,
			Duration: time.Since(start),
		}
	}

	if current.NeedsMigration(version) {
		return Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      StatusDegraded,
			Message:     "Schema migration available",
			Details:     version.String() + " → " + current.String(),
			Recoverable: true,
			Action:      ActionUserPrompt,
			Duration:    time.Since(start),
		}
	}

	return Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusHealthy,
		Message:  "Schema version OK: " + current.String(),
		Duration: time.Since(start),
	}
}

// ConfigCheck checks the configuration
type ConfigCheck struct{}

func (c *ConfigCheck) Name() string     { return "Configuration" }
func (c *ConfigCheck) Category() string { return "internal" }

func (c *ConfigCheck) Check(ctx context.Context) Result {
	start := time.Now()

	cfg := config.Get()
	if cfg == nil {
		return Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      StatusCritical,
			Message:     "Configuration not initialized",
			Recoverable: true,
			Action:      ActionAutoFix,
			Duration:    time.Since(start),
		}
	}

	return Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusHealthy,
		Message:  "Configuration OK",
		Duration: time.Since(start),
	}
}

func (c *ConfigCheck) AutoFix(ctx context.Context) error {
	return config.Init("")
}

// SessionCheck checks the CF session cookie configuration
type SessionCheck struct{}

func (c *SessionCheck) Name() string     { return "CF Session" }
func (c *SessionCheck) Category() string { return "internal" }

func (c *SessionCheck) Check(ctx context.Context) Result {
	start := time.Now()

	creds, err := config.LoadCredentials()
	if err != nil {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusDegraded,
			Message:  "Cannot load credentials",
			Details:  err.Error(),
			Duration: time.Since(start),
		}
	}

	// Check if cf_clearance is configured and valid
	if !creds.IsCFClearanceValid() {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusDegraded,
			Message:  "cf_clearance not configured or expired",
			Details:  "Extract cf_clearance from browser (DevTools > Application > Cookies)",
			Action:   ActionUserPrompt,
			Duration: time.Since(start),
		}
	}

	// Check if session cookies are configured
	if !creds.HasSessionCookies() {
		return Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      StatusDegraded,
			Message:     "Session cookies not configured",
			Details:     "Extract JSESSIONID and 39ce7 cookies from browser",
			Recoverable: true,
			Action:      ActionUserPrompt,
			Duration:    time.Since(start),
		}
	}

	// Check if ready for submission
	if !creds.IsReadyForSubmission() {
		return Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      StatusDegraded,
			Message:     "Missing handle or cookies",
			Details:     "Set CF_HANDLE and all session cookies in ~/.cf.env",
			Recoverable: true,
			Action:      ActionUserPrompt,
			Duration:    time.Since(start),
		}
	}

	return Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusHealthy,
		Message:  "Session configured (" + creds.GetCFClearanceStatus() + ")",
		Duration: time.Since(start),
	}
}

func (c *SessionCheck) IsCritical() bool { return false }
