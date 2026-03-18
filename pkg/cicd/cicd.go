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

// Package cicd provides CI/CD pipeline generation for bagboy projects.
// It supports generating GitHub Actions workflows and GitLab CI pipelines
// tailored to the detected project type.
package cicd

import (
	"fmt"
	"strings"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// File represents a generated pipeline file with its path and content.
type File struct {
	// Path is the relative path where the file should be written (e.g. ".github/workflows/release.yml").
	Path string
	// Content is the full text content of the file.
	Content string
}

// Provider generates CI/CD pipeline configuration files for a specific platform.
type Provider interface {
	// Name returns the unique lower-case provider name (e.g. "github", "gitlab").
	Name() string
	// GeneratePipeline produces the pipeline files for the given config.
	GeneratePipeline(cfg *config.Config) ([]File, error)
}

// registry holds all registered CI/CD providers.
var registry = map[string]Provider{}

// Register adds a provider to the global registry.
func Register(p Provider) {
	registry[strings.ToLower(p.Name())] = p
}

// Get returns the provider registered under name, or an error if not found.
func Get(name string) (Provider, error) {
	p, ok := registry[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown CI/CD provider %q; supported: github, gitlab", name)
	}
	return p, nil
}

// List returns the names of all registered providers.
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
