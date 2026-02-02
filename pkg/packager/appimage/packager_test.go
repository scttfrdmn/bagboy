package appimage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestAppImagePackager(t *testing.T) {
	packager := New()
	
	if packager.Name() != "appimage" {
		t.Errorf("Expected name 'appimage', got %s", packager.Name())
	}
}

func TestAppImageValidate(t *testing.T) {
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
					AppImage: config.AppImageConfig{
						Categories: []string{"Utility", "Development"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing categories",
			config: &config.Config{
				Packages: config.PackagesConfig{
					AppImage: config.AppImageConfig{},
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

func TestAppImagePack(t *testing.T) {
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
			AppImage: config.AppImageConfig{
				Categories: []string{"Utility", "Development"},
				DesktopEntry: config.AppImageDesktopConfig{
					Type:     "Application",
					Terminal: false,
				},
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
	
	// Should not error even if appimagetool/mksquashfs is not available
	// The function should return an error about tools not being found
	if err != nil && !contains(err.Error(), "neither appimagetool nor mksquashfs found") {
		t.Errorf("Pack() unexpected error = %v", err)
	}

	// If tools are available, check output
	if err == nil {
		if outputPath == "" {
			t.Error("Pack() returned empty output path")
		}
		
		// Check if AppDir was created
		appDirPath := filepath.Join("dist", "testapp.AppDir")
		if _, err := os.Stat(appDirPath); os.IsNotExist(err) {
			t.Error("AppDir was not created")
		}
	}
}

func TestCreateAppDirStructure(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "testapp")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho 'test'\n"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Packages: config.PackagesConfig{
			AppImage: config.AppImageConfig{
				Categories: []string{"Utility"},
				DesktopEntry: config.AppImageDesktopConfig{
					Type:     "Application",
					Terminal: false,
				},
			},
		},
	}

	appDir := filepath.Join(tmpDir, "testapp.AppDir")
	if err := packager.createAppDirStructure(appDir, cfg, binaryPath); err != nil {
		t.Errorf("createAppDirStructure() error = %v", err)
	}

	// Check required files exist
	requiredFiles := []string{
		filepath.Join(appDir, "AppRun"),
		filepath.Join(appDir, "usr", "bin", "testapp"),
		filepath.Join(appDir, "usr", "share", "applications", "testapp.desktop"),
		filepath.Join(appDir, "testapp.desktop"), // symlink
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Required file missing: %s", file)
		}
	}

	// Check AppRun is executable
	info, err := os.Stat(filepath.Join(appDir, "AppRun"))
	if err != nil {
		t.Errorf("Failed to stat AppRun: %v", err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Error("AppRun is not executable")
	}
}

func TestCreateAppRun(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	appRunPath := filepath.Join(tmpDir, "AppRun")
	
	cfg := &config.Config{
		Name: "testapp",
	}

	if err := packager.createAppRun(appRunPath, cfg); err != nil {
		t.Errorf("createAppRun() error = %v", err)
	}

	// Check file exists and is executable
	info, err := os.Stat(appRunPath)
	if err != nil {
		t.Errorf("AppRun file not created: %v", err)
	}

	if info.Mode().Perm()&0111 == 0 {
		t.Error("AppRun is not executable")
	}

	// Check content
	content, err := os.ReadFile(appRunPath)
	if err != nil {
		t.Errorf("Failed to read AppRun: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "#!/bin/bash") {
		t.Error("AppRun missing shebang")
	}
	if !contains(contentStr, "testapp") {
		t.Error("AppRun missing app name")
	}
}

func TestCreateDesktopFile(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	desktopPath := filepath.Join(tmpDir, "test.desktop")
	
	cfg := &config.Config{
		Name:        "testapp",
		Description: "Test application",
		Packages: config.PackagesConfig{
			AppImage: config.AppImageConfig{
				Categories: []string{"Utility", "Development"},
				DesktopEntry: config.AppImageDesktopConfig{
					Type:     "Application",
					Terminal: false,
				},
			},
		},
	}

	if err := packager.createDesktopFile(desktopPath, cfg); err != nil {
		t.Errorf("createDesktopFile() error = %v", err)
	}

	// Check content
	content, err := os.ReadFile(desktopPath)
	if err != nil {
		t.Errorf("Failed to read desktop file: %v", err)
	}

	contentStr := string(content)
	requiredFields := []string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=testapp",
		"Comment=Test application",
		"Exec=testapp",
		"Icon=testapp",
		"Categories=Utility;Development",
		"Terminal=false",
	}

	for _, field := range requiredFields {
		if !contains(contentStr, field) {
			t.Errorf("Desktop file missing required field: %s", field)
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
}

func TestCopyFile_Error(t *testing.T) {
	packager := New()
	
	// Test with non-existent source file
	err := packager.copyFile("/non/existent/file", "/tmp/dest")
	if err == nil {
		t.Error("copyFile() should fail with non-existent source file")
	}
}

func TestBuildAppImage_NoTools(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "test.AppDir")
	os.MkdirAll(appDir, 0755)
	
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	ctx := context.Background()
	_, err := packager.buildAppImage(ctx, appDir, cfg)
	
	// Should return error about missing tools
	if err == nil {
		t.Error("buildAppImage() should fail when no AppImage build tools are available")
	}
	
	if !contains(err.Error(), "neither appimagetool nor mksquashfs found") {
		t.Errorf("Expected 'neither appimagetool nor mksquashfs found' error, got: %v", err)
	}
}

func TestBuildWithAppimagetool(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "test.AppDir")
	os.MkdirAll(appDir, 0755)
	outputPath := filepath.Join(tmpDir, "test.AppImage")
	
	ctx := context.Background()
	_, err := packager.buildWithAppimagetool(ctx, appDir, outputPath)
	
	// This will fail because appimagetool is not available, but we test the code path
	if err == nil {
		t.Error("buildWithAppimagetool() should fail when appimagetool is not available")
	}
}

func TestBuildWithSquashfs(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "test.AppDir")
	os.MkdirAll(appDir, 0755)
	outputPath := filepath.Join(tmpDir, "test.AppImage")
	
	ctx := context.Background()
	_, err := packager.buildWithSquashfs(ctx, appDir, outputPath)
	
	// This will fail because mksquashfs is not available, but we test the code path
	if err == nil {
		t.Error("buildWithSquashfs() should fail when mksquashfs is not available")
	}
}

func TestCreateAppImageFromSquashfs(t *testing.T) {
	packager := New()
	
	tmpDir := t.TempDir()
	squashfsPath := filepath.Join(tmpDir, "test.squashfs")
	outputPath := filepath.Join(tmpDir, "test.AppImage")
	
	// Create a dummy squashfs file
	os.WriteFile(squashfsPath, []byte("fake squashfs content"), 0644)
	
	err := packager.createAppImageFromSquashfs(squashfsPath, outputPath)
	if err != nil {
		t.Errorf("createAppImageFromSquashfs() error = %v", err)
	}
	
	// Check output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("AppImage file was not created")
	}
	
	// Check file is executable
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Errorf("Failed to stat AppImage: %v", err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Error("AppImage is not executable")
	}
}

func TestCreateAppRun_Error(t *testing.T) {
	packager := New()
	
	// Test with invalid path
	err := packager.createAppRun("/non/existent/dir/AppRun", &config.Config{})
	if err == nil {
		t.Error("createAppRun() should fail with invalid path")
	}
}

func TestCreateDesktopFile_Error(t *testing.T) {
	packager := New()
	
	// Test with invalid path
	err := packager.createDesktopFile("/non/existent/dir/test.desktop", &config.Config{})
	if err == nil {
		t.Error("createDesktopFile() should fail with invalid path")
	}
}

func TestCreateAppDirStructure_Error(t *testing.T) {
	packager := New()
	
	// Test with invalid binary path
	err := packager.createAppDirStructure("/tmp/test.AppDir", &config.Config{Name: "test"}, "/non/existent/binary")
	if err == nil {
		t.Error("createAppDirStructure() should fail with non-existent binary")
	}
}

func TestAppImagePack_MissingBinary(t *testing.T) {
	packager := New()
	
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Binaries: map[string]string{
			"linux-amd64": "/non/existent/binary",
		},
		Packages: config.PackagesConfig{
			AppImage: config.AppImageConfig{
				Categories: []string{"Utility"},
			},
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
