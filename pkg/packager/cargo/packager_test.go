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

package cargo

import (
	"context"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestCargoPackager(t *testing.T) {
	p := New()

	if p.Name() != "cargo" {
		t.Errorf("Expected name 'cargo', got %s", p.Name())
	}

	// Test validation
	cfg := &config.Config{
		Name:     "test",
		Version:  "1.0.0",
		Homepage: "https://example.com",
	}

	err := p.Validate(cfg)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test validation failure
	cfg.Homepage = ""
	err = p.Validate(cfg)
	if err == nil {
		t.Error("Expected validation to fail without homepage")
	}
}

func TestCargoPack(t *testing.T) {
	p := New()
	cfg := &config.Config{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test app",
		Homepage:    "https://example.com",
		License:     "Apache-2.0",
		Author:      "Test Author",
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
