// Package msi implements the msi packager for bagboy.
package msi

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// Packager creates msi packages for Windows MSI installer distribution.
type Packager struct{}

// New returns a new msi Packager.
func New() *Packager {
	return &Packager{}
}

// Name returns the packager name "msi".
func (p *Packager) Name() string {
	return "msi"
}

// Validate checks that cfg has the required fields for msi packaging.
func (p *Packager) Validate(cfg *config.Config) error {
	for arch := range cfg.Binaries {
		if strings.HasPrefix(arch, "windows-") {
			return nil
		}
	}
	return fmt.Errorf("no Windows binary found for MSI creation")
}

// Pack generates msi artifacts from cfg and returns the output directory path.
// One MSI is produced per available windows-* binary (x64, arm64, x86).
// Install msitools ("brew install msitools") to build on macOS/Linux.
func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	outDir := filepath.Join("dist", "msi")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	built := 0
	for arch, binPath := range cfg.Binaries {
		if !strings.HasPrefix(arch, "windows-") {
			continue
		}

		wixArch := windowsArchToWix(arch)
		suffix := arch[len("windows-"):] // amd64, arm64, 386

		buildDir := filepath.Join("dist", "msi-build-"+suffix)
		if err := os.MkdirAll(buildDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create build directory: %w", err)
		}

		// Copy binary.
		binaryDest := filepath.Join(buildDir, cfg.Name+".exe")
		if err := p.copyFile(binPath, binaryDest); err != nil {
			return "", fmt.Errorf("failed to copy binary: %w", err)
		}

		// Generate WiX source file.
		wxsPath := filepath.Join(buildDir, cfg.Name+".wxs")
		if err := p.createWixSource(wxsPath, cfg, binaryDest); err != nil {
			return "", fmt.Errorf("failed to generate WiX file: %w", err)
		}

		outputPath := filepath.Join(outDir, fmt.Sprintf("%s-%s-%s.msi", cfg.Name, cfg.Version, suffix))
		if err := p.buildMSI(ctx, buildDir, wxsPath, outputPath, wixArch); err != nil {
			return "", err
		}
		built++
	}

	if built == 0 {
		return "", fmt.Errorf("no Windows binaries were processed")
	}

	return outDir, nil
}

// windowsArchToWix maps a windows-* architecture string to the wixl/WiX arch flag.
func windowsArchToWix(arch string) string {
	switch arch {
	case "windows-arm64":
		return "arm64"
	case "windows-386":
		return "x86"
	default: // windows-amd64
		return "x64"
	}
}

func (p *Packager) createWixSource(path string, cfg *config.Config, binaryPath string) error {
	// This template is compatible with both wixl (msitools) and the WiX toolset.
	// WixUI elements are omitted because wixl does not support WixUIExtension.
	tmpl := `<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product Id="*"
           Name="{{.Name}}"
           Language="1033"
           Version="{{.Version}}.0"
           Manufacturer="{{.AuthorName}}"
           UpgradeCode="{{.UpgradeCode}}">

    <Package InstallerVersion="200"
             Compressed="yes"
             InstallScope="perMachine"
             Description="{{.Description}}"
             Comments="{{.Description}}" />

    <MajorUpgrade DowngradeErrorMessage="A newer version of [ProductName] is already installed." />
    <MediaTemplate EmbedCab="yes" />

    <Feature Id="ProductFeature" Title="{{.Name}}" Level="1">
      <ComponentGroupRef Id="ProductComponents" />
    </Feature>

    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="ProgramFilesFolder">
        <Directory Id="INSTALLFOLDER" Name="{{.Name}}" />
      </Directory>
      <Directory Id="ProgramMenuFolder">
        <Directory Id="ApplicationProgramsFolder" Name="{{.Name}}" />
      </Directory>
    </Directory>

    <ComponentGroup Id="ProductComponents" Directory="INSTALLFOLDER">
      <Component Id="MainExecutable" Guid="{{.ComponentGuid}}">
        <File Id="MainExe"
              Source="{{.BinaryPath}}"
              KeyPath="yes"
              Name="{{.Name}}.exe" />

        <!-- Add to PATH -->
        <Environment Id="PATH" Name="PATH" Value="[INSTALLFOLDER]" Permanent="no" Part="last" Action="set" System="yes" />

        <!-- Start Menu shortcut -->
        <Shortcut Id="ApplicationStartMenuShortcut"
                  Name="{{.Name}}"
                  Description="{{.Description}}"
                  Target="[#MainExe]"
                  WorkingDirectory="INSTALLFOLDER"
                  Directory="ApplicationProgramsFolder" />

        <!-- Remove start menu folder on uninstall -->
        <RemoveFolder Id="ApplicationProgramsFolder" On="uninstall" />

        <!-- Registry key for Add/Remove Programs -->
        <RegistryValue Root="HKCU"
                       Key="Software\{{.AuthorName}}\{{.Name}}"
                       Name="installed"
                       Type="integer"
                       Value="1"
                       KeyPath="no" />
      </Component>
    </ComponentGroup>

    <!-- Custom properties -->
    <Property Id="ARPURLINFOABOUT" Value="{{.Homepage}}" />
    <Property Id="ARPCONTACT" Value="{{.AuthorName}}" />
    <Property Id="ARPHELPLINK" Value="{{.Homepage}}" />

  </Product>
</Wix>`

	t, err := template.New("wix").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	authorName := cfg.Author
	if strings.Contains(cfg.Author, "<") {
		parts := strings.Split(cfg.Author, "<")
		authorName = strings.TrimSpace(parts[0])
	}

	data := struct {
		*config.Config
		AuthorName    string
		BinaryPath    string
		UpgradeCode   string
		ComponentGuid string
	}{
		Config:        cfg,
		AuthorName:    authorName,
		BinaryPath:    binaryPath,
		UpgradeCode:   fmt.Sprintf("{%s-UPGRADE-CODE-GUID}", strings.ToUpper(cfg.Name)),
		ComponentGuid: fmt.Sprintf("{%s-COMPONENT-GUID}", strings.ToUpper(cfg.Name)),
	}

	return t.Execute(f, data)
}

