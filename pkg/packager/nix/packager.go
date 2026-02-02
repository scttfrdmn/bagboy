package nix

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
	return "nix"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Homepage == "" {
		return fmt.Errorf("homepage is required for Nix package")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	nixDir := filepath.Join("dist", "nix")
	if err := os.MkdirAll(nixDir, 0755); err != nil {
		return "", err
	}

	// Create default.nix
	defaultNixPath := filepath.Join(nixDir, "default.nix")
	if err := p.createDefaultNix(defaultNixPath, cfg); err != nil {
		return "", err
	}

	// Create flake.nix (modern Nix)
	flakePath := filepath.Join(nixDir, "flake.nix")
	if err := p.createFlake(flakePath, cfg); err != nil {
		return "", err
	}

	// Create shell.nix for development
	shellPath := filepath.Join(nixDir, "shell.nix")
	if err := p.createShell(shellPath, cfg); err != nil {
		return "", err
	}

	return nixDir, nil
}

func (p *Packager) createDefaultNix(path string, cfg *config.Config) error {
	tmpl := `{ lib
, stdenv
, fetchurl
}:

stdenv.mkDerivation rec {
  pname = "{{.Name}}";
  version = "{{.Version}}";

  src = fetchurl {
    url = "{{.BaseURL}}/{{.Name}}-linux-amd64";
    sha256 = "0000000000000000000000000000000000000000000000000000000000000000"; # TODO: Add real hash
  };

  dontUnpack = true;
  dontBuild = true;

  installPhase = ''
    install -D $src $out/bin/{{.Name}}
    chmod +x $out/bin/{{.Name}}
  '';

  meta = with lib; {
    description = "{{.Description}}";
    homepage = "{{.Homepage}}";
    license = licenses.{{.NixLicense}};
    maintainers = [ ];
    platforms = platforms.linux ++ platforms.darwin;
    mainProgram = "{{.Name}}";
  };
}`

	t, err := template.New("default").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Map common licenses to Nix license names
	nixLicense := "mit"
	switch strings.ToLower(cfg.License) {
	case "apache-2.0", "apache":
		nixLicense = "asl20"
	case "gpl-3.0", "gpl3":
		nixLicense = "gpl3"
	case "bsd-3-clause", "bsd":
		nixLicense = "bsd3"
	}

	data := struct {
		*config.Config
		BaseURL    string
		NixLicense string
	}{
		Config:     cfg,
		BaseURL:    cfg.Installer.BaseURL,
		NixLicense: nixLicense,
	}

	return t.Execute(f, data)
}

func (p *Packager) createFlake(path string, cfg *config.Config) error {
	tmpl := `{
  description = "{{.Description}}";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        {{.Name}} = pkgs.callPackage ./default.nix { };
      in
      {
        packages = {
          default = {{.Name}};
          {{.Name}} = {{.Name}};
        };

        apps = {
          default = flake-utils.lib.mkApp {
            drv = {{.Name}};
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            {{.Name}}
          ];
        };
      });
}`

	t, err := template.New("flake").Parse(tmpl)
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

func (p *Packager) createShell(path string, cfg *config.Config) error {
	tmpl := `{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    (callPackage ./default.nix { })
  ];

  shellHook = ''
    echo "{{.Name}} development environment"
    echo "Run '{{.Name}} --help' to get started"
  '';
}`

	t, err := template.New("shell").Parse(tmpl)
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
