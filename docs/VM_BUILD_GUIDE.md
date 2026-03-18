# VM Build Support - User Guide

## Overview

bagboy now supports building packages in VMs/containers, allowing you to build platform-specific packages from any OS.

## Quick Start

### 1. Enable VM Support

Add to your `bagboy.yaml`:

```yaml
vm:
  enabled: true
  provider: docker
```

### 2. Build RPM on macOS

```bash
# Build RPM package using Docker
bagboy pack --rpm

# bagboy automatically:
# 1. Detects rpmbuild is missing
# 2. Spins up Fedora container
# 3. Installs rpm-build tools
# 4. Builds your RPM
# 5. Copies artifact to dist/
```

## Configuration

### Basic Configuration

```yaml
vm:
  enabled: true
  provider: docker  # Currently only docker supported
```

### Custom Docker Image

```yaml
vm:
  enabled: true
  provider: docker
  docker:
    image: fedora:38  # Use specific Fedora version
```

## Supported Packages

Currently VM support is available for:

- ✅ **RPM** - Builds on Fedora container
- 🚧 **DEB** - Coming soon (Ubuntu container)
- 🚧 **MSI** - Coming soon (Windows container)

## Requirements

### Docker

Install Docker Desktop:
- **macOS**: https://docs.docker.com/desktop/install/mac-install/
- **Windows**: https://docs.docker.com/desktop/install/windows-install/
- **Linux**: https://docs.docker.com/engine/install/

### Check Availability

```bash
# Check if Docker is available
bagboy vm check

# Output:
# 🎯 VM Status
# ─────────────
# ✅ ✅ Docker available
```

## Examples

### Example 1: Go CLI App

```yaml
name: mycli
version: 1.0.0
description: My CLI tool
license: MIT

binaries:
  linux-amd64: dist/mycli-linux-amd64

vm:
  enabled: true
  provider: docker

packages:
  rpm:
    vendor: "My Company"
    group: "Applications/System"
```

```bash
# Build binary
GOOS=linux GOARCH=amd64 go build -o dist/mycli-linux-amd64

# Build RPM (works on macOS!)
bagboy pack --rpm
```

### Example 2: Multi-Platform Build

```yaml
name: myapp
version: 1.0.0

binaries:
  linux-amd64: dist/myapp-linux-amd64
  darwin-amd64: dist/myapp-darwin-amd64

vm:
  enabled: true
  provider: docker
  docker:
    image: fedora:38

packages:
  rpm:
    vendor: "ACME Corp"
  brew:
    test: |
      system "#{bin}/myapp --version"
```

```bash
# Build all binaries
make build-all

# Package everything
bagboy pack --all

# RPM built in Docker, Homebrew built natively
```

## How It Works

### Without VM Support

```bash
$ bagboy pack --rpm
❌ rpmbuild not found - install rpm-build package
```

### With VM Support

```bash
$ bagboy pack --rpm
🐳 Using Docker to build RPM...
📦 Pulling fedora:38...
🔨 Installing rpm-build...
✅ Built: dist/myapp-1.0.0-1.x86_64.rpm
```

### Behind the Scenes

1. **Detection**: bagboy checks if `rpmbuild` is available
2. **Fallback**: If not found and VM enabled, uses Docker
3. **Container**: Spins up Fedora container with volume mount
4. **Build**: Installs tools and builds RPM inside container
5. **Extract**: Copies RPM from container to `dist/`
6. **Cleanup**: Container removed automatically

## Performance

### First Build
- **Time**: ~2-3 minutes (image pull + build)
- **Disk**: ~500MB (Fedora image)

### Subsequent Builds
- **Time**: ~30 seconds (cached image)
- **Disk**: No additional space

## Troubleshooting

### Docker Not Available

```bash
$ bagboy pack --rpm
❌ rpmbuild not found - install rpm-build package or enable VM support
```

**Solution**: Install Docker or enable VM support in `bagboy.yaml`

### Permission Denied

```bash
❌ VM build failed: permission denied
```

**Solution**: Ensure Docker daemon is running and you have permissions

### Image Pull Failed

```bash
❌ VM build failed: failed to pull image
```

**Solution**: Check internet connection or specify different image

## Advanced Usage

### Custom Build Commands

Future feature - specify custom build commands:

```yaml
vm:
  enabled: true
  provider: docker
  docker:
    image: fedora:38
    commands:
      rpm: |
        yum install -y rpm-build rpmdevtools
        rpmbuild --define "_topdir /work/dist/rpm-build" -bb /work/dist/rpm-build/SPECS/*.spec
```

### Multiple Providers

Future feature - use different providers for different packages:

```yaml
vm:
  enabled: true
  packages:
    rpm:
      provider: docker
      image: fedora:38
    deb:
      provider: multipass
      instance: "22.04"
```

## Comparison

### Native Build
**Pros**: Fast, no overhead  
**Cons**: Requires platform-specific tools

### VM Build
**Pros**: Works anywhere, reproducible  
**Cons**: Slower first build, requires Docker

## Best Practices

1. **Enable VM for CI/CD** - Consistent builds across environments
2. **Use specific image versions** - `fedora:38` not `fedora:latest`
3. **Cache Docker images** - Speeds up subsequent builds
4. **Test locally first** - Verify VM builds work before CI

## FAQ

**Q: Does this work on Apple Silicon (M1/M2)?**  
A: Yes! Docker handles architecture translation automatically.

**Q: Can I use Podman instead of Docker?**  
A: Not yet, but it's planned for a future release.

**Q: Will this slow down my builds?**  
A: First build is slower (~2min), subsequent builds are fast (~30s).

**Q: Do I need to install rpmbuild anymore?**  
A: No! That's the whole point - bagboy handles it for you.

## Next Steps

- Try building an RPM on macOS
- Enable VM support in your CI/CD
- Check out the [examples](../examples/vm-rpm-build.yaml)

---

**Need help?** Open an issue on GitHub or check the [troubleshooting guide](TROUBLESHOOTING.md).
