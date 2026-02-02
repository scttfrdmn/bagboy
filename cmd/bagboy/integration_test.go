package main

import (
	"os"
	"strings"
	"testing"
)

// TestEndToEndWorkflow tests the complete bagboy workflow
func TestEndToEndWorkflow(t *testing.T) {
	// Create isolated test environment
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Step 1: Initialize project
	t.Run("Initialize", func(t *testing.T) {
		// Create mock project structure
		os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.21\n"), 0644)
		os.WriteFile("main.go", []byte("package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n"), 0644)

		// Run bagboy init
		cmd := rootCmd
		cmd.SetArgs([]string{"init"})
		
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("bagboy init failed: %v", err)
		}

		// Verify bagboy.yaml was created
		if _, err := os.Stat("bagboy.yaml"); os.IsNotExist(err) {
			t.Fatal("bagboy.yaml was not created")
		}

		// Verify config content
		content, err := os.ReadFile("bagboy.yaml")
		if err != nil {
			t.Fatalf("Could not read bagboy.yaml: %v", err)
		}

		configStr := string(content)
		if !strings.Contains(configStr, "name: testapp") {
			t.Error("Config missing detected app name")
		}
	})

	// Step 2: Create mock binaries
	t.Run("CreateBinaries", func(t *testing.T) {
		binaries := map[string]string{
			"testapp-linux-amd64":     "#!/bin/bash\necho 'Linux AMD64 binary'\n",
			"testapp-darwin-amd64":    "#!/bin/bash\necho 'macOS AMD64 binary'\n",
			"testapp-windows-amd64.exe": "@echo off\necho Windows AMD64 binary\n",
		}

		for name, content := range binaries {
			err := os.WriteFile(name, []byte(content), 0755)
			if err != nil {
				t.Fatalf("Failed to create binary %s: %v", name, err)
			}
		}

		// Update bagboy.yaml with binary paths
		config := `name: testapp
version: 1.0.0
description: Test application for end-to-end testing
author: Test Author <test@example.com>
homepage: https://github.com/test/testapp
license: MIT

binaries:
  linux-amd64: testapp-linux-amd64
  darwin-amd64: testapp-darwin-amd64
  windows-amd64: testapp-windows-amd64.exe

packages:
  deb:
    maintainer: test@example.com
    section: utils
    priority: optional
  rpm:
    group: Applications/System
    vendor: Test Vendor
`
		err := os.WriteFile("bagboy.yaml", []byte(config), 0644)
		if err != nil {
			t.Fatalf("Failed to update bagboy.yaml: %v", err)
		}
	})

	// Step 3: Validate configuration
	t.Run("Validate", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"validate"})
		
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("bagboy validate failed: %v", err)
		}
	})

	// Step 4: Pack individual formats
	t.Run("PackIndividual", func(t *testing.T) {
		formats := []struct {
			name     string
			flag     string
			expected string
		}{
			{"Homebrew", "--brew", "dist/testapp.rb"},
			{"Scoop", "--scoop", "dist/testapp.json"},
			{"Installer", "--installer", "dist/install.sh"},
		}

		for _, format := range formats {
			t.Run(format.name, func(t *testing.T) {
				cmd := rootCmd
				cmd.SetArgs([]string{"pack", format.flag})
				
				err := cmd.Execute()
				if err != nil {
					t.Fatalf("bagboy pack %s failed: %v", format.flag, err)
				}

				// Verify output file was created
				if _, err := os.Stat(format.expected); os.IsNotExist(err) {
					t.Errorf("Expected file %s was not created", format.expected)
				}

				// Verify file has content
				content, err := os.ReadFile(format.expected)
				if err != nil {
					t.Errorf("Could not read %s: %v", format.expected, err)
				}

				if len(content) == 0 {
					t.Errorf("File %s is empty", format.expected)
				}

				// Verify content contains app name
				if !strings.Contains(string(content), "testapp") {
					t.Errorf("File %s missing app name", format.expected)
				}
			})
		}
	})

	// Step 5: Pack multiple formats
	t.Run("PackMultiple", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"pack", "--brew", "--scoop", "--installer", "--deb"})
		
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("bagboy pack multiple failed: %v", err)
		}

		// Verify all expected files were created
		expectedFiles := []string{
			"dist/testapp.rb",
			"dist/testapp.json", 
			"dist/install.sh",
			"dist/testapp_1.0.0_amd64.deb",
		}

		for _, file := range expectedFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Expected file %s was not created", file)
			}
		}
	})

	// Step 6: Check signing status
	t.Run("CheckSigning", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"sign", "--check"})
		
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("bagboy sign --check failed: %v", err)
		}
		// This should complete without error even if signing is not configured
	})

	// Step 7: Test dry-run publish
	t.Run("PublishDryRun", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"publish", "--dry-run"})
		
		err := cmd.Execute()
		// This may fail due to missing tools or GitHub config, but should not crash
		if err != nil && !strings.Contains(err.Error(), "GitHub") && !strings.Contains(err.Error(), "rpmbuild") && !strings.Contains(err.Error(), "not found") {
			t.Errorf("Unexpected error in publish dry-run: %v", err)
		}
	})
}

