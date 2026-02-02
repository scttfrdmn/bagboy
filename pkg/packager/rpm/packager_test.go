package rpm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestRPMPackager(t *testing.T) {
	packager := New()
	
	if packager.Name() != "rpm" {
		t.Errorf("Expected name 'rpm', got %s", packager.Name())
	}
}

func TestRPMValidate(t *testing.T) {
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
					RPM: config.RPMConfig{
						Vendor: "Test Vendor",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing vendor",
			config: &config.Config{
				Packages: config.PackagesConfig{
					RPM: config.RPMConfig{},
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

func TestRPMPack(t *testing.T) {
	// Create temporary binary file
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "testapp")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho 'test app'\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create test config
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		License:     "MIT",
		Homepage:    "https://example.com",
		Author:      "Test Author <test@example.com>",
		Binaries: map[string]string{
			"linux-amd64": binaryPath,
		},
		Packages: config.PackagesConfig{
			RPM: config.RPMConfig{
				Vendor: "Test Vendor",
				Group:  "Applications/System",
			},
		},
	}

	packager := New()
	
	// Change to temp directory for test
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	ctx := context.Background()
	outputPath, err := packager.Pack(ctx, cfg)
	
	// Should not error even if rpmbuild is not available
	// The function should return an error about rpmbuild not being found
	if err != nil && !contains(err.Error(), "rpmbuild not found") {
		t.Errorf("Pack() unexpected error = %v", err)
	}

	// If rpmbuild is available, check output
	if err == nil {
		if outputPath == "" {
			t.Error("Pack() returned empty output path")
		}
		
		// Check if spec file was created
		specPath := filepath.Join("dist", "rpm-build", "SPECS", "testapp.spec")
		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			t.Error("Spec file was not created")
		}
	}
}

func TestGenerateSpec(t *testing.T) {
	packager := New()
	
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		License:     "MIT",
		Homepage:    "https://example.com",
		Packages: config.PackagesConfig{
			RPM: config.RPMConfig{
				Vendor: "Test Vendor",
				Group:  "Applications/System",
			},
		},
	}

	spec := packager.generateSpec(cfg, "/path/to/binary")
	
	// Check that spec contains required fields
	requiredFields := []string{
		"Name:           testapp",
		"Version:        1.0.0",
		"Summary:        Test application",
		"License:        MIT",
		"URL:            https://example.com",
		"Group:          Applications/System",
		"Vendor:         Test Vendor",
	}

	for _, field := range requiredFields {
		if !contains(spec, field) {
			t.Errorf("Spec file missing required field: %s", field)
		}
	}
}

func TestCopyFile(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source")
	dstPath := filepath.Join(tmpDir, "dest")
	
	testContent := "test content"
	if err := os.WriteFile(srcPath, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := packager.copyFile(srcPath, dstPath); err != nil {
		t.Errorf("copyFile() error = %v", err)
	}

	// Check destination file
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("File content mismatch. Expected %s, got %s", testContent, string(content))
	}

	// Check permissions
	info, err := os.Stat(dstPath)
	if err != nil {
		t.Errorf("Failed to stat destination file: %v", err)
	}

	if info.Mode().Perm() != 0755 {
		t.Errorf("Wrong file permissions. Expected 0755, got %v", info.Mode().Perm())
	}
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
