package chocolatey

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
	return "chocolatey"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Author == "" {
		return fmt.Errorf("author is required for chocolatey package")
	}
	// Check for Windows binary
	hasWindowsBinary := false
	for arch := range cfg.Binaries {
		if strings.HasPrefix(arch, "windows-") {
			hasWindowsBinary = true
			break
		}
	}
	if !hasWindowsBinary {
		return fmt.Errorf("no Windows binary specified for Chocolatey package")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	// Find Windows binary
	var windowsBinary string
	for arch, path := range cfg.Binaries {
		if strings.HasPrefix(arch, "windows-") {
			windowsBinary = path
			break
		}
	}
	if windowsBinary == "" {
		return "", fmt.Errorf("no Windows binary found")
	}

	// Create build directory
	buildDir := filepath.Join("dist", "chocolatey-build")
	toolsDir := filepath.Join(buildDir, "tools")
	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}

	// Copy binary to tools directory
	binaryDest := filepath.Join(toolsDir, cfg.Name+".exe")
	if err := p.copyFile(windowsBinary, binaryDest); err != nil {
		return "", fmt.Errorf("failed to copy binary: %w", err)
	}

	// Generate .nuspec file
	nuspecPath := filepath.Join(buildDir, cfg.Name+".nuspec")
	if err := p.createNuspec(nuspecPath, cfg); err != nil {
		return "", fmt.Errorf("failed to generate nuspec: %w", err)
	}

	// Generate chocolateyInstall.ps1
	installScriptPath := filepath.Join(toolsDir, "chocolateyInstall.ps1")
	if err := p.createInstallScript(installScriptPath, cfg); err != nil {
		return "", fmt.Errorf("failed to generate install script: %w", err)
	}

	// Generate chocolateyUninstall.ps1
	uninstallScriptPath := filepath.Join(toolsDir, "chocolateyUninstall.ps1")
	if err := p.createUninstallScript(uninstallScriptPath, cfg); err != nil {
		return "", fmt.Errorf("failed to generate uninstall script: %w", err)
	}

	// Build package
	return p.buildPackage(ctx, buildDir, cfg)
}

func (p *Packager) createNuspec(path string, cfg *config.Config) error {
	tmpl := `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>{{.Name}}</id>
    <version>{{.Version}}</version>
    <packageSourceUrl>{{.PackageSourceURL}}</packageSourceUrl>
    <owners>{{.AuthorName}}</owners>
    <title>{{.Name}}</title>
    <authors>{{.AuthorName}}</authors>
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
		AuthorName       string
		PackageSourceURL string
		DocsURL          string
	}{
		Config:           cfg,
		AuthorName:       p.getAuthorName(cfg),
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
$exeName = '{{.Name}}.exe'
$exePath = Join-Path $toolsDir $exeName

# Create shim for the executable
Install-BinFile -Name $packageName -Path $exePath

Write-Host "{{.Name}} has been installed successfully!" -ForegroundColor Green
Write-Host "You can now use '{{.Name}}' from any command prompt." -ForegroundColor Green`

	t, err := template.New("install").Parse(tmpl)
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

func (p *Packager) createUninstallScript(path string, cfg *config.Config) error {
	tmpl := `$ErrorActionPreference = 'Stop'
$packageName = '{{.Name}}'

# Remove the shim
Uninstall-BinFile -Name $packageName

Write-Host "{{.Name}} has been uninstalled successfully!" -ForegroundColor Green`

	t, err := template.New("uninstall").Parse(tmpl)
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

func (p *Packager) buildPackage(ctx context.Context, buildDir string, cfg *config.Config) (string, error) {
	outputPath := filepath.Join("dist", fmt.Sprintf("%s.%s.nupkg", cfg.Name, cfg.Version))

	// Try choco pack first
	if _, err := exec.LookPath("choco"); err == nil {
		return p.buildWithChoco(ctx, buildDir, outputPath, cfg)
	}

	// Try nuget pack
	if _, err := exec.LookPath("nuget"); err == nil {
		return p.buildWithNuget(ctx, buildDir, outputPath, cfg)
	}

	// Manual zip creation as fallback
	return p.buildManually(buildDir, outputPath, cfg)
}

func (p *Packager) buildWithChoco(ctx context.Context, buildDir, outputPath string, cfg *config.Config) (string, error) {
	nuspecPath := filepath.Join(buildDir, cfg.Name+".nuspec")
	
	cmd := exec.CommandContext(ctx, "choco", "pack", nuspecPath, "--outputdirectory", filepath.Dir(outputPath))
	cmd.Dir = buildDir
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("choco pack failed: %w\nOutput: %s", err, output)
	}

	return outputPath, nil
}

func (p *Packager) buildWithNuget(ctx context.Context, buildDir, outputPath string, cfg *config.Config) (string, error) {
	nuspecPath := filepath.Join(buildDir, cfg.Name+".nuspec")
	
	cmd := exec.CommandContext(ctx, "nuget", "pack", nuspecPath, "-OutputDirectory", filepath.Dir(outputPath))
	cmd.Dir = buildDir
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("nuget pack failed: %w\nOutput: %s", err, output)
	}

	return outputPath, nil
}

func (p *Packager) buildManually(buildDir, outputPath string, cfg *config.Config) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create a simple zip file (Chocolatey packages are essentially zip files with .nupkg extension)
	if _, err := exec.LookPath("zip"); err == nil {
		cmd := exec.Command("zip", "-r", outputPath, ".")
		cmd.Dir = buildDir
		
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("zip failed: %w\nOutput: %s", err, output)
		}
		
		return outputPath, nil
	}

	return "", fmt.Errorf("Chocolatey build tools not found - install Chocolatey CLI, NuGet CLI, or zip")
}

func (p *Packager) getAuthorName(cfg *config.Config) string {
	if strings.Contains(cfg.Author, "<") {
		parts := strings.Split(cfg.Author, "<")
		return strings.TrimSpace(parts[0])
	}
	return cfg.Author
}

func (p *Packager) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}