// TestCrossFormatCompatibility tests that all formats work together
func TestCrossFormatCompatibility(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Create comprehensive test setup
	setupComprehensiveTest(t)

	// Test that we can create packages for different platforms without conflicts
	t.Run("AllPlatforms", func(t *testing.T) {
		// Test platform-specific formats
		platformTests := []struct {
			name     string
			formats  []string
			platform string
		}{
			{
				name:     "macOS",
				formats:  []string{"--brew", "--dmg"},
				platform: "darwin",
			},
			{
				name:     "Windows", 
				formats:  []string{"--scoop", "--chocolatey", "--winget", "--msi", "--msix"},
				platform: "windows",
			},
			{
				name:     "Linux",
				formats:  []string{"--deb", "--rpm", "--appimage", "--snap", "--flatpak"},
				platform: "linux",
			},
		}

		for _, test := range platformTests {
			t.Run(test.name, func(t *testing.T) {
				args := append([]string{"pack"}, test.formats...)
				cmd := rootCmd
				cmd.SetArgs(args)
				
				err := cmd.Execute()
				if err != nil {
					t.Errorf("Platform %s formats failed: %v", test.name, err)
				}

				// Verify dist directory has content
				if _, err := os.Stat("dist"); os.IsNotExist(err) {
					t.Error("dist directory was not created")
				}
			})
		}
	})
}

// TestErrorRecovery tests graceful handling of various error conditions
func TestErrorRecovery(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	t.Run("MissingConfig", func(t *testing.T) {
		// Test commands without bagboy.yaml
		cmd := rootCmd
		cmd.SetArgs([]string{"pack", "--brew"})
		
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for missing config")
		}

		if !strings.Contains(err.Error(), "bagboy.yaml") {
			t.Errorf("Expected config file error, got: %v", err)
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		// Create invalid config
		invalidConfig := `name: testapp
# missing version and other required fields
invalid_yaml: [unclosed
`
		os.WriteFile("bagboy.yaml", []byte(invalidConfig), 0644)

		cmd := rootCmd
		cmd.SetArgs([]string{"validate"})
		
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for invalid config")
		}
	})

	t.Run("MissingBinaries", func(t *testing.T) {
		// Create valid config but missing binaries
		config := `name: testapp
version: 1.0.0
description: Test app
author: Test Author
license: MIT

binaries:
  linux-amd64: /non/existent/binary
`
		os.WriteFile("bagboy.yaml", []byte(config), 0644)

		cmd := rootCmd
		cmd.SetArgs([]string{"pack", "--installer"})
		
		err := cmd.Execute()
		// Should handle missing binaries gracefully
		if err != nil && !strings.Contains(err.Error(), "binary") {
			t.Errorf("Unexpected error for missing binary: %v", err)
		}
	})
}

