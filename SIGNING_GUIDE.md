# 🔐 Code Signing Guide for bagboy

## Why Code Signing Matters

### **macOS (REQUIRED)**
- **Gatekeeper**: Unsigned apps show scary warnings
- **Notarization**: Required for distribution outside App Store
- **User Trust**: Signed apps install without warnings

### **Windows (HIGHLY RECOMMENDED)**
- **SmartScreen**: Prevents "Windows protected your PC" warnings
- **User Trust**: Signed apps appear more professional
- **Enterprise**: Many organizations require signed software

### **Linux (RECOMMENDED)**
- **Package Repositories**: GPG signing enables trusted repos
- **Verification**: Users can verify package integrity
- **Distribution**: Required for official repositories

## 🆕 **Modern Signing Solutions**

### **🔗 Sigstore/Cosign (Keyless Signing)**

Sigstore provides keyless signing using OpenID Connect (OIDC) identity verification.

#### **Setup**
```bash
# Install cosign
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

# Sign with GitHub Actions OIDC (keyless)
cosign sign --yes ghcr.io/yourname/myapp:latest

# Verify signature
cosign verify --certificate-identity-regexp=".*" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  ghcr.io/yourname/myapp:latest
```

#### **bagboy Configuration**
```yaml
signing:
  sigstore:
    enabled: true
    keyless: true
    oidc_issuer: "https://token.actions.githubusercontent.com"
```

#### **Benefits**
- ✅ No certificate management
- ✅ Transparency log for auditability  
- ✅ Works great in CI/CD
- ✅ Free and open source

### **☁️ SignPath.io (Cloud Signing Service)**

SignPath.io provides cloud-based code signing with certificate management.

#### **Setup**
```bash
# 1. Create account at signpath.io
# 2. Upload your certificate or use SignPath's
# 3. Create signing policy
# 4. Get API token
```

#### **bagboy Configuration**
```yaml
signing:
  signpath:
    enabled: true
    organization_id: "your-org-id"
    project_slug: "your-project"
    api_token: "${SIGNPATH_API_TOKEN}"
```

#### **Benefits**
- ✅ Managed certificate storage
- ✅ Approval workflows
- ✅ Audit trails
- ✅ Supports Windows, macOS, Java, etc.

### **📝 Git Signing (Supply Chain Security)**

Git signing verifies commits and tags for supply chain security.

#### **Setup**
```bash
# Configure git signing
git config --global user.signingkey YOUR_GPG_KEY_ID
git config --global commit.gpgsign true
git config --global tag.gpgsign true

# Sign commits automatically
git commit -S -m "Signed commit"

# Sign tags
git tag -s v1.0.0 -m "Signed release"
```

#### **bagboy Configuration**
```yaml
signing:
  git:
    enabled: true
    gpg_key_id: "YOUR_GPG_KEY_ID"
    sign_tags: true
    sign_commits: true
```

#### **Benefits**
- ✅ Verifies commit authenticity
- ✅ Supply chain security
- ✅ GitHub shows verified badges
- ✅ Required for many security standards

## 🍎 **macOS Code Signing Setup**

### **Step 1: Join Apple Developer Program**
```bash
# Cost: $99/year
# Sign up at: https://developer.apple.com/programs/
```

### **Step 2: Create Developer ID Certificate**
```bash
# In Xcode:
# 1. Xcode → Preferences → Accounts
# 2. Add Apple ID
# 3. Manage Certificates → Create "Developer ID Application"

# Or via Apple Developer portal:
# 1. Visit developer.apple.com
# 2. Certificates → Create → Developer ID Application
# 3. Download and install in Keychain
```

### **Step 3: Configure Environment**
```bash
export APPLE_DEVELOPER_ID="Developer ID Application: Your Name (TEAMID)"
export APPLE_ID="your@apple.id"
export APPLE_TEAM_ID="YOUR_TEAM_ID"
export APPLE_APP_PASSWORD="app-specific-password"
```

### **Step 4: Sign with bagboy**
```bash
bagboy sign --check                    # Verify setup
bagboy sign --binary dist/app-darwin   # Sign binary
```

## 🪟 **Windows Code Signing Setup**

