package health

import (
	"context"
	"time"

	"github.com/harshit-vibes/cf/pkg/internal/schema"
)

// Checker orchestrates health checks
type Checker struct {
	checks []Check
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		checks: []Check{},
	}
}

// AddCheck adds a check to the checker
func (c *Checker) AddCheck(check Check) {
	c.checks = append(c.checks, check)
}

// Run executes all health checks
func (c *Checker) Run(ctx context.Context) *Report {
	start := time.Now()

	report := &Report{
		Timestamp:            start,
		OverallStatus:        StatusHealthy,
		CanProceed:           true,
		CurrentSchemaVersion: schema.CurrentVersion.String(),
	}

	// Run internal checks first
	for _, check := range c.checks {
		if check.Category() != "internal" {
			continue
		}
		c.runCheck(ctx, check, report)
		if !report.CanProceed {
			break
		}
	}

	// Run external checks if internal passed
	if report.CanProceed {
		for _, check := range c.checks {
			if check.Category() != "external" {
				continue
			}
			c.runCheck(ctx, check, report)
		}
	}

	report.Duration = time.Since(start)
	return report
}

func (c *Checker) runCheck(ctx context.Context, check Check, report *Report) {
	result := check.Check(ctx)
	report.Results = append(report.Results, result)

	switch result.Status {
	case StatusCritical:
		// Try auto-fix if available
		if result.Recoverable && result.Action == ActionAutoFix {
			if af, ok := check.(AutoFixable); ok {
				if err := af.AutoFix(ctx); err == nil {
					result.Message += " (auto-fixed)"
					result.Status = StatusHealthy
					return
				}
			}
		}

		// Check if critical
		isCritical := true
		if cr, ok := check.(Critical); ok {
			isCritical = cr.IsCritical()
		}

		if isCritical {
			report.OverallStatus = StatusCritical
			report.CanProceed = false
			report.Errors = append(report.Errors, result.Message)
		} else {
			if report.OverallStatus == StatusHealthy {
				report.OverallStatus = StatusDegraded
			}
			report.Warnings = append(report.Warnings, result.Message)
		}

	case StatusDegraded:
		if report.OverallStatus == StatusHealthy {
			report.OverallStatus = StatusDegraded
		}
		report.Warnings = append(report.Warnings, result.Message)
	}
}

// QuickCheck runs a fast subset of checks
func (c *Checker) QuickCheck(ctx context.Context) bool {
	for _, check := range c.checks {
		result := check.Check(ctx)
		if result.Status == StatusCritical {
			if cr, ok := check.(Critical); ok && cr.IsCritical() {
				return false
			}
		}
	}
	return true
}
