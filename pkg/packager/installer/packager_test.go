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
