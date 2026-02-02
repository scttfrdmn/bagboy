---
name: CI/CD Template Generation
about: Generate CI/CD pipeline templates for popular platforms
title: 'CI/CD Template Generation - v0.8.0'
labels: ['enhancement', 'v0.8.0', 'high-priority', 'ci-cd']
assignees: ''
---

## Summary
Add `bagboy ci init` command to generate CI/CD pipeline templates for popular platforms (GitHub Actions, GitLab CI, Jenkins, Azure DevOps).

## Problem Statement
Teams want to integrate bagboy into their CI/CD workflows but need to manually create pipeline configurations. We should provide:
- Ready-to-use pipeline templates
- Best practices for bagboy in CI/CD
- Platform-specific optimizations
- Automated setup for common scenarios

## Proposed Solution

### 1. CLI Command
```bash
bagboy ci init                    # Interactive template selection
bagboy ci init --platform=github # Generate GitHub Actions workflow
bagboy ci init --platform=gitlab # Generate GitLab CI configuration
bagboy ci init --platform=jenkins # Generate Jenkinsfile
bagboy ci init --platform=azure  # Generate Azure DevOps pipeline
```

### 2. Template Structure
```
templates/ci/
├── github-actions/
│   ├── release.yml
│   ├── pr-check.yml
│   └── nightly.yml
├── gitlab-ci/
│   └── .gitlab-ci.yml
├── jenkins/
│   └── Jenkinsfile
└── azure-devops/
    └── azure-pipelines.yml
```

### 3. Example GitHub Actions Template
```yaml
name: Bagboy Release Pipeline
on:
  push:
    tags: ['v*']

jobs:
  build-matrix:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        arch: [amd64, arm64]
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Install bagboy
      run: curl -fsSL bagboy.sh/install | bash
    
    - name: Validate dependencies
      run: bagboy deps check
    
    - name: Build binaries
      run: |
        GOOS=${{ matrix.os }} GOARCH=${{ matrix.arch }} \
        go build -o dist/myapp-${{ matrix.os }}-${{ matrix.arch }}
    
    - name: Package all formats
      run: bagboy pack --all
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: packages-${{ matrix.os }}-${{ matrix.arch }}
        path: dist/

  deploy:
    needs: build-matrix
    runs-on: ubuntu-latest
    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v3
    
    - name: Deploy packages
      run: bagboy publish
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Acceptance Criteria
- [ ] `bagboy ci init` command with interactive platform selection
- [ ] Support for GitHub Actions, GitLab CI, Jenkins, Azure DevOps
- [ ] Templates include dependency checking, building, packaging, and deployment
- [ ] Platform-specific optimizations (caching, matrix builds, etc.)
- [ ] Templates support both PR checks and release workflows
- [ ] Generated templates are immediately usable without modification
- [ ] Documentation for customizing generated templates
- [ ] Integration tests that validate generated templates

## Implementation Tasks
- [ ] Design CI command structure and flags
- [ ] Create template engine for CI/CD file generation
- [ ] Implement GitHub Actions templates (release, PR, nightly)
- [ ] Implement GitLab CI templates
- [ ] Implement Jenkins pipeline templates
- [ ] Implement Azure DevOps templates
- [ ] Add interactive platform selection
- [ ] Create template customization options
- [ ] Write comprehensive tests
- [ ] Update documentation with CI/CD integration guide

## Template Features
Each template should include:
- **Dependency validation** - `bagboy deps check`
- **Multi-platform builds** - Matrix strategy for OS/arch combinations
- **Artifact caching** - Platform-specific caching strategies
- **Security scanning** - Integration with security tools
- **Deployment coordination** - Staged deployment with rollback
- **Notification integration** - Slack/Teams/email notifications

## Testing Strategy
- Unit tests for template generation logic
- Integration tests that run generated templates
- Validation tests for template syntax
- End-to-end tests with actual CI/CD platforms

## Definition of Done
- All acceptance criteria met
- Templates tested on actual CI/CD platforms
- Documentation includes setup guides for each platform
- Code reviewed and approved
- Integration with existing bagboy commands working

## Related Issues
- Requires #37 (Core Dependency System) for `bagboy deps check`
- Blocks #45 (CI/CD Validation & Testing)
- Related to #42 (Container & Docker Integration)

## Estimated Effort
**Medium** (1 week)

## Priority
**High** - Critical for CI/CD adoption
