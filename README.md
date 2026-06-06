# Odyc CLI

[![Go Format Check](https://github.com/meldiron/odyc-cli/actions/workflows/formatter.yml/badge.svg)](https://github.com/meldiron/odyc-cli/actions/workflows/formatter.yml)
[![Go Lint Check](https://github.com/meldiron/odyc-cli/actions/workflows/linter.yml/badge.svg)](https://github.com/meldiron/odyc-cli/actions/workflows/linter.yml)
[![Go Tests](https://github.com/meldiron/odyc-cli/actions/workflows/tests.yml/badge.svg)](https://github.com/meldiron/odyc-cli/actions/workflows/tests.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/meldiron/odyc-cli)](https://golang.org/doc/go1.24)
[![License](https://img.shields.io/github/license/meldiron/odyc-cli)](LICENSE)

A powerful CLI tool with handy commands for Odyc.js developers. Generate code from sprites and perform various development tasks to make your life with Odyc.js easier.

![Cover](docs/cover.png)

## ✨ Features

- **Sprite Code Generation**: Convert PNG sprite images into JavaScript configuration files
- **Color Palette Analysis**: Automatically extract and optimize color palettes from sprites
- **Multi-sprite Support**: Process multiple PNG files in a single command
- **Smart Color Indexing**: Efficient color mapping with support for up to 62 unique colors
- **Flexible Output**: Generate JavaScript configuration files with customizable paths
- **Account Authentication**: Sign in securely from the terminal using the OAuth 2.1 device flow
- **Game Scaffolding**: Create a new Odyc game from a starter template with a single command
- **One-command Deploy**: Publish your game to Odyc and get a shareable, playable URL
- **Developer-friendly**: Beautiful terminal output with colored logging and progress indicators

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- PNG sprite images (for sprite generation)

### Installation

#### Using Go Install

```bash
go install github.com/meldiron/odyc-cli@latest
```

### Basic Usage

```bash
# Show help
odyc-cli
odyc-cli --help
```

## 📋 Commands

### `sprites`

Generate JavaScript configuration from sprite PNG files.

```bash
odyc-cli sprites [OPTIONS]
```

**Options:**
- `-a, --assets <path>` - Path to assets directory containing PNG files (required)
- `-o, --output <path>` - Path to output JavaScript file (required)
- `-f, --force` - Overwrite output file if it exists
- `-h, --help` - Show help for sprites command

**Example:**
```bash
odyc-cli sprites --assets ./game-sprites --output ./src/gameConfig.js --force
```

**Generated Output:**
```javascript
var gameConfig = {
    cellWidth: 6,
    cellHeight: 5,
    colors: [
        "#ff0000ff",
        "#00ff00ff",
        "#0000ffff"
    ],
    sprites: {
        "player": `
            ..12..
            .1221.
            122221
            .1221.
            ..12..
        `,
        "enemy": `
            ..13..
            .1331.
            133331
            .1331.
            ..13..
        `
    }
};
```

### `create`

Create a new game on your Odyc account and scaffold a local folder for it. Prompts for a folder name (or accepts one as an argument), creates the game via the API (using the `games.create` scope), then scaffolds a starter project and an `odyc.json` linking the folder to the game. Requires you to be signed in (`odyc-cli login`).

The scaffold is a small, modular game you can play immediately by opening `index.html` (it loads Odyc.js from a CDN) and grow from there:

```
my-game/
├── index.html      # Loads Odyc + every game file, in order. Open this to play.
├── index.js        # Entry point — starts the first scene.
├── scenes/         # One file per scene (a screen / state of the game).
│   ├── title.js
│   └── world.js
├── utils/          # Shared, reusable code (palette, sprites, helpers).
│   └── sprites.js
└── odyc.json       # Links this folder to your game on Odyc.
```

There is no build step: every `.js` file is loaded into one shared scope, so anything defined in one file is available in the others.

After creating, you need to authorize deploys for the new game by signing in again with its ID (see `login --game-id`).

```bash
odyc-cli create [folder]
```

**Example:**
```bash
odyc-cli create my-game
cd my-game
open index.html   # play locally
odyc-cli login --game-id="<id printed by create>"
# edit scenes/, utils/ and index.js
odyc-cli deploy
```

**Options:**
- `-h, --help` - Show help for create command

### `deploy`

Bundle the current folder's game code into a single file and update the linked game with it. Files are concatenated in the same order `index.html` loads them — every `*.js` in `utils/`, then every `*.js` in `scenes/`, then `index.js` last — so the deployed bundle behaves exactly like the game does locally. (Legacy single-file projects with just a `game.js` are still deployed as-is.) The game must already exist (`odyc-cli create`) and be recorded in `odyc.json`; this command only updates code, it does not create games. When finished, the CLI prints the playable URL.

Code updates are authorized per game via an OAuth 2.1 Rich Authorization Request (`type: game`, `actions: [code.write]`, `identifier: <gameId>`) granted at sign-in — run `odyc-cli login --game-id="<id>"` first. If the grant is missing, deploy returns a `403` and reminds you how to authorize.

```bash
odyc-cli deploy
```

**`odyc.json`:**
```json
{
  "gameId": "...",
  "slug": "..."
}
```

**Options:**
- `-h, --help` - Show help for deploy command

### `login`

Sign in to your Odyc account using the OAuth 2.1 device authorization flow. The CLI prints a verification URL and a code, then waits while you authorize. Press ENTER to open the URL in your browser, or open it yourself — polling starts immediately either way. Credentials are stored locally in your OS config directory (`auth.json`) with owner-only permissions.

Pass `--game-id` to additionally authorize code deploys for a specific game. This narrows the requested Rich Authorization Request to `type: game`, `actions: [code.write]`, `identifier: <gameId>`, which the deploy endpoint requires.

```bash
odyc-cli login
odyc-cli login --game-id="<game id>"
```

**Options:**
- `--game-id <id>` - Authorize code deploys for a specific game ID
- `-h, --help` - Show help for login command

### `whoami`

Show the currently signed-in account by fetching details from the OAuth `/userinfo` endpoint. Expired sessions are refreshed automatically when possible.

```bash
odyc-cli whoami
```

**Options:**
- `-h, --help` - Show help for whoami command

### `logout`

Sign out by revoking the current tokens at the authorization server (best effort) and removing the locally stored credentials.

```bash
odyc-cli logout
```

**Options:**
- `-h, --help` - Show help for logout command

## 🏗️ Architecture

### Project Structure

```
odyc-cli/
├── cmd/                  # Commands implementation
├── .github/              # GitHub Actions workflows
└── main.go               # Application entrypoint
```

### Architecture Overview

The CLI is built using the [Cobra](https://github.com/spf13/cobra) framework for command-line interface management and [charmbracelet/log](https://github.com/charmbracelet/log) for beautiful terminal logging.

**Core Components:**
- **Main Entry Point** (`main.go`): Sets up logging styles and executes commands
- **Root Command** (`cmd/root.go`): Defines the base CLI structure and help information  
- **Commands** (`cmd/*.go`): All available commands

## 🛠️ Development

### Prerequisites for Contributors

- Go 1.24 or higher
- golangci-lint (for linting)
- Git

### Getting Started

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/odyc-cli.git
   cd odyc-cli
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Build and Run**
   ```bash
   go build -o odyc-cli .
   ./odyc-cli --help
   ```

### Running Tests

```bash
# Build binary
go build -o odyc-cli .

# Run all tests
go test ./...
```

### Code Formatting

We use `go fmt` to ensure consistent code formatting.

```bash
# Format all Go files
go fmt ./...
```

### Code Linter

We use `golangci-lint` for comprehensive code linting.

```bash
# Install golangci-lint (if not already installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Continuous Integration

Our CI/CD pipeline includes:

- **Format Check**: Ensures all Go code is properly formatted with `go fmt`
- **Lint Check**: Runs `golangci-lint` to catch potential issues and enforce coding standards
- **Automated Testing**: Runs on every push and pull request to `main` and `develop` branches

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to:

- Report bugs and request features
- Submit pull requests
- Follow our coding standards
- Set up your development environment

## 📜 Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

### Core Dependencies
- [Go](https://golang.org/) - The programming language used to build CLI tool
- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework for Go
- [Charmbracelet Log](https://github.com/charmbracelet/log) - Beautiful structured logging
- [Charmbracelet Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal UIs
- [Testify](https://github.com/stretchr/testify) - Testing toolkit with assertions and mocks
- [golangci-lint](https://golangci-lint.run/) - Comprehensive Go linter

## 📞 Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/meldiron/odyc-cli/issues) page
2. Create a new issue if your problem isn't already reported
3. Follow the issue template to provide necessary details

---

Made with ❤️ for the Odyc.js community