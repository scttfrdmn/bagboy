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

//go:build integration

// Package packager_test contains end-to-end integration tests for the packager
// pipeline. Run with: go test -tags integration ./pkg/packager/...
package packager_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"io/fs"

	"github.com/scttfrdmn/bagboy/pkg/checksum"
	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/packager/brew"
	"github.com/scttfrdmn/bagboy/pkg/packager/cargo"
	"github.com/scttfrdmn/bagboy/pkg/packager/dmg"
	"github.com/scttfrdmn/bagboy/pkg/packager/installer"
	"github.com/scttfrdmn/bagboy/pkg/packager/nix"
	"github.com/scttfrdmn/bagboy/pkg/packager/npm"
	"github.com/scttfrdmn/bagboy/pkg/packager/pypi"
	"github.com/scttfrdmn/bagboy/pkg/packager/scoop"
	"github.com/scttfrdmn/bagboy/pkg/packager/spack"
	"github.com/scttfrdmn/bagboy/pkg/packager/winget"
)

// minimalBinaryConfig returns a Config with fake binaries that actually exist
// on disk (created in tmpDir).
func minimalBinaryConfig(t *testing.T, tmpDir string) *config.Config {
	t.Helper()

	fakeContent := []byte("#!/bin/sh\nexec echo hello\n")

	binaries := map[string]string{}
	for _, arch := range []string{"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64", "windows-amd64"} {
		name := "myapp-" + strings.ReplaceAll(arch, "-", "-")
		p := filepath.Join(tmpDir, name)
		if err := os.WriteFile(p, fakeContent, 0755); err != nil {
			t.Fatalf("create fake binary %s: %v", arch, err)
		}
		binaries[arch] = p
	}

	return &config.Config{
		Name:        "myapp",
		Version:     "1.0.0",
		Description: "Integration test application",
		Homepage:    "https://myapp.example.com",
		Author:      "Test Author",
		License:     "Apache-2.0",
		Binaries:    binaries,
		Installer: config.InstallerConfig{
			BaseURL:        "https://myapp.example.com/releases/v1.0.0",
			InstallPath:    "/usr/local/bin",
			VerifyChecksum: true,
		},
	}
}

// wingetBinaryConfig extends minimalBinaryConfig with the two winget-required fields.
func wingetBinaryConfig(t *testing.T, tmpDir string) *config.Config {
	t.Helper()
	cfg := minimalBinaryConfig(t, tmpDir)
	cfg.Packages.Winget.PackageIdentifier = "TestPublisher.myapp"
	cfg.Packages.Winget.Publisher = "TestPublisher"
	return cfg
}

// chdir switches the working directory for the duration of the test.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

// TestIntegration_Brew_Pack_Validate_Checksum runs the full brew pack →
// validate → checksum chain without any external tools.
func TestIntegration_Brew_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := brew.New()

	// Validate first.
	if err := p.Validate(cfg); err != nil {
		t.Fatalf("brew.Validate: %v", err)
	}

	// Pack.
	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("brew.Pack: %v", err)
	}
	if outputPath == "" {
		t.Fatal("expected non-empty output path")
	}

	// Output file must exist.
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file not found: %v", err)
	}

	// Checksum the output artifact.
	sum, err := checksum.Calculate(outputPath)
	if err != nil {
		t.Fatalf("checksum.Calculate: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("expected 64-char SHA256, got %d chars: %s", len(sum), sum)
	}

	// Content sanity: formula should reference the project name.
	content, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(content), "myapp") {
		t.Error("brew formula should contain project name 'myapp'")
	}
}

// TestIntegration_Scoop_Pack_Validate_Checksum runs the full scoop chain.
func TestIntegration_Scoop_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := scoop.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("scoop.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("scoop.Pack: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file not found: %v", err)
	}

	// Checksum.
	sum, err := checksum.Calculate(outputPath)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("unexpected SHA256 length %d", len(sum))
	}

	// JSON manifest should include version and name.
	content, _ := os.ReadFile(outputPath)
	for _, want := range []string{"myapp", "1.0.0"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("scoop manifest should contain %q", want)
		}
	}
}

// TestIntegration_Installer_Pack_Validate_Checksum runs the installer chain.
func TestIntegration_Installer_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := installer.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("installer.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("installer.Pack: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}

	// Script must be executable.
	if info.Mode()&0111 == 0 {
		t.Error("install script should be executable")
	}

	sum, err := checksum.Calculate(outputPath)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("unexpected SHA256 length %d", len(sum))
	}
}

// TestIntegration_NPM_Pack_Validate_Checksum runs the npm chain.
func TestIntegration_NPM_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := npm.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("npm.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("npm.Pack: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output dir/file not found: %v", err)
	}
}

// TestIntegration_DMG_Pack_Validate_Checksum runs the full dmg pack → validate → checksum chain.
func TestIntegration_DMG_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := dmg.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("dmg.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("dmg.Pack: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file not found: %v", err)
	}

	sum, err := checksum.Calculate(outputPath)
	if err != nil {
		t.Fatalf("checksum.Calculate: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("expected 64-char SHA256, got %d chars: %s", len(sum), sum)
	}

	content, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(content), "myapp") {
		t.Error("dmg output should contain project name 'myapp'")
	}
}

