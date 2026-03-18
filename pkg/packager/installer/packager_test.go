/*
Copyright 2026 Scott Friedman

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package installer

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestInstallerPackager(t *testing.T) {
	p := New()

	if p.Name() != "installer" {
		t.Errorf("Expected name 'installer', got %s", p.Name())
	}

	// Test validation
	cfg := &config.Config{
		Name:    "test",
		Version: "1.0.0",
		Installer: config.InstallerConfig{
			BaseURL: "https://example.com/releases",
		},
	}

	err := p.Validate(cfg)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test validation failure
	cfg.Installer.BaseURL = ""
	err = p.Validate(cfg)
	if err == nil {
		t.Error("Expected validation to fail without base_url")
	}
}

func TestInstallerPack(t *testing.T) {
	p := New()
	cfg := &config.Config{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test app",
		Installer: config.InstallerConfig{
			BaseURL:        "https://example.com/releases",
			InstallPath:    "/usr/local/bin",
			VerifyChecksum: true,
		},
	}

	ctx := context.Background()
	output, err := p.Pack(ctx, cfg)
	if err != nil {
		t.Errorf("Pack failed: %v", err)
	}

	if output == "" {
		t.Error("Expected output path")
	}

	// Check if file was created and is executable
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Errorf("Output file not created: %s", output)
	}

	// Clean up
	os.Remove(output)
}

func TestInstallerPack_ContentVerification(t *testing.T) {
	// Run in a temp dir so the output file is cleaned up automatically.
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	p := New()
	cfg := &config.Config{
		Name:    "myapp",
		Version: "2.3.4",
		Installer: config.InstallerConfig{
			BaseURL:        "https://dl.example.com/v2.3.4",
			InstallPath:    "/usr/local/bin",
			VerifyChecksum: true,
		},
	}

	outputPath, err := p.Pack(ctx, cfg)
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	script := string(content)

	for _, want := range []string{
		"myapp",
		"2.3.4",
		"https://dl.example.com/v2.3.4",
		"/usr/local/bin",
		"sha256sum",   // VerifyChecksum=true
		"#!/bin/bash", // shebang
	} {
		if !strings.Contains(script, want) {
			t.Errorf("expected %q in install script", want)
		}
	}

	// File should be executable.
	info, _ := os.Stat(outputPath)
	if info.Mode()&0111 == 0 {
		t.Error("install script should be executable")
	}
}

func TestInstallerPack_NoChecksumBlock(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	p := New()
	cfg := &config.Config{
		Name:    "app",
		Version: "1.0.0",
		Installer: config.InstallerConfig{
			BaseURL:        "https://example.com",
			VerifyChecksum: false,
		},
	}

	outputPath, err := p.Pack(ctx, cfg)
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}

	content, _ := os.ReadFile(outputPath)
	if strings.Contains(string(content), "sha256sum") {
		t.Error("expected no sha256sum block when VerifyChecksum=false")
	}
}

// ctx is a package-level background context for test convenience.
var ctx = context.Background()
