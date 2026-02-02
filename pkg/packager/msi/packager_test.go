package msi

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestMSIPackager(t *testing.T) {
	packager := New()
	
	if packager.Name() != "msi" {
		t.Errorf("Expected name 'msi', got %s", packager.Name())
	}
}

func TestMSIValidate(t *testing.T) {
	packager := New()
	
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config with Windows binary",
			config: &config.Config{
				Binaries: map[string]string{
					"windows-amd64": "dist/app.exe",
				},
			},
			wantErr: false,
		},
		{
			name: "missing Windows binary",
			config: &config.Config{
				Binaries: map[string]string{
					"linux-amd64": "dist/app",
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

func TestMSIPack(t *testing.T) {
	// Create temporary binary file
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "testapp.exe")
	if err := os.WriteFile(binaryPath, []byte("fake exe content"), 0755); err != nil {
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
			"windows-amd64": binaryPath,
		},
	}

	packager := New()
	
	// Change to temp directory for test
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	ctx := context.Background()
	outputPath, err := packager.Pack(ctx, cfg)
	
	// Should not error even if WiX/go-msi is not available
	// The function should return an error about tools not being found
	if err != nil && !contains(err.Error(), "MSI build tools not found") {
		t.Errorf("Pack() unexpected error = %v", err)
	}

	// If tools are available, check output
	if err == nil {
		if outputPath == "" {
			t.Error("Pack() returned empty output path")
		}
		
		// Check if WiX file was created
		wxsPath := filepath.Join("dist", "msi-build", "testapp.wxs")
		if _, err := os.Stat(wxsPath); os.IsNotExist(err) {
			t.Error("WiX source file was not created")
		}
	}
}

func TestCreateWixSource(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	wxsPath := filepath.Join(tmpDir, "test.wxs")
	binaryPath := filepath.Join(tmpDir, "test.exe")
	
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		License:     "MIT",
		Homepage:    "https://example.com",
		Author:      "Test Author <test@example.com>",
	}

	if err := packager.createWixSource(wxsPath, cfg, binaryPath); err != nil {
		t.Errorf("createWixSource() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(wxsPath); os.IsNotExist(err) {
		t.Error("WiX source file was not created")
	}

	// Check content
	content, err := os.ReadFile(wxsPath)
	if err != nil {
		t.Errorf("Failed to read WiX file: %v", err)
	}

	contentStr := string(content)
	requiredElements := []string{
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
		"<Wix xmlns=\"http://schemas.microsoft.com/wix/2006/wi\">",
		"Name=\"testapp\"",
		"Version=\"1.0.0.0\"",
		"Manufacturer=\"Test Author\"",
		"Description=\"Test application\"",
	}

	for _, element := range requiredElements {
		if !contains(contentStr, element) {
			t.Errorf("WiX file missing required element: %s", element)
		}
	}
}

func TestGenerateUpgradeCode(t *testing.T) {
	packager := New()
	
	cfg1 := &config.Config{Name: "testapp"}
	cfg2 := &config.Config{Name: "testapp"}
	cfg3 := &config.Config{Name: "different"}

	code1 := packager.generateUpgradeCode(cfg1)
	code2 := packager.generateUpgradeCode(cfg2)
	code3 := packager.generateUpgradeCode(cfg3)

	// Same name should generate same code
	if code1 != code2 {
		t.Errorf("Same app name should generate same upgrade code. Got %s and %s", code1, code2)
	}

	// Different names should generate different codes
	if code1 == code3 {
		t.Errorf("Different app names should generate different upgrade codes. Both got %s", code1)
	}

	// Should be in GUID format
	if !contains(code1, "{") || !contains(code1, "}") {
		t.Errorf("Upgrade code should be in GUID format. Got %s", code1)
	}
}

func TestGetAuthorName(t *testing.T) {
	packager := New()
	
	tests := []struct {
		name     string
		author   string
		expected string
	}{
		{
			name:     "author with email",
			author:   "John Doe <john@example.com>",
			expected: "John Doe",
		},
		{
			name:     "author without email",
			author:   "Jane Smith",
			expected: "Jane Smith",
		},
		{
			name:     "author with extra spaces",
			author:   "  Bob Wilson  <bob@test.com>",
			expected: "Bob Wilson",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Author: tt.author}
			result := packager.getAuthorName(cfg)
			if result != tt.expected {
				t.Errorf("getAuthorName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.exe")
	dstPath := filepath.Join(tmpDir, "dest.exe")
	
	testContent := "fake exe content"
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
