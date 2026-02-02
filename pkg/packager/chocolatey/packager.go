package chocolatey

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
	return "chocolatey"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Author == "" {
		return fmt.Errorf("author is required for chocolatey package")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	// Create package directory
	pkgDir := filepath.Join("dist", "chocolatey", cfg.Name)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return "", err
	}

	// Create nuspec file
	nuspecPath := filepath.Join(pkgDir, cfg.Name+".nuspec")
	if err := p.createNuspec(nuspecPath, cfg); err != nil {
		return "", err
	}

	// Create tools directory and install script
	toolsDir := filepath.Join(pkgDir, "tools")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		return "", err
	}

	installPath := filepath.Join(toolsDir, "chocolateyinstall.ps1")
	if err := p.createInstallScript(installPath, cfg); err != nil {
		return "", err
	}

	return pkgDir, nil
}

func (p *Packager) createNuspec(path string, cfg *config.Config) error {
	tmpl := `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>{{.Name}}</id>
    <version>{{.Version}}</version>
    <packageSourceUrl>{{.PackageSourceURL}}</packageSourceUrl>
    <owners>{{.Author}}</owners>
    <title>{{.Name}}</title>
    <authors>{{.Author}}</authors>
    <projectUrl>{{.Homepage}}</projectUrl>
    <docsUrl>{{.DocsURL}}</docsUrl>
    <tags>{{.Name}} cli tool</tags>
    <summary>{{.Description}}</summary>
    <description>{{.Description}}</description>
    <licenseUrl>{{.Homepage}}/blob/main/LICENSE</licenseUrl>
    <requireLicenseAcceptance>false</requireLicenseAcceptance>
  </metadata>
  <files>
    <file src="tools\**" target="tools" />
  </files>
</package>`

	t, err := template.New("nuspec").Parse(tmpl)
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
		PackageSourceURL string
		DocsURL          string
	}{
		Config:           cfg,
		PackageSourceURL: cfg.Packages.Chocolatey.PackageSourceURL,
		DocsURL:          cfg.Packages.Chocolatey.DocsURL,
	}

	if data.PackageSourceURL == "" {
		data.PackageSourceURL = cfg.Homepage
	}
	if data.DocsURL == "" {
		data.DocsURL = cfg.Homepage
	}

	return t.Execute(f, data)
}

func (p *Packager) createInstallScript(path string, cfg *config.Config) error {
	tmpl := `$ErrorActionPreference = 'Stop'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$packageName = '{{.Name}}'
$url64 = '{{.BaseURL}}/{{.Name}}-windows-amd64.exe'

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  fileType      = 'exe'
  url64bit      = $url64
  softwareName  = '{{.Name}}*'
  checksum64    = 'TODO_CHECKSUM'
  checksumType64= 'sha256'
  silentArgs    = '/S'
  validExitCodes= @(0)
}

Install-ChocolateyPackage @packageArgs`

	t, err := template.New("install").Parse(tmpl)
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
