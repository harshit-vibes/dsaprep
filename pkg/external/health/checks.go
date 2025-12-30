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

	handle := config.GetCFHandle()
	if handle == "" {
		return health.Result{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   health.StatusDegraded,
			Message:  "CF handle not configured",
			Details:  "Run: cf config set cf_handle YOUR_HANDLE",
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
	users, err := c.client.GetUserInfo(ctx, []string{handle})
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
			Details:  handle,
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

// Helper functions

func formatRating(rating int) string {
	if rating == 0 {
		return "unrated"
	}
	return fmt.Sprintf("%d", rating)
}
