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
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Requirement represents a build/deployment requirement
type Requirement struct {
	Name        string
	Command     string
	InstallCmd  string
	MacInstall  string
	LinuxInstall string
	WindowsInstall string
	Required    bool
	Description string
}

// RequirementChecker checks system requirements for package formats
type RequirementChecker struct {
	requirements map[string][]Requirement
}

// NewRequirementChecker creates a new requirement checker
func NewRequirementChecker() *RequirementChecker {
	rc := &RequirementChecker{
		requirements: make(map[string][]Requirement),
	}
	rc.initializeRequirements()
	return rc
}

func (rc *RequirementChecker) initializeRequirements() {
	// DMG requirements
	rc.requirements["dmg"] = []Requirement{
		{
			Name:        "hdiutil",
			Command:     "hdiutil",
			Required:    true,
			Description: "macOS disk image utility (built-in on macOS)",
			MacInstall:  "Built-in on macOS",
			LinuxInstall: "Not available on Linux",
			WindowsInstall: "Not available on Windows",
		},
	}

	// MSI requirements
	rc.requirements["msi"] = []Requirement{
		{
			Name:        "WiX Toolset",
			Command:     "candle",
			Required:    true,
			Description: "Windows Installer XML toolset",
			WindowsInstall: "Download from https://wixtoolset.org/",
			MacInstall:  "Not available on macOS",
			LinuxInstall: "Not available on Linux",
		},
	}

	// DEB requirements
	rc.requirements["deb"] = []Requirement{
		{
			Name:        "dpkg-deb",
			Command:     "dpkg-deb",
			Required:    false,
			Description: "Debian package builder (optional, bagboy has built-in support)",
			LinuxInstall: "sudo apt-get install dpkg-dev",
			MacInstall:  "brew install dpkg",
			WindowsInstall: "Not available on Windows",
		},
	}

	// RPM requirements
	rc.requirements["rpm"] = []Requirement{
		{
			Name:        "rpmbuild",
			Command:     "rpmbuild",
			Required:    false,
			Description: "RPM package builder (optional for full RPM support)",
			LinuxInstall: "sudo yum install rpm-build",
			MacInstall:  "brew install rpm",
			WindowsInstall: "Not available on Windows",
		},
	}

	// AppImage requirements
	rc.requirements["appimage"] = []Requirement{
		{
			Name:        "appimagetool",
			Command:     "appimagetool",
			Required:    false,
			Description: "AppImage creation tool (optional, bagboy has built-in support)",
			LinuxInstall: "wget https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage",
			MacInstall:  "Not typically used on macOS",
			WindowsInstall: "Not available on Windows",
		},
	}

	// Docker requirements
	rc.requirements["docker"] = []Requirement{
		{
			Name:        "Docker",
			Command:     "docker",
			Required:    true,
			Description: "Docker container platform",
			MacInstall:  "brew install --cask docker",
			LinuxInstall: "curl -fsSL https://get.docker.com | sh",
			WindowsInstall: "Download Docker Desktop from docker.com",
		},
	}

	// Snap requirements
	rc.requirements["snap"] = []Requirement{
		{
			Name:        "snapcraft",
			Command:     "snapcraft",
			Required:    true,
			Description: "Snap package builder",
			LinuxInstall: "sudo snap install snapcraft --classic",
			MacInstall:  "Not available on macOS",
			WindowsInstall: "Not available on Windows",
		},
	}

	// Code signing requirements
	rc.requirements["signing"] = []Requirement{
		{
			Name:        "Code Signing",
			Required:    true,
			Description: "Code signing for trusted software distribution",
			MacInstall:  "Join Apple Developer Program, install Xcode",
			LinuxInstall: "Install GPG: sudo apt-get install gnupg",
			WindowsInstall: "Install Windows SDK, purchase code signing certificate",
		},
	}
}

// CheckRequirements checks if requirements are met for given package formats
func (rc *RequirementChecker) CheckRequirements(formats []string) map[string]RequirementStatus {
	results := make(map[string]RequirementStatus)
	
	for _, format := range formats {
		results[format] = rc.checkFormatRequirements(format)
	}
	
	return results
}

// RequirementStatus represents the status of requirements for a format
type RequirementStatus struct {
	Format       string
	Available    bool
	Missing      []Requirement
	Optional     []Requirement
	Instructions []string
}

func (rc *RequirementChecker) checkFormatRequirements(format string) RequirementStatus {
	status := RequirementStatus{
		Format:    format,
		Available: true,
	}
	
	requirements, exists := rc.requirements[format]
	if !exists {
		// No special requirements
		return status
	}
	
	for _, req := range requirements {
		if rc.isCommandAvailable(req.Command) {
			continue
		}
		
		if req.Required {
			status.Available = false
			status.Missing = append(status.Missing, req)
		} else {
			status.Optional = append(status.Optional, req)
		}
		
		// Add installation instructions
		instruction := rc.getInstallInstruction(req)
		if instruction != "" {
			status.Instructions = append(status.Instructions, instruction)
		}
	}
	
	return status
}

func (rc *RequirementChecker) isCommandAvailable(command string) bool {
	if command == "" {
		return true
	}
	_, err := exec.LookPath(command)
	return err == nil
}

func (rc *RequirementChecker) getInstallInstruction(req Requirement) string {
	switch runtime.GOOS {
	case "darwin":
		if req.MacInstall != "" {
			return fmt.Sprintf("%s: %s", req.Name, req.MacInstall)
		}
	case "linux":
		if req.LinuxInstall != "" {
			return fmt.Sprintf("%s: %s", req.Name, req.LinuxInstall)
		}
	case "windows":
		if req.WindowsInstall != "" {
			return fmt.Sprintf("%s: %s", req.Name, req.WindowsInstall)
		}
	}
	return ""
}

// PrintRequirementReport prints a formatted requirement report
func (rc *RequirementChecker) PrintRequirementReport(results map[string]RequirementStatus) {
	fmt.Println("📋 Package Format Requirements Check")
	fmt.Println("=====================================")
	
	for format, status := range results {
		fmt.Printf("\n🔧 %s:\n", strings.ToUpper(format))
		
		if status.Available && len(status.Missing) == 0 {
			fmt.Println("  ✅ Ready to build")
		} else {
			if len(status.Missing) > 0 {
				fmt.Println("  ❌ Missing required dependencies:")
				for _, req := range status.Missing {
					fmt.Printf("    • %s - %s\n", req.Name, req.Description)
				}
			}
			
			if len(status.Optional) > 0 {
				fmt.Println("  ⚠️  Optional dependencies not found:")
				for _, req := range status.Optional {
					fmt.Printf("    • %s - %s\n", req.Name, req.Description)
				}
			}
		}
		
		if len(status.Instructions) > 0 {
			fmt.Println("  📝 Installation instructions:")
			for _, instruction := range status.Instructions {
				fmt.Printf("    %s\n", instruction)
			}
		}
	}
	
	fmt.Println("\n💡 Note: bagboy includes built-in support for most formats")
	fmt.Println("   External tools are only needed for advanced features")
}
