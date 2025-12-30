# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`cf` is a Codeforces CLI tool and Go SDK for competitive programming. It provides:
- Codeforces API client with caching and rate limiting
- Web scraper for parsing problem statements and samples
- Local workspace management with versioned schemas
- Interactive TUI dashboard (Bubble Tea)

## Build & Development Commands

```bash
# Development build (fast, no version info)
make dev

# Production build with version info
make build

# Build and run
make run

# Install to $GOPATH/bin
make install

# Install to /usr/local/bin
make install-local

# Cross-compile for all platforms
make build-all
```

## Testing

```bash
# Run all unit tests
make test

# Run tests with coverage report (generates coverage.html)
make test-coverage

# Run a single test file
go test -v ./pkg/internal/config/...

# Run a specific test function
go test -v -run TestRealAPI_GetUserInfo ./pkg/integration/...

# Run integration tests (requires credentials in ~/.cf.env)
go test -tags=integration -v ./pkg/integration/...
```

## Linting

```bash
# Run golangci-lint (requires golangci-lint installed)
make lint

# Format code
make fmt

# Download and tidy dependencies
make deps
```

## Architecture

### Package Structure

```
pkg/
├── cmd/                    # CLI commands (Cobra)
│   └── root.go             # Main entrypoint, all commands registered here
├── tui/                    # Terminal UI (Bubble Tea)
│   ├── app.go              # Main TUI application model
│   ├── views/              # Individual view models (dashboard, problems, etc.)
│   └── styles/             # Styling (lipgloss)
├── internal/               # Private packages
│   ├── config/             # Configuration management (Viper)
│   │   └── config.go       # Config (~/.cf/config.yaml) - handle, cookie, preferences
│   ├── health/             # Health check system
│   ├── workspace/          # Local workspace management
│   ├── schema/             # Versioned data schemas
│   │   └── v1/             # Schema version 1.x types
│   └── errors/             # Custom error types
└── external/               # Public-facing packages (can be imported as SDK)
    ├── cfapi/              # Codeforces API client
    │   ├── client.go       # API client with rate limiting & caching
    │   ├── types.go        # API response types
    │   └── cache.go        # In-memory cache
    ├── cfweb/              # Web scraper
    │   ├── parser.go       # HTML parser for problems
    │   ├── selectors.go    # CSS selectors for CF pages
    │   ├── session.go      # HTTP session with cookies
    │   └── submitter.go    # Solution submission
    └── health/             # External health checks
```

### Key Patterns

**CLI Command Structure**: Commands are defined in `pkg/cmd/` using Cobra. New commands should be added to `init()` in `root.go`.

**Configuration**: Single config file `~/.cf/config.yaml`:
- `cf_handle` - Codeforces username
- `cookie` - Browser cookie string for authenticated requests (JSESSIONID, 39ce7, cf_clearance)
- `difficulty.min/max` - Problem difficulty range for recommendations
- `daily_goal` - Practice target
- `workspace_path` - Workspace directory

**Schema Versioning**: Workspace data uses versioned schemas in `pkg/internal/schema/`. The current version is in `schema.CurrentVersion`. New schema versions should maintain backward compatibility.

**Health Checks**: Startup health checks in `pkg/internal/health/` verify configuration and connectivity. External checks (API, web) are in `pkg/external/health/`.

**API Client Options**: The cfapi client uses functional options pattern:
```go
client := cfapi.NewClient(
    cfapi.WithCacheTTL(10*time.Minute),
    cfapi.WithHTTPClient(customClient),
)
```
Note: CF API endpoints are public and don't require API credentials.

### TUI Architecture

The TUI uses Bubble Tea's Model-View-Update pattern:
- `App` struct in `pkg/tui/app.go` is the main model
- Views are in `pkg/tui/views/` (dashboard, problems, submissions, profile, settings)
- Each view implements `SetSize`, `Update`, and `View` methods
- Tab switching is handled by keyboard shortcuts (1-5, Tab/Shift+Tab)

## Testing Conventions

- Unit tests are in `*_test.go` files alongside source
- Mock implementations for testing are in `mock_test.go` files
- Integration tests use `//go:build integration` tag and require real credentials
- Tests use temporary directories via `t.TempDir()`

## Configuration Files

| File | Purpose |
|------|---------|
| `~/.cf/config.yaml` | User preferences (handle, cookie, difficulty, etc.) |
| `workspace.yaml` | Workspace manifest (in workspace root) |
| `.golangci.yml` | Linter configuration |
| `.goreleaser.yaml` | Release automation |