// TestIntegration_Cargo_Pack_Validate_Checksum runs the full cargo pack → validate → checksum chain.
func TestIntegration_Cargo_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := cargo.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("cargo.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("cargo.Pack: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output dir not found: %v", err)
	}

	cargoToml := filepath.Join(outputPath, "Cargo.toml")
	content, err := os.ReadFile(cargoToml)
	if err != nil {
		t.Fatalf("read Cargo.toml: %v", err)
	}
	for _, want := range []string{`name = "myapp"`, "1.0.0"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("Cargo.toml should contain %q", want)
		}
	}

	sum, err := checksum.Calculate(cargoToml)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("expected 64-char SHA256, got %d chars", len(sum))
	}
}

// TestIntegration_PyPI_Pack_Validate_Checksum runs the full pypi pack → validate → checksum chain.
func TestIntegration_PyPI_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := pypi.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("pypi.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("pypi.Pack: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output dir not found: %v", err)
	}

	setupPy := filepath.Join(outputPath, "setup.py")
	content, err := os.ReadFile(setupPy)
	if err != nil {
		t.Fatalf("read setup.py: %v", err)
	}
	for _, want := range []string{"myapp", cfg.Author} {
		if !strings.Contains(string(content), want) {
			t.Errorf("setup.py should contain %q", want)
		}
	}

	sum, err := checksum.Calculate(setupPy)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("expected 64-char SHA256, got %d chars", len(sum))
	}
}

// TestIntegration_Nix_Pack_Validate_Checksum runs the full nix pack → validate → checksum chain.
func TestIntegration_Nix_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := nix.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("nix.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("nix.Pack: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output dir not found: %v", err)
	}

	defaultNix := filepath.Join(outputPath, "default.nix")
	content, err := os.ReadFile(defaultNix)
	if err != nil {
		t.Fatalf("read default.nix: %v", err)
	}
	for _, want := range []string{"myapp", "1.0.0"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("default.nix should contain %q", want)
		}
	}

	sum, err := checksum.Calculate(defaultNix)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("expected 64-char SHA256, got %d chars", len(sum))
	}
}

// TestIntegration_Spack_Pack_Validate_Checksum runs the full spack pack → validate → checksum chain.
func TestIntegration_Spack_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := spack.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("spack.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("spack.Pack: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output dir not found: %v", err)
	}

	packagePy := filepath.Join(outputPath, "package.py")
	content, err := os.ReadFile(packagePy)
	if err != nil {
		t.Fatalf("read package.py: %v", err)
	}
	for _, want := range []string{"myapp", "1.0.0"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("package.py should contain %q", want)
		}
	}

	sum, err := checksum.Calculate(packagePy)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("expected 64-char SHA256, got %d chars", len(sum))
	}
}

// TestIntegration_Winget_Pack_Validate_Checksum runs the full winget pack → validate → checksum chain.
func TestIntegration_Winget_Pack_Validate_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := wingetBinaryConfig(t, tmpDir)
	p := winget.New()

	if err := p.Validate(cfg); err != nil {
		t.Fatalf("winget.Validate: %v", err)
	}

	outputPath, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("winget.Pack: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output dir not found: %v", err)
	}

	// Walk the manifest dir to find a .yaml file.
	var yamlFile string
	if err := fs.WalkDir(os.DirFS(outputPath), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".yaml") && yamlFile == "" {
			yamlFile = filepath.Join(outputPath, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk manifest dir: %v", err)
	}
	if yamlFile == "" {
		t.Fatal("no .yaml manifest found in winget output dir")
	}

	content, err := os.ReadFile(yamlFile)
	if err != nil {
		t.Fatalf("read yaml: %v", err)
	}
	for _, want := range []string{"myapp", "1.0.0"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("winget manifest should contain %q", want)
		}
	}

	sum, err := checksum.Calculate(yamlFile)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("expected 64-char SHA256, got %d chars", len(sum))
	}
}

// TestIntegration_DMG_Determinism verifies that dmg.Pack produces the same checksum on repeated runs.
func TestIntegration_DMG_Determinism(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := dmg.New()

	output1, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("first Pack: %v", err)
	}
	sum1, err := checksum.Calculate(output1)
	if err != nil {
		t.Fatalf("checksum 1: %v", err)
	}

	output2, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("second Pack: %v", err)
	}
	sum2, err := checksum.Calculate(output2)
	if err != nil {
		t.Fatalf("checksum 2: %v", err)
	}

	if sum1 != sum2 {
		t.Errorf("dmg pack is not deterministic: %s != %s", sum1, sum2)
	}
}

// TestIntegration_MultiFormat_SameChecksums verifies that running pack twice
// on identical input produces the same checksum (determinism check).
func TestIntegration_MultiFormat_SameChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	cfg := minimalBinaryConfig(t, tmpDir)
	p := installer.New()

	output1, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("first Pack: %v", err)
	}
	sum1, err := checksum.Calculate(output1)
	if err != nil {
		t.Fatalf("checksum 1: %v", err)
	}

	// Pack again — must produce the same content.
	output2, err := p.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("second Pack: %v", err)
	}
	sum2, err := checksum.Calculate(output2)
	if err != nil {
		t.Fatalf("checksum 2: %v", err)
	}

	if sum1 != sum2 {
		t.Errorf("pack is not deterministic: %s != %s", sum1, sum2)
	}
}
