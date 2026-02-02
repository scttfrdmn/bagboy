# Code Signing Guide

## Overview
Code signing ensures the authenticity and integrity of your software. bagboy supports comprehensive code signing across all major platforms.

## Platform Support

### macOS Code Signing ✅
- **Developer ID Application** certificates
- **Notarization** with Apple
- **Gatekeeper** bypass
- **Keychain** integration

### Windows Code Signing ✅
- **Authenticode** signing
- **SmartScreen** reputation
- **Certificate stores**
- **Timestamp servers**

### Linux Code Signing ✅
- **GPG** signing
- **Package repository** trust
- **Signature verification**

### Modern Solutions ✅
- **Sigstore/Cosign** keyless signing
- **SignPath.io** cloud signing
- **Git** commit/tag signing

## macOS Code Signing

### Prerequisites
1. **Apple Developer Program** membership ($99/year)
2. **Xcode** or Command Line Tools
3. **Developer ID Application** certificate

### Setup Steps

#### 1. Create Certificate
**Option A: Xcode**
1. Open Xcode → Preferences → Accounts
2. Add Apple ID → Manage Certificates
3. Create "Developer ID Application" certificate

**Option B: Apple Developer Portal**
1. Visit [developer.apple.com](https://developer.apple.com)
2. Certificates → Create → Developer ID Application
3. Download and install in Keychain

#### 2. Configure Environment
```bash
export APPLE_DEVELOPER_ID="Developer ID Application: Your Name (TEAM123456)"
export APPLE_ID="your@email.com"
export APPLE_TEAM_ID="TEAM123456"
export APPLE_APP_PASSWORD="app-specific-password"
```

#### 3. Generate App-Specific Password
1. Visit [appleid.apple.com](https://appleid.apple.com)
2. Sign In → App-Specific Passwords
3. Generate password for "bagboy"

#### 4. Configure bagboy.yaml
```yaml
signing:
  macos:
    identity: "Developer ID Application: Your Name (TEAM123456)"
    notarize: true
    apple_id: ""      # Set via APPLE_ID env var
    team_id: ""       # Set via APPLE_TEAM_ID env var
    app_password: ""  # Set via APPLE_APP_PASSWORD env var
```

### Verification
```bash
bagboy sign --check
codesign -v -d dist/myapp-darwin-amd64
spctl -a -v dist/myapp-darwin-amd64
```

## Windows Code Signing

### Prerequisites
1. **Code signing certificate** from trusted CA
2. **Windows SDK** (for signtool.exe)
3. **Certificate** installed in Windows Certificate Store

### Setup Steps

#### 1. Purchase Certificate
Recommended Certificate Authorities:
- **DigiCert** - Industry standard
- **Sectigo** - Cost-effective
- **GlobalSign** - Worldwide recognition

#### 2. Install Windows SDK
```bash
# Via Chocolatey
choco install windows-sdk-10-version-2004-all

# Via Visual Studio Installer
# Select "Windows 10 SDK" component
```

#### 3. Install Certificate
```bash
# Import PFX file
certlm.msc
# Or via PowerShell
Import-PfxCertificate -FilePath cert.pfx -CertStoreLocation Cert:\LocalMachine\My
```

#### 4. Get Certificate Thumbprint
```bash
# PowerShell
Get-ChildItem -Path Cert:\LocalMachine\My | Where-Object {$_.Subject -like "*Your Name*"}
```

#### 5. Configure Environment
```bash
export WINDOWS_CERT_THUMBPRINT="1234567890ABCDEF1234567890ABCDEF12345678"
export WINDOWS_TIMESTAMP_URL="http://timestamp.digicert.com"
```

#### 6. Configure bagboy.yaml
```yaml
signing:
  windows:
    certificate_thumbprint: ""  # Set via WINDOWS_CERT_THUMBPRINT env var
    timestamp_url: "http://timestamp.digicert.com"
```

### Verification
```bash
bagboy sign --check
signtool verify /pa dist/myapp-windows-amd64.exe
```

## Linux Code Signing

### Prerequisites
1. **GPG** installed
2. **GPG key pair** generated
3. **Public key** distributed

### Setup Steps

#### 1. Install GPG
```bash
# Ubuntu/Debian
sudo apt-get install gnupg

# CentOS/RHEL
sudo yum install gnupg2

# macOS
brew install gnupg
```

#### 2. Generate GPG Key
```bash
gpg --gen-key
# Follow prompts:
# - Real name: Your Name
# - Email: your@email.com
# - Passphrase: (secure passphrase)
```

#### 3. Export Public Key
```bash
# Get key ID
gpg --list-keys

# Export public key
gpg --export --armor your@email.com > public-key.asc

# Upload to keyservers
gpg --send-keys YOUR_KEY_ID
```

#### 4. Configure Environment
```bash
export GPG_KEY_ID="YOUR_KEY_ID"
export GPG_PASSPHRASE="your-passphrase"  # Optional, will prompt if not set
```

#### 5. Configure bagboy.yaml
```yaml
signing:
  linux:
    gpg_key_id: ""    # Set via GPG_KEY_ID env var
```

### Verification
```bash
bagboy sign --check
gpg --verify dist/myapp-linux-amd64.sig dist/myapp-linux-amd64
```

## Sigstore (Keyless Signing)

### Overview
Sigstore provides keyless signing using OpenID Connect (OIDC) identity verification.

### Setup Steps

#### 1. Install Cosign
```bash
# Via Homebrew
brew install cosign

# Via Go
go install github.com/sigstore/cosign/cmd/cosign@latest
```

#### 2. Configure bagboy.yaml
```yaml
signing:
  sigstore:
    enabled: true
    keyless: true
    oidc_issuer: "https://token.actions.githubusercontent.com"  # For GitHub Actions
```

#### 3. GitHub Actions Setup
```yaml
# .github/workflows/release.yml
permissions:
  id-token: write  # Required for OIDC
  contents: write

steps:
  - name: Sign and publish
    env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    run: bagboy publish
```

### Verification
```bash
cosign verify --certificate-identity=your@email.com \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  dist/myapp-linux-amd64
```

## SignPath.io (Cloud Signing)

### Overview
SignPath.io provides cloud-based code signing without managing certificates locally.

### Setup Steps

#### 1. Create SignPath Account
1. Visit [signpath.io](https://signpath.io)
2. Create account and organization
3. Create signing policy and project

#### 2. Get API Token
1. SignPath Dashboard → API Tokens
2. Create token with signing permissions

#### 3. Configure Environment
```bash
export SIGNPATH_API_TOKEN="your-api-token"
export SIGNPATH_ORGANIZATION_ID="your-org-id"
export SIGNPATH_PROJECT_SLUG="your-project"
```

#### 4. Configure bagboy.yaml
```yaml
signing:
  signpath:
    enabled: true
    organization_id: ""  # Set via env var
    project_slug: ""     # Set via env var
    api_token: ""        # Set via SIGNPATH_API_TOKEN env var
```

## Git Signing

### Overview
Sign Git commits and tags for verification.

### Setup Steps

#### 1. Configure Git
```bash
git config --global user.signingkey YOUR_GPG_KEY_ID
git config --global commit.gpgsign true
git config --global tag.gpgsign true
```

#### 2. Configure bagboy.yaml
```yaml
signing:
  git:
    enabled: true
    gpg_key_id: ""       # Set via GPG_KEY_ID env var
    sign_tags: true
    sign_commits: false  # Usually handled by Git config
```

## Multi-Platform Workflow

### Complete Signing Setup
```yaml
signing:
  # macOS signing
  macos:
    identity: "Developer ID Application: Your Name"
    notarize: true
    
  # Windows signing
  windows:
    certificate_thumbprint: ""
    timestamp_url: "http://timestamp.digicert.com"
    
  # Linux signing
  linux:
    gpg_key_id: ""
    
  # Modern solutions
  sigstore:
    enabled: true
    keyless: true
    
  signpath:
    enabled: false  # Alternative to traditional signing
    
  git:
    enabled: true
    sign_tags: true
```

### Environment Variables
```bash
# macOS
export APPLE_DEVELOPER_ID="Developer ID Application: Your Name"
export APPLE_ID="your@email.com"
export APPLE_APP_PASSWORD="app-specific-password"
export APPLE_TEAM_ID="TEAM123456"

# Windows
export WINDOWS_CERT_THUMBPRINT="certificate-thumbprint"

# Linux
export GPG_KEY_ID="your-gpg-key-id"

# Sigstore (GitHub Actions)
# Automatically provided by GitHub

# SignPath.io
export SIGNPATH_API_TOKEN="your-api-token"
export SIGNPATH_ORGANIZATION_ID="your-org-id"
export SIGNPATH_PROJECT_SLUG="your-project"
```

## CI/CD Integration

### GitHub Actions
```yaml
name: Release
on:
  push:
    tags: ['v*']

permissions:
  id-token: write  # For Sigstore
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup signing
        env:
          APPLE_DEVELOPER_ID: ${{ secrets.APPLE_DEVELOPER_ID }}
          APPLE_ID: ${{ secrets.APPLE_ID }}
          APPLE_APP_PASSWORD: ${{ secrets.APPLE_APP_PASSWORD }}
          WINDOWS_CERT_THUMBPRINT: ${{ secrets.WINDOWS_CERT_THUMBPRINT }}
          GPG_KEY_ID: ${{ secrets.GPG_KEY_ID }}
          GPG_PRIVATE_KEY: ${{ secrets.GPG_PRIVATE_KEY }}
        run: |
          echo "$GPG_PRIVATE_KEY" | gpg --import
          
      - name: Build and sign
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: bagboy publish
```

## Troubleshooting

### Common Issues

#### macOS: "Developer cannot be verified"
**Solution**: Enable notarization and ensure certificate is valid.

#### Windows: "Windows protected your PC"
**Solution**: Use EV certificate or build reputation over time.

#### Linux: "Signature verification failed"
**Solution**: Ensure GPG key is properly distributed and trusted.

### Debug Commands
```bash
# Check signing setup
bagboy sign --check

# Verify signatures
codesign -v -d binary                    # macOS
signtool verify /pa binary.exe          # Windows
gpg --verify binary.sig binary          # Linux
cosign verify binary                     # Sigstore
```

### Best Practices

1. **Secure Key Storage**
   - Use hardware security modules (HSM) when possible
   - Never commit private keys to version control
   - Use environment variables for sensitive data

2. **Certificate Management**
   - Monitor certificate expiration dates
   - Have backup certificates ready
   - Document renewal procedures

3. **Verification**
   - Always verify signatures after signing
   - Test on clean systems
   - Monitor signature validation in the wild

4. **Automation**
   - Integrate signing into CI/CD pipelines
   - Use secure secret management
   - Implement signing verification tests

## Cost Considerations

### Certificate Costs (Annual)
- **Apple Developer Program**: $99
- **Windows Code Signing**: $200-500
- **EV Certificates**: $300-800
- **SignPath.io**: $20-200/month
- **Linux GPG**: Free

### ROI Benefits
- **User Trust**: Reduces security warnings
- **Distribution**: Required for some app stores
- **Compliance**: Meets enterprise requirements
- **Reputation**: Builds software credibility

For more information, run `bagboy sign --check` to verify your signing setup.
