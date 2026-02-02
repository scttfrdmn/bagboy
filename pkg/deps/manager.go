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

package deps

import (
	"fmt"
	"strings"
)

// DependencyManager handles dependency resolution and validation
type DependencyManager struct {
	Runtime         []Dependency          `yaml:"runtime"`
	Build           []Dependency          `yaml:"build"`
	PackageSpecific map[string]PackageDeps `yaml:"package_specific"`
}

// Dependency represents a single dependency with version constraints
type Dependency struct {
	Name      string   `yaml:"name"`
	Version   string   `yaml:"version"`
	Platforms []string `yaml:"platforms"`
	Optional  bool     `yaml:"optional"`
}

// PackageDeps represents package-format-specific dependencies
type PackageDeps struct {
	Depends    []string `yaml:"depends"`
	Recommends []string `yaml:"recommends"`
	Suggests   []string `yaml:"suggests"`
	Conflicts  []string `yaml:"conflicts"`
}

// ValidationResult represents the result of dependency validation
type ValidationResult struct {
	Valid   bool
	Missing []Dependency
	Errors  []string
}

// NewDependencyManager creates a new dependency manager
func NewDependencyManager() *DependencyManager {
	return &DependencyManager{
		PackageSpecific: make(map[string]PackageDeps),
	}
}

// ValidateAll validates all dependencies for the current platform
func (dm *DependencyManager) ValidateAll() *ValidationResult {
	result := &ValidationResult{
		Valid: true,
	}
	
	// Validate runtime dependencies
	for _, dep := range dm.Runtime {
		if !dm.isDependencyAvailable(dep) {
			result.Valid = false
			result.Missing = append(result.Missing, dep)
			if !dep.Optional {
				result.Errors = append(result.Errors, 
					fmt.Sprintf("Required dependency '%s' not found", dep.Name))
			}
		}
	}
	
	// Validate build dependencies
	for _, dep := range dm.Build {
		if !dm.isDependencyAvailable(dep) && !dep.Optional {
			result.Valid = false
			result.Missing = append(result.Missing, dep)
			result.Errors = append(result.Errors, 
				fmt.Sprintf("Required build dependency '%s' not found", dep.Name))
		}
	}
	
	return result
}

// GetDependenciesForFormat returns dependencies for a specific package format
func (dm *DependencyManager) GetDependenciesForFormat(format string) PackageDeps {
	if deps, exists := dm.PackageSpecific[format]; exists {
		return deps
	}
	return PackageDeps{}
}

// isDependencyAvailable checks if a dependency is available on the system
func (dm *DependencyManager) isDependencyAvailable(dep Dependency) bool {
	// TODO: Implement actual dependency checking logic
	// This would check system package managers, command availability, etc.
	return true // Placeholder
}

// ParseVersionConstraint parses version constraint strings like ">=1.0.0", "~1.2.0"
func ParseVersionConstraint(constraint string) (*VersionConstraint, error) {
	constraint = strings.TrimSpace(constraint)
	
	if constraint == "" {
		return &VersionConstraint{Operator: "any"}, nil
	}
	
	operators := []string{">=", "<=", "==", "!=", "~", "^", ">", "<"}
	
	for _, op := range operators {
		if strings.HasPrefix(constraint, op) {
			version := strings.TrimSpace(constraint[len(op):])
			return &VersionConstraint{
				Operator: op,
				Version:  version,
			}, nil
		}
	}
	
	// Exact version match
	return &VersionConstraint{
		Operator: "==",
		Version:  constraint,
	}, nil
}

// VersionConstraint represents a version constraint
type VersionConstraint struct {
	Operator string // >=, <=, ==, !=, ~, ^, >, <, any
	Version  string
}

// Satisfies checks if a version satisfies the constraint
func (vc *VersionConstraint) Satisfies(version string) bool {
	// TODO: Implement semantic version comparison
	// This would use a proper semver library
	return true // Placeholder
}

// Example usage in bagboy.yaml:
/*
dependencies:
  runtime:
    - name: "libc"
      version: ">=2.17"
      platforms: ["linux"]
      optional: false
    - name: "vcredist"
      version: ">=14.0"
      platforms: ["windows"]
      optional: false
  
  build:
    - name: "gcc"
      version: ">=7.0"
      optional: true
    - name: "docker"
      version: ">=20.0"
      optional: true

  package_specific:
    deb:
      depends: ["libc6 (>= 2.17)", "libssl3"]
      recommends: ["curl"]
    rpm:
      requires: ["glibc >= 2.17", "openssl-libs"]
    brew:
      depends_on: ["openssl@3"]
*/
