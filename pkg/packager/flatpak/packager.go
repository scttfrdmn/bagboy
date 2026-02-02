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

type Packager struct{}

func New() *Packager {
	return &Packager{}
}

func (p *Packager) Name() string {
	return "flatpak"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Homepage == "" {
		return fmt.Errorf("homepage is required for flatpak manifest")
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

	appId := fmt.Sprintf("dev.bagboy.%s", strings.Title(cfg.Name))

	manifest := map[string]interface{}{
		"app-id":          appId,
		"runtime":         "org.freedesktop.Platform",
		"runtime-version": "22.08",
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
