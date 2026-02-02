package snap

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
	return "snap"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Description == "" {
		return fmt.Errorf("description is required for snap package")
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

	snapDir := filepath.Join("dist", "snap")
	if err := os.MkdirAll(snapDir, 0755); err != nil {
		return "", err
	}

	// Create snapcraft.yaml
	snapcraftPath := filepath.Join(snapDir, "snapcraft.yaml")
	if err := p.createSnapcraft(snapcraftPath, cfg, linuxBinary); err != nil {
		return "", err
	}

	return snapDir, nil
}

func (p *Packager) createSnapcraft(path string, cfg *config.Config, binaryPath string) error {
	tmpl := `name: {{.Name}}
version: '{{.Version}}'
summary: {{.Description}}
description: |
  {{.Description}}
  
  {{.Homepage}}

grade: stable
confinement: strict
base: core22

apps:
  {{.Name}}:
    command: bin/{{.Name}}
    plugs:
      - home
      - network
      - network-bind

parts:
  {{.Name}}:
    plugin: dump
    source: .
    organize:
      '{{.BinaryName}}': bin/{{.Name}}
    stage:
      - bin/{{.Name}}
    prime:
      - bin/{{.Name}}`

	t, err := template.New("snapcraft").Parse(tmpl)
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
		BinaryName string
	}{
		Config:     cfg,
		BinaryName: filepath.Base(binaryPath),
	}

	return t.Execute(f, data)
}
