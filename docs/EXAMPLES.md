# Examples and Tutorials

## Quick Start Examples

### Go CLI Application
```bash
# 1. Create Go project
mkdir mygoapp && cd mygoapp
go mod init github.com/yourname/mygoapp

# 2. Create main.go
cat > main.go << 'EOF'
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    version := flag.Bool("version", false, "Show version")
    flag.Parse()
    
    if *version {
        fmt.Println("mygoapp v1.0.0")
        return
    }
    
    fmt.Println("Hello from mygoapp!")
}
EOF

# 3. Build for multiple platforms
mkdir -p dist
GOOS=darwin GOARCH=amd64 go build -o dist/mygoapp-darwin-amd64
GOOS=linux GOARCH=amd64 go build -o dist/mygoapp-linux-amd64
GOOS=windows GOARCH=amd64 go build -o dist/mygoapp-windows-amd64.exe

# 4. Initialize bagboy
bagboy init

# 5. Create packages
bagboy pack --all

# 6. Publish
bagboy publish
```

### Node.js Application
```bash
# 1. Create Node.js project
mkdir mynodeapp && cd mynodeapp
npm init -y

# 2. Create CLI script
mkdir bin
cat > bin/mynodeapp << 'EOF'
#!/usr/bin/env node
const args = process.argv.slice(2);

if (args.includes('--version')) {
    console.log('mynodeapp v1.0.0');
    process.exit(0);
}

console.log('Hello from mynodeapp!');
EOF

chmod +x bin/mynodeapp

# 3. Build with pkg
npm install -g pkg
pkg bin/mynodeapp --targets node18-macos-x64,node18-linux-x64,node18-win-x64 --out-path dist

# 4. Initialize bagboy
bagboy init

# 5. Update bagboy.yaml for Node.js
cat > bagboy.yaml << 'EOF'
name: mynodeapp
version: 1.0.0
description: My Node.js CLI application
binaries:
  darwin-amd64: dist/mynodeapp-macos
  linux-amd64: dist/mynodeapp-linux
  windows-amd64: dist/mynodeapp-win.exe
packages:
  npm:
    main: bin/mynodeapp
    bin:
      mynodeapp: bin/mynodeapp
EOF

# 6. Create packages
bagboy pack --all
```

### Rust Application
```bash
# 1. Create Rust project
cargo new myrustapp --bin
cd myrustapp

# 2. Update src/main.rs
cat > src/main.rs << 'EOF'
use std::env;

fn main() {
    let args: Vec<String> = env::args().collect();
    
    if args.len() > 1 && args[1] == "--version" {
        println!("myrustapp v0.1.0");
        return;
    }
    
    println!("Hello from myrustapp!");
}
EOF

# 3. Build for multiple targets
rustup target add x86_64-apple-darwin
rustup target add x86_64-pc-windows-gnu

cargo build --release
cargo build --release --target x86_64-apple-darwin
cargo build --release --target x86_64-pc-windows-gnu

# 4. Copy binaries
mkdir -p dist
cp target/release/myrustapp dist/myrustapp-linux-amd64
cp target/x86_64-apple-darwin/release/myrustapp dist/myrustapp-darwin-amd64
cp target/x86_64-pc-windows-gnu/release/myrustapp.exe dist/myrustapp-windows-amd64.exe

# 5. Initialize bagboy
bagboy init

# 6. Create packages
bagboy pack --all
```

## Advanced Examples

### Multi-Binary Application
```yaml
# bagboy.yaml
name: mytools
version: 2.0.0
description: Collection of useful tools
binaries:
  darwin-amd64: dist/mytools-darwin-amd64
  linux-amd64: dist/mytools-linux-amd64
  windows-amd64: dist/mytools-windows-amd64.exe

# Additional binaries
additional_binaries:
  helper:
    darwin-amd64: dist/helper-darwin-amd64
    linux-amd64: dist/helper-linux-amd64
    windows-amd64: dist/helper-windows-amd64.exe

packages:
  brew:
    test: |
      system "#{bin}/mytools --version"
      system "#{bin}/helper --version"
  deb:
    depends: ["libc6", "libssl3"]
```

### GUI Application with Assets
```yaml
# bagboy.yaml
name: myguiapp
version: 1.5.0
description: My GUI application
binaries:
  darwin-amd64: dist/myguiapp-darwin-amd64
  linux-amd64: dist/myguiapp-linux-amd64
  windows-amd64: dist/myguiapp-windows-amd64.exe

assets:
  - src: assets/icon.png
    dst: share/icons/myguiapp.png
  - src: assets/myguiapp.desktop
    dst: share/applications/myguiapp.desktop

packages:
  appimage:
    categories: [Graphics, Photography]
    icon: assets/icon.png
    desktop_entry:
      name: My GUI App
      comment: A great GUI application
      terminal: false
      type: Application
  
  dmg:
    background: assets/dmg-background.png
    icon_size: 128
    window_size: [600, 400]
    
  msi:
    install_scope: perUser
    shortcuts:
      - name: My GUI App
        target: myguiapp.exe
        folder: ProgramMenuFolder
```

