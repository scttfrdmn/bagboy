package chocolatey

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestChocolateyPackager(t *testing.T) {
	packager := New()
	
	if packager.Name() != "chocolatey" {
		t.Errorf("Expected name 'chocolatey', got %s", packager.Name())
	}
}

func TestChocolateyValidate(t *testing.T) {
	packager := New()
	
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Author: "Test Author",
				Binaries: map[string]string{
					"windows-amd64": "dist/app.exe",
				},
			},
			wantErr: false,
		},
		{
			name: "missing author",
			config: &config.Config{
				Binaries: map[string]string{
					"windows-amd64": "dist/app.exe",
				},
			},
			wantErr: true,
		},
		{
			name: "missing Windows binary",
			config: &config.Config{
				Author: "Test Author",
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

func TestChocolateyPack(t *testing.T) {
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
		Packages: config.PackagesConfig{
			Chocolatey: config.ChocolateyConfig{
				PackageSourceURL: "https://github.com/test/testapp",
				DocsURL:          "https://example.com/docs",
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
	
	// Should not error even if choco/nuget/zip is not available
	// The function should return an error about tools not being found
	if err != nil && !contains(err.Error(), "Chocolatey build tools not found") && !contains(err.Error(), "zip failed") {
		t.Errorf("Pack() unexpected error = %v", err)
	}

	// If tools are available, check output
	if err == nil {
		if outputPath == "" {
			t.Error("Pack() returned empty output path")
		}
		
		// Check if nuspec file was created
		nuspecPath := filepath.Join("dist", "chocolatey-build", "testapp.nuspec")
		if _, err := os.Stat(nuspecPath); os.IsNotExist(err) {
			t.Error("Nuspec file was not created")
		}
	}
}

func TestCreateNuspec(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	nuspecPath := filepath.Join(tmpDir, "test.nuspec")
	
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		License:     "MIT",
		Homepage:    "https://example.com",
		Author:      "Test Author <test@example.com>",
		Packages: config.PackagesConfig{
			Chocolatey: config.ChocolateyConfig{
				PackageSourceURL: "https://github.com/test/testapp",
				DocsURL:          "https://example.com/docs",
			},
		},
	}

	if err := packager.createNuspec(nuspecPath, cfg); err != nil {
		t.Errorf("createNuspec() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(nuspecPath); os.IsNotExist(err) {
		t.Error("Nuspec file was not created")
	}

	// Check content
	content, err := os.ReadFile(nuspecPath)
	if err != nil {
		t.Errorf("Failed to read nuspec file: %v", err)
	}

	contentStr := string(content)
	requiredElements := []string{
		"<?xml version=\"1.0\" encoding=\"utf-8\"?>",
		"<package xmlns=\"http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd\">",
		"<id>testapp</id>",
		"<version>1.0.0</version>",
		"<authors>Test Author</authors>",
		"<description>Test application</description>",
		"<packageSourceUrl>https://github.com/test/testapp</packageSourceUrl>",
		"<docsUrl>https://example.com/docs</docsUrl>",
	}

	for _, element := range requiredElements {
		if !contains(contentStr, element) {
			t.Errorf("Nuspec file missing required element: %s", element)
		}
	}
}

func TestCreateInstallScript(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "chocolateyInstall.ps1")
	
	cfg := &config.Config{
		Name: "testapp",
	}

	if err := packager.createInstallScript(scriptPath, cfg); err != nil {
		t.Errorf("createInstallScript() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("Install script was not created")
	}

	// Check content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Errorf("Failed to read install script: %v", err)
	}

	contentStr := string(content)
	requiredElements := []string{
		"$ErrorActionPreference = 'Stop'",
		"$packageName = 'testapp'",
		"Install-BinFile",
		"testapp has been installed successfully!",
	}

	for _, element := range requiredElements {
		if !contains(contentStr, element) {
			t.Errorf("Install script missing required element: %s", element)
		}
	}
}

func TestCreateUninstallScript(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "chocolateyUninstall.ps1")
	
	cfg := &config.Config{
		Name: "testapp",
	}

	if err := packager.createUninstallScript(scriptPath, cfg); err != nil {
		t.Errorf("createUninstallScript() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("Uninstall script was not created")
	}

	// Check content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Errorf("Failed to read uninstall script: %v", err)
	}

	contentStr := string(content)
	requiredElements := []string{
		"$ErrorActionPreference = 'Stop'",
		"$packageName = 'testapp'",
		"Uninstall-BinFile",
		"testapp has been uninstalled successfully!",
	}

	for _, element := range requiredElements {
		if !contains(contentStr, element) {
			t.Errorf("Uninstall script missing required element: %s", element)
		}
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

func TestCopyFile_Error(t *testing.T) {
	packager := New()
	
	// Test with non-existent source file
	err := packager.copyFile("/non/existent/file", "/tmp/dest")
	if err == nil {
		t.Error("copyFile() should fail with non-existent source file")
	}
}

func TestBuildWithChoco(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	os.MkdirAll(buildDir, 0755)
	outputPath := filepath.Join(tmpDir, "test.nupkg")
	
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	// Create a dummy nuspec file
	nuspecPath := filepath.Join(buildDir, "testapp.nuspec")
	os.WriteFile(nuspecPath, []byte("<?xml version=\"1.0\"?><package></package>"), 0644)

	ctx := context.Background()
	_, err := packager.buildWithChoco(ctx, buildDir, outputPath, cfg)
	
	// This will fail because choco is not available, but we test the code path
	if err == nil {
		t.Error("buildWithChoco() should fail when choco is not available")
	}
}

func TestBuildWithNuget(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	os.MkdirAll(buildDir, 0755)
	outputPath := filepath.Join(tmpDir, "test.nupkg")
	
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	// Create a dummy nuspec file
	nuspecPath := filepath.Join(buildDir, "testapp.nuspec")
	os.WriteFile(nuspecPath, []byte("<?xml version=\"1.0\"?><package></package>"), 0644)

	ctx := context.Background()
	_, err := packager.buildWithNuget(ctx, buildDir, outputPath, cfg)
	
	// This will fail because nuget is not available, but we test the code path
	if err == nil {
		t.Error("buildWithNuget() should fail when nuget is not available")
	}
}

func TestBuildPackage_NoTools(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	os.MkdirAll(buildDir, 0755)
	
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	ctx := context.Background()
	_, err := packager.buildPackage(ctx, buildDir, cfg)
	
	// Should return error about missing tools or zip failure
	if err == nil {
		t.Error("buildPackage() should fail when no Chocolatey build tools are available")
	}
	
	// Accept either missing tools error or zip failure
	if !contains(err.Error(), "Chocolatey build tools not found") && !contains(err.Error(), "zip failed") {
		t.Errorf("Expected 'Chocolatey build tools not found' or 'zip failed' error, got: %v", err)
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

func TestChocolateyPack_MissingBinary(t *testing.T) {
	packager := New()
	
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Author:      "Test Author",
		Binaries: map[string]string{
			"windows-amd64": "/non/existent/binary.exe",
		},
	}

	ctx := context.Background()
	_, err := packager.Pack(ctx, cfg)
	
	if err == nil {
		t.Error("Pack() should fail with missing binary file")
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
