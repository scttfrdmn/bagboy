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

package docker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestDockerPackager(t *testing.T) {
	p := New()

	if p.Name() != "docker" {
		t.Errorf("Expected name 'docker', got %s", p.Name())
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

func TestDockerPack(t *testing.T) {
	p := New()
	cfg := &config.Config{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test app",
		Homepage:    "https://example.com",
		License:     "Apache-2.0",
		Author:      "Test Author",
		Binaries:    map[string]string{"linux-amd64": "test-binary"},
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

func TestDockerValidation_NoDescription(t *testing.T) {
	p := New()
	cfg := &config.Config{
		Name:    "myapp",
		Version: "1.0.0",
		// Description missing
		Binaries: map[string]string{"linux-amd64": "bin"},
	}
	if err := p.Validate(cfg); err == nil {
		t.Error("expected error when description is missing")
	}
}

func TestDockerPack_DockerfileContent(t *testing.T) {
	testDir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	p := New()
	cfg := &config.Config{
		Name:        "myapp",
		Version:     "1.2.3",
		Description: "My cool app",
		Homepage:    "https://myapp.example.com",
		Author:      "Jane Doe",
		Binaries:    map[string]string{"linux-amd64": "dist/myapp-linux-amd64"},
	}

	outputDir, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}

	// Verify Dockerfile exists and contains expected content.
	dockerfile := filepath.Join(outputDir, "Dockerfile")
	content, err := os.ReadFile(dockerfile)
	if err != nil {
		t.Fatalf("Dockerfile not created: %v", err)
	}
	df := string(content)
	for _, want := range []string{
		"FROM alpine:latest",
		"FROM scratch",
		"ENTRYPOINT [\"/myapp\"]",
		"myapp",
		"1.2.3",
		"My cool app",
		"Jane Doe",
	} {
		if !strings.Contains(df, want) {
			t.Errorf("Dockerfile should contain %q", want)
		}
	}

	// Verify .dockerignore exists.
	if _, err := os.Stat(filepath.Join(outputDir, ".dockerignore")); err != nil {
		t.Error(".dockerignore should be created")
	}

	// Verify docker-compose.yml exists and contains the image name.
	compose, _ := os.ReadFile(filepath.Join(outputDir, "docker-compose.yml"))
	if !strings.Contains(string(compose), "myapp") {
		t.Error("docker-compose.yml should reference image 'myapp'")
	}

	// Verify build script is executable.
	buildScript := filepath.Join(outputDir, "build.sh")
	info, err := os.Stat(buildScript)
	if err != nil {
		t.Fatalf("build.sh not created: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("build.sh should be executable")
	}
}

func TestDockerPack_NoLinuxBinary(t *testing.T) {
	testDir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	p := New()
	cfg := &config.Config{
		Name:        "myapp",
		Version:     "1.0.0",
		Description: "app",
		// Only darwin binary — no linux
		Binaries: map[string]string{"darwin-amd64": "dist/myapp-darwin-amd64"},
	}

	_, err := p.Pack(context.Background(), cfg)
	if err == nil {
		t.Error("expected error when no linux binary is provided")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "linux") {
		t.Errorf("expected linux-related error, got: %v", err)
	}
}
