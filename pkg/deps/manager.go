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
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/scttfrdmn/bagboy/pkg/config"
)

// Manager handles dependency resolution and installation
type Manager struct {
	config *config.Config
	cache  *Cache
}

// NewManager creates a new dependency manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
		cache:  NewCache(),
	}
}

// Check verifies all dependencies are available
func (m *Manager) Check(ctx context.Context) (map[string]DependencyStatus, error) {
	results := make(map[string]DependencyStatus)
	
	// Check system dependencies
	for platform, deps := range m.config.Dependencies.System {
		if platform == runtime.GOOS {
			for _, dep := range deps {
				status := m.checkSystemDependency(dep)
				results[dep] = status
			}
		}
	}
	
	// Check package manager dependencies
	pm := m.detectPackageManager()
	if deps, ok := m.config.Dependencies.PackageManagers[pm]; ok {
		for _, dep := range deps {
			status := m.checkPackageManagerDependency(pm, dep)
			results[dep] = status
		}
	}
	
	// Check runtime dependencies
	for runtime, version := range m.config.Dependencies.Runtime {
		status := m.checkRuntimeDependency(runtime, version)
		results[runtime] = status
	}
	
	return results, nil
}

// Install installs missing dependencies
func (m *Manager) Install(ctx context.Context, deps []string) error {
	pm := m.detectPackageManager()
	
	for _, dep := range deps {
		if err := m.installDependency(pm, dep); err != nil {
			return fmt.Errorf("failed to install %s: %w", dep, err)
		}
	}
	
	return nil
}

// List returns all configured dependencies
func (m *Manager) List() []Dependency {
	var deps []Dependency
	
	// System dependencies
	for platform, sysDeps := range m.config.Dependencies.System {
		for _, dep := range sysDeps {
			deps = append(deps, Dependency{
				Name:     dep,
				Type:     "system",
				Platform: platform,
			})
		}
	}
	
	// Package manager dependencies
	for pm, pmDeps := range m.config.Dependencies.PackageManagers {
		for _, dep := range pmDeps {
			deps = append(deps, Dependency{
				Name:           dep,
				Type:           "package_manager",
				PackageManager: pm,
			})
		}
	}
	
	// Runtime dependencies
	for runtime, version := range m.config.Dependencies.Runtime {
		deps = append(deps, Dependency{
			Name:    runtime,
			Type:    "runtime",
			Version: version,
		})
	}
	
	return deps
}

// Resolve handles dependency conflicts and version constraints
func (m *Manager) Resolve(ctx context.Context) (*ResolutionResult, error) {
	result := &ResolutionResult{
		Resolved:  make(map[string]string),
		Conflicts: make([]Conflict, 0),
	}
	
	// Collect all dependencies with their constraints
	depMap := make(map[string][]string)
	
	// System dependencies (current platform only)
	platform := runtime.GOOS
	if deps, ok := m.config.Dependencies.System[platform]; ok {
		for _, dep := range deps {
			depMap[dep] = append(depMap[dep], "system")
		}
	}
	
	// Package manager dependencies (current platform)
	pm := m.detectPackageManager()
	if deps, ok := m.config.Dependencies.PackageManagers[pm]; ok {
		for _, dep := range deps {
			depMap[dep] = append(depMap[dep], pm)
		}
	}
	
	// Runtime dependencies with version constraints
	for runtime, version := range m.config.Dependencies.Runtime {
		if existing, exists := depMap[runtime]; exists {
			// Check for version conflicts
			for _, existingSource := range existing {
				if existingSource != version {
					result.Conflicts = append(result.Conflicts, Conflict{
						Dependency: runtime,
						Versions:   []string{existingSource, version},
						Reason:     "Version constraint mismatch",
					})
				}
			}
		}
		depMap[runtime] = append(depMap[runtime], version)
		result.Resolved[runtime] = version
	}
	
	// Resolve system and package manager dependencies
	for dep, sources := range depMap {
		if len(sources) > 1 {
			// Multiple sources - check for conflicts
			unique := make(map[string]bool)
			for _, source := range sources {
				unique[source] = true
			}
			if len(unique) > 1 {
				var versions []string
				for version := range unique {
					versions = append(versions, version)
				}
				result.Conflicts = append(result.Conflicts, Conflict{
					Dependency: dep,
					Versions:   versions,
					Reason:     "Multiple sources with different requirements",
				})
			}
		}
		
		// Use the first source as resolved version
		if len(sources) > 0 {
			result.Resolved[dep] = sources[0]
		}
	}
	
	return result, nil
}

