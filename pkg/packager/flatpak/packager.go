// Package flatpak implements the flatpak packager for bagboy.
package flatpak

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// Packager creates flatpak packages for Linux Flatpak app bundle distribution.
type Packager struct{}

// New returns a new flatpak Packager.
func New() *Packager {
	return &Packager{}
}

// Name returns the packager name "flatpak".
func (p *Packager) Name() string {
	return "flatpak"
}

// Validate checks that cfg has the required fields for flatpak packaging.
func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Homepage == "" {
		return fmt.Errorf("homepage is required for flatpak manifest")
	}
	return nil
}

// Pack generates flatpak artifacts from cfg and returns the output path.
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

	appId := fmt.Sprintf("dev.bagboy.%s", strings.Title(cfg.Name))

	manifest := map[string]interface{}{
		"app-id":          appId,
		"runtime":         "org.freedesktop.Platform",
		"runtime-version": "24.08",
		"sdk":             "org.freedesktop.Sdk",
		"command":         cfg.Name,
		"finish-args": []string{
			"--share=network",
			"--filesystem=home",
		},
		"modules": []map[string]interface{}{
			{
				"name":        cfg.Name,
				"buildsystem": "simple",
				"build-commands": []string{
					fmt.Sprintf("install -Dm755 %s /app/bin/%s", filepath.Base(linuxBinary), cfg.Name),
				},
				"sources": []map[string]interface{}{
					{
						"type": "file",
						"path": filepath.Base(linuxBinary),
					},
				},
			},
		},
	}

	outputPath := filepath.Join("dist", fmt.Sprintf("%s.json", appId))
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return "", err
	}

	return outputPath, nil
}
