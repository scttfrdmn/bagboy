# API Reference

## Core Interfaces

### Packager Interface
```go
type Packager interface {
    Pack(ctx context.Context, cfg *config.Config) (string, error)
    Name() string
    Validate(cfg *config.Config) error
}
```

### Registry
```go
type Registry struct {
    packagers map[string]Packager
}

func NewRegistry() *Registry
func (r *Registry) Register(p Packager)
func (r *Registry) Get(name string) (Packager, bool)
func (r *Registry) List() []string
func (r *Registry) Count() int
func (r *Registry) PackAll(ctx context.Context, cfg *config.Config) (map[string]string, error)
```

## Configuration

### Config Structure
```go
type Config struct {
    Name        string            `yaml:"name"`
    Version     string            `yaml:"version"`
    Description string            `yaml:"description"`
    Homepage    string            `yaml:"homepage"`
    License     string            `yaml:"license"`
    Author      string            `yaml:"author"`
    Binaries    map[string]string `yaml:"binaries"`
    GitHub      GitHubConfig      `yaml:"github"`
    Installer   InstallerConfig   `yaml:"installer"`
    Packages    PackagesConfig    `yaml:"packages"`
    Signing     SigningConfig     `yaml:"signing"`
}
```

### GitHub Configuration
```go
type GitHubConfig struct {
    Owner    string        `yaml:"owner"`
    Repo     string        `yaml:"repo"`
    TokenEnv string        `yaml:"token_env"`
    Release  ReleaseConfig `yaml:"release"`
    Tap      TapConfig     `yaml:"tap"`
    Bucket   BucketConfig  `yaml:"bucket"`
    Winget   WingetConfig  `yaml:"winget"`
}
```

### Signing Configuration
```go
type SigningConfig struct {
    MacOS    MacOSSigningConfig    `yaml:"macos"`
    Windows  WindowsSigningConfig  `yaml:"windows"`
    Linux    LinuxSigningConfig    `yaml:"linux"`
    Sigstore SigstoreConfig       `yaml:"sigstore"`
    SignPath SignPathConfig       `yaml:"signpath"`
    Git      GitSigningConfig     `yaml:"git"`
}
```

## Error Handling

### BagboyError
```go
type BagboyError struct {
    Type        ErrorType `json:"type"`
    Code        string    `json:"code"`
    Message     string    `json:"message"`
    Details     string    `json:"details,omitempty"`
    Suggestions []string  `json:"suggestions,omitempty"`
    Cause       error     `json:"-"`
}
```

### Error Types
```go
const (
    ErrorTypeValidation    ErrorType = "validation"
    ErrorTypeConfiguration ErrorType = "configuration"
    ErrorTypeDependency    ErrorType = "dependency"
    ErrorTypeFileSystem    ErrorType = "filesystem"
    ErrorTypeNetwork       ErrorType = "network"
    ErrorTypeExternal      ErrorType = "external"
    ErrorTypeInternal      ErrorType = "internal"
)
```

## UI Utilities

### Progress Tracking
```go
type ProgressBar struct {
    total   int
    current int
    width   int
    prefix  string
}

func NewProgressBar(total int, prefix string) *ProgressBar
func (pb *ProgressBar) Update(current int)
func (pb *ProgressBar) Increment()
func (pb *ProgressBar) Finish()
```

### Spinner
```go
type Spinner struct {
    chars   []string
    current int
    message string
    active  bool
}

func NewSpinner(message string) *Spinner
func (s *Spinner) Start()
func (s *Spinner) Stop()
```

### Table Display
```go
type Table struct {
    headers []string
    rows    [][]string
    widths  []int
}

func NewTable(headers []string) *Table
func (t *Table) AddRow(row []string)
func (t *Table) Print()
```

### Message Functions
```go
func Success(message string)
func Warning(message string)
func Error(message string)
func Info(message string)
func Header(message string)
func Confirm(message string) bool
func Select(message string, options []string) int
```

## Packager Implementations

### Homebrew Packager
```go
type BrewPackager struct{}

func New() *BrewPackager
func (b *BrewPackager) Name() string
func (b *BrewPackager) Pack(ctx context.Context, cfg *config.Config) (string, error)
func (b *BrewPackager) Validate(cfg *config.Config) error
```

### Scoop Packager
```go
type ScoopPackager struct{}

func New() *ScoopPackager
func (s *ScoopPackager) Name() string
func (s *ScoopPackager) Pack(ctx context.Context, cfg *config.Config) (string, error)
func (s *ScoopPackager) Validate(cfg *config.Config) error
```

### DEB Packager
```go
type DEBPackager struct{}

func New() *DEBPackager
func (d *DEBPackager) Name() string
func (d *DEBPackager) Pack(ctx context.Context, cfg *config.Config) (string, error)
func (d *DEBPackager) Validate(cfg *config.Config) error
```

## Signing

### Signer Interface
```go
type Signer struct {
    config *config.Config
}

func NewSigner(cfg *config.Config) *Signer
func (s *Signer) SignAllBinaries(ctx context.Context) error
func (s *Signer) SignBinary(ctx context.Context, binaryPath string) error
func (s *Signer) CheckSetup(ctx context.Context) error
```

## Benchmarking

### Profiler
```go
type Profiler struct {
    results map[string]BenchmarkResult
}

type BenchmarkResult struct {
    Name           string
    Duration       time.Duration
    MemoryUsage    int64
    AllocCount     int64
    ThroughputOps  int64
}

func NewProfiler() *Profiler
func (p *Profiler) BenchmarkPackager(packager packager.Packager, cfg *config.Config) BenchmarkResult
func (p *Profiler) BenchmarkAll(registry *packager.Registry, cfg *config.Config) map[string]BenchmarkResult
```

## Requirements Checking

### Checker
```go
type Checker struct{}

func NewChecker() *Checker
func (c *Checker) CheckAll() map[string]RequirementResult
func (c *Checker) CheckFormat(format string) RequirementResult
```

### Requirement Result
```go
type RequirementResult struct {
    Available bool
    Version   string
    Path      string
    Error     error
}
```

## Deployment

### Deployer
```go
type Deployer struct {
    config *config.Config
}

func NewDeployer(cfg *config.Config) *Deployer
func (d *Deployer) DeployAll(ctx context.Context) error
func (d *Deployer) DeployToTarget(ctx context.Context, target string) error
```

## Usage Examples

### Creating a Custom Packager
```go
type MyPackager struct{}

func (m *MyPackager) Name() string {
    return "myformat"
}

func (m *MyPackager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
    // Implementation here
    return "output/path", nil
}

func (m *MyPackager) Validate(cfg *config.Config) error {
    // Validation logic
    return nil
}

// Register with registry
registry := packager.NewRegistry()
registry.Register(&MyPackager{})
```

### Using the API Programmatically
```go
package main

import (
    "context"
    "github.com/scttfrdmn/bagboy/pkg/config"
    "github.com/scttfrdmn/bagboy/pkg/packager"
    "github.com/scttfrdmn/bagboy/pkg/packager/brew"
)

func main() {
    cfg, err := config.Load("bagboy.yaml")
    if err != nil {
        panic(err)
    }

    registry := packager.NewRegistry()
    registry.Register(brew.New())

    ctx := context.Background()
    results, err := registry.PackAll(ctx, cfg)
    if err != nil {
        panic(err)
    }

    for name, path := range results {
        fmt.Printf("Created %s: %s\n", name, path)
    }
}
```
