# Package Format Guide

## Overview
bagboy supports 20+ package formats across different platforms and ecosystems. This guide provides detailed information about each format.

## Package Managers

### Homebrew (macOS)
**Format**: Ruby formula  
**Extension**: `.rb`  
**Platform**: macOS, Linux

#### Configuration
```yaml
packages:
  brew:
    test: |
      system "#{bin}/myapp --version"
    dependencies: []
    conflicts_with: []
```

#### Generated Files
- `Formula/myapp.rb` - Homebrew formula
- Automatically calculates SHA256 checksums
- Supports multiple architectures

#### Installation
```bash
brew install yourname/tap/myapp
```

### Scoop (Windows)
**Format**: JSON manifest  
**Extension**: `.json`  
**Platform**: Windows

#### Configuration
```yaml
packages:
  scoop:
    bin: myapp.exe
    shortcuts: [[myapp.exe, MyApp]]
    checkver:
      github: yourname/myapp
```

#### Generated Files
- `bucket/myapp.json` - Scoop manifest
- Supports portable and installer packages
- Auto-update capabilities

#### Installation
```bash
scoop bucket add yourname https://github.com/yourname/scoop-bucket
scoop install myapp
```

### Chocolatey (Windows)
**Format**: NuGet package  
**Extension**: `.nupkg`  
**Platform**: Windows

#### Configuration
```yaml
packages:
  chocolatey:
    package_source_url: https://github.com/yourname/myapp
    docs_url: https://myapp.com/docs
    bug_tracker_url: https://github.com/yourname/myapp/issues
```

#### Generated Files
- `myapp.nuspec` - Package specification
- `tools/chocolateyInstall.ps1` - Installation script
- `myapp.nupkg` - Final package

#### Installation
```bash
choco install myapp
```

### Winget (Windows)
**Format**: YAML manifests  
**Extension**: `.yaml`  
**Platform**: Windows

#### Configuration
```yaml
packages:
  winget:
    package_identifier: YourName.MyApp
    publisher: Your Name
    minimum_os_version: 10.0.0.0
```

#### Generated Files
- `version.yaml` - Version manifest
- `installer.yaml` - Installer manifest
- `locale.yaml` - Localization manifest

#### Installation
```bash
winget install YourName.MyApp
```

## Linux Packages

### DEB (Debian/Ubuntu)
**Format**: Debian package  
**Extension**: `.deb`  
**Platform**: Debian, Ubuntu, derivatives

#### Configuration
```yaml
packages:
  deb:
    maintainer: you@example.com
    section: utils
    priority: optional
    depends: ["libc6"]
    recommends: []
    suggests: []
```

#### Generated Files
- `control` - Package metadata
- `myapp_1.0.0_amd64.deb` - Final package

#### Installation
```bash
sudo dpkg -i myapp_1.0.0_amd64.deb
sudo apt-get install -f  # Fix dependencies
```

### RPM (RedHat/CentOS)
**Format**: RPM package  
**Extension**: `.rpm`  
**Platform**: RedHat, CentOS, Fedora, SUSE

#### Configuration
```yaml
packages:
  rpm:
    group: Applications/System
    vendor: Your Name
    requires: ["glibc"]
    provides: []
```

#### Generated Files
- `myapp.spec` - RPM specification
- `myapp-1.0.0-1.x86_64.rpm` - Final package

#### Installation
```bash
sudo rpm -i myapp-1.0.0-1.x86_64.rpm
# or
sudo yum install myapp-1.0.0-1.x86_64.rpm
```

### AppImage (Universal Linux)
**Format**: Portable application  
**Extension**: `.AppImage`  
**Platform**: Most Linux distributions

#### Configuration
```yaml
packages:
  appimage:
    categories: [Utility, Development]
    icon: assets/icon.png
    desktop_entry:
      terminal: false
      type: Application
```

#### Generated Files
- `AppDir/` - Application directory structure
- `myapp-1.0.0-x86_64.AppImage` - Portable executable

#### Installation
```bash
chmod +x myapp-1.0.0-x86_64.AppImage
./myapp-1.0.0-x86_64.AppImage
```

### Snap (Ubuntu)
**Format**: Snap package  
**Extension**: `.snap`  
**Platform**: Ubuntu, many Linux distributions

#### Configuration
```yaml
packages:
  snap:
    grade: stable
    confinement: strict
    apps:
      myapp:
        command: bin/myapp
```

#### Generated Files
- `snap/snapcraft.yaml` - Snap configuration
- `myapp_1.0.0_amd64.snap` - Final package

#### Installation
```bash
sudo snap install myapp_1.0.0_amd64.snap --dangerous
```

### Flatpak (Linux)
**Format**: Flatpak application  
**Extension**: `.flatpak`  
**Platform**: Most Linux distributions

#### Configuration
```yaml
packages:
  flatpak:
    runtime: org.freedesktop.Platform
    runtime_version: "22.08"
    sdk: org.freedesktop.Sdk
```