// GenerateLockFile creates a lock file with resolved dependencies
func (m *Manager) GenerateLockFile(ctx context.Context) (*LockFile, error) {
	resolution, err := m.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	
	lockFile := &LockFile{
		Version:      "1.0",
		Generated:    time.Now().UTC(),
		Dependencies: make(map[string]LockEntry),
	}
	
	// Check current versions of resolved dependencies
	for dep, constraint := range resolution.Resolved {
		status := m.checkAnyDependency(dep)
		
		entry := LockEntry{
			Version:    status.Version,
			Constraint: constraint,
			Source:     m.getDepSource(dep),
			Resolved:   time.Now().UTC(),
		}
		
		if !status.Available {
			entry.Version = "not-installed"
		}
		
		lockFile.Dependencies[dep] = entry
	}
	
	return lockFile, nil
}

// WriteLockFile writes the lock file to disk
func (m *Manager) WriteLockFile(ctx context.Context, path string) error {
	lockFile, err := m.GenerateLockFile(ctx)
	if err != nil {
		return err
	}
	
	data, err := yaml.Marshal(lockFile)
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

func (m *Manager) checkAnyDependency(dep string) DependencyStatus {
	// Try system dependency first
	if m.commandExists(dep) {
		return DependencyStatus{Available: true, Version: "system"}
	}
	
	// Try package manager
	pm := m.detectPackageManager()
	status := m.checkPackageManagerDependency(pm, dep)
	if status.Available {
		return status
	}
	
	// Try runtime dependency
	if version, ok := m.config.Dependencies.Runtime[dep]; ok {
		return m.checkRuntimeDependency(dep, version)
	}
	
	return DependencyStatus{Available: false}
}

func (m *Manager) getDepSource(dep string) string {
	// Check which source this dependency comes from
	platform := runtime.GOOS
	if deps, ok := m.config.Dependencies.System[platform]; ok {
		for _, sysDep := range deps {
			if sysDep == dep {
				return "system"
			}
		}
	}
	
	pm := m.detectPackageManager()
	if deps, ok := m.config.Dependencies.PackageManagers[pm]; ok {
		for _, pmDep := range deps {
			if pmDep == dep {
				return pm
			}
		}
	}
	
	if _, ok := m.config.Dependencies.Runtime[dep]; ok {
		return "runtime"
	}
	
	return "unknown"
}

func (m *Manager) detectPackageManager() string {
	switch runtime.GOOS {
	case "darwin":
		if m.commandExists("brew") {
			return "homebrew"
		}
	case "linux":
		if m.commandExists("apt-get") {
			return "apt"
		}
		if m.commandExists("yum") {
			return "yum"
		}
		if m.commandExists("dnf") {
			return "dnf"
		}
	case "windows":
		if m.commandExists("choco") {
			return "chocolatey"
		}
		if m.commandExists("scoop") {
			return "scoop"
		}
	}
	return "unknown"
}

func (m *Manager) commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (m *Manager) checkSystemDependency(dep string) DependencyStatus {
	// Check cache first
	cacheKey := fmt.Sprintf("system_%s_%s", runtime.GOOS, dep)
	if cached, found := m.cache.Get(cacheKey); found {
		return *cached
	}
	
	// Simple check - see if command exists
	status := DependencyStatus{Available: false}
	if m.commandExists(dep) {
		status = DependencyStatus{
			Available: true,
			Version:   "system",
		}
	}
	
	// Cache result for 5 minutes
	m.cache.Set(cacheKey, status, 5*time.Minute)
	return status
}

func (m *Manager) checkPackageManagerDependency(pm, dep string) DependencyStatus {
	switch pm {
	case "homebrew":
		return m.checkBrewPackage(dep)
	case "apt":
		return m.checkAptPackage(dep)
	default:
		return DependencyStatus{Available: false}
	}
}

func (m *Manager) checkBrewPackage(pkg string) DependencyStatus {
	cmd := exec.Command("brew", "list", pkg)
	if err := cmd.Run(); err == nil {
		return DependencyStatus{Available: true, Version: "installed"}
	}
	return DependencyStatus{Available: false}
}

func (m *Manager) checkAptPackage(pkg string) DependencyStatus {
	cmd := exec.Command("dpkg", "-l", pkg)
	if err := cmd.Run(); err == nil {
		return DependencyStatus{Available: true, Version: "installed"}
	}
	return DependencyStatus{Available: false}
}

func (m *Manager) checkRuntimeDependency(runtime, version string) DependencyStatus {
	switch runtime {
	case "node":
		return m.checkNodeVersion(version)
	case "python":
		return m.checkPythonVersion(version)
	case "go":
		return m.checkGoVersion(version)
	default:
		return DependencyStatus{Available: false}
	}
}

func (m *Manager) checkNodeVersion(constraint string) DependencyStatus {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return DependencyStatus{Available: false}
	}
	
	version := strings.TrimSpace(string(output))
	return DependencyStatus{
		Available: true,
		Version:   version,
		Satisfies: m.satisfiesConstraint(version, constraint),
	}
}