func (p *Packager) createBuildScript(path string, cfg *config.Config) error {
	tmpl := `@echo off
REM Build script for {{.Name}} MSI installer
REM Requires WiX Toolset: https://wixtoolset.org/

set APP_NAME={{.Name}}
set VERSION={{.Version}}
set WXS_FILE=%APP_NAME%.wxs
set WIXOBJ_FILE=%APP_NAME%.wixobj
set MSI_FILE=%APP_NAME%-%VERSION%.msi

echo Building MSI installer for %APP_NAME% v%VERSION%...

REM Check if WiX is installed
where candle >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo ERROR: WiX Toolset not found in PATH
    echo Please install WiX Toolset from https://wixtoolset.org/
    exit /b 1
)

REM Compile WiX source
echo Compiling WiX source...
candle %WXS_FILE% -out %WIXOBJ_FILE%
if %ERRORLEVEL% neq 0 (
    echo ERROR: Failed to compile WiX source
    exit /b 1
)

REM Link to create MSI
echo Creating MSI installer...
light %WIXOBJ_FILE% -out %MSI_FILE% -ext WixUIExtension
if %ERRORLEVEL% neq 0 (
    echo ERROR: Failed to create MSI
    exit /b 1
)

REM Clean up
del %WIXOBJ_FILE%

echo.
echo ✅ Created %MSI_FILE%
echo.
echo Usage:
echo   msiexec /i %MSI_FILE%           (Install)
echo   msiexec /x %MSI_FILE%           (Uninstall)
echo   %MSI_FILE%                      (Interactive install)`

	t, err := template.New("build").Parse(tmpl)
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

func (p *Packager) createPowerShellScript(path string, cfg *config.Config) error {
	tmpl := `# PowerShell build script for {{.Name}} MSI installer
# Requires WiX Toolset: https://wixtoolset.org/

param(
    [switch]$Clean
)

$AppName = "{{.Name}}"
$Version = "{{.Version}}"
$WxsFile = "$AppName.wxs"
$WixObjFile = "$AppName.wixobj"
$MsiFile = "$AppName-$Version.msi"

Write-Host "Building MSI installer for $AppName v$Version..." -ForegroundColor Green

# Clean up previous build
if ($Clean -or (Test-Path $WixObjFile)) {
    Remove-Item $WixObjFile -ErrorAction SilentlyContinue
}
if ($Clean -or (Test-Path $MsiFile)) {
    Remove-Item $MsiFile -ErrorAction SilentlyContinue
}

# Check if WiX is installed
try {
    $null = Get-Command candle -ErrorAction Stop
    $null = Get-Command light -ErrorAction Stop
} catch {
    Write-Error "WiX Toolset not found in PATH. Please install from https://wixtoolset.org/"
    exit 1
}

# Compile WiX source
Write-Host "Compiling WiX source..." -ForegroundColor Yellow
& candle $WxsFile -out $WixObjFile
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to compile WiX source"
    exit 1
}

# Link to create MSI
Write-Host "Creating MSI installer..." -ForegroundColor Yellow
& light $WixObjFile -out $MsiFile -ext WixUIExtension
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to create MSI"
    exit 1
}

# Clean up
Remove-Item $WixObjFile -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "✅ Created $MsiFile" -ForegroundColor Green
Write-Host ""
Write-Host "Usage:" -ForegroundColor Cyan
Write-Host "  msiexec /i $MsiFile           (Install)" -ForegroundColor White
Write-Host "  msiexec /x $MsiFile           (Uninstall)" -ForegroundColor White
Write-Host "  ./$MsiFile                    (Interactive install)" -ForegroundColor White`

	t, err := template.New("ps1").Parse(tmpl)
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

// buildMSI tries available MSI build tools in priority order:
//  1. wixl (msitools) — "brew install msitools" — works on macOS and Linux for any arch
//  2. WiX candle/light — Windows only
//  3. go-msi — any platform
func (p *Packager) buildMSI(ctx context.Context, buildDir, wxsPath, outputPath, wixArch string) error {
	// wixl (msitools): cross-platform, supports --arch x64|arm64|x86
	if _, err := exec.LookPath("wixl"); err == nil {
		return p.buildWithWixl(ctx, wxsPath, outputPath, wixArch)
	}

	// WiX candle/light: Windows only
	if runtime.GOOS == "windows" {
		if err := p.buildWithWix(ctx, buildDir, wxsPath, outputPath); err == nil {
			return nil
		}
	}

	// go-msi: any platform, wraps WiX
	if _, err := exec.LookPath("go-msi"); err == nil {
		cfg := &config.Config{} // go-msi path reuses existing wix.json
		_, err := p.buildWithGoMSI(ctx, buildDir, cfg, outputPath)
		return err
	}

	return fmt.Errorf("MSI build tools not found - install WiX Toolset (Windows), go-msi, or msitools (brew install msitools)")
}

// buildWithWixl calls wixl from the msitools package.
// Install on macOS: brew install msitools
// Install on Linux: apt install msitools / dnf install msitools
func (p *Packager) buildWithWixl(ctx context.Context, wxsPath, outputPath, arch string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "wixl", "--arch", arch, "-o", outputPath, wxsPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wixl failed: %w\n%s", err, out)
	}
	return nil
}

func (p *Packager) buildWithWix(ctx context.Context, buildDir, wxsPath, outputPath string) error {
	if _, err := exec.LookPath("candle"); err != nil {
		return fmt.Errorf("candle not found")
	}
	if _, err := exec.LookPath("light"); err != nil {
		return fmt.Errorf("light not found")
	}

	wixobjPath := strings.TrimSuffix(wxsPath, ".wxs") + ".wixobj"

	candleCmd := exec.CommandContext(ctx, "candle", "-out", wixobjPath, wxsPath)
	candleCmd.Dir = buildDir
	if output, err := candleCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("candle failed: %w\nOutput: %s", err, output)
	}

	lightCmd := exec.CommandContext(ctx, "light", "-out", outputPath, wixobjPath, "-ext", "WixUIExtension")
	lightCmd.Dir = buildDir
	if output, err := lightCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("light failed: %w\nOutput: %s", err, output)
	}

	return nil
}

