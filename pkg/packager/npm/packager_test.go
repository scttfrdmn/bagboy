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

package npm

import (
	"context"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestNpmPackager(t *testing.T) {
	p := New()

	if p.Name() != "npm" {
		t.Errorf("Expected name 'npm', got %s", p.Name())
	}

	// Test validation
	cfg := &config.Config{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test app",
		Binaries:    map[string]string{"linux-amd64": "test-binary"},
	}

	err := p.Validate(cfg)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test validation failure
	cfg.Description = ""
	err = p.Validate(cfg)
	if err == nil {
		t.Error("Expected validation to fail without description")
	}
}

func TestNpmPack(t *testing.T) {
	p := New()
	cfg := &config.Config{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test app",
		Homepage:    "https://example.com",
		License:     "Apache-2.0",
		Author:      "Test Author",
		Binaries:    map[string]string{"linux-amd64": "test-binary"},
		Installer: config.InstallerConfig{
			BaseURL: "https://example.com/releases",
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
}
