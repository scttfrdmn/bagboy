package pypi

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
	return "pypi"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Author == "" {
		return fmt.Errorf("author is required for PyPI package")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	pypiDir := filepath.Join("dist", "pypi")
	if err := os.MkdirAll(pypiDir, 0755); err != nil {
		return "", err
	}

	// Create setup.py
	setupPath := filepath.Join(pypiDir, "setup.py")
	if err := p.createSetupPy(setupPath, cfg); err != nil {
		return "", err
	}

	// Create pyproject.toml (modern Python packaging)
	pyprojectPath := filepath.Join(pypiDir, "pyproject.toml")
	if err := p.createPyprojectToml(pyprojectPath, cfg); err != nil {
		return "", err
	}

	// Create package directory
	pkgDir := filepath.Join(pypiDir, strings.ReplaceAll(cfg.Name, "-", "_"))
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return "", err
	}

	// Create __init__.py
	initPath := filepath.Join(pkgDir, "__init__.py")
	initContent := fmt.Sprintf(`"""
%s

%s
"""

__version__ = "%s"
`, cfg.Name, cfg.Description, cfg.Version)
	if err := os.WriteFile(initPath, []byte(initContent), 0644); err != nil {
		return "", err
	}

	// Create main.py (CLI entry point)
	mainPath := filepath.Join(pkgDir, "main.py")
	if err := p.createMainPy(mainPath, cfg); err != nil {
		return "", err
	}

	return pypiDir, nil
}

func (p *Packager) createSetupPy(path string, cfg *config.Config) error {
	tmpl := `#!/usr/bin/env python3
"""Setup script for {{.Name}}."""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="{{.Name}}",
    version="{{.Version}}",
    author="{{.AuthorName}}",
    author_email="{{.AuthorEmail}}",
    description="{{.Description}}",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="{{.Homepage}}",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
    ],
    python_requires=">=3.8",
    entry_points={
        "console_scripts": [
            "{{.Name}}={{.PackageName}}.main:main",
        ],
    },
    install_requires=[
        "requests>=2.25.0",
    ],
)`

	t, err := template.New("setup").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Parse author name and email
	authorName := cfg.Author
	authorEmail := ""
	if strings.Contains(cfg.Author, "<") && strings.Contains(cfg.Author, ">") {
		parts := strings.Split(cfg.Author, "<")
		authorName = strings.TrimSpace(parts[0])
		authorEmail = strings.Trim(parts[1], ">")
	}

	data := struct {
		*config.Config
		AuthorName  string
		AuthorEmail string
		PackageName string
	}{
		Config:      cfg,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		PackageName: strings.ReplaceAll(cfg.Name, "-", "_"),
	}

	return t.Execute(f, data)
}

func (p *Packager) createPyprojectToml(path string, cfg *config.Config) error {
	tmpl := `[build-system]
requires = ["setuptools>=45", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "{{.Name}}"
version = "{{.Version}}"
description = "{{.Description}}"
readme = "README.md"
license = {text = "{{.License}}"}
authors = [
    {name = "{{.AuthorName}}", email = "{{.AuthorEmail}}"},
]
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Operating System :: OS Independent",
    "Programming Language :: Python :: 3",
]
requires-python = ">=3.8"
dependencies = [
    "requests>=2.25.0",
]

[project.urls]
Homepage = "{{.Homepage}}"
Repository = "{{.Homepage}}"

[project.scripts]
{{.Name}} = "{{.PackageName}}.main:main"`

	t, err := template.New("pyproject").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Parse author name and email
	authorName := cfg.Author
	authorEmail := ""
	if strings.Contains(cfg.Author, "<") && strings.Contains(cfg.Author, ">") {
		parts := strings.Split(cfg.Author, "<")
		authorName = strings.TrimSpace(parts[0])
		authorEmail = strings.Trim(parts[1], ">")
	}

	data := struct {
		*config.Config
		AuthorName  string
		AuthorEmail string
		PackageName string
	}{
		Config:      cfg,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		PackageName: strings.ReplaceAll(cfg.Name, "-", "_"),
	}

	return t.Execute(f, data)
}

func (p *Packager) createMainPy(path string, cfg *config.Config) error {
	tmpl := `#!/usr/bin/env python3
"""
Main entry point for {{.Name}} CLI tool.
Downloads and executes the appropriate binary for the current platform.
"""

import os
import sys
import platform
import subprocess
import urllib.request
from pathlib import Path

def get_binary_url():
    """Get the download URL for the current platform."""
    system = platform.system().lower()
    machine = platform.machine().lower()
    
    # Map Python platform names to our binary names
    if machine in ['x86_64', 'amd64']:
        arch = 'amd64'
    elif machine in ['aarch64', 'arm64']:
        arch = 'arm64'
    else:
        arch = 'amd64'  # fallback
    
    if system == 'darwin':
        binary_name = '{{.Name}}-darwin-' + arch
    elif system == 'linux':
        binary_name = '{{.Name}}-linux-' + arch
    elif system == 'windows':
        binary_name = '{{.Name}}-windows-' + arch + '.exe'
    else:
        raise RuntimeError(f"Unsupported platform: {system}")
    
    return f"{{.BaseURL}}/{binary_name}"

def get_binary_path():
    """Get the local path where the binary should be stored."""
    cache_dir = Path.home() / '.cache' / '{{.Name}}'
    cache_dir.mkdir(parents=True, exist_ok=True)
    
    system = platform.system().lower()
    ext = '.exe' if system == 'windows' else ''
    return cache_dir / f'{{.Name}}{ext}'

def download_binary():
    """Download the binary if it doesn't exist."""
    binary_path = get_binary_path()
    
    if binary_path.exists():
        return binary_path
    
    print(f"Downloading {{.Name}} binary...")
    url = get_binary_url()
    
    try:
        urllib.request.urlretrieve(url, binary_path)
        binary_path.chmod(0o755)
        print(f"Downloaded to {binary_path}")
        return binary_path
    except Exception as e:
        print(f"Failed to download binary: {e}")
        print("This is a mock implementation. In production, would download from:", url)
        # Create mock binary for demo
        binary_path.write_text(f'#!/bin/bash\necho "Mock {{.Name}} binary"\n')
        binary_path.chmod(0o755)
        return binary_path

def main():
    """Main entry point."""
    try:
        binary_path = download_binary()
        # Execute the binary with all arguments
        result = subprocess.run([str(binary_path)] + sys.argv[1:])
        sys.exit(result.returncode)
    except KeyboardInterrupt:
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()`

	t, err := template.New("main").Parse(tmpl)
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
		BaseURL string
	}{
		Config:  cfg,
		BaseURL: cfg.Installer.BaseURL,
	}

	return t.Execute(f, data)
}
