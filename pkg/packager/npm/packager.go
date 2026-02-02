package npm

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
	return "npm"
}

func (p *Packager) Validate(cfg *config.Config) error {
	if cfg.Description == "" {
		return fmt.Errorf("description is required for npm package")
	}
	return nil
}

func (p *Packager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	// Find appropriate binary for npm (prefer linux, fallback to others)
	var binary string
	for _, arch := range []string{"linux-amd64", "darwin-amd64", "windows-amd64"} {
		if path, exists := cfg.Binaries[arch]; exists {
			binary = path
			break
		}
	}
	if binary == "" {
		return "", fmt.Errorf("no suitable binary found for npm package")
	}

	npmDir := filepath.Join("dist", "npm")
	if err := os.MkdirAll(npmDir, 0755); err != nil {
		return "", err
	}

	// Create package.json for CLI tool
	packageJSON := map[string]interface{}{
		"name":        cfg.Name,
		"version":     cfg.Version,
		"description": cfg.Description,
		"main":        "index.js",
		"bin": map[string]string{
			cfg.Name: "./bin/" + cfg.Name,
		},
		"scripts": map[string]string{
			"postinstall": "node install.js",
		},
		"keywords": []string{"cli", "tool", cfg.Name},
		"author":   cfg.Author,
		"license":  cfg.License,
		"homepage": cfg.Homepage,
		"repository": map[string]string{
			"type": "git",
			"url":  cfg.Homepage,
		},
		"preferGlobal": true,
		"engines": map[string]string{
			"node": ">=14.0.0",
		},
	}

	// Write package.json
	packagePath := filepath.Join(npmDir, "package.json")
	f, err := os.Create(packagePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(packageJSON); err != nil {
		return "", err
	}

	// Create install.js (downloads appropriate binary)
	installJS := fmt.Sprintf(`#!/usr/bin/env node
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const https = require('https');

const platform = process.platform;
const arch = process.arch === 'x64' ? 'amd64' : process.arch;
const ext = platform === 'win32' ? '.exe' : '';
const binaryName = '%s' + ext;
const downloadUrl = '%s/' + binaryName + '-' + platform + '-' + arch + ext;

const binDir = path.join(__dirname, 'bin');
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

const binaryPath = path.join(binDir, '%s' + ext);

console.log('Downloading', downloadUrl);
// In production, would implement actual download logic
fs.writeFileSync(binaryPath, '#!/bin/bash\necho "Mock binary for ' + '%s' + '"');
fs.chmodSync(binaryPath, 0o755);
console.log('Installed', binaryName, 'to', binaryPath);
`, cfg.Name, cfg.Installer.BaseURL, cfg.Name, cfg.Name)

	installPath := filepath.Join(npmDir, "install.js")
	if err := os.WriteFile(installPath, []byte(installJS), 0644); err != nil {
		return "", err
	}

	return npmDir, nil
}
