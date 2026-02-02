# Troubleshooting Guide

## Common Issues and Solutions

### Configuration Issues

#### "No bagboy configuration file found"
**Symptoms:**
```
❌ No bagboy configuration file found
```

**Causes:**
- No `bagboy.yaml` file in current directory
- File named incorrectly (must be `bagboy.yaml` or `bagboy.yml`)
- Running command from wrong directory

**Solutions:**
```bash
# Create new configuration
bagboy init

# Check current directory
ls -la bagboy.yaml

# Verify you're in the right directory
pwd
```

#### "Configuration validation failed"
**Symptoms:**
```
❌ Configuration validation failed
```

**Common Causes & Solutions:**

**Missing required fields:**
```yaml
# ❌ Invalid - missing required fields
name: myapp

# ✅ Valid - all required fields
name: myapp
version: 1.0.0
description: My application
binaries:
  linux-amd64: dist/myapp-linux-amd64
```

**Invalid YAML syntax:**
```yaml
# ❌ Invalid - incorrect indentation
name: myapp
version: 1.0.0
binaries:
linux-amd64: dist/myapp  # Wrong indentation

# ✅ Valid - correct indentation
name: myapp
version: 1.0.0
binaries:
  linux-amd64: dist/myapp
```

**Debug Steps:**
```bash
# Validate configuration
bagboy validate --verbose

# Check YAML syntax
python -c "import yaml; yaml.safe_load(open('bagboy.yaml'))"
```

### Binary Issues

#### "Binary file not found"
**Symptoms:**
```
❌ Binary file not found: dist/myapp-linux-amd64
```

**Solutions:**
```bash
# Check if binary exists
ls -la dist/

# Build binaries first
make build  # or your build command

# Update paths in bagboy.yaml
vim bagboy.yaml
```

#### "Permission denied"
**Symptoms:**
```
❌ Permission denied: dist/myapp-linux-amd64
```

**Solutions:**
```bash
# Make binary executable
chmod +x dist/myapp-linux-amd64

# Check file permissions
ls -la dist/
```

### Package Format Issues

#### "rpmbuild not found"
**Symptoms:**
```
❌ rpmbuild not found - install rpm-build package
```

**Solutions:**
```bash
# Ubuntu/Debian
sudo apt-get install rpm

# CentOS/RHEL
sudo yum install rpm-build

# macOS
brew install rpm
```

#### "Docker daemon not running"
**Symptoms:**
```
❌ Cannot connect to Docker daemon
```

**Solutions:**
```bash
# Start Docker
sudo systemctl start docker  # Linux
open -a Docker              # macOS

# Check Docker status
docker version
```

#### "Snap not supported"
**Symptoms:**
```
❌ Snap packaging not supported on this system
```

**Solutions:**
```bash
# Install snapd
sudo apt-get install snapd  # Ubuntu/Debian
sudo yum install snapd      # CentOS/RHEL

# Enable snapd
sudo systemctl enable --now snapd.socket
```

### GitHub Integration Issues

#### "GitHub token not found"
**Symptoms:**
```
❌ GitHub token not found
```

**Solutions:**
```bash
# Set GitHub token
export GITHUB_TOKEN="your-token-here"

# Create GitHub token
# 1. Go to GitHub Settings → Developer settings → Personal access tokens
# 2. Generate new token with 'repo' permissions
# 3. Copy token and set environment variable

# Verify token
echo $GITHUB_TOKEN
```

#### "Repository not found"
**Symptoms:**
```
❌ Repository yourname/myapp not found
```

**Solutions:**
```bash
# Check repository exists
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/yourname/myapp

# Verify owner/repo in bagboy.yaml
grep -A2 "github:" bagboy.yaml

# Create repository if needed
gh repo create yourname/myapp --public
```

#### "Permission denied to repository"
**Symptoms:**
```
❌ Permission denied to repository
```

**Solutions:**
```bash
# Check token permissions
# Token needs 'repo' scope for private repos
# Token needs 'public_repo' scope for public repos

# Verify token has correct permissions
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/user/repos
```

### Code Signing Issues

#### macOS: "Developer cannot be verified"
**Symptoms:**
- macOS shows "cannot be verified" dialog
- Gatekeeper blocks execution

**Solutions:**
```bash
# Check signing status
bagboy sign --check

# Verify certificate
security find-identity -v -p codesigning

# Check notarization
spctl -a -v dist/myapp-darwin-amd64

# Re-sign if needed
codesign --force --sign "Developer ID Application: Your Name" dist/myapp-darwin-amd64
```

#### Windows: "Windows protected your PC"
**Symptoms:**
- SmartScreen warning appears
- Users can't run application easily

**Solutions:**
```bash
# Check certificate
signtool verify /pa dist/myapp-windows-amd64.exe

# Use EV certificate for immediate reputation
# Or build reputation over time with standard certificate

# Verify timestamp
signtool verify /v dist/myapp-windows-amd64.exe
```

#### Linux: "GPG signature verification failed"
**Symptoms:**
```
❌ GPG signature verification failed
```

**Solutions:**
```bash
# Check GPG setup
gpg --list-keys

# Import public key
gpg --import public-key.asc

# Trust key
gpg --edit-key YOUR_KEY_ID
# Type: trust
# Select: 5 (ultimate trust)
# Type: quit

# Re-sign
gpg --detach-sign --armor dist/myapp-linux-amd64
```

### Performance Issues

