package apptainer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestApptainerPackager(t *testing.T) {
	// Create test binary
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "test-linux-amd64")
	if err := os.WriteFile(testBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application for HPC",
		Author:      "Test Author <test@example.com>",
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

	// Check if definition file was created
	defPath := filepath.Join(outputPath, "testapp.def")
	if _, err := os.Stat(defPath); os.IsNotExist(err) {
		t.Error("Definition file not created")
	}

	// Check if build script was created
	buildScript := filepath.Join(outputPath, "build.sh")
	if _, err := os.Stat(buildScript); os.IsNotExist(err) {
		t.Error("Build script not created")
	}

	// Read and verify definition content
	content, err := os.ReadFile(defPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "testapp") {
		t.Error("Definition missing app name")
	}
	if !contains(contentStr, "1.0.0") {
		t.Error("Definition missing version")
	}
	if !contains(contentStr, "Bootstrap: library") {
		t.Error("Definition missing bootstrap")
	}
}

func TestApptainerPackagerValidation(t *testing.T) {
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

	// Test with no Linux binary
	cfg.Description = "Test app"
	cfg.Binaries = map[string]string{
		"darwin-amd64": "test-darwin",
	}

	_, err := packager.Pack(context.Background(), cfg)
	if err == nil {
		t.Error("Expected pack to fail with no Linux binary")
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
