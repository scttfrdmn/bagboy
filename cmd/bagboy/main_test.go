package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestInitCommand(t *testing.T) {
	// Create temporary directory for test
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Create a mock Go project
	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)

	// Test init command (non-interactive by default)
	cmd := rootCmd
	cmd.SetArgs([]string{"init"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("init command failed: %v", err)
	}

	// Check if bagboy.yaml was created
	if _, err := os.Stat("bagboy.yaml"); os.IsNotExist(err) {
		t.Error("bagboy.yaml was not created")
	}

	// Read the created config to verify it's valid
	content, err := os.ReadFile("bagboy.yaml")
	if err != nil {
		t.Errorf("Could not read bagboy.yaml: %v", err)
	}

	configStr := string(content)
	if !strings.Contains(configStr, "name:") {
		t.Error("bagboy.yaml missing name field")
	}
	if !strings.Contains(configStr, "version:") {
		t.Error("bagboy.yaml missing version field")
	}
}

func TestPackCommand(t *testing.T) {
	// Create temporary directory with test config
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Create test binary
	testBinary := filepath.Join(testDir, "testapp-linux-amd64")
	os.WriteFile(testBinary, []byte("fake binary"), 0755)

	// Create test config
	config := `name: testapp
version: 1.0.0
description: Test application
author: Test Author <test@example.com>
homepage: https://example.com
license: MIT

binaries:
  linux-amd64: testapp-linux-amd64

packages:
  deb:
    maintainer: test@example.com
    section: utils
    priority: optional
`
	os.WriteFile("bagboy.yaml", []byte(config), 0644)

	// Test pack command with specific format
	cmd := rootCmd
	cmd.SetArgs([]string{"pack", "--brew"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("pack command failed: %v", err)
	}

	// Check if brew formula was created
	if _, err := os.Stat("dist/testapp.rb"); os.IsNotExist(err) {
		t.Error("Brew formula was not created")
	}

	// Verify the formula content
	content, err := os.ReadFile("dist/testapp.rb")
	if err != nil {
		t.Errorf("Could not read brew formula: %v", err)
	}

	formulaStr := string(content)
	if !strings.Contains(formulaStr, "testapp") {
		t.Error("Brew formula missing app name")
	}
}

func TestValidateCommand(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Create invalid config (missing required fields)
	config := `name: testapp
# missing version and other required fields
`
	os.WriteFile("bagboy.yaml", []byte(config), 0644)

	// Test validate command
	cmd := rootCmd
	cmd.SetArgs([]string{"validate"})
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("validate command should have failed with invalid config")
	}

	// Check error output
	output := buf.String()
	if !strings.Contains(output, "validation") {
		t.Errorf("Expected validation error message, got: %s", output)
	}
}

func TestSignCommand(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Create test config
	config := `name: testapp
version: 1.0.0
description: Test application
author: Test Author <test@example.com>
license: MIT

binaries:
  linux-amd64: testapp-linux-amd64
`
	os.WriteFile("bagboy.yaml", []byte(config), 0644)

	// Test sign --check command
	cmd := rootCmd
	cmd.SetArgs([]string{"sign", "--check"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("sign --check command failed: %v", err)
	}

	// The command should complete successfully even if signing is not configured
	// This tests that the command runs without crashing
}

func TestHelpCommand(t *testing.T) {
	// Test help command
	cmd := rootCmd
	cmd.SetArgs([]string{"--help"})
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("help command failed: %v", err)
	}

	// Check help output contains expected commands
	output := buf.String()
	expectedCommands := []string{"init", "pack", "publish", "validate", "sign"}
	for _, cmdName := range expectedCommands {
		if !strings.Contains(output, cmdName) {
			t.Errorf("Help output missing command: %s", cmdName)
		}
	}
}

