# bagboy Documentation

## Table of Contents
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Package Formats](#package-formats)
- [Code Signing](#code-signing)
- [GitHub Integration](#github-integration)
- [CLI Reference](#cli-reference)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Quick Start

1. **Install bagboy**
```bash
curl -fsSL bagboy.sh/install | bash
```

2. **Initialize your project**
```bash
cd your-project
bagboy init
```

3. **Build your binaries** for target platforms
```bash
# Example for Go project
GOOS=darwin GOARCH=amd64 go build -o dist/myapp-darwin-amd64
GOOS=linux GOARCH=amd64 go build -o dist/myapp-linux-amd64
GOOS=windows GOARCH=amd64 go build -o dist/myapp-windows-amd64.exe
```

4. **Create packages**
```bash
bagboy pack --all
```

5. **Publish everywhere**
```bash
bagboy publish
```

## Installation

### Via curl (Recommended)
```bash
curl -fsSL bagboy.sh/install | bash
```

### Via Homebrew
```bash
brew install scttfrdmn/tap/bagboy
```

### From Source
```bash
git clone https://github.com/scttfrdmn/bagboy
cd bagboy
make build
sudo cp bin/bagboy /usr/local/bin/
```

## Configuration

### Basic Configuration
Create `bagboy.yaml` in your project root:

```yaml
name: myapp
version: 1.0.0
description: My awesome application
homepage: https://myapp.com
license: MIT
author: Your Name <you@example.com>

binaries:
  darwin-amd64: dist/myapp-darwin-amd64
  linux-amd64: dist/myapp-linux-amd64
  windows-amd64: dist/myapp-windows-amd64.exe
```

### GitHub Integration
```yaml
github:
  owner: yourname
  repo: myapp
  token_env: GITHUB_TOKEN
  
  release:
    enabled: true
    generate_notes: true
  
  tap:
    enabled: true
    repo: yourname/homebrew-tap
    auto_create: true
    auto_commit: true
    auto_push: true
```

### Code Signing
```yaml
signing:
  macos:
    identity: "Developer ID Application: Your Name"
    notarize: true
  windows:
    certificate_thumbprint: ""  # Set via env var
  linux:
    gpg_key_id: ""  # Set via env var
```

## Package Formats

### Package Managers
- **Homebrew** (macOS) - Formula generation
- **Scoop** (Windows) - Manifest generation
- **Chocolatey** (Windows) - Package creation
- **Winget** (Windows) - Manifest generation

### Linux Packages
- **DEB** (Debian/Ubuntu) - Binary packages
- **RPM** (RedHat/CentOS) - Binary packages
- **AppImage** (Universal Linux) - Portable applications
- **Snap** (Ubuntu) - Containerized packages
- **Flatpak** (Linux) - Sandboxed applications

### Containers
- **Docker** - Container images
- **Apptainer** - HPC containers

### Language Packages
- **npm** (Node.js) - JavaScript packages
- **PyPI** (Python) - Python packages
- **Cargo** (Rust) - Rust crates
- **Nix** - Functional package manager
- **Spack** - HPC package manager

### Platform Installers
- **DMG** (macOS) - Disk images
- **MSI** (Windows) - Windows Installer
- **MSIX** (Windows) - Modern Windows packages
- **curl|bash** - Universal installer scripts

## Code Signing

### macOS Code Signing
1. **Join Apple Developer Program** ($99/year)
2. **Create Developer ID Certificate**
3. **Configure environment**:
```bash
export APPLE_DEVELOPER_ID="Developer ID Application: Your Name"
export APPLE_ID="your@email.com"
export APPLE_APP_PASSWORD="app-specific-password"
export APPLE_TEAM_ID="TEAM123456"
```

### Windows Code Signing
1. **Purchase code signing certificate**
2. **Install Windows SDK**
3. **Configure environment**:
```bash
export WINDOWS_CERT_THUMBPRINT="certificate-thumbprint"
```

### Linux Code Signing
1. **Generate GPG key**:
```bash
gpg --gen-key
```
2. **Configure environment**:
```bash
export GPG_KEY_ID="your-key-id"
```

## GitHub Integration

### Automatic Releases
bagboy can automatically:
- Create GitHub releases
- Upload all package artifacts
- Generate release notes
- Update Homebrew taps
- Update Scoop buckets
- Submit Winget PRs

### Setup
1. **Create GitHub token** with repo permissions
2. **Set environment variable**:
```bash
export GITHUB_TOKEN="your-token"
```
3. **Configure bagboy.yaml** (see Configuration section)

## CLI Reference

### Commands

#### `bagboy init`
Initialize a new bagboy project with auto-detection.
```bash
bagboy init                    # Auto-detect project
bagboy init --interactive      # Interactive setup
```

#### `bagboy pack`
Create packages for distribution.
```bash
bagboy pack --all              # All formats
bagboy pack --brew --scoop     # Specific formats
bagboy pack --deb --rpm        # Linux packages
bagboy pack --sign             # With code signing
```

#### `bagboy validate`
Validate configuration file.
```bash
bagboy validate                # Basic validation
bagboy validate --verbose      # Detailed info
```

#### `bagboy publish`
Complete publishing workflow.
```bash
bagboy publish                 # Full workflow
bagboy publish --dry-run       # Preview only
bagboy publish --skip-github   # Skip GitHub ops
```

#### `bagboy sign`
Code signing operations.
```bash
bagboy sign --check            # Check setup
bagboy sign --binary app       # Sign specific binary
```

### Command Aliases
- `pack` → `p`, `package`, `build`
- `init` → `i`, `new`, `create`
- `validate` → `v`, `check`, `verify`
- `publish` → `pub`, `release`, `deploy`

## Examples

### Go Application
```yaml
name: mygoapp
version: 1.2.3
description: My Go application
binaries:
  darwin-amd64: dist/mygoapp-darwin-amd64
  linux-amd64: dist/mygoapp-linux-amd64
  windows-amd64: dist/mygoapp-windows-amd64.exe
```

### Node.js Application
```yaml
name: mynodeapp
version: 1.0.0
description: My Node.js application
binaries:
  darwin-amd64: dist/mynodeapp-macos
  linux-amd64: dist/mynodeapp-linux
  windows-amd64: dist/mynodeapp-win.exe
packages:
  npm:
    main: index.js
    bin:
      mynodeapp: bin/mynodeapp
```

### Rust Application
```yaml
name: myrustapp
version: 0.1.0
description: My Rust application
binaries:
  darwin-amd64: target/release/myrustapp
  linux-amd64: target/release/myrustapp
  windows-amd64: target/release/myrustapp.exe
packages:
  cargo:
    edition: "2021"
    categories: ["command-line-utilities"]
```

## Troubleshooting

### Common Issues

#### "No bagboy configuration file found"
**Solution**: Run `bagboy init` to create a configuration file.

#### "Binary file not found"
**Solution**: Build your binaries first, then update paths in `bagboy.yaml`.

#### "GitHub token not found"
**Solution**: Set `GITHUB_TOKEN` environment variable.

#### "Code signing failed"
**Solution**: Check signing setup with `bagboy sign --check`.

#### "Package format not supported"
**Solution**: Check if required tools are installed (e.g., `rpmbuild` for RPM).

### Debug Mode
Enable verbose output for troubleshooting:
```bash
bagboy pack --all --verbose
bagboy validate --verbose
```

### Getting Help
- **Documentation**: https://bagboy.dev
- **GitHub Issues**: https://github.com/scttfrdmn/bagboy/issues
- **CLI Help**: `bagboy --help` or `bagboy [command] --help`

## Performance

### Benchmarks
bagboy includes built-in performance benchmarking:
```bash
bagboy benchmark               # Run all benchmarks
```

Typical performance (Apple M4 Pro):
- **Scoop**: ~44,000 ns/op (fastest)
- **Homebrew**: ~156,000 ns/op
- **DEB**: ~970,000 ns/op (most complex)

### Optimization Tips
1. **Parallel processing** - bagboy automatically parallelizes package creation
2. **Binary size** - Smaller binaries = faster packaging
3. **Incremental builds** - Only rebuild changed packages
4. **Local caching** - bagboy caches intermediate files

## Best Practices

### Project Structure
```
your-project/
├── bagboy.yaml              # Configuration
├── dist/                    # Built binaries
│   ├── myapp-darwin-amd64
│   ├── myapp-linux-amd64
│   └── myapp-windows-amd64.exe
├── assets/                  # Icons, metadata
│   └── icon.png
└── scripts/                 # Build scripts
    └── build.sh
```

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Build binaries
  run: make build-all

- name: Package and publish
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: bagboy publish
```

### Security
- **Never commit tokens** to version control
- **Use environment variables** for sensitive data
- **Enable code signing** for production releases
- **Verify checksums** for distributed binaries

---

For more information, visit [bagboy.dev](https://bagboy.dev) or run `bagboy --help`.
