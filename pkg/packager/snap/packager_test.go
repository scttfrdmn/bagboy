package snap

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestSnapPackager(t *testing.T) {
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

	// Check if snapcraft.yaml was created
	snapcraftPath := filepath.Join(outputPath, "snapcraft.yaml")
	if _, err := os.Stat(snapcraftPath); os.IsNotExist(err) {
		t.Error("snapcraft.yaml not created")
	}

	// Read and verify snapcraft.yaml content
	content, err := os.ReadFile(snapcraftPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "name: testapp") {
		t.Error("snapcraft.yaml missing app name")
	}
	if !contains(contentStr, "version: '1.0.0'") {
		t.Error("snapcraft.yaml missing version")
	}
}

func TestSnapPackagerValidation(t *testing.T) {
	packager := New()

	// Test with no description
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Binaries: map[string]string{
			"linux-amd64": "test-linux",
		},
	}

	if err := packager.Validate(cfg); err == nil {
		t.Error("Expected validation to fail with no description")
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
