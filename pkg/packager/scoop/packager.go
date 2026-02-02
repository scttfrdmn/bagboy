package scoop

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

type Packager struct{}

func New() *Packager {
	return &Packager{}
}

func (p *Packager) Name() string {
	return "scoop"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Homepage == "" {
		return fmt.Errorf("homepage is required for scoop manifest")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	manifest := map[string]interface{}{
		"version":     cfg.Version,
		"description": cfg.Description,
		"homepage":    cfg.Homepage,
		"license":     cfg.License,
		"url":         fmt.Sprintf("%s/%s-windows-amd64.exe", cfg.Installer.BaseURL, cfg.Name),
		"hash":        "sha256:TODO", // Would need actual hash
		"bin":         cfg.Name + ".exe",
	}

	if cfg.Packages.Scoop.Bin != "" {
		manifest["bin"] = cfg.Packages.Scoop.Bin
	}

	if len(cfg.Packages.Scoop.Shortcuts) > 0 {
		manifest["shortcuts"] = cfg.Packages.Scoop.Shortcuts
	}

	outputPath := filepath.Join("dist", cfg.Name+".json")
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
