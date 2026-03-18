# Changelog

All notable changes to bagboy will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.0] - 2026-03-18

### Added
- `bagboy validate --format <name>` and `--all-formats` flags for per-packager validation results displayed in a table
- `bagboy validate --check-deps` flag to show system dependency availability alongside validation
- `bagboy init --template <name>` flag to generate `bagboy.yaml` from project-type presets
- `bagboy init --list-templates` to display all available templates
- `bagboy deploy --generate-ci [github|gitlab]` for automated CI/CD pipeline file generation
- `pkg/templates` package with embedded YAML templates for `go-cli`, `go-service`, `node-cli`, `rust-cli` project types
- `pkg/cicd` package with `Provider` interface and global registry
- `pkg/cicd/github` provider: generates `.github/workflows/release.yml`
- `pkg/cicd/gitlab` provider: generates `.gitlab-ci.yml`
- Expanded test coverage: `pkg/vm` mock-provider tests covering all `Manager` paths
- CLI integration tests for `cmd_sign`, `cmd_vm`, `cmd_check` subcommands
- `pkg/signing` tests: SignPath field validation, mock-cosign PATH tests, notarize gating
- `pkg/github` httptest-backed tests: `CreateRelease`, `UpdateTap`, `UpdateBucket`, 404/422 error handling
- `pkg/packager/{installer,dmg,docker}` content-verification and error-path tests
- `pkg/packager/integration_test.go` with `//go:build integration` tag for pack→validate→checksum chain

### Fixed
- `bagboy validate` no longer silently skips per-format validation

## [0.8.0-dev] - 2026-03-18

### Added
- `internal/version` package with ldflags-injectable Version, Commit, Date, BuiltBy
- `.golangci.yml` with revive, gocyclo, errcheck, staticcheck, and more
- `.github/workflows/release.yml` for tag-based releases with pre-release detection
- `CHANGELOG.md` in keepachangelog format
- `context.Context` parameter to `config.Load()` for future cancellation support
- Context-aware `Spinner.Start(ctx)` prevents goroutine leaks on cancellation
- Structured logging via `log/slog` with `--verbose` / `--debug` persistent flags
- Godoc comments on all exported symbols across all packages
- Signal context (`SIGINT`/`SIGTERM`) in `main()` via `signal.NotifyContext`

### Changed
- Split 1 200-line `cmd/bagboy/main.go` into 14 focused files to fix gocyclo violations
- `runPackFlags` replaced 21 sequential `if` branches with a map dispatch (complexity: ~21 → ~3)
- Go matrix in CI updated to `['1.22', '1.23', '1.24']`
- `actions/setup-go@v4` → `@v5` (built-in cache, dropped separate `actions/cache@v3`)
- `golangci/golangci-lint-action@v3` → `@v6`
- `codecov/codecov-action@v3` → `@v4`
- CI build step now uses `make build` (includes ldflags)
- Added `fetch-depth: 0` to all CI checkout steps (required for `git describe`)
- Makefile `build` and `build-all` targets now inject VERSION, COMMIT, DATE via ldflags
- `bagboy version` now prints real commit and date from build-time injection

### Fixed
- Root-level compiled `bagboy` binary now excluded via `.gitignore`
- Goroutine leak in `Spinner.Start()` when caller context is cancelled

### Removed
- Tracking/planning `.md` files from repository root (preserved in git history)
- Hardcoded version string `"0.7.0-dev"` in main.go

## [0.7.0] - 2026-02-08

### Added
- VM build environment support via Docker, Multipass, and Vagrant providers
- `bagboy vm check` subcommand for verifying VM provider availability
- `pkg/checksum` package for SHA256 checksum generation and verification
- `pkg/vm` package with Provider interface and Manager
- VM build configuration in `bagboy.yaml` (`vm.provider`, `vm.docker.image`)
- Enhanced dependency resolution with conflict detection
- Dependency cache for repeated resolution performance
- Dependency lock file (`bagboy.lock`) generation
- `bagboy deps resolve` subcommand for conflict resolution

### Changed
- Improved test coverage across all packages
- Enhanced error messages with recovery suggestions

## [0.6.0] - 2025-11-01

### Added
- Core dependency management system (`pkg/deps`)
- `bagboy deps check`, `list`, `install` subcommands
- System dependency detection per platform (Linux, macOS, Windows)
- Package manager dependency support (brew, apt, choco, etc.)
- Runtime requirement validation

## [0.5.0] - 2025-09-01

### Added
- SignPath.io cloud-based code signing support
- Sigstore/Cosign keyless signing integration
- Git commit and tag signing configuration
- `bagboy sign` subcommand with multi-provider support

## [0.4.0] - 2025-07-01

### Added
- `bagboy deploy` subcommand for registry and repository deployment
- `bagboy benchmark` subcommand for performance profiling
- `pkg/deploy` orchestration package
- `pkg/benchmark` profiling package with benchmark suite

## [0.3.0] - 2025-05-01

### Added
- `bagboy validate` subcommand for configuration validation
- `bagboy check` subcommand for system requirement checking
- `pkg/requirements` package for external tool detection
- `pkg/errors` package with structured errors and recovery suggestions

## [0.2.0] - 2025-03-01

### Added
- `bagboy publish` complete workflow command (pack + GitHub release)
- GitHub API integration (`pkg/github`) for releases, taps, buckets, Winget PRs
- Homebrew tap auto-update after release
- Scoop bucket auto-update after release
- Winget community PR submission

## [0.1.0] - 2025-01-01

### Added
- Initial release of bagboy
- `bagboy init` with auto-detection for Go, Node.js, Rust, Python projects
- `bagboy pack` with 20 package format support:
  - Package managers: Homebrew, Scoop, Chocolatey, Winget
  - Linux packages: DEB, RPM, AppImage, Snap, Flatpak
  - Containers: Docker, Apptainer
  - Language packages: npm, PyPI, Cargo, Nix, Spack
  - Platform installers: DMG, MSI, MSIX, curl\|bash
- Single `bagboy.yaml` configuration file
- `pkg/config` with YAML parsing and validation
- `pkg/packager` interface and Registry pattern
- `pkg/ui` terminal utilities (progress bars, spinners, tables)
- Apache 2.0 license

[Unreleased]: https://github.com/scttfrdmn/bagboy/compare/v0.8.0-dev...HEAD
[0.8.0-dev]: https://github.com/scttfrdmn/bagboy/compare/v0.7.0...v0.8.0-dev
[0.7.0]: https://github.com/scttfrdmn/bagboy/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/scttfrdmn/bagboy/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/scttfrdmn/bagboy/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/scttfrdmn/bagboy/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/scttfrdmn/bagboy/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/scttfrdmn/bagboy/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/scttfrdmn/bagboy/releases/tag/v0.1.0
