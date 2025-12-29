// Package health provides external health checks for Codeforces services
package health

import (
	"context"
	"fmt"
	"time"

	"github.com/harshit-vibes/cf/pkg/external/cfapi"
	"github.com/harshit-vibes/cf/pkg/external/cfweb"
	"github.com/harshit-vibes/cf/pkg/internal/config"
	"github.com/harshit-vibes/cf/pkg/internal/health"
)

// CFAPICheck checks the Codeforces API availability
type CFAPICheck struct {
	client *cfapi.Client
}

// NewCFAPICheck creates a new CF API check
func NewCFAPICheck(client *cfapi.Client) *CFAPICheck {
	return &CFAPICheck{client: client}
}

func (c *CFAPICheck) Name() string     { return "CF API" }
func (c *CFAPICheck) Category() string { return "external" }

func (c *CFAPICheck) Check(ctx context.Context) health.Result {
	start := time.Now()

	if c.client == nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "API client not initialized",
			Duration: time.Since(start),
		}
	}

	// Try to ping the API
	err := c.client.Ping(ctx)
	if err != nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "CF API unreachable",
			Details:  err.Error(),
			Action:   health.ActionRetry,
			Duration: time.Since(start),
		}
	}

	return health.Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   health.StatusHealthy,
		Message:  "CF API OK",
		Duration: time.Since(start),
	}
}

func (c *CFAPICheck) IsCritical() bool { return false }

// CFWebCheck checks the Codeforces web page structure
type CFWebCheck struct {
	parser *cfweb.Parser
}

// NewCFWebCheck creates a new CF web structure check
func NewCFWebCheck(parser *cfweb.Parser) *CFWebCheck {
	return &CFWebCheck{parser: parser}
}

func (c *CFWebCheck) Name() string     { return "CF Web Structure" }
func (c *CFWebCheck) Category() string { return "external" }

func (c *CFWebCheck) Check(ctx context.Context) health.Result {
	start := time.Now()

	if c.parser == nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "Parser not initialized",
			Duration: time.Since(start),
		}
	}

	// Verify page structure matches selectors
	err := c.parser.VerifyPageStructure()
	if err != nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "CF page structure changed",
			Details:  err.Error() + " (selector version: " + cfweb.CurrentVersion.Version + ")",
			Action:   health.ActionManualFix,
			Duration: time.Since(start),
		}
	}

	return health.Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   health.StatusHealthy,
		Message:  "CF web structure OK (v" + cfweb.CurrentVersion.Version + ")",
		Duration: time.Since(start),
	}
}

func (c *CFWebCheck) IsCritical() bool { return false }

// CFSessionCheck checks the CF session validity
type CFSessionCheck struct {
	session *cfweb.Session
}

// NewCFSessionCheck creates a new session check
func NewCFSessionCheck(session *cfweb.Session) *CFSessionCheck {
	return &CFSessionCheck{session: session}
}

func (c *CFSessionCheck) Name() string     { return "CF Session" }
func (c *CFSessionCheck) Category() string { return "external" }

func (c *CFSessionCheck) Check(ctx context.Context) health.Result {
	start := time.Now()

	if c.session == nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "No active session",
			Details:  "Session cookies required for submission",
			Action:   health.ActionUserPrompt,
			Duration: time.Since(start),
		}
	}

	if !c.session.IsAuthenticated() {
		return health.Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      health.StatusDegraded,
			Message:     "Session not authenticated",
			Details:     "Extract session cookies from browser (DevTools > Application > Cookies)",
			Recoverable: true,
			Action:      health.ActionUserPrompt,
			Duration:    time.Since(start),
		}
	}

	// Validate session is still working
	err := c.session.Validate()
	if err != nil {
		return health.Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      health.StatusDegraded,
			Message:     "Session invalid",
			Details:     err.Error(),
			Recoverable: true,
			Action:      health.ActionUserPrompt,
			Duration:    time.Since(start),
		}
	}

	return health.Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   health.StatusHealthy,
		Message:  "Session valid (" + c.session.Handle() + ")",
		Duration: time.Since(start),
	}
}

func (c *CFSessionCheck) IsCritical() bool { return false }

// CFHandleCheck checks if the configured handle exists
type CFHandleCheck struct {
	client *cfapi.Client
}

// NewCFHandleCheck creates a new handle check
func NewCFHandleCheck(client *cfapi.Client) *CFHandleCheck {
	return &CFHandleCheck{client: client}
}