func (m *Manager) checkPythonVersion(constraint string) DependencyStatus {
	cmd := exec.Command("python3", "--version")
	output, err := cmd.Output()
	if err != nil {
		return DependencyStatus{Available: false}
	}
	
	version := strings.TrimSpace(string(output))
	return DependencyStatus{
		Available: true,
		Version:   version,
		Satisfies: m.satisfiesConstraint(version, constraint),
	}
}

func (m *Manager) checkGoVersion(constraint string) DependencyStatus {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return DependencyStatus{Available: false}
	}
	
	version := strings.TrimSpace(string(output))
	return DependencyStatus{
		Available: true,
		Version:   version,
		Satisfies: m.satisfiesConstraint(version, constraint),
	}
}

func (m *Manager) satisfiesConstraint(version, constraint string) bool {
	// Simple constraint checking - just check if version contains constraint
	return strings.Contains(version, strings.TrimPrefix(constraint, ">="))
}

func (m *Manager) installDependency(pm, dep string) error {
	switch pm {
	case "homebrew":
		return exec.Command("brew", "install", dep).Run()
	case "apt":
		return exec.Command("sudo", "apt-get", "install", "-y", dep).Run()
	case "chocolatey":
		return exec.Command("choco", "install", dep, "-y").Run()
	default:
		return fmt.Errorf("unsupported package manager: %s", pm)
	}
}

// Types for dependency management
type Dependency struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Platform       string `json:"platform,omitempty"`
	PackageManager string `json:"package_manager,omitempty"`
	Version        string `json:"version,omitempty"`
}

type DependencyStatus struct {
	Available bool   `json:"available"`
	Version   string `json:"version,omitempty"`
	Satisfies bool   `json:"satisfies,omitempty"`
	Error     string `json:"error,omitempty"`
}

type ResolutionResult struct {
	Resolved  map[string]string `json:"resolved"`
	Conflicts []Conflict        `json:"conflicts"`
}

type Conflict struct {
	Dependency string   `json:"dependency"`
	Versions   []string `json:"versions"`
	Reason     string   `json:"reason"`
}

// LockFile represents a dependency lock file
type LockFile struct {
	Version      string               `yaml:"version"`
	Generated    time.Time            `yaml:"generated"`
	Dependencies map[string]LockEntry `yaml:"dependencies"`
}

// LockEntry represents a locked dependency
type LockEntry struct {
	Version    string    `yaml:"version"`
	Constraint string    `yaml:"constraint,omitempty"`
	Source     string    `yaml:"source"`
	Resolved   time.Time `yaml:"resolved"`
}
