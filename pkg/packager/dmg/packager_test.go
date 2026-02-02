package dmg

import (
	"context"
	"os"
	"path/filepath"
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
