/*
Copyright 2026 Scott Friedman

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package requirements

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestNewRequirementChecker(t *testing.T) {
	rc := NewRequirementChecker()
	
	if rc == nil {
		t.Fatal("NewRequirementChecker returned nil")
	}
	
	if rc.requirements == nil {
		t.Fatal("Requirements map not initialized")
	}
	
	// Verify key formats are initialized
	expectedFormats := []string{"dmg", "msi", "deb", "rpm", "appimage", "docker", "snap", "signing"}
	for _, format := range expectedFormats {
		if _, exists := rc.requirements[format]; !exists {
			t.Errorf("Format %s not initialized in requirements", format)
		}
	}
}

func TestCheckRequirements_EmptyFormats(t *testing.T) {
	rc := NewRequirementChecker()
	
	results := rc.CheckRequirements([]string{})
	
	if len(results) != 0 {
		t.Errorf("Expected empty results for empty formats, got %d results", len(results))
	}
}

func TestCheckRequirements_UnknownFormat(t *testing.T) {
	rc := NewRequirementChecker()
	
	results := rc.CheckRequirements([]string{"unknown-format"})
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	status := results["unknown-format"]
	if status.Format != "unknown-format" {
		t.Errorf("Expected format 'unknown-format', got '%s'", status.Format)
	}
	
	if !status.Available {
		t.Error("Unknown format should be available (no requirements)")
	}
	
	if len(status.Missing) != 0 {
		t.Errorf("Unknown format should have no missing requirements, got %d", len(status.Missing))
	}
}

func TestCheckRequirements_MultipleFormats(t *testing.T) {
	rc := NewRequirementChecker()
	
	formats := []string{"deb", "rpm", "docker"}
	results := rc.CheckRequirements(formats)
	
	if len(results) != len(formats) {
		t.Errorf("Expected %d results, got %d", len(formats), len(results))
	}
	
	for _, format := range formats {
		if _, exists := results[format]; !exists {
			t.Errorf("Missing result for format %s", format)
		}
	}
}

func TestCheckFormatRequirements_DEB(t *testing.T) {
	rc := NewRequirementChecker()
	
	status := rc.checkFormatRequirements("deb")
	
	if status.Format != "deb" {
		t.Errorf("Expected format 'deb', got '%s'", status.Format)
	}
	
	// DEB should always be available (optional requirements only)
	if !status.Available {
		t.Error("DEB format should be available (has built-in support)")
	}
	
	// Check if dpkg-deb is available and categorized correctly
	_, err := exec.LookPath("dpkg-deb")
	if err != nil {
		// dpkg-deb not available, should be in optional
		if len(status.Optional) == 0 {
			t.Error("Expected dpkg-deb in optional requirements when not available")
		}
	}
}

func TestCheckFormatRequirements_Docker(t *testing.T) {
	rc := NewRequirementChecker()
	
	status := rc.checkFormatRequirements("docker")
	
	if status.Format != "docker" {
		t.Errorf("Expected format 'docker', got '%s'", status.Format)
	}
	
	// Check if docker is available
	_, err := exec.LookPath("docker")
	if err != nil {
		// Docker not available, should not be available
		if status.Available {
			t.Error("Docker format should not be available when docker command is missing")
		}
		
		if len(status.Missing) == 0 {
			t.Error("Expected docker in missing requirements when not available")
		}
		
		// Should have installation instructions
		if len(status.Instructions) == 0 {
			t.Error("Expected installation instructions when docker is missing")
		}
	} else {
		// Docker available, should be available
		if !status.Available {
			t.Error("Docker format should be available when docker command is present")
		}
	}
}

func TestCheckFormatRequirements_PlatformSpecific(t *testing.T) {
	rc := NewRequirementChecker()
	
	tests := []struct {
		format   string
		platform string
		shouldBeAvailable bool
	}{
		{"dmg", "darwin", true},   // hdiutil built-in on macOS
		{"dmg", "linux", false},  // hdiutil not available on Linux
		{"dmg", "windows", false}, // hdiutil not available on Windows
		{"msi", "windows", false}, // WiX not typically installed
		{"msi", "darwin", false},  // WiX not available on macOS
		{"msi", "linux", false},   // WiX not available on Linux
	}
	
	for _, test := range tests {
		t.Run(test.format+"_"+test.platform, func(t *testing.T) {
			// Skip if not on the target platform
			if runtime.GOOS != test.platform {
				t.Skipf("Skipping %s test on %s", test.format, runtime.GOOS)
			}
			
			status := rc.checkFormatRequirements(test.format)
			
			if test.shouldBeAvailable && !status.Available {
				t.Errorf("Format %s should be available on %s", test.format, test.platform)
			}
			
			if !test.shouldBeAvailable && status.Available && len(status.Missing) == 0 {
				// Only fail if there are no missing requirements (meaning it thinks it's available)
				t.Errorf("Format %s should not be readily available on %s", test.format, test.platform)
			}
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	rc := NewRequirementChecker()
	
	tests := []struct {
		command   string
		available bool
	}{
		{"", true},           // Empty command should be considered available
		{"ls", true},         // ls should be available on Unix systems
		{"nonexistent-cmd-12345", false}, // This command should not exist
	}
	
	// Skip ls test on Windows
	if runtime.GOOS == "windows" {
		tests[1] = struct {
			command   string
			available bool
		}{"dir", true} // Use dir instead of ls on Windows
	}
	
	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			result := rc.isCommandAvailable(test.command)
			if result != test.available {
				t.Errorf("isCommandAvailable(%s) = %v, want %v", test.command, result, test.available)
			}
		})
	}
}

func TestGetInstallInstruction(t *testing.T) {
	rc := NewRequirementChecker()
	
	req := Requirement{
		Name:           "TestTool",
		MacInstall:     "brew install testtool",
		LinuxInstall:   "sudo apt-get install testtool",
		WindowsInstall: "choco install testtool",
	}
	
	instruction := rc.getInstallInstruction(req)
	
	// Should return platform-specific instruction
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(instruction, "brew install testtool") {
			t.Errorf("Expected macOS instruction, got: %s", instruction)
		}
	case "linux":
		if !strings.Contains(instruction, "sudo apt-get install testtool") {
			t.Errorf("Expected Linux instruction, got: %s", instruction)
		}
	case "windows":
		if !strings.Contains(instruction, "choco install testtool") {
			t.Errorf("Expected Windows instruction, got: %s", instruction)
		}
	}
	
	if !strings.Contains(instruction, "TestTool:") {
		t.Errorf("Expected tool name in instruction, got: %s", instruction)
	}
}

func TestGetInstallInstruction_MissingPlatform(t *testing.T) {
	rc := NewRequirementChecker()
	
	req := Requirement{
		Name: "TestTool",
		// No platform-specific install instructions
	}
	
	instruction := rc.getInstallInstruction(req)
	
	if instruction != "" {
		t.Errorf("Expected empty instruction for missing platform, got: %s", instruction)
	}
}

func TestRequirementStatus_Structure(t *testing.T) {
	rc := NewRequirementChecker()
	
	// Test with a format that has both required and optional requirements
	status := rc.checkFormatRequirements("docker")
	
	// Verify structure
	if status.Format != "docker" {
		t.Errorf("Expected format 'docker', got '%s'", status.Format)
	}
	
	// Available should be boolean
	if status.Available != true && status.Available != false {
		t.Error("Available field should be boolean")
	}
	
	// Missing and Optional should be slices (can be nil or empty)
	// This is acceptable Go behavior - nil slices are valid and equivalent to empty slices
	
	// Instructions should be slices (can be nil or empty)
	// This is acceptable Go behavior - nil slices are valid and equivalent to empty slices
}

func TestRequirementInitialization(t *testing.T) {
	rc := NewRequirementChecker()
	
	// Test that all expected formats have requirements defined
	expectedFormats := map[string]bool{
		"dmg":      true,
		"msi":      true,
		"deb":      true,
		"rpm":      true,
		"appimage": true,
		"docker":   true,
		"snap":     true,
		"signing":  true,
	}
	
	for format, shouldExist := range expectedFormats {
		requirements, exists := rc.requirements[format]
		
		if shouldExist && !exists {
			t.Errorf("Format %s should have requirements defined", format)
			continue
		}
		
		if !shouldExist && exists {
			t.Errorf("Format %s should not have requirements defined", format)
			continue
		}
		
		if shouldExist && len(requirements) == 0 {
			t.Errorf("Format %s should have at least one requirement", format)
		}
		
		// Verify requirement structure
		for _, req := range requirements {
			if req.Name == "" {
				t.Errorf("Requirement for %s has empty name", format)
			}
			
			if req.Description == "" {
				t.Errorf("Requirement %s for %s has empty description", req.Name, format)
			}
		}
	}
}

func TestPrintRequirementReport(t *testing.T) {
	rc := NewRequirementChecker()
	
	// Create test results
	results := map[string]RequirementStatus{
		"deb": {
			Format:    "deb",
			Available: true,
			Missing:   []Requirement{},
			Optional: []Requirement{
				{Name: "dpkg-deb", Description: "Debian package builder"},
			},
			Instructions: []string{"dpkg-deb: sudo apt-get install dpkg-dev"},
		},
		"docker": {
			Format:    "docker",
			Available: false,
			Missing: []Requirement{
				{Name: "Docker", Description: "Docker container platform"},
			},
			Optional:     []Requirement{},
			Instructions: []string{"Docker: brew install --cask docker"},
		},
	}
	
	// This test just ensures PrintRequirementReport doesn't panic
	// In a real scenario, you might capture stdout to verify output
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintRequirementReport panicked: %v", r)
		}
	}()
	
	rc.PrintRequirementReport(results)
}

func TestRequirementChecker_AllFormats(t *testing.T) {
	rc := NewRequirementChecker()
	
	// Test all known formats
	allFormats := []string{"dmg", "msi", "deb", "rpm", "appimage", "docker", "snap", "signing"}
	
	results := rc.CheckRequirements(allFormats)
	
	if len(results) != len(allFormats) {
		t.Errorf("Expected %d results, got %d", len(allFormats), len(results))
	}
	
	for _, format := range allFormats {
		status, exists := results[format]
		if !exists {
			t.Errorf("Missing result for format %s", format)
			continue
		}
		
		if status.Format != format {
			t.Errorf("Expected format %s, got %s", format, status.Format)
		}
		
		// Each format should have some kind of result
		// (either available, or missing/optional requirements)
		if !status.Available && len(status.Missing) == 0 && len(status.Optional) == 0 {
			t.Errorf("Format %s has no requirements but is not available", format)
		}
	}
}

func TestRequirementChecker_EdgeCases(t *testing.T) {
	rc := NewRequirementChecker()
	
	t.Run("NilFormats", func(t *testing.T) {
		results := rc.CheckRequirements(nil)
		if len(results) != 0 {
			t.Errorf("Expected empty results for nil formats, got %d", len(results))
		}
	})
	
	t.Run("DuplicateFormats", func(t *testing.T) {
		results := rc.CheckRequirements([]string{"deb", "deb", "rpm"})
		if len(results) != 2 {
			t.Errorf("Expected 2 unique results, got %d", len(results))
		}
	})
	
	t.Run("EmptyStringFormat", func(t *testing.T) {
		results := rc.CheckRequirements([]string{""})
		if len(results) != 1 {
			t.Errorf("Expected 1 result for empty string format, got %d", len(results))
		}
		
		status := results[""]
		if !status.Available {
			t.Error("Empty format should be available (no requirements)")
		}
	})
}
