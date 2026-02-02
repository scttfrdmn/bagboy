package winget

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestWingetPackager(t *testing.T) {
	packager := New()
	
	if packager.Name() != "winget" {
		t.Errorf("Expected name 'winget', got %s", packager.Name())
	}
}

func TestWingetValidate(t *testing.T) {
	packager := New()
	
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Binaries: map[string]string{
					"windows-amd64": "dist/app.exe",
				},
				Packages: config.PackagesConfig{
					Winget: config.WingetPkgConfig{
						PackageIdentifier: "Publisher.AppName",
						Publisher:         "Publisher",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing package identifier",
			config: &config.Config{
				Binaries: map[string]string{
					"windows-amd64": "dist/app.exe",
				},
				Packages: config.PackagesConfig{
					Winget: config.WingetPkgConfig{
						Publisher: "Publisher",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing publisher",
			config: &config.Config{
				Binaries: map[string]string{
					"windows-amd64": "dist/app.exe",
				},
				Packages: config.PackagesConfig{
					Winget: config.WingetPkgConfig{
						PackageIdentifier: "Publisher.AppName",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing Windows binary",
			config: &config.Config{
				Binaries: map[string]string{
					"linux-amd64": "dist/app",
				},
				Packages: config.PackagesConfig{
					Winget: config.WingetPkgConfig{
						PackageIdentifier: "Publisher.AppName",
						Publisher:         "Publisher",
					},
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

func TestWingetPack(t *testing.T) {
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
			Winget: config.WingetPkgConfig{
				PackageIdentifier: "TestPublisher.TestApp",
				Publisher:         "Test Publisher",
				MinimumOSVersion:  "10.0.0.0",
			},
		},
		Installer: config.InstallerConfig{
			BaseURL: "https://github.com/test/testapp/releases/download/v1.0.0",
		},
	}

	packager := New()
	
	// Change to temp directory for test
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	ctx := context.Background()
	outputPath, err := packager.Pack(ctx, cfg)
	
	if err != nil {
		t.Errorf("Pack() error = %v", err)
	}

	if outputPath == "" {
		t.Error("Pack() returned empty output path")
	}

	// Check if manifest files were created
	expectedFiles := []string{
		filepath.Join(outputPath, "TestPublisher.TestApp.yaml"),
		filepath.Join(outputPath, "TestPublisher.TestApp.installer.yaml"),
		filepath.Join(outputPath, "TestPublisher.TestApp.locale.en-US.yaml"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected manifest file not created: %s", file)
		}
	}
}

func TestCreateVersionManifest(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.yaml")
	
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Packages: config.PackagesConfig{
			Winget: config.WingetPkgConfig{
				PackageIdentifier: "TestPublisher.TestApp",
			},
		},
	}

	if err := packager.createVersionManifest(manifestPath, cfg); err != nil {
		t.Errorf("createVersionManifest() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Version manifest file was not created")
	}

	// Check content
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Errorf("Failed to read version manifest: %v", err)
	}

	contentStr := string(content)
	requiredFields := []string{
		"PackageIdentifier: TestPublisher.TestApp",
		"PackageVersion: 1.0.0",
		"DefaultLocale: en-US",
		"ManifestType: version",
		"ManifestVersion: 1.4.0",
	}

	for _, field := range requiredFields {
		if !contains(contentStr, field) {
			t.Errorf("Version manifest missing required field: %s", field)
		}
	}
}

func TestCreateInstallerManifest(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.installer.yaml")
	
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Packages: config.PackagesConfig{
			Winget: config.WingetPkgConfig{
				PackageIdentifier: "TestPublisher.TestApp",
				MinimumOSVersion:  "10.0.0.0",
			},
		},
		Installer: config.InstallerConfig{
			BaseURL: "https://github.com/test/testapp/releases/download/v1.0.0",
		},
	}

	if err := packager.createInstallerManifest(manifestPath, cfg); err != nil {
		t.Errorf("createInstallerManifest() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Installer manifest file was not created")
	}

	// Check content
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Errorf("Failed to read installer manifest: %v", err)
	}

	contentStr := string(content)
	requiredFields := []string{
		"PackageIdentifier: TestPublisher.TestApp",
		"PackageVersion: 1.0.0",
		"MinimumOSVersion: 10.0.0.0",
		"Architecture: x64",
		"InstallerType: exe",
		"ManifestType: installer",
	}

	for _, field := range requiredFields {
		if !contains(contentStr, field) {
			t.Errorf("Installer manifest missing required field: %s", field)
		}
	}
}

func TestCreateLocaleManifest(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "test.locale.en-US.yaml")
	
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		License:     "MIT",
		Homepage:    "https://example.com",
		Author:      "Test Author",
		Packages: config.PackagesConfig{
			Winget: config.WingetPkgConfig{
				PackageIdentifier: "TestPublisher.TestApp",
				Publisher:         "Test Publisher",
			},
		},
	}

	if err := packager.createLocaleManifest(manifestPath, cfg); err != nil {
		t.Errorf("createLocaleManifest() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Locale manifest file was not created")
	}

	// Check content
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Errorf("Failed to read locale manifest: %v", err)
	}

	contentStr := string(content)
	requiredFields := []string{
		"PackageIdentifier: TestPublisher.TestApp",
		"PackageVersion: 1.0.0",
		"PackageLocale: en-US",
		"Publisher: Test Publisher",
		"PackageName: testapp",
		"License: MIT",
		"ShortDescription: Test application",
		"PackageUrl: https://example.com",
		"ManifestType: defaultLocale",
	}

	for _, field := range requiredFields {
		if !contains(contentStr, field) {
			t.Errorf("Locale manifest missing required field: %s", field)
		}
	}
}

func TestWriteTemplate(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.yaml")
	
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Packages: config.PackagesConfig{
			Winget: config.WingetPkgConfig{
				PackageIdentifier: "TestPublisher.TestApp",
				Publisher:         "Test Publisher",
			},
		},
	}

	template := `PackageIdentifier: {{.PackageIdentifier}}
PackageVersion: {{.Version}}
Publisher: {{.Publisher}}`

	if err := packager.writeTemplate(templatePath, template, cfg); err != nil {
		t.Errorf("writeTemplate() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Error("Template file was not created")
	}

	// Check content
	content, err := os.ReadFile(templatePath)
	if err != nil {
		t.Errorf("Failed to read template file: %v", err)
	}

	contentStr := string(content)
	expectedContent := []string{
		"PackageIdentifier: TestPublisher.TestApp",
		"PackageVersion: 1.0.0",
		"Publisher: Test Publisher",
	}

	for _, expected := range expectedContent {
		if !contains(contentStr, expected) {
			t.Errorf("Template file missing expected content: %s", expected)
		}
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
