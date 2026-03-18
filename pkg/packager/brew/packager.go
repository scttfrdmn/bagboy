// Package brew implements the brew packager for bagboy.
package brew

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// Packager creates brew packages for Homebrew macOS distribution.
type Packager struct{}

// New returns a new brew Packager.
func New() *Packager {
	return &Packager{}
}

// Name returns the packager name "brew".
func (p *Packager) Name() string {
	return "brew"
}

// Validate checks that cfg has the required fields for brew packaging.
func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Homepage == "" {
		return fmt.Errorf("homepage is required for brew formula")
	}
	return nil
}

// Pack generates brew artifacts from cfg and returns the output path.
func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	tmpl := `class {{.ClassName}} < Formula
  desc "{{.Description}}"
  homepage "{{.Homepage}}"
  version "{{.Version}}"
  license "{{.License}}"

  {{range $arch, $binary := .Binaries}}
  {{if eq $arch "darwin-amd64"}}
  if Hardware::CPU.intel?
    url "{{$.BaseURL}}/{{$.Name}}-darwin-amd64"
    sha256 "TODO_CHECKSUM_AMD64"
  end
  {{end}}
  {{if eq $arch "darwin-arm64"}}
  if Hardware::CPU.arm?
    url "{{$.BaseURL}}/{{$.Name}}-darwin-arm64"
    sha256 "TODO_CHECKSUM_ARM64"
  end
  {{end}}
  {{end}}

  def install
    bin.install "{{.Name}}"
  end

  {{if .Test}}
  test do
    {{.Test}}
  end
  {{end}}
end`

	t, err := template.New("formula").Parse(tmpl)
	if err != nil {
		return "", err
	}

	// Interpolate {{.Version}} in the base URL so download URLs are fully resolved.
	baseURL := strings.ReplaceAll(cfg.Installer.BaseURL, "{{.Version}}", cfg.Version)

	data := struct {
		*config.Config
		ClassName string
		BaseURL   string
		Test      string
	}{
		Config:    cfg,
		ClassName: capitalize(cfg.Name),
		BaseURL:   baseURL,
		Test:      cfg.Packages.Brew.Test,
	}

	outputPath := filepath.Join("dist", cfg.Name+".rb")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return "", err
	}

	return outputPath, nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
