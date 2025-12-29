// Package errors provides centralized error handling for cf
package errors

import "fmt"

// Category represents the error category
type Category int

const (
	CatInternal Category = iota // Our fault
	CatExternal                 // CF API/web issue
	CatUser                     // User config issue
	CatNetwork                  // Connectivity
)

// String returns the category name
func (c Category) String() string {
	switch c {
	case CatInternal:
		return "internal"
	case CatExternal:
		return "external"
	case CatUser:
		return "user"
	case CatNetwork:
		return "network"
	default:
		return "unknown"
	}
}

// RecoveryAction represents what can be done to fix the error
type RecoveryAction int

const (
	ActionNone       RecoveryAction = iota // Nothing to do
	ActionAutoFix                          // Can fix automatically
	ActionUserPrompt                       // Need user input
	ActionManualFix                        // User must fix manually
	ActionRetry                            // Retry may help
	ActionFatal                            // Cannot continue
)

// AppError represents an application error
type AppError struct {
	Code        string
	Category    Category
	Message     string
	Details     string
	Suggestion  string
	Recoverable bool
	Action      RecoveryAction
	Cause       error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Error codes
const (
	// Internal errors
	ErrSchemaInvalid      = "SCHEMA_INVALID"
	ErrSchemaIncompatible = "SCHEMA_INCOMPATIBLE"
	ErrWorkspaceCorrupt   = "WORKSPACE_CORRUPT"
	ErrMigrationFailed    = "MIGRATION_FAILED"

	// User errors
	ErrEnvMissing        = "ENV_MISSING"
	ErrEnvCorrupt        = "ENV_CORRUPT"
	ErrHandleNotSet      = "HANDLE_NOT_SET"
	ErrCredentialsMissing = "CREDENTIALS_MISSING"
	ErrSessionExpired    = "SESSION_EXPIRED"
	ErrWorkspaceNotFound = "WORKSPACE_NOT_FOUND"

	// External errors
	ErrCFAPIDown          = "CF_API_DOWN"
	ErrCFAPIRateLimit     = "CF_API_RATE_LIMIT"
	ErrCFWebChanged       = "CF_WEB_CHANGED"
	ErrCFLoginFailed      = "CF_LOGIN_FAILED"
	ErrCFSubmitFailed     = "CF_SUBMIT_FAILED"
	ErrCFParseFailed      = "CF_PARSE_FAILED"

	// Network errors
	ErrNetworkOffline  = "NETWORK_OFFLINE"
	ErrNetworkTimeout  = "NETWORK_TIMEOUT"
	ErrNetworkDNS      = "NETWORK_DNS"
)

// Registry of known errors with their default properties
var Registry = map[string]*AppError{
	ErrEnvMissing: {
		Code:        ErrEnvMissing,
		Category:    CatUser,
		Message:     "Configuration file not found",
		Suggestion:  "Run 'cf init' to create configuration",
		Recoverable: true,
		Action:      ActionAutoFix,
	},
	ErrEnvCorrupt: {
		Code:        ErrEnvCorrupt,
		Category:    CatUser,
		Message:     "Configuration file is corrupted",
		Suggestion:  "Delete ~/.cf.env and run 'cf init'",
		Recoverable: true,
		Action:      ActionAutoFix,
	},
	ErrHandleNotSet: {
		Code:        ErrHandleNotSet,
		Category:    CatUser,
		Message:     "Codeforces handle not configured",
		Suggestion:  "Set CF_HANDLE in ~/.cf.env",
		Recoverable: false,
		Action:      ActionUserPrompt,
	},
	ErrCredentialsMissing: {
		Code:        ErrCredentialsMissing,
		Category:    CatUser,
		Message:     "Codeforces credentials not configured",
		Suggestion:  "Set CF_API_KEY and CF_API_SECRET in ~/.cf.env",
		Recoverable: false,
		Action:      ActionUserPrompt,
	},
	ErrSessionExpired: {
		Code:        ErrSessionExpired,
		Category:    CatUser,
		Message:     "Your Codeforces session has expired",
		Suggestion:  "Press 'L' to re-login",
		Recoverable: true,
		Action:      ActionUserPrompt,
	},
	ErrCFAPIDown: {
		Code:        ErrCFAPIDown,
		Category:    CatExternal,
		Message:     "Codeforces API is not responding",
		Suggestion:  "Check your internet or try again later",
		Recoverable: true,
		Action:      ActionRetry,
	},
	ErrCFAPIRateLimit: {
		Code:        ErrCFAPIRateLimit,
		Category:    CatExternal,
		Message:     "Codeforces API rate limit exceeded",
		Suggestion:  "Wait a moment and try again",
		Recoverable: true,
		Action:      ActionRetry,
	},
	ErrCFWebChanged: {
		Code:        ErrCFWebChanged,
		Category:    CatExternal,
		Message:     "Codeforces website structure has changed",
		Suggestion:  "Please update cf to the latest version",
		Recoverable: false,
		Action:      ActionManualFix,
	},
	ErrCFLoginFailed: {
		Code:        ErrCFLoginFailed,
		Category:    CatExternal,
		Message:     "Failed to login to Codeforces",
		Suggestion:  "Check your credentials in ~/.cf.env",
		Recoverable: true,
		Action:      ActionUserPrompt,
	},
	ErrSchemaIncompatible: {
		Code:        ErrSchemaIncompatible,
		Category:    CatInternal,
		Message:     "Workspace schema is incompatible",
		Suggestion:  "Run 'cf migrate' to upgrade",
		Recoverable: true,
		Action:      ActionUserPrompt,
	},
	ErrWorkspaceNotFound: {
		Code:        ErrWorkspaceNotFound,
		Category:    CatUser,
		Message:     "Workspace not initialized",
		Suggestion:  "Run 'cf init' in your workspace directory",
		Recoverable: true,
		Action:      ActionAutoFix,
	},
	ErrNetworkOffline: {
		Code:        ErrNetworkOffline,
		Category:    CatNetwork,
		Message:     "No internet connection",
		Suggestion:  "Check your network connection",
		Recoverable: true,
		Action:      ActionRetry,
	},
}

// New creates a new AppError from the registry
func New(code string) *AppError {
	if template, ok := Registry[code]; ok {
		return &AppError{
			Code:        template.Code,
			Category:    template.Category,
			Message:     template.Message,
			Suggestion:  template.Suggestion,
			Recoverable: template.Recoverable,
			Action:      template.Action,
		}
	}
	return &AppError{
		Code:     code,
		Category: CatInternal,
		Message:  "Unknown error",
	}
}

// Wrap wraps an existing error with context
func Wrap(code string, cause error) *AppError {
	err := New(code)
	err.Cause = cause
	err.Details = cause.Error()
	return err
}

// WithDetails adds details to an error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithSuggestion adds a suggestion to an error
func (e *AppError) WithSuggestion(suggestion string) *AppError {
	e.Suggestion = suggestion
	return e
}
