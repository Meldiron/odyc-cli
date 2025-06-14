# Contributing to Odyc CLI

Thank you for your interest in contributing to Odyc CLI! We welcome contributions from everyone and are grateful for every pull request, bug report, and feature suggestion.

## 🚀 Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.24+**: [Download and install Go](https://golang.org/doc/install)
- **Git**: [Install Git](https://git-scm.com/downloads)
- **golangci-lint**: [Install golangci-lint](https://golangci-lint.run/usage/install/)

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/odyc-cli.git
   cd odyc-cli
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/meldiron/odyc-cli.git
   ```
4. **Install dependencies**:
   ```bash
   go mod download
   ```
5. **Build and test** the project:
   ```bash
   go build -o odyc .
   ./odyc --help
   ```

## 🐛 Reporting Bugs

When reporting bugs, please include:

1. **Clear title** describing the issue
2. **Steps to reproduce** the bug
3. **Expected behavior** vs **actual behavior**
4. **Environment details**:
   - Go version (`go version`)
   - Operating system
   - Odyc CLI version
5. **Sample files** if the bug is related to sprite processing
6. **Error messages** and stack traces (if any)

### Bug Report Template

```markdown
**Bug Description**
A clear and concise description of what the bug is.

**Steps to Reproduce**
1. Run command '...'
2. With files '...'
3. See error

**Expected Behavior**
What you expected to happen.

**Actual Behavior**
What actually happened.

**Environment**
- Go version: 
- OS: 
- Odyc CLI version: 

**Additional Context**
Add any other context about the problem here.
```

## 💡 Suggesting Features

We love feature suggestions! When proposing new features:

1. **Check existing issues** to avoid duplicates
2. **Describe the problem** your feature would solve
3. **Explain your proposed solution** in detail
4. **Consider alternatives** and mention them
5. **Think about backward compatibility**

### Feature Request Template

```markdown
**Feature Summary**
A brief description of the feature you'd like to see.

**Problem Statement**
What problem does this feature solve?

**Proposed Solution**
Detailed description of how you envision this feature working.

**Alternatives Considered**
What other approaches did you consider?

**Additional Context**
Any other relevant information.
```

## 🔧 Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make your changes** following our coding standards

3. **Test your changes**:
   ```bash
   # Run tests
   go test ./...
   
   # Build and test manually
   go build -o odyc .
   ./odyc sprites --assets ./test-assets --output ./test.js
   ```

4. **Format and lint**:
   ```bash
   # Format code
   go fmt ./...
   # or
   ./format.sh
   
   # Run linter
   golangci-lint run
   # or  
   ./lint.sh
   ```

5. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new sprite processing feature"
   ```

### Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or modifying tests
- `chore:` - Maintenance tasks

**Examples:**
```
feat: add support for JPEG sprite processing
fix: handle empty sprite directories gracefully
docs: update installation instructions
refactor: simplify color indexing algorithm
```

### Pull Request Process

1. **Update your branch** with upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push your changes**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a Pull Request** on GitHub with:
   - Clear title and description
   - Reference to related issues (if any)
   - Screenshots/examples (if applicable)
   - Confirmation that tests pass

4. **Address review feedback** promptly

5. **Squash commits** if requested before merging

### Pull Request Template

When creating a PR, please include:

```markdown
## Description
Brief description of changes made.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring
- [ ] Other (please describe)

## Testing
- [ ] I have tested these changes locally
- [ ] I have added/updated tests as needed
- [ ] All existing tests pass

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated (if needed)
- [ ] No breaking changes (or clearly documented)

## Related Issues
Fixes #(issue number)
```

## 🎯 Coding Standards

### Code Style

- Follow standard Go conventions and idioms
- Use `go fmt` for consistent formatting
- Write clear, descriptive variable and function names
- Add comments for complex logic
- Keep functions small and focused

### Error Handling

- Always handle errors appropriately
- Use descriptive error messages
- Wrap errors with context when needed:
  ```go
  if err != nil {
      return fmt.Errorf("failed to process sprite %s: %w", filename, err)
  }
  ```

### Logging

- Use the project's logging system (`charmbracelet/log`)
- Choose appropriate log levels:
  - `log.Error()` - Critical errors
  - `log.Warn()` - Warnings that don't stop execution
  - `log.Info()` - General information
  - `log.Debug()` - Detailed debugging information

### Testing

- Write tests for new functionality
- Maintain or improve code coverage
- Use table-driven tests when appropriate
- Test both happy path and error cases

**Test Example:**
```go
func TestSpriteProcessing(t *testing.T) {
    tests := []struct {
        name        string
        inputFile   string
        expected    SpriteMetadata
        expectError bool
    }{
        {
            name:        "valid sprite",
            inputFile:   "test-sprite.png",
            expected:    SpriteMetadata{/* ... */},
            expectError: false,
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 📝 Documentation

- Update documentation for new features
- Include examples in docstrings
- Update README.md if needed
- Write clear commit messages

## 🔍 Code Review Process

All submissions require code review. Here's what reviewers look for:

### Technical Review
- Code correctness and efficiency
- Proper error handling
- Test coverage
- Documentation completeness

### Style Review
- Consistent formatting
- Clear naming conventions
- Appropriate comments
- Go idioms and best practices

### Design Review
- API design consistency
- Backward compatibility
- Performance implications
- Security considerations

## 🏷️ Release Process

Releases follow semantic versioning (SemVer):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## 📚 Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Cobra Documentation](https://cobra.dev/)
- [Conventional Commits](https://www.conventionalcommits.org/)

## ❓ Getting Help

If you need help:

1. Check existing [Issues](https://github.com/meldiron/odyc-cli/issues)
2. Read the [README](README.md)
3. Create a new issue with the "question" label
4. Join community discussions

## 🙏 Recognition

Contributors are recognized through:
- GitHub contributor graphs
- Release notes acknowledgments
- Community shout-outs

Thank you for contributing to Odyc CLI! 🎉

---

*This contributing guide is inspired by open source best practices and may be updated as the project evolves.*