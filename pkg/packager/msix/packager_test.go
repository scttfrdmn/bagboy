package msix

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestMSIXPackager(t *testing.T) {
	// Create test binary
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "test-windows-amd64.exe")
	if err := os.WriteFile(testBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Author:      "Test Author <test@example.com>",
		Binaries: map[string]string{
			"windows-amd64": testBinary,
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

	// Check if MSIX file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("MSIX file not created: %s", outputPath)
	}

	// Check if manifest was created
	manifestPath := filepath.Join("dist", "msix", "AppxManifest.xml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("AppxManifest.xml not created")
	}

	// Check if build script was created
	buildScript := filepath.Join("dist", "msix", "build-msix.ps1")
	if _, err := os.Stat(buildScript); os.IsNotExist(err) {
		t.Error("Build script not created")
	}

	// Read and verify manifest content
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "testapp") {
		t.Error("Manifest missing app name")
	}
	if !contains(contentStr, "1.0.0") {
		t.Error("Manifest missing version")
	}
}

func TestMSIXPackagerValidation(t *testing.T) {
	packager := New()

	// Test with no Windows binary
	cfg := &config.Config{
		Name:     "testapp",
		Version:  "1.0.0",
		Binaries: map[string]string{
			"linux-amd64": "test-linux",
		},
	}

	if err := packager.Validate(cfg); err == nil {
		t.Error("Expected validation to fail with no Windows binary")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
