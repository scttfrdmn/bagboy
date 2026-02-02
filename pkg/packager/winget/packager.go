package winget

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

type Packager struct{}

func New() *Packager {
	return &Packager{}
}

func (p *Packager) Name() string {
	return "winget"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Packages.Winget.PackageIdentifier == "" {
		return fmt.Errorf("winget.package_identifier is required")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	// Create manifests directory structure
	parts := strings.Split(cfg.Packages.Winget.PackageIdentifier, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid package identifier format")
	}

	manifestDir := filepath.Join("dist", "winget", "manifests", strings.ToLower(parts[0][:1]), parts[0], parts[1], cfg.Version)
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return "", err
	}

	// Create version manifest
	versionPath := filepath.Join(manifestDir, fmt.Sprintf("%s.yaml", cfg.Packages.Winget.PackageIdentifier))
	if err := p.createVersionManifest(versionPath, cfg); err != nil {
		return "", err
	}

	// Create installer manifest
	installerPath := filepath.Join(manifestDir, fmt.Sprintf("%s.installer.yaml", cfg.Packages.Winget.PackageIdentifier))
	if err := p.createInstallerManifest(installerPath, cfg); err != nil {
		return "", err
	}

	// Create locale manifest
	localePath := filepath.Join(manifestDir, fmt.Sprintf("%s.locale.en-US.yaml", cfg.Packages.Winget.PackageIdentifier))
	if err := p.createLocaleManifest(localePath, cfg); err != nil {
		return "", err
	}

	return manifestDir, nil
}

func (p *Packager) createVersionManifest(path string, cfg *config.Config) error {
	tmpl := `PackageIdentifier: {{.PackageIdentifier}}
PackageVersion: {{.Version}}
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.4.0`

	return p.writeTemplate(path, tmpl, cfg)
}

func (p *Packager) createInstallerManifest(path string, cfg *config.Config) error {
	tmpl := `PackageIdentifier: {{.PackageIdentifier}}
PackageVersion: {{.Version}}
MinimumOSVersion: {{.MinimumOSVersion}}
Installers:
- Architecture: x64
  InstallerType: exe
  InstallerUrl: {{.BaseURL}}/{{.Name}}-windows-amd64.exe
  InstallerSha256: TODO_CHECKSUM
  InstallerSwitches:
    Silent: /S
    SilentWithProgress: /S
ManifestType: installer
ManifestVersion: 1.4.0`

	return p.writeTemplate(path, tmpl, cfg)
}

func (p *Packager) createLocaleManifest(path string, cfg *config.Config) error {
	tmpl := `PackageIdentifier: {{.PackageIdentifier}}
PackageVersion: {{.Version}}
PackageLocale: en-US
Publisher: {{.Publisher}}
PackageName: {{.Name}}
License: {{.License}}
ShortDescription: {{.Description}}
PackageUrl: {{.Homepage}}
ManifestType: defaultLocale
ManifestVersion: 1.4.0`

	return p.writeTemplate(path, tmpl, cfg)
}

func (p *Packager) writeTemplate(path, tmpl string, cfg *config.Config) error {
	t, err := template.New("manifest").Parse(tmpl)
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
		PackageIdentifier string
		Publisher         string
		MinimumOSVersion  string
		BaseURL           string
	}{
		Config:            cfg,
		PackageIdentifier: cfg.Packages.Winget.PackageIdentifier,
		Publisher:         cfg.Packages.Winget.Publisher,
		MinimumOSVersion:  cfg.Packages.Winget.MinimumOSVersion,
		BaseURL:           cfg.Installer.BaseURL,
	}

	if data.Publisher == "" {
		data.Publisher = cfg.Author
	}
	if data.MinimumOSVersion == "" {
		data.MinimumOSVersion = "10.0.0.0"
	}

	return t.Execute(f, data)
}