### **Step 1: Purchase Certificate**
```bash
# Recommended CAs:
# - DigiCert: ~$400/year
# - Sectigo: ~$200/year  
# - GlobalSign: ~$300/year
```

### **Step 2: Install Windows SDK**
```powershell
# Download from: https://developer.microsoft.com/windows/downloads/windows-sdk/
# Or via Visual Studio Installer
```

### **Step 3: Install Certificate**
```powershell
# Import .pfx file to Windows Certificate Store
# Or use Hardware Security Module (HSM)
```

### **Step 4: Configure Environment**
```powershell
$env:WINDOWS_CERT_THUMBPRINT = "YOUR_CERTIFICATE_THUMBPRINT"
```

### **Step 5: Sign with bagboy**
```powershell
bagboy sign --check                      # Verify setup
bagboy sign --binary dist/app.exe        # Sign binary
```

## 🐧 **Linux GPG Signing Setup**

### **Step 1: Generate GPG Key**
```bash
gpg --gen-key
# Follow prompts to create key
```

### **Step 2: Export and Share Public Key**
```bash
# Export public key
gpg --export --armor your@email.com > public-key.asc

# Upload to keyservers
gpg --send-keys YOUR_KEY_ID
```

### **Step 3: Configure Environment**
```bash
export GPG_KEY_ID="YOUR_KEY_ID"
```

### **Step 4: Sign with bagboy**
```bash
bagboy sign --check                    # Verify setup
bagboy sign --binary dist/app-linux    # Sign binary
```

## 🚀 **bagboy Signing Integration**

### **bagboy Signing Integration**
```yaml
# bagboy.yaml - Complete signing configuration
signing:
  # Traditional platform signing
  macos:
    identity: "Developer ID Application: Your Name"
    notarize: true
  windows:
    certificate_thumbprint: "${WINDOWS_CERT_THUMBPRINT}"
    timestamp_url: "http://timestamp.digicert.com"
  linux:
    gpg_key_id: "${GPG_KEY_ID}"
    
  # Modern signing solutions
  sigstore:
    enabled: true
    keyless: true
    oidc_issuer: "https://token.actions.githubusercontent.com"
  signpath:
    enabled: false
    organization_id: "your-org-id"
    project_slug: "your-project"
    api_token: "${SIGNPATH_API_TOKEN}"
  git:
    enabled: true
    gpg_key_id: "${GPG_KEY_ID}"
    sign_tags: true
    sign_commits: false
```

### **Commands**
```bash
# Check signing setup
bagboy sign --check

# Sign specific binary
bagboy sign --binary dist/myapp-darwin-amd64

# Sign all binaries (future feature)
bagboy sign --all

# Pack with signing
bagboy pack --all --sign
```

## 📋 **Platform-Specific Requirements**

### **macOS Requirements**
- ✅ **Apple Developer Account** ($99/year)
- ✅ **Developer ID Application Certificate**
- ✅ **Xcode Command Line Tools** (for codesign)
- ✅ **App-Specific Password** (for notarization)

### **Windows Requirements**
- ✅ **Code Signing Certificate** ($200-400/year)
- ✅ **Windows SDK** (for signtool.exe)
- ✅ **Certificate in Windows Store** or HSM

### **Linux Requirements**
- ✅ **GPG Key Pair** (free)
- ✅ **GPG installed** (usually pre-installed)
- ✅ **Public key on keyservers** (for verification)

## 🎯 **Best Practices**

### **Security**
- Store certificates securely (HSM for Windows, Keychain for macOS)
- Use environment variables for sensitive data
- Never commit certificates or passwords to git

### **Automation**
- Set up signing in CI/CD pipelines
- Use dedicated signing machines/services
- Implement signing as part of release process

### **Verification**
- Always verify signatures after signing
- Test signed binaries on clean systems
- Monitor certificate expiration dates

## 🔄 **Complete Signed Release Workflow**

```bash
# 1. Build binaries
make build-all

# 2. Check signing setup
bagboy sign --check

# 3. Sign all binaries
for binary in dist/*; do
  bagboy sign --binary "$binary"
done

# 4. Create packages (with signed binaries)
bagboy pack --all

# 5. Deploy to repositories
bagboy deploy --targets brew,npm,docker
```

**With proper code signing, your software will install without warnings and build user trust across all platforms!** 🔐✨
