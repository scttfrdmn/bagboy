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
	"testing"
	"time"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestManager(t *testing.T) {
	cfg := &config.Config{
		Dependencies: config.DependenciesConfig{
			System: map[string][]string{
				"linux": {"curl", "git"},
				"darwin": {"curl", "git"},
			},
			PackageManagers: map[string][]string{
				"homebrew": {"openssl", "curl"},
				"apt":      {"libssl-dev", "libcurl4-openssl-dev"},
			},
			Runtime: map[string]string{
				"node": ">=18.0.0",
				"go": ">=1.19",
			},
		},
	}

	manager := NewManager(cfg)

	t.Run("List dependencies", func(t *testing.T) {
		deps := manager.List()
		if len(deps) == 0 {
			t.Error("Expected dependencies to be listed")
		}

		// Check that we have system, package manager, and runtime deps
		hasSystem := false
		hasPackageManager := false
		hasRuntime := false

		for _, dep := range deps {
			switch dep.Type {
			case "system":
				hasSystem = true
			case "package_manager":
				hasPackageManager = true
			case "runtime":
				hasRuntime = true
			}
		}

		if !hasSystem {
			t.Error("Expected system dependencies")
		}
		if !hasPackageManager {
			t.Error("Expected package manager dependencies")
		}
		if !hasRuntime {
			t.Error("Expected runtime dependencies")
		}
	})

	t.Run("Check dependencies", func(t *testing.T) {
		ctx := context.Background()
		results, err := manager.Check(ctx)
		if err != nil {
			t.Fatalf("Check failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected dependency check results")
		}

		// Verify results have proper structure
		for name, status := range results {
			if name == "" {
				t.Error("Dependency name should not be empty")
			}
			// Status can be available or not, both are valid
			_ = status
		}
	})

	t.Run("Resolve dependencies", func(t *testing.T) {
		ctx := context.Background()
		result, err := manager.Resolve(ctx)
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}

		if result == nil {
			t.Error("Expected resolution result")
		}

		if result.Resolved == nil {
			t.Error("Expected resolved dependencies map")
		}

		if result.Conflicts == nil {
			t.Error("Expected conflicts slice (even if empty)")
		}
	})
}

func TestDependencyDetection(t *testing.T) {
	cfg := &config.Config{
		Dependencies: config.DependenciesConfig{},
	}

	manager := NewManager(cfg)

	t.Run("Detect package manager", func(t *testing.T) {
		pm := manager.detectPackageManager()
		// Should return a string (even if "unknown")
		if pm == "" {
			t.Error("Package manager detection should return a value")
		}
	})

	t.Run("Command exists check", func(t *testing.T) {
		// Test with a command that should exist on most systems
		exists := manager.commandExists("echo")
		if !exists {
			t.Error("echo command should exist on most systems")
		}

		// Test with a command that likely doesn't exist
		exists = manager.commandExists("nonexistentcommand12345")
		if exists {
			t.Error("nonexistent command should not be found")
		}
	})
}

func TestDependencyResolution(t *testing.T) {
	cfg := &config.Config{
		Dependencies: config.DependenciesConfig{
			System: map[string][]string{
				"linux": {"curl", "git"},
			},
			PackageManagers: map[string][]string{
				"apt": {"libssl-dev"},
			},
			Runtime: map[string]string{
				"node": ">=18.0.0",
				"go":   ">=1.19",
			},
		},
	}

	manager := NewManager(cfg)
	ctx := context.Background()

	t.Run("Resolve dependencies", func(t *testing.T) {
		result, err := manager.Resolve(ctx)
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected resolution result")
		}

		if len(result.Resolved) == 0 {
			t.Error("Expected resolved dependencies")
		}

		// Check that runtime dependencies are resolved with constraints
		if nodeVersion, ok := result.Resolved["node"]; !ok || nodeVersion != ">=18.0.0" {
			t.Errorf("Expected node >=18.0.0, got %s", nodeVersion)
		}
	})

	t.Run("Generate lock file", func(t *testing.T) {
		lockFile, err := manager.GenerateLockFile(ctx)
		if err != nil {
			t.Fatalf("GenerateLockFile failed: %v", err)
		}

		if lockFile == nil {
			t.Fatal("Expected lock file")
		}

		if lockFile.Version == "" {
			t.Error("Expected lock file version")
		}

		if len(lockFile.Dependencies) == 0 {
			t.Error("Expected lock file dependencies")
		}
	})
}

