# Dependency Management Guide

bagboy provides comprehensive dependency management for your software packages, ensuring all required dependencies are properly declared and resolved across different platforms and package managers.

## Overview

The dependency system supports:
- **System dependencies** - OS-level libraries and tools
- **Package manager dependencies** - Homebrew, APT, YUM, Chocolatey, Scoop
- **Runtime dependencies** - Node.js, Python, Go version requirements
- **Dependency resolution** - Automatic conflict detection and resolution
- **Lock files** - Reproducible builds with `bagboy.lock`

## Configuration

Add dependencies to your `bagboy.yaml`:

```yaml
name: myapp
version: 1.0.0

dependencies:
  # System-level dependencies by platform
  system:
    linux: ["libc6", "libssl3"]
    macos: ["openssl"]
    windows: ["vcredist2019"]
  
  # Package manager specific dependencies
  package_managers:
    homebrew: ["openssl@3", "curl"]
    apt: ["libssl-dev", "libcurl4-openssl-dev"]
    yum: ["openssl-devel", "libcurl-devel"]
    chocolatey: ["openssl", "curl"]
    scoop: ["openssl", "curl"]
    
  # Runtime version requirements
  runtime:
    node: ">=18.0.0"
    python: ">=3.8"
    go: ">=1.19"
```

## Commands

### Check Dependencies

Verify all dependencies are available:

```bash
bagboy deps check
```

Output:
```
✅ node: v20.0.0 (satisfies >=18.0.0)
✅ python: 3.11.0 (satisfies >=3.8)
⚠️  openssl: not found
```

### List Dependencies

Display configured dependency tree:

```bash
bagboy deps list
```

Output:
```
Dependencies for myapp v1.0.0
┌──────────┬─────────────┬────────────┬──────────┐
│ Name     │ Type        │ Constraint │ Status   │
├──────────┼─────────────┼────────────┼──────────┤
│ node     │ runtime     │ >=18.0.0   │ ✅ Found │
│ python   │ runtime     │ >=3.8      │ ✅ Found │
│ openssl  │ homebrew    │ latest     │ ⚠️  Missing │
└──────────┴─────────────┴────────────┴──────────┘
```

### Install Dependencies

Install missing dependencies automatically:

```bash
bagboy deps install
```

This will:
1. Detect your OS and package manager
2. Install missing dependencies
3. Verify installation
4. Update lock file

### Resolve Conflicts

Handle dependency conflicts:

```bash
bagboy deps resolve
```

This generates a `bagboy.lock` file with resolved versions.

## Version Constraints

bagboy supports semantic versioning constraints:

| Constraint | Meaning | Example |
|------------|---------|---------|
| `>=1.0.0` | Greater than or equal | `>=18.0.0` |
| `<=2.0.0` | Less than or equal | `<=3.11` |
| `~1.2.3` | Patch updates only | `~1.2.0` matches 1.2.x |
| `^1.2.3` | Minor updates only | `^1.2.0` matches 1.x.x |
| `1.2.3` | Exact version | `1.2.3` only |

## Lock Files

The `bagboy.lock` file ensures reproducible builds:

```yaml
version: "1.0"
generated: 2026-02-08T18:00:00Z
dependencies:
  node:
    version: v20.0.0
    constraint: ">=18.0.0"
    source: runtime
    resolved: 2026-02-08T18:00:00Z
  openssl:
    version: 3.1.0
    constraint: latest
    source: homebrew
    resolved: 2026-02-08T18:00:00Z
```

Commit this file to version control for consistent builds across environments.

## Platform-Specific Dependencies

Declare different dependencies per platform:

```yaml
dependencies:
  system:
    linux: ["libssl3", "libcurl4"]
    macos: ["openssl@3"]
    windows: ["vcredist2019", "openssl"]
```

bagboy automatically selects the correct dependencies for the target platform.

## Package Manager Mapping

bagboy intelligently maps dependencies across package managers:

| Dependency | Homebrew | APT | YUM | Chocolatey |
|------------|----------|-----|-----|------------|
| OpenSSL | `openssl@3` | `libssl-dev` | `openssl-devel` | `openssl` |
| cURL | `curl` | `libcurl4-openssl-dev` | `libcurl-devel` | `curl` |
| Python | `python@3.11` | `python3` | `python3` | `python` |

## Examples

### Node.js Application

```yaml
dependencies:
  runtime:
    node: ">=18.0.0"
  package_managers:
    homebrew: ["node"]
    apt: ["nodejs", "npm"]
    chocolatey: ["nodejs"]
```

### Python Application

```yaml
dependencies:
  runtime:
    python: ">=3.8"
  system:
    linux: ["python3-dev", "python3-pip"]
    macos: ["python@3.11"]
    windows: ["python"]
```

### Go Application with C Dependencies

```yaml
dependencies:
  runtime:
    go: ">=1.19"
  system:
    linux: ["gcc", "libc6-dev"]
    macos: ["gcc"]
    windows: ["mingw"]
```

### Native Application with SSL

```yaml
dependencies:
  system:
    linux: ["libssl3", "libcrypto3"]
    macos: ["openssl@3"]
    windows: ["openssl"]
  package_managers:
    apt: ["libssl-dev"]
    yum: ["openssl-devel"]
    homebrew: ["openssl@3"]
```

## Best Practices

1. **Be Specific**: Use version constraints to avoid breaking changes
2. **Test Across Platforms**: Verify dependencies on Linux, macOS, and Windows
3. **Commit Lock Files**: Ensure reproducible builds
4. **Document Requirements**: Add comments explaining why dependencies are needed
5. **Minimize Dependencies**: Only include what's truly required

## Troubleshooting

### Dependency Not Found

```bash
⚠️  openssl: not found
```

**Solution**: Install manually or run `bagboy deps install`

### Version Conflict

```bash
❌ Conflict: node requires >=18.0.0 but found v16.0.0
```

**Solution**: Upgrade the conflicting dependency or adjust constraints

### Lock File Out of Date

```bash
⚠️  bagboy.lock is out of date
```

**Solution**: Run `bagboy deps resolve` to regenerate

## Integration with Packaging

Dependencies are automatically included in package metadata:

- **DEB/RPM**: Added to `Depends:` field
- **Homebrew**: Added to `depends_on` in formula
- **Chocolatey**: Added to `<dependencies>` in nuspec
- **Docker**: Added to `RUN` commands in Dockerfile

## CI/CD Integration

Check dependencies in your CI pipeline:

```yaml
# GitHub Actions
- name: Check dependencies
  run: bagboy deps check
  
- name: Install dependencies
  run: bagboy deps install
```

## Advanced Features

### Dependency Caching

bagboy caches dependency resolution results for faster builds:

```bash
# Clear cache
rm -rf ~/.bagboy/cache/deps
```

### Custom Package Managers

Support for additional package managers coming in future releases:
- Nix
- Spack
- Conda
- vcpkg

---

For more information, see:
- [Configuration Guide](README.md)
- [Package Formats](PACKAGE_FORMATS.md)
- [Examples](EXAMPLES.md)
