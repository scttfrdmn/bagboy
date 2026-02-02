---
name: Core Dependency System
about: Implement comprehensive dependency management system
title: 'Core Dependency System - v0.7.0'
labels: ['enhancement', 'v0.7.0', 'high-priority', 'dependencies']
assignees: ''
---

## Summary
Implement core dependency management system with validation and resolution capabilities across all package formats.

## Problem Statement
Currently bagboy has limited dependency handling - only basic declarations in some package formats. We need a comprehensive system that:
- Validates dependencies across platforms
- Supports version constraints
- Maps dependencies between package formats
- Provides clear error messages for missing dependencies

## Proposed Solution

### 1. Enhanced Configuration Schema
```yaml
# bagboy.yaml
dependencies:
  runtime:
    - name: "libc"
      version: ">=2.17"
      platforms: ["linux"]
      optional: false
    - name: "vcredist"
      version: ">=14.0"
      platforms: ["windows"]
      optional: false
  
  build:
    - name: "gcc"
      version: ">=7.0"
      optional: true
    - name: "docker"
      version: ">=20.0"
      optional: true

  package_specific:
    deb:
      depends: ["libc6 (>= 2.17)", "libssl3"]
      recommends: ["curl"]
    rpm:
      requires: ["glibc >= 2.17", "openssl-libs"]
    brew:
      depends_on: ["openssl@3"]
```

### 2. Core Implementation
```go
// pkg/deps/manager.go
type DependencyManager struct {
    Runtime         []Dependency `yaml:"runtime"`
    Build           []Dependency `yaml:"build"`
    PackageSpecific map[string]PackageDeps `yaml:"package_specific"`
}

type Dependency struct {
    Name      string   `yaml:"name"`
    Version   string   `yaml:"version"`
    Platforms []string `yaml:"platforms"`
    Optional  bool     `yaml:"optional"`
}

type PackageDeps struct {
    Depends    []string `yaml:"depends"`
    Recommends []string `yaml:"recommends"`
    Suggests   []string `yaml:"suggests"`
    Conflicts  []string `yaml:"conflicts"`
}
```

## Acceptance Criteria
- [ ] Add `dependencies` section to bagboy.yaml schema with validation
- [ ] Support runtime, build, and package-specific dependencies
- [ ] Implement dependency validation engine with clear error messages
- [ ] Add version constraint parsing (>=, <=, ~, ^, exact)
- [ ] Cross-platform dependency mapping (Linux/macOS/Windows)
- [ ] Integration with existing config loading system
- [ ] Comprehensive unit tests with >80% coverage
- [ ] Documentation with examples for all dependency types

## Implementation Tasks
- [ ] Design and implement dependency data structures
- [ ] Add dependency parsing to config loader
- [ ] Implement version constraint validation
- [ ] Create platform-specific dependency resolution
- [ ] Add dependency validation to `bagboy validate` command
- [ ] Write comprehensive tests
- [ ] Update documentation and examples

## Testing Strategy
- Unit tests for dependency parsing and validation
- Integration tests with various version constraint formats
- Cross-platform testing for platform-specific dependencies
- Error handling tests for invalid configurations

## Definition of Done
- All acceptance criteria met
- Tests passing with >80% coverage
- Documentation updated
- Code reviewed and approved
- Integration with existing bagboy commands working

## Related Issues
- Depends on completion of v0.6.0 quality improvements
- Blocks #38 (Dependency Commands)
- Blocks #39 (Package Format Integration)

## Estimated Effort
**Large** (1-2 weeks)

## Priority
**High** - Foundation for all dependency management features
