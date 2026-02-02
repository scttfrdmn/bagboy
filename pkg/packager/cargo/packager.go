package cargo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

type Packager struct{}

func New() *Packager {
	return &Packager{}
}

func (p *Packager) Name() string {
	return "cargo"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Homepage == "" {
		return fmt.Errorf("homepage is required for Cargo package")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	cargoDir := filepath.Join("dist", "cargo")
	if err := os.MkdirAll(cargoDir, 0755); err != nil {
		return "", err
	}

	// Create Cargo.toml
	cargoTomlPath := filepath.Join(cargoDir, "Cargo.toml")
	if err := p.createCargoToml(cargoTomlPath, cfg); err != nil {
		return "", err
	}

	// Create src/main.rs (wrapper that downloads binary)
	srcDir := filepath.Join(cargoDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return "", err
	}

	mainRsPath := filepath.Join(srcDir, "main.rs")
	if err := p.createMainRs(mainRsPath, cfg); err != nil {
		return "", err
	}

	// Create README.md
	readmePath := filepath.Join(cargoDir, "README.md")
	if err := p.createReadme(readmePath, cfg); err != nil {
		return "", err
	}

	return cargoDir, nil
}

func (p *Packager) createCargoToml(path string, cfg *config.Config) error {
	tmpl := `[package]
name = "{{.Name}}"
version = "{{.Version}}"
edition = "2021"
description = "{{.Description}}"
homepage = "{{.Homepage}}"
repository = "{{.Homepage}}"
license = "{{.License}}"
authors = ["{{.Author}}"]
keywords = ["cli", "tool"]
categories = ["command-line-utilities"]

[[bin]]
name = "{{.Name}}"
path = "src/main.rs"

[dependencies]
reqwest = { version = "0.11", features = ["blocking"] }
tokio = { version = "1.0", features = ["full"] }
dirs = "5.0"
flate2 = "1.0"
tar = "0.4"`

	t, err := template.New("cargo").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, cfg)
}

func (p *Packager) createMainRs(path string, cfg *config.Config) error {
	tmpl := `//! {{.Name}} - {{.Description}}
//! 
//! This is a Rust wrapper that downloads and executes the appropriate binary
//! for your platform from GitHub releases.

use std::env;
use std::fs;
use std::path::PathBuf;
use std::process::Command;

const VERSION: &str = "{{.Version}}";
const BASE_URL: &str = "{{.BaseURL}}";
const BIN_NAME: &str = "{{.Name}}";

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let binary_path = get_or_download_binary().await?;
    
    // Execute the binary with all arguments
    let args: Vec<String> = env::args().skip(1).collect();
    let status = Command::new(&binary_path)
        .args(&args)
        .status()?;
    
    std::process::exit(status.code().unwrap_or(1));
}

async fn get_or_download_binary() -> Result<PathBuf, Box<dyn std::error::Error>> {
    let cache_dir = dirs::cache_dir()
        .ok_or("Could not find cache directory")?
        .join("{{.Name}}");
    
    fs::create_dir_all(&cache_dir)?;
    
    let binary_name = get_binary_name();
    let binary_path = cache_dir.join(&binary_name);
    
    if binary_path.exists() {
        return Ok(binary_path);
    }
    
    println!("Downloading {{.Name}} v{VERSION}...");
    
    let download_url = format!("{BASE_URL}/{binary_name}");
    let response = reqwest::blocking::get(&download_url)?;
    
    if !response.status().is_success() {
        return Err(format!("Failed to download: {}", response.status()).into());
    }
    
    let bytes = response.bytes()?;
    fs::write(&binary_path, bytes)?;
    
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mut perms = fs::metadata(&binary_path)?.permissions();
        perms.set_mode(0o755);
        fs::set_permissions(&binary_path, perms)?;
    }
    
    println!("✓ Downloaded to {}", binary_path.display());
    Ok(binary_path)
}

fn get_binary_name() -> String {
    let os = if cfg!(target_os = "macos") {
        "darwin"
    } else if cfg!(target_os = "linux") {
        "linux"
    } else if cfg!(target_os = "windows") {
        "windows"
    } else {
        "linux" // fallback
    };
    
    let arch = if cfg!(target_arch = "x86_64") {
        "amd64"
    } else if cfg!(target_arch = "aarch64") {
        "arm64"
    } else {
        "amd64" // fallback
    };
    
    let ext = if cfg!(target_os = "windows") { ".exe" } else { "" };
    
    format!("{BIN_NAME}-{os}-{arch}{ext}")
}`

	t, err := template.New("main").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		*config.Config
		BaseURL string
	}{
		Config:  cfg,
		BaseURL: cfg.Installer.BaseURL,
	}

	return t.Execute(f, data)
}

func (p *Packager) createReadme(path string, cfg *config.Config) error {
	content := fmt.Sprintf("# %s\n\n%s\n\n## Installation\n\n```bash\ncargo install %s\n```\n\n## Usage\n\n```bash\n%s --help\n```\n\nThis package downloads the appropriate binary for your platform from [%s](%s).\n\n## License\n\n%s\n", cfg.Name, cfg.Description, cfg.Name, cfg.Name, cfg.Homepage, cfg.Homepage, cfg.License)

	return os.WriteFile(path, []byte(content), 0644)
}
