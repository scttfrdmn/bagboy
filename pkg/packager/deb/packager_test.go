package deb

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestDEBPackager(t *testing.T) {
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
		Author:      "Test Author <test@example.com>",
		Packages: config.PackagesConfig{
			Deb: config.DebConfig{
				Maintainer: "test@example.com",
				Section:    "utils",
				Priority:   "optional",
			},
		},
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

	// Check if DEB file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("DEB file not created: %s", outputPath)
	}
}

func TestDEBPackager_Name(t *testing.T) {
	packager := New()
	if packager.Name() != "deb" {
		t.Errorf("Expected name 'deb', got %s", packager.Name())
	}
}

func TestDEBPackager_Validate(t *testing.T) {
	packager := New()
	
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Packages: config.PackagesConfig{
					Deb: config.DebConfig{
						Maintainer: "test@example.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing maintainer",
			config: &config.Config{
				Packages: config.PackagesConfig{
					Deb: config.DebConfig{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := packager.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateControlFile(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	controlPath := filepath.Join(tmpDir, "control")
	
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Packages: config.PackagesConfig{
			Deb: config.DebConfig{
				Maintainer: "test@example.com",
				Section:    "utils",
				Priority:   "optional",
			},
		},
	}

	if err := packager.createControlFile(controlPath, cfg); err != nil {
		t.Errorf("createControlFile() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(controlPath); os.IsNotExist(err) {
		t.Error("Control file was not created")
	}

	// Check content
	content, err := os.ReadFile(controlPath)
	if err != nil {
		t.Errorf("Failed to read control file: %v", err)
	}

	contentStr := string(content)
	requiredFields := []string{
		"Package: testapp",
		"Version: 1.0.0",
		"Description: Test application",
		"Maintainer: test@example.com",
		"Section: utils",
		"Priority: optional",
	}

	for _, field := range requiredFields {
		if !contains(contentStr, field) {
			t.Errorf("Control file missing required field: %s", field)
		}
	}
}

func TestCreateTarGz(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	
	// Create test directory structure
	testDir := filepath.Join(tmpDir, "test")
	os.MkdirAll(testDir, 0755)
	
	// Create test files
	testFile1 := filepath.Join(testDir, "file1.txt")
	testFile2 := filepath.Join(testDir, "file2.txt")
	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)
	
	outputPath := filepath.Join(tmpDir, "test.tar.gz")
	
	if err := packager.createTarGz(testDir, outputPath, []string{}); err != nil {
		t.Errorf("createTarGz() error = %v", err)
	}

	// Check output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Tar.gz file was not created")
	}
	
	// Check file size is reasonable (not empty)
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Errorf("Failed to stat tar.gz file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Tar.gz file is empty")
	}
}

func TestCreateTarGz_Error(t *testing.T) {
	packager := New()
	
	// Test with non-existent source directory
	err := packager.createTarGz("/non/existent/dir", "/tmp/test.tar.gz", []string{})
	if err == nil {
		t.Error("createTarGz() should fail with non-existent source directory")
	}
}

func TestAddFileToAr(t *testing.T) {
	// This test is more complex since addFileToAr requires an ar.Writer
	// We'll test it indirectly through the Pack method which uses it
	packager := New()
	
	tmpDir := t.TempDir()
	testBinary := filepath.Join(tmpDir, "test-linux-amd64")
	if err := os.WriteFile(testBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Packages: config.PackagesConfig{
			Deb: config.DebConfig{
				Maintainer: "test@example.com",
			},
		},
		Binaries: map[string]string{
			"linux-amd64": testBinary,
		},
	}

	// Change to test directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// This will test addFileToAr indirectly
	_, err := packager.Pack(context.Background(), cfg)
	if err != nil {
		t.Errorf("Pack() error = %v", err)
	}
}

func TestAddFileToAr_Error(t *testing.T) {
	// Test addFileToAr error handling indirectly through Pack with missing binary
	packager := New()
	
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Binaries: map[string]string{
			"linux-amd64": "/non/existent/binary",
		},
		Packages: config.PackagesConfig{
			Deb: config.DebConfig{
				Maintainer: "test@example.com",
			},
		},
	}

	ctx := context.Background()
	_, err := packager.Pack(ctx, cfg)
	
	if err == nil {
		t.Error("Pack() should fail with missing binary file")
	}
}

func TestDEBPack_MissingBinary(t *testing.T) {
	// This test is already covered by TestAddFileToAr_Error above
	// Removing duplicate
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
