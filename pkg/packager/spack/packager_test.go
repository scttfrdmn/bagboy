package spack

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestSpackPackager(t *testing.T) {
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
		Homepage:    "https://github.com/test/testapp",
		License:     "MIT",
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

	// Check if package.py was created
	packagePath := filepath.Join(outputPath, "package.py")
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		t.Error("package.py not created")
	}

	// Check if instructions were created
	instructionsPath := filepath.Join(outputPath, "INSTALL.md")
	if _, err := os.Stat(instructionsPath); os.IsNotExist(err) {
		t.Error("INSTALL.md not created")
	}

	// Read and verify package.py content
	content, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !contains(contentStr, "class Testapp(Package):") {
		t.Error("Package missing class definition")
	}
	if !contains(contentStr, "Test application for HPC") {
		t.Error("Package missing description")
	}
	if !contains(contentStr, "MIT") {
		t.Error("Package missing license")
	}
}

func TestSpackPackagerValidation(t *testing.T) {
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
