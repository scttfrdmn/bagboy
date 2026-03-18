# Contributing to bagboy

Thank you for your interest in contributing to bagboy! This guide will help you get started.

## Quick Start

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/bagboy.git
cd bagboy

# Build the project
make build

# Run tests
make test

# Run a specific test
go test ./pkg/packager/brew -v
```

## Development Setup

### Prerequisites

- Go 1.19 or later
- Git
- Make

### Building

```bash
make build          # Build binary to bin/bagboy
make install        # Install to /usr/local/bin
make clean          # Clean build artifacts
```

### Testing

```bash
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make test-verbose   # Run tests with verbose output
```

## Project Structure

```
bagboy/
├── cmd/bagboy/           # CLI application entry point
├── pkg/
│   ├── config/          # Configuration handling
│   ├── packager/        # Package format implementations
│   │   ├── brew/        # Homebrew formulas
│   │   ├── scoop/       # Scoop manifests
│   │   ├── deb/         # Debian packages
│   │   └── ...          # Other packagers
│   ├── github/          # GitHub API integration
│   ├── signing/         # Code signing
│   ├── deps/            # Dependency management
│   ├── ui/              # CLI user interface
│   └── errors/          # Error handling
├── docs/                # Documentation
└── examples/            # Example configurations
```

## Adding a New Packager

To add support for a new package format:

### 1. Create Package Directory

```bash
mkdir -p pkg/packager/myformat
```

### 2. Implement the Packager Interface

Create `pkg/packager/myformat/packager.go`:

```go
package myformat

import (
    "context"
    "github.com/scttfrdmn/bagboy/pkg/config"
)

type Packager struct {
    config *config.Config
}

func New(cfg *config.Config) *Packager {
    return &Packager{config: cfg}
}

func (p *Packager) Name() string {
    return "myformat"
}

func (p *Packager) Pack(ctx context.Context, outputDir string) (string, error) {
    // Implementation here
    return outputPath, nil
}

func (p *Packager) Validate() error {
    // Validation logic
    return nil
}
```

### 3. Add Tests

Create `pkg/packager/myformat/packager_test.go`:

```go
package myformat

import (
    "context"
    "testing"
    "github.com/scttfrdmn/bagboy/pkg/config"
)

func TestNew(t *testing.T) {
    cfg := &config.Config{Name: "test", Version: "1.0.0"}
    p := New(cfg)
    if p == nil {
        t.Fatal("New returned nil")
    }
}

func TestPack(t *testing.T) {
    cfg := &config.Config{
        Name: "testapp",
        Version: "1.0.0",
    }
    
    p := New(cfg)
    ctx := context.Background()
    outputDir := t.TempDir()
    
    path, err := p.Pack(ctx, outputDir)
    if err != nil {
        t.Fatalf("Pack failed: %v", err)
    }
    
    if path == "" {
        t.Error("Expected output path")
    }
}
```

### 4. Register the Packager

Add to `cmd/bagboy/main.go`:

```go
import "github.com/scttfrdmn/bagboy/pkg/packager/myformat"

// In the packager registry
packagers["myformat"] = myformat.New(cfg)
```

### 5. Add Documentation

Update `docs/PACKAGE_FORMATS.md` with your new format.

## Code Style

### Go Conventions

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions small and focused
- Write descriptive variable names

### Error Handling

Use the structured error system:

```go
import "github.com/scttfrdmn/bagboy/pkg/errors"

// Return structured errors
if err != nil {
    return errors.Wrap(err, errors.CategoryPackaging, 
        "failed to create package",
        "Check that output directory is writable")
}
```

### Testing

- Aim for 60%+ test coverage
- Test happy path and error cases
- Use table-driven tests for multiple scenarios
- Clean up temp files with `t.TempDir()`

Example:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "result", false},
        {"invalid input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Pull Request Process

### 1. Create a Branch

```bash
git checkout -b feature/my-new-feature
```

### 2. Make Changes

- Write code
- Add tests
- Update documentation
- Run tests locally

### 3. Commit

Use conventional commit messages:

```bash
git commit -m "feat: add support for XYZ package format"
git commit -m "fix: resolve checksum calculation bug"
git commit -m "docs: update installation instructions"
```

Commit types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding tests
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `chore`: Maintenance tasks

### 4. Push and Create PR

```bash
git push origin feature/my-new-feature
```

Then create a pull request on GitHub with:
- Clear description of changes
- Link to related issues
- Screenshots (if UI changes)
- Test results

### 5. Code Review

- Address reviewer feedback
- Keep PR focused and small
- Update tests as needed

## Testing Guidelines

### Unit Tests

Test individual functions and methods:

```go
func TestPackagerName(t *testing.T) {
    p := New(&config.Config{})
    if p.Name() != "expected" {
        t.Errorf("got %s, want expected", p.Name())
    }
}
```

### Integration Tests

Test complete workflows:

```go
func TestEndToEndPackaging(t *testing.T) {
    // Setup
    cfg := loadTestConfig()
    outputDir := t.TempDir()
    
    // Execute
    packager := New(cfg)
    path, err := packager.Pack(context.Background(), outputDir)
    
    // Verify
    if err != nil {
        t.Fatalf("Pack failed: %v", err)
    }
    
    // Check output exists
    if _, err := os.Stat(path); err != nil {
        t.Errorf("Output file not created: %v", err)
    }
}
```

### Coverage Goals

- Overall project: 60%+
- New packages: 70%+
- Critical paths: 80%+

Check coverage:

```bash
go test ./... -cover
go tool cover -html=coverage.out
```

## Documentation

### Code Comments

```go
// PackageManager handles package creation and distribution.
// It supports multiple output formats and signing options.
type PackageManager struct {
    config *Config
}

// Pack creates a package in the specified format.
// Returns the path to the created package or an error.
func (pm *PackageManager) Pack(format string) (string, error) {
    // Implementation
}
```

### README Updates

Update README.md when adding:
- New package formats
- New commands
- New configuration options

### Documentation Files

Update relevant docs in `docs/`:
- `README.md` - Main documentation
- `PACKAGE_FORMATS.md` - Package format details
- `CODE_SIGNING.md` - Signing instructions
- `EXAMPLES.md` - Usage examples
- `TROUBLESHOOTING.md` - Common issues

## Release Process

Releases are managed by maintainers:

1. Update version in code
2. Update CHANGELOG.md
3. Create git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tag: `git push origin v1.0.0`
5. GitHub Actions builds and publishes release

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/scttfrdmn/bagboy/issues)
- **Discussions**: [GitHub Discussions](https://github.com/scttfrdmn/bagboy/discussions)
- **Documentation**: [docs/](docs/)

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Assume good intentions

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to bagboy! 🎉
