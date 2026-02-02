package brew

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

type Packager struct{}

func New() *Packager {
	return &Packager{}
}

func (p *Packager) Name() string {
	return "brew"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Homepage == "" {
		return fmt.Errorf("homepage is required for brew formula")
	}
	return nil
}

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

	data := struct {
		*config.Config
		ClassName string
		BaseURL   string
		Test      string
	}{
		Config:    cfg,
		ClassName: capitalize(cfg.Name),
		BaseURL:   cfg.Installer.BaseURL,
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
