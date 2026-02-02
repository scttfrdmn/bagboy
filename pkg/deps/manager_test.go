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

func TestConstraintSatisfaction(t *testing.T) {
	cfg := &config.Config{}
	manager := NewManager(cfg)

	tests := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"v18.0.0", ">=18.0.0", true},
		{"v17.0.0", ">=18.0.0", false},
		{"go version go1.19.0", ">=1.19", true},
		{"Python 3.8.0", ">=3.8", true},
	}

	for _, test := range tests {
		result := manager.satisfiesConstraint(test.version, test.constraint)
		if result != test.expected {
			t.Errorf("satisfiesConstraint(%q, %q) = %v, expected %v",
				test.version, test.constraint, result, test.expected)
		}
	}
}
