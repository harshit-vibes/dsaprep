# DSA Prep

A beautiful terminal UI application for practicing competitive programming problems from Codeforces.

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## Features

- 🎯 Browse 8000+ Codeforces problems with filtering and search
- ⏱️ Track practice sessions with built-in timer
- 📊 View your Codeforces statistics and rating history
- 🎲 Get random problems matching your skill level
- 🎨 Beautiful TUI with Codeforces rank colors

## Installation

### Using Go (requires Go 1.22+)

```bash
go install github.com/harshit-vibes/dsaprep/cmd/dsaprep@latest
```

### Using Homebrew (macOS/Linux)

```bash
brew tap harshit-vibes/tap
brew install dsaprep
```

### Download Binary

Download pre-built binaries from the [Releases](https://github.com/harshit-vibes/dsaprep/releases) page.

#### macOS (Apple Silicon)
```bash
curl -Lo dsaprep.tar.gz https://github.com/harshit-vibes/dsaprep/releases/latest/download/dsaprep_Darwin_arm64.tar.gz
tar -xzf dsaprep.tar.gz
sudo mv dsaprep /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -Lo dsaprep.tar.gz https://github.com/harshit-vibes/dsaprep/releases/latest/download/dsaprep_Darwin_amd64.tar.gz
tar -xzf dsaprep.tar.gz
sudo mv dsaprep /usr/local/bin/
```

#### Linux
```bash
curl -Lo dsaprep.tar.gz https://github.com/harshit-vibes/dsaprep/releases/latest/download/dsaprep_Linux_amd64.tar.gz
tar -xzf dsaprep.tar.gz
sudo mv dsaprep /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/harshit-vibes/dsaprep.git
cd dsaprep
make install
```

## Usage

### Interactive TUI

Launch the interactive terminal UI:

```bash
dsaprep
```

**Navigation:**
- `1` / `d` - Dashboard
- `2` / `p` - Problems browser
- `3` / `P` - Practice mode
- `4` / `s` - Statistics
- `↑/↓` or `j/k` - Navigate
- `Enter` - Select
- `/` - Search
- `?` - Help
- `q` - Quit

### CLI Commands

```bash
# Get a random problem in your rating range
dsaprep random

# Get a random problem with specific rating
dsaprep random --min 1200 --max 1600

# Get a random DP problem
dsaprep random --tag dp

# View a specific problem
dsaprep problem 1A
dsaprep problem 1234B --open  # Open in browser

# View Codeforces stats
dsaprep stats tourist
dsaprep stats --rating       # Show rating history
dsaprep stats --submissions  # Show recent submissions

# Configuration
dsaprep config get                          # Show all config
dsaprep config set cf_handle <your_handle>  # Set your CF handle
dsaprep config set difficulty.min 1200      # Set min rating
dsaprep config set difficulty.max 1800      # Set max rating
dsaprep config set daily_goal 10            # Set daily goal
```

## Configuration

Configuration is stored in `~/.dsaprep/config.yaml`:

```yaml
cf_handle: your_codeforces_handle
difficulty:
  min: 800
  max: 1600
daily_goal: 5
theme: dark
```

## Screenshots

*Coming soon*

## Development

```bash
# Clone the repository
git clone https://github.com/harshit-vibes/dsaprep.git
cd dsaprep

# Install dependencies
make deps

# Development build
make dev

# Run
make run

# Run tests
make test

# Build for all platforms
make build-all
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