#### "Packaging is slow"
**Symptoms:**
- `bagboy pack --all` takes a long time
- High CPU/memory usage

**Solutions:**
```bash
# Run benchmark to identify bottlenecks
bagboy benchmark

# Use specific formats only
bagboy pack --brew --scoop  # Instead of --all

# Optimize binary sizes
go build -ldflags="-s -w" -o dist/myapp

# Use parallel building
make -j$(nproc) build-all
```

#### "Out of disk space"
**Symptoms:**
```
❌ No space left on device
```

**Solutions:**
```bash
# Check disk usage
df -h

# Clean up old packages
rm -rf dist/*.old

# Use smaller base images for containers
# alpine instead of ubuntu in Dockerfile

# Compress binaries
upx dist/myapp-*  # If UPX is available
```

### Network Issues

#### "Connection timeout"
**Symptoms:**
```
❌ Connection timeout to api.github.com
```

**Solutions:**
```bash
# Check internet connection
ping api.github.com

# Use proxy if needed
export HTTP_PROXY=http://proxy:8080
export HTTPS_PROXY=http://proxy:8080

# Retry with timeout
bagboy publish --timeout 300s
```

#### "Rate limit exceeded"
**Symptoms:**
```
❌ GitHub API rate limit exceeded
```

**Solutions:**
```bash
# Check rate limit status
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/rate_limit

# Wait for reset (shown in response)
# Or use different token

# Use authenticated requests (higher limits)
export GITHUB_TOKEN="your-token"
```

### Platform-Specific Issues

#### macOS: "Command not found: bagboy"
**Solutions:**
```bash
# Add to PATH
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Or install via Homebrew
brew install scttfrdmn/tap/bagboy
```

#### Windows: "bagboy is not recognized"
**Solutions:**
```powershell
# Add to PATH
$env:PATH += ";C:\Program Files\bagboy"

# Or use Scoop
scoop bucket add scttfrdmn https://github.com/scttfrdmn/scoop-bucket
scoop install bagboy
```

#### Linux: "Permission denied: /usr/local/bin/bagboy"
**Solutions:**
```bash
# Install to user directory
curl -fsSL bagboy.sh/install | INSTALL_PATH=~/.local/bin bash

# Or use sudo for system install
curl -fsSL bagboy.sh/install | sudo bash

# Add to PATH if needed
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

## Debug Mode

### Enable Verbose Output
```bash
# Verbose validation
bagboy validate --verbose

# Debug package creation
bagboy pack --all --debug

# Trace GitHub operations
GITHUB_DEBUG=1 bagboy publish
```

### Log Files
```bash
# Check system logs
journalctl -u bagboy  # systemd systems

# Application logs
tail -f ~/.bagboy/logs/bagboy.log

# Package manager logs
tail -f /var/log/dpkg.log      # DEB packages
tail -f /var/log/yum.log       # RPM packages
```

## Environment Debugging

### Check Environment
```bash
# Verify environment variables
env | grep -E "(GITHUB|APPLE|WINDOWS|GPG)_"

# Check PATH
echo $PATH

# Verify tools
which docker
which rpmbuild
which codesign
which signtool
```

### System Information
```bash
# OS information
uname -a
cat /etc/os-release  # Linux
sw_vers             # macOS
systeminfo          # Windows

# Architecture
uname -m
arch

# Available tools
bagboy check --formats all
```

## Getting Help

### Built-in Help
```bash
# General help
bagboy --help

# Command-specific help
bagboy pack --help
bagboy sign --help

# Check system requirements
bagboy check
```

### Community Support
- **GitHub Issues**: https://github.com/scttfrdmn/bagboy/issues
- **Documentation**: https://bagboy.dev
- **Examples**: https://github.com/scttfrdmn/bagboy/tree/main/examples

### Reporting Issues
When reporting issues, include:

1. **bagboy version**: `bagboy version`
2. **Operating system**: `uname -a`
3. **Configuration file**: `cat bagboy.yaml`
4. **Error message**: Full error output
5. **Steps to reproduce**: Exact commands run

### Issue Template
```markdown
## Bug Report

**bagboy version**: 
**OS**: 
**Architecture**: 

**Configuration**:
```yaml
# Paste bagboy.yaml here
```

**Steps to reproduce**:
1. 
2. 
3. 

**Expected behavior**:

**Actual behavior**:

**Error output**:
```
# Paste error here
```

**Additional context**:
```

## Performance Troubleshooting

### Benchmark Analysis
```bash
# Run comprehensive benchmarks
bagboy benchmark

# Analyze results
# Look for:
# - Slow packagers (>1s duration)
# - High memory usage (>100MB)
# - Low throughput (<1000 ops/sec)
```

### Optimization Tips
1. **Binary Size**: Smaller binaries = faster packaging
2. **Parallel Processing**: bagboy automatically parallelizes
3. **Selective Packaging**: Use specific formats instead of `--all`
4. **Local Caching**: bagboy caches intermediate files
5. **SSD Storage**: Use fast storage for build directories

### Memory Issues
```bash
# Monitor memory usage
top -p $(pgrep bagboy)

# Reduce memory usage
export GOMAXPROCS=1  # Limit Go runtime
ulimit -v 1048576    # Limit virtual memory
```

Remember: Most issues can be resolved by checking the basics first - configuration syntax, file paths, and required tools. Use `bagboy validate --verbose` and `bagboy sign --check` for comprehensive diagnostics.