### Server Application with Docker
```yaml
# bagboy.yaml
name: myserver
version: 3.0.0
description: My web server application
binaries:
  linux-amd64: dist/myserver-linux-amd64

packages:
  docker:
    base_image: alpine:latest
    expose: [8080, 8443]
    volumes: ["/data", "/config"]
    env:
      PORT: "8080"
      LOG_LEVEL: "info"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      
  apptainer:
    base_image: ubuntu:22.04
    runscript: |
      #!/bin/bash
      exec /usr/local/bin/myserver "$@"
    environment:
      PORT: 8080
      
installer:
  base_url: https://github.com/yourname/myserver/releases/download/v{{.Version}}
  install_path: /usr/local/bin
  post_install: |
    # Create systemd service
    sudo tee /etc/systemd/system/myserver.service > /dev/null << 'EOF'
    [Unit]
    Description=My Server
    After=network.target
    
    [Service]
    Type=simple
    User=myserver
    ExecStart=/usr/local/bin/myserver
    Restart=always
    
    [Install]
    WantedBy=multi-user.target
    EOF
    
    sudo systemctl daemon-reload
    sudo systemctl enable myserver
```

## GitHub Integration Examples

### Complete GitHub Workflow
```yaml
# bagboy.yaml
name: myapp
version: 1.0.0
description: My application with full GitHub integration

github:
  owner: yourname
  repo: myapp
  token_env: GITHUB_TOKEN
  
  release:
    enabled: true
    draft: false
    prerelease: false
    generate_notes: true
    
  tap:
    enabled: true
    repo: yourname/homebrew-tap
    auto_create: true
    auto_commit: true
    auto_push: true
    commit_message: "feat: update {{.Name}} to {{.Version}}"
    
  bucket:
    enabled: true
    repo: yourname/scoop-bucket
    auto_create: true
    auto_commit: true
    auto_push: true
    
  winget:
    enabled: true
    auto_pr: true
    fork_repo: yourname/winget-pkgs
```

### GitHub Actions Integration
```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags: ['v*']

permissions:
  contents: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Build binaries
        run: |
          make build-all
          
      - name: Package and publish
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          APPLE_DEVELOPER_ID: ${{ secrets.APPLE_DEVELOPER_ID }}
          APPLE_ID: ${{ secrets.APPLE_ID }}
          APPLE_APP_PASSWORD: ${{ secrets.APPLE_APP_PASSWORD }}
        run: |
          bagboy publish
```

## Code Signing Examples

### macOS Notarization
```yaml
# bagboy.yaml
signing:
  macos:
    identity: "Developer ID Application: Your Name (TEAM123456)"
    notarize: true
    apple_id: ""      # Set via APPLE_ID env var
    team_id: ""       # Set via APPLE_TEAM_ID env var
    app_password: ""  # Set via APPLE_APP_PASSWORD env var
    entitlements: |
      <?xml version="1.0" encoding="UTF-8"?>
      <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
      <plist version="1.0">
      <dict>
          <key>com.apple.security.cs.allow-jit</key>
          <true/>
          <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
          <true/>
      </dict>
      </plist>
```

### Multi-Platform Signing
```yaml
# bagboy.yaml
signing:
  macos:
    identity: "Developer ID Application: Your Name"
    notarize: true
    
  windows:
    certificate_thumbprint: ""  # Set via env var
    timestamp_url: "http://timestamp.digicert.com"
    
  linux:
    gpg_key_id: ""  # Set via env var
    
  sigstore:
    enabled: true
    keyless: true
    oidc_issuer: "https://token.actions.githubusercontent.com"
```

## CI/CD Integration Examples

### GitLab CI
```yaml
# .gitlab-ci.yml
stages:
  - build
  - package
  - deploy

variables:
  GO_VERSION: "1.21"

build:
  stage: build
  image: golang:${GO_VERSION}
  script:
    - make build-all
  artifacts:
    paths:
      - dist/

package:
  stage: package
  image: alpine:latest
  before_script:
    - apk add --no-cache curl bash
    - curl -fsSL bagboy.sh/install | bash
  script:
    - bagboy pack --all
  artifacts:
    paths:
      - dist/

deploy:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache curl bash
    - curl -fsSL bagboy.sh/install | bash
  script:
    - bagboy publish
  only:
    - tags
```

