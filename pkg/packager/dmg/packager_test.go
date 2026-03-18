package dmg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestDMGPackager(t *testing.T) {
	// Create test binary
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "test-darwin-amd64")
	if err := os.WriteFile(testBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Binaries: map[string]string{
			"darwin-amd64": testBinary,
		},
	}

	// Change to test directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	packager := New()

	// Test validation
	if err := packager.Validate(cfg); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test packing
	outputPath, err := packager.Pack(context.Background(), cfg)
	if err != nil {
		t.Errorf("Pack failed: %v", err)
	}

	if outputPath == "" {
		t.Error("Expected output path")
	}

	// Check if DMG file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("DMG file not created: %s", outputPath)
	}

	// Check if build script was created
	buildScript := filepath.Join("dist", "dmg", "build-dmg.sh")
	if _, err := os.Stat(buildScript); os.IsNotExist(err) {
		t.Error("Build script not created")
	}
}

func TestDMGPackagerValidation(t *testing.T) {
	packager := New()

	// Test with no macOS binary
	cfg := &config.Config{
		Name:     "testapp",
		Version:  "1.0.0",
		Binaries: map[string]string{
			"linux-amd64": "test-linux",
		},
	}

	if err := packager.Validate(cfg); err == nil {
		t.Error("Expected validation to fail with no macOS binary")
	}
}

func TestDMGPackager_Name(t *testing.T) {
	p := New()
	if p.Name() != "dmg" {
		t.Errorf("expected 'dmg', got %q", p.Name())
	}
}

func TestDMGPack_BuildScriptContent(t *testing.T) {
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "myapp-darwin-amd64")
	if err := os.WriteFile(testBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:        "myapp",
		Version:     "3.0.0",
		Description: "My Application",
		Binaries: map[string]string{
			"darwin-amd64": testBinary,
		},
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	packager := New()
	outputPath, err := packager.Pack(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}

	// Verify mock DMG file mentions app name and version.
	dmgContent, _ := os.ReadFile(outputPath)
	for _, want := range []string{"myapp", "3.0.0"} {
		if !strings.Contains(string(dmgContent), want) {
			t.Errorf("DMG file should contain %q", want)
		}
	}

	// Verify build script was created and contains hdiutil commands.
	buildScript := filepath.Join("dist", "dmg", "build-dmg.sh")
	scriptContent, err := os.ReadFile(buildScript)
	if err != nil {
		t.Fatalf("build script not created: %v", err)
	}
	script := string(scriptContent)
	for _, want := range []string{"hdiutil", "myapp", "3.0.0", "#!/bin/bash"} {
		if !strings.Contains(script, want) {
			t.Errorf("build script should contain %q", want)
		}
	}

	// Verify build script is executable.
	info, _ := os.Stat(buildScript)
	if info.Mode()&0111 == 0 {
		t.Error("build script should be executable")
	}

	// Verify DS_Store template was created.
	dsStore := filepath.Join("dist", "dmg", "DS_Store_template")
	if _, err := os.Stat(dsStore); err != nil {
		t.Error("DS_Store template should be created")
	}
}

func TestDMGPack_CopiesBinaryToContents(t *testing.T) {
	testDir := t.TempDir()
	binaryContent := []byte("fake macos binary content")
	testBinary := filepath.Join(testDir, "myapp-darwin-arm64")
	if err := os.WriteFile(testBinary, binaryContent, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:    "myapp",
		Version: "1.0.0",
		Binaries: map[string]string{
			"darwin-arm64": testBinary,
		},
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(testDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	p := New()
	if _, err := p.Pack(context.Background(), cfg); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	// Binary should have been copied into dist/dmg/contents/.
	copiedBinary := filepath.Join("dist", "dmg", "contents", "myapp")
	copied, err := os.ReadFile(copiedBinary)
	if err != nil {
		t.Fatalf("binary not copied to contents: %v", err)
	}
	if string(copied) != string(binaryContent) {
		t.Error("copied binary content mismatch")
	}
}