// TestPackageValidation tests that generated packages have correct structure
func TestPackageValidation(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	setupComprehensiveTest(t)

	t.Run("BrewFormula", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"pack", "--brew"})
		
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("brew pack failed: %v", err)
		}

		// Validate brew formula structure
		content, err := os.ReadFile("dist/testapp.rb")
		if err != nil {
			t.Fatalf("Could not read brew formula: %v", err)
		}

		formula := string(content)
		requiredElements := []string{
			"class Testapp",
			"desc \"Test application\"",
			"homepage \"https://github.com/test/testapp\"",
			"version \"1.0.0\"",
			"def install",
		}

		for _, element := range requiredElements {
			if !strings.Contains(formula, element) {
				t.Errorf("Brew formula missing: %s", element)
			}
		}
	})

	t.Run("ScoopManifest", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"pack", "--scoop"})
		
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("scoop pack failed: %v", err)
		}

		// Validate scoop manifest structure
		content, err := os.ReadFile("dist/testapp.json")
		if err != nil {
			t.Fatalf("Could not read scoop manifest: %v", err)
		}

		manifest := string(content)
		requiredElements := []string{
			"\"version\": \"1.0.0\"",
			"\"description\": \"Test application\"",
			"\"homepage\": \"https://github.com/test/testapp\"",
			"\"license\": \"MIT\"",
		}

		for _, element := range requiredElements {
			if !strings.Contains(manifest, element) {
				t.Errorf("Scoop manifest missing: %s", element)
			}
		}
	})

	t.Run("InstallScript", func(t *testing.T) {
		cmd := rootCmd
		cmd.SetArgs([]string{"pack", "--installer"})
		
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("installer pack failed: %v", err)
		}

		// Validate install script structure
		content, err := os.ReadFile("dist/install.sh")
		if err != nil {
			t.Fatalf("Could not read install script: %v", err)
		}

		script := string(content)
		requiredElements := []string{
			"#!/bin/bash",
			"set -e",
			"testapp",
			"1.0.0",
			"OS=",
			"ARCH=",
		}

		for _, element := range requiredElements {
			if !strings.Contains(script, element) {
				t.Errorf("Install script missing: %s", element)
			}
		}

		// Verify script is executable
		info, err := os.Stat("dist/install.sh")
		if err != nil {
			t.Fatalf("Could not stat install script: %v", err)
		}

		if info.Mode()&0111 == 0 {
			t.Error("Install script is not executable")
		}
	})
}

// Helper function to set up comprehensive test environment
func setupComprehensiveTest(t *testing.T) {
	// Create mock binaries
	binaries := map[string]string{
		"testapp-linux-amd64":     "#!/bin/bash\necho 'Linux AMD64'\n",
		"testapp-linux-arm64":     "#!/bin/bash\necho 'Linux ARM64'\n", 
		"testapp-darwin-amd64":    "#!/bin/bash\necho 'macOS AMD64'\n",
		"testapp-darwin-arm64":    "#!/bin/bash\necho 'macOS ARM64'\n",
		"testapp-windows-amd64.exe": "@echo off\necho Windows AMD64\n",
	}

	for name, content := range binaries {
		err := os.WriteFile(name, []byte(content), 0755)
		if err != nil {
			t.Fatalf("Failed to create binary %s: %v", name, err)
		}
	}

	// Create comprehensive config
	config := `name: testapp
version: 1.0.0
description: Test application for comprehensive testing
author: Test Author <test@example.com>
homepage: https://github.com/test/testapp
license: MIT

binaries:
  linux-amd64: testapp-linux-amd64
  linux-arm64: testapp-linux-arm64
  darwin-amd64: testapp-darwin-amd64
  darwin-arm64: testapp-darwin-arm64
  windows-amd64: testapp-windows-amd64.exe

packages:
  deb:
    maintainer: test@example.com
    section: utils
    priority: optional
  rpm:
    group: Applications/System
    vendor: Test Vendor
  chocolatey:
    package_source_url: https://github.com/test/testapp
    docs_url: https://github.com/test/testapp/docs
  winget:
    package_identifier: TestAuthor.TestApp
    publisher: Test Author
    minimum_os_version: 10.0.0.0
`

	err := os.WriteFile("bagboy.yaml", []byte(config), 0644)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
}