func (c *CFHandleCheck) Name() string     { return "CF Handle" }
func (c *CFHandleCheck) Category() string { return "external" }

func (c *CFHandleCheck) Check(ctx context.Context) health.Result {
	start := time.Now()

	creds, err := config.LoadCredentials()
	if err != nil || creds == nil || !creds.HasHandle() {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "CF handle not configured",
			Details:  "Set CF_HANDLE in ~/.cf.env",
			Action:   health.ActionUserPrompt,
			Duration: time.Since(start),
		}
	}

	if c.client == nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "Cannot verify handle",
			Details:  "API client not initialized",
			Duration: time.Since(start),
		}
	}

	// Verify handle exists on CF
	users, err := c.client.GetUserInfo(ctx, []string{creds.CFHandle})
	if err != nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "Cannot verify handle",
			Details:  err.Error(),
			Action:   health.ActionRetry,
			Duration: time.Since(start),
		}
	}

	if len(users) == 0 {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusCritical,
			Message:  "Handle not found on CF",
			Details:  creds.CFHandle,
			Action:   health.ActionManualFix,
			Duration: time.Since(start),
		}
	}

	user := users[0]
	return health.Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   health.StatusHealthy,
		Message:  user.Handle + " (" + user.Rank + ", " + formatRating(user.Rating) + ")",
		Duration: time.Since(start),
	}
}

func (c *CFHandleCheck) IsCritical() bool { return false }

// NetworkCheck checks basic network connectivity
type NetworkCheck struct{}

func (c *NetworkCheck) Name() string     { return "Network" }
func (c *NetworkCheck) Category() string { return "external" }

func (c *NetworkCheck) Check(ctx context.Context) health.Result {
	start := time.Now()

	// Try to fetch CF homepage
	parser := cfweb.NewParserWithClient(nil)
	_, err := parser.ParseProblem(1, "A") // Test with a known problem

	if err != nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "Network connectivity issue",
			Details:  "Cannot reach codeforces.com",
			Action:   health.ActionRetry,
			Duration: time.Since(start),
		}
	}

	return health.Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   health.StatusHealthy,
		Message:  "Network OK",
		Duration: time.Since(start),
	}
}

func (c *NetworkCheck) IsCritical() bool { return false }

// CFClearanceCheck checks the Cloudflare cf_clearance token status
type CFClearanceCheck struct{}

// NewCFClearanceCheck creates a new cf_clearance check
func NewCFClearanceCheck() *CFClearanceCheck {
	return &CFClearanceCheck{}
}

func (c *CFClearanceCheck) Name() string     { return "CF Clearance" }
func (c *CFClearanceCheck) Category() string { return "external" }

func (c *CFClearanceCheck) Check(ctx context.Context) health.Result {
	start := time.Now()

	creds, err := config.LoadCredentials()
	if err != nil {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "Cannot load credentials",
			Details:  err.Error(),
			Duration: time.Since(start),
		}
	}

	if creds.CFClearance == "" {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "cf_clearance not configured",
			Details:  "Run 'cf token refresh' or set CF_CLEARANCE manually",
			Action:   health.ActionUserPrompt,
			Duration: time.Since(start),
		}
	}

	if !creds.IsCFClearanceValid() {
		return health.Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      health.StatusDegraded,
			Message:     "cf_clearance expired",
			Details:     "Run 'cf token refresh' to update",
			Recoverable: true,
			Action:      health.ActionUserPrompt,
			Duration:    time.Since(start),
		}
	}

	// Check if expiring soon (within 5 minutes)
	expiresIn := creds.CFClearanceExpiresIn()
	if expiresIn < 5*time.Minute {
		return health.Result{
			Name:        c.Name(),
			Category:    c.Category(),
			Status:      health.StatusDegraded,
			Message:     "cf_clearance expiring soon",
			Details:     fmt.Sprintf("Expires in %s", formatDuration(expiresIn)),
			Recoverable: true,
			Action:      health.ActionUserPrompt,
			Duration:    time.Since(start),
		}
	}

	return health.Result{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   health.StatusHealthy,
		Message:  fmt.Sprintf("cf_clearance valid (%s remaining)", formatDuration(expiresIn)),
		Duration: time.Since(start),
	}
}

func (c *CFClearanceCheck) IsCritical() bool { return false }

// Helper functions

func formatRating(rating int) string {
	if rating == 0 {
		return "unrated"
	}
	return fmt.Sprintf("%d", rating)
}

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
