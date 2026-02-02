package apptainer

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
	return "apptainer"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Description == "" {
		return fmt.Errorf("description is required for Apptainer definition")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	// Find Linux binary (Apptainer primarily runs on Linux)
	var linuxBinary string
	for arch, path := range cfg.Binaries {
		if strings.HasPrefix(arch, "linux-") {
			linuxBinary = path
			break
		}
	}
	if linuxBinary == "" {
		return "", fmt.Errorf("no Linux binary found for Apptainer")
	}

	apptainerDir := filepath.Join("dist", "apptainer")
	if err := os.MkdirAll(apptainerDir, 0755); err != nil {
		return "", err
	}

	// Create Apptainer definition file
	defPath := filepath.Join(apptainerDir, fmt.Sprintf("%s.def", cfg.Name))
	if err := p.createDefinitionFile(defPath, cfg, linuxBinary); err != nil {
		return "", err
	}

	// Create build script
	buildScriptPath := filepath.Join(apptainerDir, "build.sh")
	if err := p.createBuildScript(buildScriptPath, cfg); err != nil {
		return "", err
	}

	return apptainerDir, nil
}

func (p *Packager) createDefinitionFile(path string, cfg *config.Config, binaryPath string) error {
	tmpl := `Bootstrap: library
From: ubuntu:22.04

%labels
    Author {{.Author}}
    Version {{.Version}}
    Description {{.Description}}
    {{if .Homepage}}URL {{.Homepage}}{{end}}

%help
    {{.Description}}
    
    Usage:
        apptainer exec {{.Name}}.sif {{.Name}} [options]
    
    {{if .Homepage}}More info: {{.Homepage}}{{end}}

%files
    {{.BinaryName}} /usr/local/bin/{{.Name}}

%post
    # Update system
    apt-get update && apt-get install -y \
        ca-certificates \
        && rm -rf /var/lib/apt/lists/*
    
    # Make binary executable
    chmod +x /usr/local/bin/{{.Name}}
    
    # Create symlink for convenience
    ln -sf /usr/local/bin/{{.Name}} /usr/bin/{{.Name}}

%environment
    export PATH="/usr/local/bin:$PATH"

%runscript
    exec /usr/local/bin/{{.Name}} "$@"

%test
    {{.Name}} --version || {{.Name}} --help || echo "{{.Name}} installed successfully"`

	t, err := template.New("definition").Parse(tmpl)
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

func (p *Packager) createBuildScript(path string, cfg *config.Config) error {
	tmpl := `#!/bin/bash
# Build script for {{.Name}} Apptainer container

set -e

DEFINITION_FILE="{{.Name}}.def"
OUTPUT_FILE="{{.Name}}.sif"

echo "Building Apptainer container for {{.Name}} v{{.Version}}..."

# Check if apptainer is installed
if ! command -v apptainer &> /dev/null; then
    echo "Error: Apptainer is not installed"
    echo "Install from: https://apptainer.org/docs/admin/main/installation.html"
    exit 1
fi

# Build the container
echo "Building container image..."
apptainer build --fakeroot "$OUTPUT_FILE" "$DEFINITION_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Successfully built $OUTPUT_FILE"
    echo ""
    echo "Usage:"
    echo "  apptainer exec $OUTPUT_FILE {{.Name}} --help"
    echo "  apptainer run $OUTPUT_FILE [args]"
    echo ""
    echo "For HPC environments:"
    echo "  srun apptainer exec $OUTPUT_FILE {{.Name}} [args]"
    echo "  sbatch --wrap=\"apptainer exec $OUTPUT_FILE {{.Name}} [args]\""
else
    echo "❌ Build failed"
    exit 1
fi`

	t, err := template.New("build").Parse(tmpl)
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
