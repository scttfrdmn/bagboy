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

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Name:     "test",
				Version:  "1.0.0",
				Binaries: map[string]string{"linux-amd64": "test"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: &Config{
				Version:  "1.0.0",
				Binaries: map[string]string{"linux-amd64": "test"},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			config: &Config{
				Name:     "test",
				Binaries: map[string]string{"linux-amd64": "test"},
			},
			wantErr: true,
		},
		{
			name: "missing binaries",
			config: &Config{
				Name:    "test",
				Version: "1.0.0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	configContent := `name: test
version: 1.0.0
description: Test app
binaries:
  linux-amd64: test-binary`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test successful load
	cfg, err := Load(configPath)
	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}

	if cfg.Name != "test" {
		t.Errorf("Expected name 'test', got %s", cfg.Name)
	}

	if cfg.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", cfg.Version)
	}

	// Test load non-existent file
	_, err = Load("non-existent.yaml")
	if err == nil {
		t.Error("Expected error loading non-existent file")
	}

	// Test load invalid YAML
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	err = os.WriteFile(invalidPath, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}

	_, err = Load(invalidPath)
	if err == nil {
		t.Error("Expected error loading invalid YAML")
	}
}

func TestFindConfigFile(t *testing.T) {
	// Test in directory with no config
	originalDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	_, err := FindConfigFile()
	if err == nil {
		t.Error("Expected error when no config file found")
	}

	// Test with bagboy.yaml
	err = os.WriteFile("bagboy.yaml", []byte("name: test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	path, err := FindConfigFile()
	if err != nil {
		t.Errorf("FindConfigFile() failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}

	// Test with bagboy.yml
	os.Remove("bagboy.yaml")
	err = os.WriteFile("bagboy.yml", []byte("name: test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err = FindConfigFile()
	if err != nil {
		t.Errorf("FindConfigFile() failed with .yml: %v", err)
	}
}