func TestDependencyCache(t *testing.T) {
	cache := NewCache()

	t.Run("Cache operations", func(t *testing.T) {
		status := DependencyStatus{
			Available: true,
			Version:   "1.0.0",
		}

		// Set cache entry
		err := cache.Set("test-dep", status, time.Minute)
		if err != nil {
			t.Fatalf("Cache set failed: %v", err)
		}

		// Get cache entry
		cached, found := cache.Get("test-dep")
		if !found {
			t.Error("Expected cached entry to be found")
		}

		if cached.Available != status.Available {
			t.Error("Cached status doesn't match")
		}

		if cached.Version != status.Version {
			t.Error("Cached version doesn't match")
		}
	})

	t.Run("Cache expiration", func(t *testing.T) {
		status := DependencyStatus{Available: true}

		// Set with very short TTL
		err := cache.Set("expire-test", status, time.Nanosecond)
		if err != nil {
			t.Fatalf("Cache set failed: %v", err)
		}

		// Wait for expiration
		time.Sleep(time.Millisecond)

		// Should not find expired entry
		_, found := cache.Get("expire-test")
		if found {
			t.Error("Expected expired entry to not be found")
		}
	})
}

func TestDependencyInjection(t *testing.T) {
	cfg := &config.Config{
		Dependencies: config.DependenciesConfig{
			System: map[string][]string{
				"linux":  {"curl", "git"},
				"darwin": {"curl", "git"},
			},
			PackageManagers: map[string][]string{
				"apt":      {"libssl-dev"},
				"homebrew": {"openssl"},
			},
			Runtime: map[string]string{
				"node": ">=18.0.0",
			},
		},
	}

	injector := NewInjector(cfg)

	t.Run("DEB dependency injection", func(t *testing.T) {
		deps := injector.InjectDEBDependencies()
		if len(deps) == 0 {
			t.Error("Expected DEB dependencies")
		}

		// Should include both system and apt dependencies
		hasSystemDep := false
		hasAptDep := false
		for _, dep := range deps {
			if dep == "curl" || dep == "git" {
				hasSystemDep = true
			}
			if dep == "libssl-dev" {
				hasAptDep = true
			}
		}

		if !hasSystemDep {
			t.Error("Expected system dependencies in DEB injection")
		}
		if !hasAptDep {
			t.Error("Expected APT dependencies in DEB injection")
		}
	})

	t.Run("Homebrew dependency injection", func(t *testing.T) {
		deps := injector.InjectBrewDependencies()
		if len(deps) == 0 {
			t.Error("Expected Homebrew dependencies")
		}

		// Should include both system and homebrew dependencies
		hasSystemDep := false
		hasBrewDep := false
		for _, dep := range deps {
			if dep == "curl" || dep == "git" {
				hasSystemDep = true
			}
			if dep == "openssl" {
				hasBrewDep = true
			}
		}

		if !hasSystemDep {
			t.Error("Expected system dependencies in Homebrew injection")
		}
		if !hasBrewDep {
			t.Error("Expected Homebrew dependencies in injection")
		}
	})

	t.Run("Runtime dependencies", func(t *testing.T) {
		runtime := injector.GetRuntimeDependencies()
		if len(runtime) == 0 {
			t.Error("Expected runtime dependencies")
		}

		if nodeVersion, ok := runtime["node"]; !ok || nodeVersion != ">=18.0.0" {
			t.Error("Expected node runtime dependency")
		}
	})
}
