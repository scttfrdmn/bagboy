package flatpak

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestFlatpakPackager(t *testing.T) {
	// Create test binary
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "test-linux-amd64")
	if err := os.WriteFile(testBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Homepage:    "https://example.com",
		Binaries: map[string]string{
			"linux-amd64": testBinary,
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

	// Check if manifest file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Flatpak manifest not created: %s", outputPath)
	}

	// Read and verify manifest content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Errorf("Invalid JSON manifest: %v", err)
	}

	if manifest["app-id"] != "dev.bagboy.Testapp" {
		t.Errorf("Expected app-id 'dev.bagboy.Testapp', got %v", manifest["app-id"])
	}

	if manifest["command"] != "testapp" {
		t.Errorf("Expected command 'testapp', got %v", manifest["command"])
	}
}

func TestFlatpakPackagerValidation(t *testing.T) {
	packager := New()

	// Test with no homepage
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Binaries: map[string]string{
			"linux-amd64": "test-linux",
		},
	}

	if err := packager.Validate(cfg); err == nil {
		t.Error("Expected validation to fail with no homepage")
	}

	// Test with no Linux binary
	cfg.Homepage = "https://example.com"
	cfg.Binaries = map[string]string{
		"darwin-amd64": "test-darwin",
	}

	_, err := packager.Pack(context.Background(), cfg)
	if err == nil {
		t.Error("Expected pack to fail with no Linux binary")
	}
}
