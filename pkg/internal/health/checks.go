package health

import (
	"context"
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/schema"
	"github.com/harshit-vibes/cf/pkg/internal/workspace"
)

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

	if !config.HasHandle() {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusDegraded,
			Message:  "CF handle not configured",
			Details:  "Run: cf config set cf_handle YOUR_HANDLE",
			Action:   ActionUserPrompt,
			Duration: time.Since(start),
		}
	}

	return Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusHealthy,
		Message:  "Configuration OK (@" + config.GetCFHandle() + ")",
		Duration: time.Since(start),
	}
}

func (c *ConfigCheck) AutoFix(ctx context.Context) error {
	return config.Init("")
}

// CookieCheck checks if the browser cookie is configured
type CookieCheck struct{}

func (c *CookieCheck) Name() string     { return "Cookie" }
func (c *CookieCheck) Category() string { return "internal" }

func (c *CookieCheck) Check(ctx context.Context) Result {
	start := time.Now()

	if !config.HasCookie() {
		return Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   StatusDegraded,
			Message:  "Browser cookie not configured",
			Details:  "Run: cf config set cookie 'YOUR_COOKIE_STRING'",
			Action:   ActionUserPrompt,
			Duration: time.Since(start),
		}
	}

	return Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   StatusHealthy,
		Message:  "Cookie configured",
		Duration: time.Since(start),
	}
}

func (c *CookieCheck) IsCritical() bool { return false }

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
	handle := config.GetCFHandle()
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
