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

package packager

import (
	"context"
	"fmt"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// MockPackager for testing
type MockPackager struct {
	name      string
	shouldErr bool
}

func (m *MockPackager) Name() string {
	return m.name
}

func (m *MockPackager) Pack(ctx context.Context, cfg *config.Config) (string, error) {
	if m.shouldErr {
		return "", fmt.Errorf("mock error")
	}
	return "mock-output", nil
}

func (m *MockPackager) Validate(cfg *config.Config) error {
	if m.shouldErr {
		return fmt.Errorf("mock validation error")
	}
	return nil
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test registration
	mock := &MockPackager{name: "mock"}
	registry.Register(mock)

	// Test retrieval
	retrieved, ok := registry.Get("mock")
	if !ok {
		t.Error("Expected to find registered packager")
	}
	if retrieved.Name() != "mock" {
		t.Errorf("Expected name 'mock', got %s", retrieved.Name())
	}

	// Test list
	names := registry.List()
	if len(names) != 1 || names[0] != "mock" {
		t.Errorf("Expected ['mock'], got %v", names)
	}
}

func TestPackAll(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&MockPackager{name: "good", shouldErr: false})

	cfg := &config.Config{
		Name:     "test",
		Version:  "1.0.0",
		Binaries: map[string]string{"linux-amd64": "test"},
	}

	ctx := context.Background()
	results, err := registry.PackAll(ctx, cfg)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results["good"] != "mock-output" {
		t.Errorf("Expected 'mock-output', got %s", results["good"])
	}

	// Test with packager that fails validation (should be skipped)
	registry2 := NewRegistry()
	registry2.Register(&MockPackager{name: "bad-validation", shouldErr: true})

	results, err = registry2.PackAll(ctx, cfg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results (validation failed), got %d", len(results))
	}
}
