---
name: VM Build Support
about: Add VM/container support for cross-platform package building
title: '[Feature] VM Build Support for Cross-Platform Packaging'
labels: enhancement, build-system
assignees: ''
---

## Problem

Currently, bagboy requires platform-specific tools to build certain package formats:
- RPM packages need `rpmbuild` (Linux only)
- MSI packages need WiX Toolset (Windows only)  
- DEB packages need `dpkg-deb` (Linux/macOS)

This prevents users from building all package formats on their development machine.

## Proposed Solution

Add VM/container support to automatically provision build environments for non-native platforms.

```bash
# Build RPM on macOS
bagboy pack --rpm --vm

# Build all formats with VMs
bagboy pack --all --vm
```

## Use Cases

### 1. macOS Developer
- Wants to build DEB/RPM packages without Linux machine
- Needs to test Linux packages locally

### 2. Linux Developer  
- Wants to build MSI/Chocolatey packages without Windows
- Needs to test Windows installers

### 3. CI/CD Pipeline
- Single runner builds all package formats
- Reproducible build environments
- No manual tool installation

## Implementation Approach

### Phase 1: Docker Support (MVP)
- Docker provider for Linux packages (DEB, RPM, AppImage)
- Fast, lightweight, widely available
- Works in CI/CD

**Configuration:**
```yaml
build:
  vm:
    enabled: true
    provider: docker
    images:
      linux: "ubuntu:22.04"
```

**Commands:**
```bash
bagboy pack --deb --vm
bagboy pack --rpm --vm-provider docker
bagboy vm setup
bagboy vm clean
```

### Phase 2: Multipass Support
- Ubuntu VMs via Multipass
- Cross-platform (macOS, Linux, Windows)
- Official Ubuntu tool

### Phase 3: Vagrant Support
- Full VM support for Windows builds
- Code signing capabilities
- Advanced use cases

### Phase 4: Cloud Integration
- GitHub Actions templates
- GitLab CI integration
- Scalable builds

## Benefits

- ✅ Build any package format from any OS
- ✅ No manual tool installation
- ✅ Reproducible builds
- ✅ Isolated environments
- ✅ Better CI/CD support

## Technical Details

### Docker Implementation
1. Create Dockerfiles for each package type
2. Mount source/dist directories
3. Run packaging commands in container
4. Copy artifacts back to host

### VM Management
- Cache images/boxes for speed
- Clean up after builds
- Parallel builds when possible
- Clear error messages

### Configuration
```yaml
build:
  vm:
    enabled: true
    provider: docker  # docker, multipass, vagrant, cloud
    
    docker:
      images:
        linux: "ubuntu:22.04"
      volumes:
        - "./dist:/dist"
    
    packages:
      rpm:
        provider: docker
        image: "fedora:38"
      deb:
        provider: docker
        image: "ubuntu:22.04"
```

## Acceptance Criteria

- [ ] Can build RPM packages on macOS using Docker
- [ ] Can build DEB packages on Windows using Docker
- [ ] VM setup is automatic (no manual steps)
- [ ] Artifacts are copied back to host
- [ ] Clear error messages for missing providers
- [ ] Documentation with examples
- [ ] CI/CD integration examples
- [ ] < 2 minute overhead for VM setup

## Related Issues

- Closes #XX (RPM building on macOS)
- Related to #XX (Cross-platform builds)
- Depends on #XX (Dependency management)

## Documentation Needed

- [ ] VM build guide in docs/
- [ ] Configuration reference
- [ ] CI/CD examples
- [ ] Troubleshooting section
- [ ] Provider comparison

## Timeline

- Week 1-2: Docker support (MVP)
- Week 3: Multipass support
- Week 4: Vagrant support  
- Week 5-6: Cloud integration
- Week 7: Documentation
- Week 8: Testing

## Questions

1. Should VM support be opt-in or automatic?
2. Which Docker images should be default?
3. Should we cache VM images between builds?
4. How to handle code signing in VMs?

## References

- [Full Proposal](../VM_BUILD_SUPPORT_PROPOSAL.md)
- [Docker Documentation](https://docs.docker.com/)
- [Multipass Documentation](https://multipass.run/)
- [Vagrant Documentation](https://www.vagrantup.com/)