### Jenkins Pipeline
```groovy
// Jenkinsfile
pipeline {
    agent any
    
    environment {
        GITHUB_TOKEN = credentials('github-token')
        APPLE_DEVELOPER_ID = credentials('apple-developer-id')
    }
    
    stages {
        stage('Build') {
            steps {
                sh 'make build-all'
            }
        }
        
        stage('Package') {
            steps {
                sh 'curl -fsSL bagboy.sh/install | bash'
                sh 'bagboy pack --all'
            }
        }
        
        stage('Deploy') {
            when {
                tag pattern: "v\\d+\\.\\d+\\.\\d+", comparator: "REGEXP"
            }
            steps {
                sh 'bagboy publish'
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'dist/**', fingerprint: true
        }
    }
}
```

## Custom Packager Example

### Creating a Custom Format
```go
// pkg/packager/custom/packager.go
package custom

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/scttfrdmn/bagboy/pkg/config"
)

type CustomPackager struct{}

func New() *CustomPackager {
    return &CustomPackager{}
}

func (c *CustomPackager) Name() string {
    return "custom"
}

func (c *CustomPackager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
    outputPath := filepath.Join("dist", fmt.Sprintf("%s-custom.pkg", cfg.Name))
    
    // Create custom package format
    file, err := os.Create(outputPath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    // Write custom format data
    fmt.Fprintf(file, "Name: %s\n", cfg.Name)
    fmt.Fprintf(file, "Version: %s\n", cfg.Version)
    fmt.Fprintf(file, "Description: %s\n", cfg.Description)
    
    return outputPath, nil
}

func (c *CustomPackager) Validate(cfg *config.Config) error {
    if cfg.Name == "" {
        return fmt.Errorf("name is required for custom format")
    }
    return nil
}
```

### Using Custom Packager
```go
// main.go
package main

import (
    "context"
    
    "github.com/scttfrdmn/bagboy/pkg/config"
    "github.com/scttfrdmn/bagboy/pkg/packager"
    "github.com/scttfrdmn/bagboy/pkg/packager/custom"
)

func main() {
    cfg, err := config.Load("bagboy.yaml")
    if err != nil {
        panic(err)
    }
    
    registry := packager.NewRegistry()
    registry.Register(custom.New())
    
    ctx := context.Background()
    output, err := registry.Get("custom").Pack(ctx, cfg)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Created custom package: %s\n", output)
}
```

## Performance Optimization Examples

### Parallel Building
```bash
#!/bin/bash
# build-parallel.sh

# Build all platforms in parallel
(
    GOOS=darwin GOARCH=amd64 go build -o dist/myapp-darwin-amd64 &
    GOOS=darwin GOARCH=arm64 go build -o dist/myapp-darwin-arm64 &
    GOOS=linux GOARCH=amd64 go build -o dist/myapp-linux-amd64 &
    GOOS=linux GOARCH=arm64 go build -o dist/myapp-linux-arm64 &
    GOOS=windows GOARCH=amd64 go build -o dist/myapp-windows-amd64.exe &
    wait
)

# Package in parallel
bagboy pack --all
```

### Optimized Binary Size
```yaml
# bagboy.yaml with size optimization
name: myapp
version: 1.0.0

build:
  ldflags: "-s -w -X main.version={{.Version}}"
  tags: ["netgo", "osusergo"]
  
packages:
  docker:
    base_image: scratch  # Minimal image
    multi_stage: true
```

## Testing Examples

### Package Testing
```bash
#!/bin/bash
# test-packages.sh

# Test Homebrew formula
brew install --build-from-source dist/myapp.rb
myapp --version
brew uninstall myapp

# Test DEB package
sudo dpkg -i dist/myapp_1.0.0_amd64.deb
myapp --version
sudo dpkg -r myapp

# Test installer script
curl -fsSL file://$(pwd)/dist/install.sh | bash
myapp --version
```

### Integration Testing
```go
// integration_test.go
package main

import (
    "os/exec"
    "testing"
)

func TestPackageCreation(t *testing.T) {
    cmd := exec.Command("bagboy", "pack", "--brew", "--deb")
    if err := cmd.Run(); err != nil {
        t.Fatalf("Failed to create packages: %v", err)
    }
    
    // Verify files exist
    files := []string{
        "dist/myapp.rb",
        "dist/myapp_1.0.0_amd64.deb",
    }
    
    for _, file := range files {
        if _, err := os.Stat(file); os.IsNotExist(err) {
            t.Errorf("Expected file %s was not created", file)
        }
    }
}
```

These examples demonstrate the flexibility and power of bagboy across different languages, platforms, and use cases. Start with the basic examples and gradually incorporate more advanced features as needed.