#### Generated Files
- `com.yourname.MyApp.yaml` - Flatpak manifest
- `com.yourname.MyApp.flatpak` - Final package

#### Installation
```bash
flatpak install com.yourname.MyApp.flatpak
```

## Containers

### Docker
**Format**: Container image  
**Extension**: `Dockerfile`  
**Platform**: Cross-platform

#### Configuration
```yaml
packages:
  docker:
    base_image: alpine:latest
    expose: [8080]
    volumes: ["/data"]
    env:
      PORT: "8080"
```

#### Generated Files
- `Dockerfile` - Container definition
- `.dockerignore` - Ignore file

#### Usage
```bash
docker build -t myapp:1.0.0 .
docker run myapp:1.0.0
```

### Apptainer (HPC)
**Format**: Container image  
**Extension**: `.sif`  
**Platform**: HPC environments

#### Configuration
```yaml
packages:
  apptainer:
    base_image: ubuntu:22.04
    runscript: |
      #!/bin/bash
      exec /usr/local/bin/myapp "$@"
```

#### Generated Files
- `myapp.def` - Apptainer definition
- `myapp.sif` - Container image

#### Usage
```bash
apptainer run myapp.sif
```

## Language Packages

### npm (Node.js)
**Format**: npm package  
**Extension**: `.tgz`  
**Platform**: Cross-platform

#### Configuration
```yaml
packages:
  npm:
    main: index.js
    bin:
      myapp: bin/myapp
    keywords: ["cli", "tool"]
```

#### Generated Files
- `package.json` - Package metadata
- `myapp-1.0.0.tgz` - Package archive

#### Installation
```bash
npm install -g myapp
```

### PyPI (Python)
**Format**: Python package  
**Extension**: `.whl`, `.tar.gz`  
**Platform**: Cross-platform

#### Configuration
```yaml
packages:
  pypi:
    python_requires: ">=3.8"
    classifiers:
      - "Programming Language :: Python :: 3"
    entry_points:
      console_scripts:
        - myapp = myapp.main:main
```

#### Generated Files
- `setup.py` - Package setup
- `pyproject.toml` - Modern Python packaging
- `dist/myapp-1.0.0-py3-none-any.whl` - Wheel package

#### Installation
```bash
pip install myapp
```

### Cargo (Rust)
**Format**: Rust crate  
**Extension**: `.crate`  
**Platform**: Cross-platform

#### Configuration
```yaml
packages:
  cargo:
    edition: "2021"
    categories: ["command-line-utilities"]
    keywords: ["cli", "tool"]
```

#### Generated Files
- `Cargo.toml` - Crate metadata
- `myapp-1.0.0.crate` - Crate archive

#### Installation
```bash
cargo install myapp
```

## Platform Installers

### DMG (macOS)
**Format**: Disk image  
**Extension**: `.dmg`  
**Platform**: macOS

#### Configuration
```yaml
packages:
  dmg:
    background: assets/background.png
    icon_size: 128
    window_size: [600, 400]
```

#### Generated Files
- `myapp-1.0.0.dmg` - Disk image

#### Installation
Double-click to mount, drag to Applications folder.

### MSI (Windows)
**Format**: Windows Installer  
**Extension**: `.msi`  
**Platform**: Windows

#### Configuration
```yaml
packages:
  msi:
    upgrade_code: "{12345678-1234-1234-1234-123456789012}"
    install_scope: perMachine
```

#### Generated Files
- `myapp-1.0.0.msi` - Windows Installer package

#### Installation
```bash
msiexec /i myapp-1.0.0.msi
```

### MSIX (Windows)
**Format**: Modern Windows package  
**Extension**: `.msix`  
**Platform**: Windows 10+

#### Configuration
```yaml
packages:
  msix:
    publisher: "CN=Your Name"
    capabilities: []
```

#### Generated Files
- `AppxManifest.xml` - Package manifest
- `myapp-1.0.0.msix` - MSIX package

#### Installation
```bash
Add-AppxPackage myapp-1.0.0.msix
```

## Universal Installer

### curl|bash Script
**Format**: Shell script  
**Extension**: `.sh`  
**Platform**: Unix-like systems

#### Configuration
```yaml
installer:
  base_url: https://github.com/yourname/myapp/releases/download/v{{.Version}}
  install_path: /usr/local/bin
  detect_os: true
  verify_checksum: true
```

#### Generated Files
- `install.sh` - Universal installer script

#### Usage
```bash
curl -fsSL https://myapp.com/install.sh | bash
```

## Best Practices

### Cross-Platform Compatibility
- Use consistent naming across formats
- Handle platform-specific dependencies
- Test on target platforms

### Security
- Always verify checksums
- Sign packages when possible
- Use HTTPS for downloads

### Performance
- Optimize binary sizes
- Use compression when available
- Consider parallel packaging

### Maintenance
- Keep dependencies minimal
- Document package-specific requirements
- Test installation procedures
