package appimage

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
	return "appimage"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if len(cfg.Packages.AppImage.Categories) == 0 {
		return fmt.Errorf("appimage.categories is required")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	// Find Linux binary
	var linuxBinary string
	for arch, path := range cfg.Binaries {
		if strings.HasPrefix(arch, "linux-") {
			linuxBinary = path
			break
		}
	}
	if linuxBinary == "" {
		return "", fmt.Errorf("no Linux binary found")
	}

	appDir := filepath.Join("dist", cfg.Name+".AppDir")
	if err := os.RemoveAll(appDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	// Create AppDir structure
	if err := p.createAppDirStructure(appDir, cfg, linuxBinary); err != nil {
		return "", err
	}

	// Build AppImage
	return p.buildAppImage(ctx, appDir, cfg)
}

func (p *Packager) createAppDirStructure(appDir string, cfg *config.Config, binaryPath string) error {
	// Create directories
	dirs := []string{
		filepath.Join(appDir, "usr", "bin"),
		filepath.Join(appDir, "usr", "share", "applications"),
		filepath.Join(appDir, "usr", "share", "icons", "hicolor", "256x256", "apps"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Copy binary
	binDest := filepath.Join(appDir, "usr", "bin", cfg.Name)
	if err := p.copyFile(binaryPath, binDest); err != nil {
		return err
	}
	if err := os.Chmod(binDest, 0755); err != nil {
		return err
	}

	// Create AppRun
	appRunPath := filepath.Join(appDir, "AppRun")
	if err := p.createAppRun(appRunPath, cfg); err != nil {
		return err
	}

	// Create desktop file
	desktopPath := filepath.Join(appDir, "usr", "share", "applications", cfg.Name+".desktop")
	if err := p.createDesktopFile(desktopPath, cfg); err != nil {
		return err
	}

	// Create symlinks for AppImage convention
	if err := os.Symlink("usr/share/applications/"+cfg.Name+".desktop", filepath.Join(appDir, cfg.Name+".desktop")); err != nil {
		return err
	}

	return nil
}

func (p *Packager) createAppRun(path string, cfg *config.Config) error {
	tmpl := `#!/bin/bash
# AppRun script for {{.Name}}
HERE="$(dirname "$(readlink -f "${0}")")"
export PATH="${HERE}/usr/bin:${PATH}"
exec "${HERE}/usr/bin/{{.Name}}" "$@"`

	t, err := template.New("apprun").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	return t.Execute(f, cfg)
}

func (p *Packager) createDesktopFile(path string, cfg *config.Config) error {
	tmpl := `[Desktop Entry]
Type={{.Type}}
Name={{.Name}}
Comment={{.Description}}
Exec={{.Name}}
Icon={{.Name}}
Categories={{.Categories}}
Terminal={{.Terminal}}`

	t, err := template.New("desktop").Parse(tmpl)
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
		Type       string
		Categories string
		Terminal   string
	}{
		Config:     cfg,
		Type:       cfg.Packages.AppImage.DesktopEntry.Type,
		Categories: strings.Join(cfg.Packages.AppImage.Categories, ";"),
		Terminal:   fmt.Sprintf("%t", cfg.Packages.AppImage.DesktopEntry.Terminal),
	}

	if data.Type == "" {
		data.Type = "Application"
	}

	return t.Execute(f, data)
}

func (p *Packager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

func (p *Packager) buildAppImage(ctx context.Context, appDir string, cfg *config.Config) (string, error) {
	outputPath := filepath.Join("dist", fmt.Sprintf("%s-%s-x86_64.AppImage", cfg.Name, cfg.Version))

	// Try appimagetool first
	if _, err := exec.LookPath("appimagetool"); err == nil {
		return p.buildWithAppimagetool(ctx, appDir, outputPath)
	}

	// Fallback to manual squashfs creation
	if _, err := exec.LookPath("mksquashfs"); err == nil {
		return p.buildWithSquashfs(ctx, appDir, outputPath)
	}

	return "", fmt.Errorf("neither appimagetool nor mksquashfs found - install AppImageKit or squashfs-tools")
}

func (p *Packager) buildWithAppimagetool(ctx context.Context, appDir, outputPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "appimagetool", appDir, outputPath)
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("appimagetool failed: %w\nOutput: %s", err, output)
	}

	return outputPath, nil
}

func (p *Packager) buildWithSquashfs(ctx context.Context, appDir, outputPath string) (string, error) {
	// Create squashfs filesystem
	squashfsPath := outputPath + ".squashfs"
	cmd := exec.CommandContext(ctx, "mksquashfs", appDir, squashfsPath, "-root-owned", "-noappend")
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mksquashfs failed: %w\nOutput: %s", err, output)
	}

	// Create AppImage by prepending runtime (simplified version)
	if err := p.createAppImageFromSquashfs(squashfsPath, outputPath); err != nil {
		os.Remove(squashfsPath)
		return "", err
	}

	os.Remove(squashfsPath)
	return outputPath, nil
}

func (p *Packager) createAppImageFromSquashfs(squashfsPath, outputPath string) error {
	// This is a simplified version - in production would need proper AppImage runtime
	squashfsData, err := os.ReadFile(squashfsPath)
	if err != nil {
		return err
	}

	// Create a basic AppImage header (simplified)
	header := fmt.Sprintf("#!/bin/sh\n# AppImage created by bagboy\n# This is a simplified AppImage - use appimagetool for production\necho 'AppImage would execute here'\n")
	
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(header); err != nil {
		return err
	}

	if _, err := file.Write(squashfsData); err != nil {
		return err
	}

	return os.Chmod(outputPath, 0755)
}
