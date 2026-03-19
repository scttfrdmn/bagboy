# Changelog

All notable changes to bagboy will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.6] - 2026-03-18

### Added
- `pkg/packager/msi`: cross-platform MSI builds via `wixl` from `msitools` (`brew install msitools` on macOS, `apt install msitools` on Linux) — no Wine or Windows required
- `pkg/packager/msi`: multi-arch support — one MSI is produced per available `windows-*` binary (`windows-amd64` → x64, `windows-arm64` → arm64, `windows-386` → x86); output is `dist/msi/`
- `pkg/packager/msi`: removed `WixUIExtension`/`WixUI_InstallDir`/`WixVariable` from WiX template for `wixl` compatibility (WiX toolset builds are unaffected)

## [0.9.5] - 2026-03-18

### Fixed
- `cmd/bagboy publish`: pack errors (e.g. "MSI build tools not found") now emit a `⚠️ Skipping` warning and continue rather than aborting the entire publish workflow (closes #41)
- `pkg/github`: `CreateRelease` now detects HTTP 422 `tag_name: already_exists` and falls back to fetching the existing release, then uploads assets to it — re-running `publish` after a partial failure no longer requires manually deleting the release (closes #42)

## [0.9.4] - 2026-03-18

### Fixed
- `pkg/packager/brew`: `{{.Version}}` in `installer.base_url` was written literally into formula download URLs instead of being substituted with the configured version (closes #37)
- `cmd/bagboy publish`: ignored the `packages:` config section and attempted to build all 20 formats unconditionally; now only builds the formats explicitly listed under `packages:` — falls back to all formats when the section is absent (closes #38)
- `pkg/packager/chocolatey`: nupkg output filename used dot separator (`name.version.nupkg`) causing a `zip I/O error` because the parent directory did not exist; changed to dash separator (`name-version.nupkg`) (closes #39)
- `pkg/packager/dmg`: on macOS with `hdiutil` available, `Pack()` now executes the `hdiutil` pipeline to produce a real compressed disk image instead of a 103-byte scaffold stub; non-macOS and CI environments retain the scaffold behavior (closes #40)

### Internal
- `pkg/config`: `PackagesConfig` gains a custom `UnmarshalYAML` that tracks which format keys were explicitly present in YAML; new `ConfiguredNames()` method exposes the list
- `pkg/packager`: `Registry.PackSelected()` packs only a named subset of registered packagers

## [0.9.3] - 2026-03-18

### Tests
- `pkg/packager`: integration tests for dmg, cargo, pypi, nix, spack, and winget packagers — all pure-Go and macOS-compatible with no external tools required
- `Makefile`: add `make test-integration` target

## [0.9.2] - 2026-03-18

### Security
- `pkg/packager/deb`: shell-quote `sourceDir`/`outputPath` in `buildDebWithVM` to prevent shell injection via crafted project names
- `pkg/packager/rpm`: shell-quote `buildDir`/`specPath` in `buildRPMWithVM` (same class of issue)
- `pkg/signing`: upgrade Windows timestamp authority URL from `http://` to `https://timestamp.digicert.com` to prevent MITM on timestamp responses
- `pkg/config`: add regexp validation for `name` and `version` fields in `Validate()` to block path traversal characters before they reach `filepath.Join` calls in packagers
- `pkg/config`: emit `slog.Warn` when `app_password` or `api_token` fields appear to contain plaintext secrets rather than env-var references

## [0.9.1] - 2026-03-18

### Fixed
- Docker Dockerfile: pin Alpine base image to `3.21` (was unpinned `latest`)
- Docker Compose template: remove deprecated `version:` field (Docker Compose V2)
- Debian/Apptainer builder image: `ubuntu:22.04` → `ubuntu:24.04` (current LTS)
- RPM builder image: `fedora:38` → `fedora:41` (38 EOL Nov 2023)
- npm `engines.node`: `>=14.0.0` → `>=20.0.0` (Node 14 and 18 both EOL by Mar 2026)
- Snap: `core22` → `core24` base (Ubuntu 24.04, released May 2024)
- Flatpak: runtime-version `22.08` → `24.08`
- Cargo: `reqwest` `0.11` → `0.12`
- PyPI: add Python 3.12/3.13 classifiers; raise `python_requires`/`requires-python` to `>=3.9`; `setuptools>=45` → `>=68`
- MSIX: `MaxVersionTested` `10.0.22000.0` → `10.0.26100.0` (Windows 11 24H2)
- `pkg/ui` Spinner: data race on `active` field resolved with `sync/atomic`
- `cmd/bagboy` tests: Cobra flag-state leakage between test cases resolved with `resetAllFlags` helper; flag values are now reset to defaults before every `Execute()` call
- `pkg/packager/deb` tests: skip gracefully when `dpkg-deb` is not on PATH (macOS, CI)
- Integration tests: `--deb` in `PackMultiple` conditioned on `dpkg-deb` availability; cross-platform error allowlist broadened to cover `appimagetool`/`mksquashfs`; description-string assertions corrected to match fixture data

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
