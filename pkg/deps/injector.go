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
	"runtime"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// Injector handles dependency injection into package formats
type Injector struct {
	config *config.Config
}

// NewInjector creates a new dependency injector
func NewInjector(cfg *config.Config) *Injector {
	return &Injector{config: cfg}
}

// InjectDEBDependencies adds dependencies to DEB package configuration
func (i *Injector) InjectDEBDependencies() []string {
	var deps []string
	
	// Add system dependencies for Linux
	if linuxDeps, ok := i.config.Dependencies.System["linux"]; ok {
		deps = append(deps, linuxDeps...)
	}
	
	// Add APT package manager dependencies
	if aptDeps, ok := i.config.Dependencies.PackageManagers["apt"]; ok {
		deps = append(deps, aptDeps...)
	}
	
	return deps
}

// InjectRPMDependencies adds dependencies to RPM package configuration
func (i *Injector) InjectRPMDependencies() []string {
	var deps []string
	
	// Add system dependencies for Linux
	if linuxDeps, ok := i.config.Dependencies.System["linux"]; ok {
		deps = append(deps, linuxDeps...)
	}
	
	// Add YUM/DNF package manager dependencies
	if yumDeps, ok := i.config.Dependencies.PackageManagers["yum"]; ok {
		deps = append(deps, yumDeps...)
	}
	if dnfDeps, ok := i.config.Dependencies.PackageManagers["dnf"]; ok {
		deps = append(deps, dnfDeps...)
	}
	
	return deps
}

// InjectBrewDependencies adds dependencies to Homebrew formula
func (i *Injector) InjectBrewDependencies() []string {
	var deps []string
	
	// Add system dependencies for macOS
	if macDeps, ok := i.config.Dependencies.System["darwin"]; ok {
		deps = append(deps, macDeps...)
	}
	
	// Add Homebrew package manager dependencies
	if brewDeps, ok := i.config.Dependencies.PackageManagers["homebrew"]; ok {
		deps = append(deps, brewDeps...)
	}
	
	return deps
}

// InjectDockerDependencies adds dependencies to Docker image
func (i *Injector) InjectDockerDependencies() map[string][]string {
	result := make(map[string][]string)
	
	// Add system dependencies for Linux (Docker containers are Linux-based)
	if linuxDeps, ok := i.config.Dependencies.System["linux"]; ok {
		result["system"] = linuxDeps
	}
	
	// Add APT dependencies (most Docker images use Debian/Ubuntu)
	if aptDeps, ok := i.config.Dependencies.PackageManagers["apt"]; ok {
		result["apt"] = aptDeps
	}
	
	return result
}

// GetPlatformDependencies returns dependencies for the current platform
func (i *Injector) GetPlatformDependencies() []string {
	var deps []string
	
	// Add system dependencies for current platform
	if sysDeps, ok := i.config.Dependencies.System[runtime.GOOS]; ok {
		deps = append(deps, sysDeps...)
	}
	
	return deps
}

// GetRuntimeDependencies returns runtime dependencies with versions
func (i *Injector) GetRuntimeDependencies() map[string]string {
	return i.config.Dependencies.Runtime
}