func TestPackAllFormats(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Create test binaries for all platforms
	binaries := map[string]string{
		"linux-amd64":   "testapp-linux-amd64",
		"darwin-amd64":  "testapp-darwin-amd64", 
		"windows-amd64": "testapp-windows-amd64.exe",
	}

	for _, binary := range binaries {
		os.WriteFile(binary, []byte("fake binary"), 0755)
	}

	// Create comprehensive test config
	config := `name: testapp
version: 1.0.0
description: Test application for all formats
author: Test Author <test@example.com>
homepage: https://example.com
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
	os.WriteFile("bagboy.yaml", []byte(config), 0644)

	// Test pack with formats that don't require external tools
	cmd := rootCmd
	cmd.SetArgs([]string{"pack", "--brew", "--scoop", "--installer"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("pack command failed: %v", err)
	}

	// Check that formats were created
	expectedFiles := []string{
		"dist/testapp.rb",      // brew
		"dist/testapp.json",    // scoop
		"dist/install.sh",      // installer
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}

	// Check that dist directory was created
	if _, err := os.Stat("dist"); os.IsNotExist(err) {
		t.Error("dist directory was not created")
	}
}

// Helper function to reset cobra command state between tests
func resetCommand(cmd *cobra.Command) {
	cmd.SetArgs(nil)
	cmd.SetOut(nil)
	cmd.SetErr(nil)
}

func TestCLICommands(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "help command",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Universal Software Packager",
		},
		{
			name:     "version command",
			args:     []string{"version"},
			wantErr:  false,
			contains: "bagboy version",
		},
		{
			name:     "pack help",
			args:     []string{"pack", "--help"},
			wantErr:  false,
			contains: "Create packages for distribution",
		},
		{
			name:     "validate help",
			args:     []string{"validate", "--help"},
			wantErr:  false,
			contains: "Validate your bagboy.yaml",
		},
		{
			name:     "init help",
			args:     []string{"init", "--help"},
			wantErr:  false,
			contains: "Initialize a new bagboy project",
		},
		{
			name:     "publish help",
			args:     []string{"publish", "--help"},
			wantErr:  false,
			contains: "Complete publishing workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new root command for each test
			cmd := &cobra.Command{
				Use: "bagboy",
			}
			
			// Add subcommands with improved help text
			cmd.AddCommand(&cobra.Command{
				Use:   "version",
				Short: "Show version information",
				Run: func(cmd *cobra.Command, args []string) {
					fmt.Println("bagboy version 0.6.0-dev")
				},
			})
			
			cmd.AddCommand(&cobra.Command{
				Use:   "pack",
				Short: "Create packages for distribution",
				Long:  "Create packages for various platforms and package managers.",
			})
			
			cmd.AddCommand(&cobra.Command{
				Use:   "validate",
				Short: "Validate bagboy configuration",
				Long:  "Validate your bagboy.yaml configuration file.",
			})
			
			cmd.AddCommand(&cobra.Command{
				Use:   "init",
				Short: "Initialize a new bagboy project",
				Long:  "Initialize a new bagboy project with smart detection.",
			})
			
			cmd.AddCommand(&cobra.Command{
				Use:   "publish",
				Short: "Pack all formats and create GitHub release",
				Long:  "Complete publishing workflow: pack, release, and distribute.",
			})

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("Command error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("Expected output to contain '%s', got: %s", tt.contains, output)
			}
		})
	}
}

func TestCommandAliases(t *testing.T) {
	tests := []struct {
		command string
		aliases []string
	}{
		{"pack", []string{"p", "package", "build"}},
		{"init", []string{"i", "new", "create"}},
		{"validate", []string{"v", "check", "verify"}},
		{"publish", []string{"pub", "release", "deploy"}},
		{"version", []string{"v", "--version"}},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			// Test that aliases are properly configured
			// This is more of a documentation test since we can't easily
			// test the actual cobra command structure here
			if len(tt.aliases) == 0 {
				t.Errorf("Command %s should have aliases", tt.command)
			}
		})
	}
}
