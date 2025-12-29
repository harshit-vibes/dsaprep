# cf - Codeforces CLI

A command-line tool and Go SDK for competitive programming with Codeforces integration.

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
[![CI](https://github.com/harshit-vibes/cf/actions/workflows/ci.yml/badge.svg)](https://github.com/harshit-vibes/cf/actions/workflows/ci.yml)

## Features

- **Codeforces API Client** - Full-featured API client with caching and rate limiting
- **Web Scraper** - Parse problem statements, samples, and metadata
- **Workspace Management** - Organize problems locally with versioned schemas
- **Health Checks** - Verify system configuration and connectivity
- **TUI Dashboard** - Beautiful terminal UI (coming soon)
- **Cross-Platform** - Works on Linux, macOS, and Windows

## Installation

### Using Go (requires Go 1.22+)

```bash
go install github.com/harshit-vibes/cf/cmd/cf@latest
```

### Download Binary

Download pre-built binaries from the [Releases](https://github.com/harshit-vibes/cf/releases) page.

**macOS (Apple Silicon)**
```bash
curl -Lo cf https://github.com/harshit-vibes/cf/releases/latest/download/cf-darwin-arm64
chmod +x cf
sudo mv cf /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -Lo cf https://github.com/harshit-vibes/cf/releases/latest/download/cf-darwin-amd64
chmod +x cf
sudo mv cf /usr/local/bin/
```

**Linux (amd64)**
```bash
curl -Lo cf https://github.com/harshit-vibes/cf/releases/latest/download/cf-linux-amd64
chmod +x cf
sudo mv cf /usr/local/bin/
```

**Linux (arm64)**
```bash
curl -Lo cf https://github.com/harshit-vibes/cf/releases/latest/download/cf-linux-arm64
chmod +x cf
sudo mv cf /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/harshit-vibes/cf.git
cd cf
make build
sudo mv build/cf /usr/local/bin/
```

## Quick Start

```bash
# Initialize a workspace
cf init ~/codeforces-practice

# Check system health
cf health

# Parse a problem from Codeforces
cf problem parse 1 A

# View your profile
cf user info

# List recent contests
cf contest list

# Show your statistics
cf stats
```

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `cf init [path]` | Initialize a new workspace |
| `cf health` | Check system health and configuration |
| `cf version` | Show version information |

### Problem Commands (`cf problem`, `cf p`)

| Command | Description |
|---------|-------------|
| `cf problem parse <contest> <index>` | Parse a problem from Codeforces |
| `cf problem list [--tag TAG] [--min-rating N] [--max-rating N]` | List problems with filters |
| `cf problem fetch <contest> [index]` | Fetch problem(s) to workspace |

```bash
# Parse problem A from contest 1
cf problem parse 1 A

# List DP problems rated 1200-1400
cf problem list --tag dp --min-rating 1200 --max-rating 1400

# Fetch all problems from contest 1234
cf problem fetch 1234
```

### User Commands (`cf user`, `cf u`)

| Command | Description |
|---------|-------------|
| `cf user info [handle]` | Show user profile information |
| `cf user submissions [handle] [--limit N]` | Show recent submissions |
| `cf user rating [handle]` | Show rating history |

```bash
# View your profile
cf user info

# View tourist's submissions
cf user submissions tourist --limit 20

# View your rating history
cf user rating
```

### Contest Commands (`cf contest`, `cf c`)

| Command | Description |
|---------|-------------|
| `cf contest list [--gym] [--limit N]` | List contests |
| `cf contest problems <contest_id>` | Show contest problems |

```bash
# List upcoming contests
cf contest list --limit 10

# Show problems from contest 1234
cf contest problems 1234
```

### Statistics (`cf stats`)

```bash
# View your practice statistics
cf stats

# View another user's stats
cf stats tourist
```

### Configuration (`cf config`)

| Command | Description |
|---------|-------------|
| `cf config get [key]` | Show configuration value(s) |
| `cf config set <key> <value>` | Set a configuration value |
| `cf config path` | Show config file paths |

```bash
# View all configuration
cf config get

# Set your CF handle
cf config set cf_handle your_handle

# Set difficulty range
cf config set difficulty.min 1000
cf config set difficulty.max 1600
```

### Workspace Structure

After running `cf init`, your workspace looks like:

```
workspace/
├── workspace.yaml      # Workspace manifest
├── problems/           # Problem metadata and statements
├── submissions/        # Your solutions
└── stats/              # Progress tracking
```

## Configuration

### Config Files

Configuration is stored in:
- `~/.cf/config.yaml` - Main configuration
- `~/.cf.env` - API credentials

### Setting Up Credentials

Create `~/.cf.env`:

```bash
# Get your API key from https://codeforces.com/settings/api
CF_HANDLE=your_username
CF_API_KEY=your_api_key
CF_API_SECRET=your_api_secret

# Optional: Session cookies for submissions
CF_JSESSIONID=
CF_39CE7=
```

### Configuration Options

Edit `~/.cf/config.yaml`:

```yaml
cf_handle: your_handle
difficulty:
  min: 800
  max: 1400
daily_goal: 3
workspace_path: /path/to/workspace
```

## Migration from dsaprep

If you were using the previous version (dsaprep), your configuration will be automatically migrated on first run:

- `~/.dsaprep/` → `~/.cf/`
- `~/.dsaprep.env` → `~/.cf.env`

Your original files are preserved.

## Using as a Go SDK

### Installation

```bash
go get github.com/harshit-vibes/cf
```

### Codeforces API Client

```go
package main

import (
    "fmt"
    "github.com/harshit-vibes/cf/pkg/external/cfapi"
)

func main() {
    // Create client (no auth needed for public endpoints)
    client := cfapi.NewClient()

    // Get user info
    users, err := client.GetUserInfo([]string{"tourist"})
    if err != nil {
        panic(err)
    }
    fmt.Printf("Rating: %d\n", users[0].Rating)

    // Get problems
    problems, err := client.GetProblems([]string{"dp"})
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d DP problems\n", len(problems))
}
```

### Web Parser

```go
package main

import (
    "fmt"
    "github.com/harshit-vibes/cf/pkg/external/cfweb"
)

func main() {
    parser := cfweb.NewParser()

    // Parse a problem
    problem, err := parser.ParseProblem(1, "A")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Problem: %s\n", problem.Name)
    fmt.Printf("Rating: %d\n", problem.Rating)
    fmt.Printf("Samples: %d\n", len(problem.Samples))
}
```

### Workspace Management

```go
package main

import (
    "github.com/harshit-vibes/cf/pkg/internal/workspace"
)

func main() {
    // Open workspace
    ws := workspace.New("./my-workspace")

    // Initialize if needed
    if !ws.Exists() {
        ws.Init("My Practice", "my_cf_handle")
    }

    // List problems
    problems, _ := ws.ListProblems()
    for _, p := range problems {
        fmt.Printf("%s: %s\n", p.ID, p.Name)
    }
}
```

## Development

### Building

```bash
# Development build
make dev

# Production build
make build

# Cross-compile
make build-all
```

### Testing

```bash
# Run tests
make test

# With coverage
make test-coverage
```

### Project Structure

```
cf/
├── cmd/cf/              # CLI entry point
├── pkg/
│   ├── cmd/             # CLI commands
│   ├── tui/             # TUI components (coming soon)
│   ├── internal/
│   │   ├── config/      # Configuration
│   │   ├── health/      # Health checks
│   │   ├── workspace/   # Workspace management
│   │   ├── schema/      # Data schemas
│   │   └── errors/      # Error handling
│   └── external/
│       ├── cfapi/       # Codeforces API
│       ├── cfweb/       # Web scraping
│       └── health/      # External checks
├── Makefile
└── README.md
```

## Roadmap

- [ ] TUI Dashboard
- [ ] Problem recommendations
- [ ] Solution submission
- [ ] Contest participation
- [ ] Progress analytics

## License

MIT
