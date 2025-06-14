# Odyc CLI

[![Go Format Check](https://github.com/meldiron/odyc-cli/actions/workflows/formatter.yml/badge.svg)](https://github.com/meldiron/odyc-cli/actions/workflows/formatter.yml)
[![Go Lint Check](https://github.com/meldiron/odyc-cli/actions/workflows/linter.yml/badge.svg)](https://github.com/meldiron/odyc-cli/actions/workflows/linter.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/meldiron/odyc-cli)](https://golang.org/doc/go1.24)
[![License](https://img.shields.io/github/license/meldiron/odyc-cli)](LICENSE)

A powerful CLI tool with handy commands for Odyc.js developers. Generate code from sprites and perform various development tasks to make your life with Odyc.js easier.

## ✨ Features

- **Sprite Code Generation**: Convert PNG sprite images into JavaScript configuration files
- **Color Palette Analysis**: Automatically extract and optimize color palettes from sprites
- **Multi-sprite Support**: Process multiple PNG files in a single command
- **Smart Color Indexing**: Efficient color mapping with support for up to 62 unique colors
- **Flexible Output**: Generate JavaScript configuration files with customizable paths
- **Developer-friendly**: Beautiful terminal output with colored logging and progress indicators

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- PNG sprite images (for sprite generation)

### Installation

#### From Source

```bash
git clone https://github.com/meldiron/odyc-cli.git
cd odyc-cli
go build -o odyc .
```

#### Using Go Install

```bash
go install github.com/meldiron/odyc-cli@latest
```

### Basic Usage

```bash
# Generate sprite configuration from PNG files
odyc sprites --assets ./sprites --output ./config.js

# Force overwrite existing output file
odyc sprites --assets ./sprites --output ./config.js --force

# Show help
odyc --help
odyc sprites --help
```

## 📋 Commands

### `sprites`

Generate JavaScript configuration from sprite PNG files.

```bash
odyc sprites [OPTIONS]
```

**Options:**
- `-a, --assets <path>` - Path to assets directory containing PNG files (required)
- `-o, --output <path>` - Path to output JavaScript file (required)
- `-f, --force` - Overwrite output file if it exists
- `-h, --help` - Show help for sprites command

**Example:**
```bash
odyc sprites --assets ./game-sprites --output ./src/gameConfig.js --force
```

**Generated Output:**
```javascript
var gameConfig = {
    cellWidth: 32,
    cellHeight: 32,
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

## 🏗️ Architecture

### Project Structure

```
odyc-cli/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command and CLI setup
│   └── sprites.go         # Sprites command implementation
├── .github/               # GitHub Actions workflows
│   └── workflows/
│       ├── formatter.yml  # Go format checking
│       └── linter.yml     # Go linting with golangci-lint
├── tmp/                   # Temporary files (ignored)
├── main.go               # Application entry point
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── format.sh             # Local formatting script
├── lint.sh               # Local linting script
└── README.md             # This documentation
```

### Architecture Overview

The CLI is built using the [Cobra](https://github.com/spf13/cobra) framework for command-line interface management and [charmbracelet/log](https://github.com/charmbracelet/log) for beautiful terminal logging.

**Core Components:**
- **Main Entry Point** (`main.go`): Sets up logging styles and executes commands
- **Root Command** (`cmd/root.go`): Defines the base CLI structure and help information  
- **Sprites Command** (`cmd/sprites.go`): Handles PNG sprite processing and JavaScript generation
- **Image Processing**: Uses Go's standard `image` package for PNG decoding and pixel analysis
- **Color Management**: Intelligent color palette extraction with efficient indexing system

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

3. **Build and Test**
   ```bash
   go build -o odyc .
   ./odyc --help
   ```

### Running Locally

#### Development Build
```bash
# Build the binary
go build -o odyc .

# Run with your changes
./odyc sprites --assets ./test-sprites --output ./test-output.js
```

#### Direct Execution
```bash
# Run without building binary
go run . sprites --assets ./test-sprites --output ./test-output.js
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Code Formatting

We use `go fmt` to ensure consistent code formatting.

```bash
# Format all Go files
go fmt ./...

# Or use the provided script
./format.sh
```

**Automated Formatting Check:**
The GitHub Actions workflow automatically checks if all files are properly formatted on every push and pull request.

### Linting

We use `golangci-lint` for comprehensive code linting.

```bash
# Install golangci-lint (if not already installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Or use the provided script
./lint.sh
```

**Automated Linting:**
The GitHub Actions workflow automatically runs `golangci-lint` on every push and pull request to ensure code quality.

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

- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework for Go
- [Charm](https://charm.sh/) - Beautiful terminal applications and logging
- [golangci-lint](https://golangci-lint.run/) - Comprehensive Go linter

## 📞 Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/meldiron/odyc-cli/issues) page
2. Create a new issue if your problem isn't already reported
3. Follow the issue template to provide necessary details

---

Made with ❤️ for the Odyc.js community