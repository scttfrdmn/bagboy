package spack

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
	return "spack"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Homepage == "" {
		return fmt.Errorf("homepage is required for Spack package")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	spackDir := filepath.Join("dist", "spack")
	if err := os.MkdirAll(spackDir, 0755); err != nil {
		return "", err
	}

	// Create package.py file
	packagePath := filepath.Join(spackDir, "package.py")
	if err := p.createPackageFile(packagePath, cfg); err != nil {
		return "", err
	}

	// Create installation instructions
	instructionsPath := filepath.Join(spackDir, "INSTALL.md")
	if err := p.createInstructions(instructionsPath, cfg); err != nil {
		return "", err
	}

	return spackDir, nil
}

func (p *Packager) createPackageFile(path string, cfg *config.Config) error {
	tmpl := `# Copyright 2013-2024 Lawrence Livermore National Security, LLC and other
# Spack Project Developers. See the top-level COPYRIGHT file for details.
#
# SPDX-License-Identifier: (Apache-2.0 OR MIT)

from spack.package import *


class {{.ClassName}}(Package):
    """{{.Description}}"""

    homepage = "{{.Homepage}}"
    url = "{{.DownloadURL}}"
    git = "{{.GitURL}}"

    maintainers("{{.Maintainer}}")

    license("{{.License}}")

    version("{{.Version}}", sha256="{{.SHA256}}")

    # Dependencies
    depends_on("c", type="build")  # generated

    def install(self, spec, prefix):
        # Install the binary
        install("{{.BinaryName}}", prefix.bin)

    def setup_run_environment(self, env):
        env.prepend_path("PATH", self.prefix.bin)

    @run_after("install")
    def check_install(self):
        """Run basic checks on the installed package."""
        {{.Name}} = Executable(join_path(self.prefix.bin, "{{.Name}}"))
        {{.Name}}("--version", output=str.split, error=str.split)`

	t, err := template.New("package").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Find a Linux binary for the package
	var linuxBinary string
	for arch, path := range cfg.Binaries {
		if strings.HasPrefix(arch, "linux-") {
			linuxBinary = path
			break
		}
	}

	// Extract maintainer from author
	maintainer := cfg.Author
	if strings.Contains(cfg.Author, "<") {
		parts := strings.Split(cfg.Author, "<")
		maintainer = strings.TrimSpace(parts[0])
	}

	data := struct {
		*config.Config
		ClassName   string
		DownloadURL string
		GitURL      string
		Maintainer  string
		SHA256      string
		BinaryName  string
	}{
		Config:      cfg,
		ClassName:   strings.Title(cfg.Name),
		DownloadURL: fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s-linux-amd64.tar.gz", cfg.Name, cfg.Version, cfg.Name),
		GitURL:      cfg.Homepage,
		Maintainer:  maintainer,
		SHA256:      "0000000000000000000000000000000000000000000000000000000000000000",  // Placeholder
		BinaryName:  filepath.Base(linuxBinary),
	}

	return t.Execute(f, data)
}

func (p *Packager) createInstructions(path string, cfg *config.Config) error {
	tmpl := `# Spack Package Installation

## Adding to Spack

1. **Copy package to Spack repository:**
   ` + "```bash" + `
   cp package.py $SPACK_ROOT/var/spack/repos/builtin/packages/{{.Name}}/package.py
   ` + "```" + `

2. **Or create a custom repository:**
   ` + "```bash" + `
   spack repo create {{.Name}}-repo
   cp package.py {{.Name}}-repo/packages/{{.Name}}/package.py
   spack repo add {{.Name}}-repo
   ` + "```" + `

## Installation

` + "```bash" + `
# Install with default settings
spack install {{.Name}}

# Install with specific compiler
spack install {{.Name}} %gcc@11.2.0

# Install with MPI support (if applicable)
spack install {{.Name}} +mpi

# Load the package
spack load {{.Name}}
` + "```" + `

## Usage in HPC Environment

` + "```bash" + `
# In job script
#!/bin/bash
#SBATCH --job-name={{.Name}}-job
#SBATCH --nodes=1
#SBATCH --time=01:00:00

# Load Spack environment
source $SPACK_ROOT/share/spack/setup-env.sh
spack load {{.Name}}

# Run your application
{{.Name}} [options]
` + "```" + `

## Package Information

- **Name**: {{.Name}}
- **Version**: {{.Version}}
- **Description**: {{.Description}}
- **Homepage**: {{.Homepage}}
- **License**: {{.License}}

## Notes

- Update the SHA256 checksum in package.py with the actual hash
- Modify dependencies as needed for your specific requirements
- Consider adding variants for different build options
- Test installation on your target HPC system`

	t, err := template.New("instructions").Parse(tmpl)
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