func (p *Packager) buildWithGoMSI(ctx context.Context, buildDir string, cfg *config.Config, outputPath string) (string, error) {
	goMSIConfig := fmt.Sprintf(`{
  "product-name": "%s",
  "company-name": "%s",
  "upgrade-code": "%s",
  "version": "%s",
  "license": "LICENSE",
  "description": "%s",
  "start-menu-shortcut": true,
  "env-vars": [],
  "files": {
    "guid": "*",
    "items": ["%s"]
  }
}`, cfg.Name, p.getAuthorName(cfg), p.generateUpgradeCode(cfg), cfg.Version, cfg.Description, cfg.Name+".exe")

	configPath := filepath.Join(buildDir, "wix.json")
	if err := os.WriteFile(configPath, []byte(goMSIConfig), 0644); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "go-msi", "make", "--msi", outputPath, "--version", cfg.Version)
	cmd.Dir = buildDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("go-msi failed: %w\nOutput: %s", err, output)
	}

	return outputPath, nil
}

func (p *Packager) generateUpgradeCode(cfg *config.Config) string {
	hash := 0
	for _, c := range cfg.Name {
		hash = hash*31 + int(c)
	}
	return fmt.Sprintf("{%08X-0000-0000-0000-000000000000}", hash&0xFFFFFFFF)
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
