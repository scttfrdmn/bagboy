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

// Package templates provides predefined bagboy.yaml configuration templates
// for common project types. Templates are embedded at compile time and can be
// applied via `bagboy init --template <name>`.
package templates

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed templates/go-cli.yaml
var goCLITemplate string

//go:embed templates/go-service.yaml
var goServiceTemplate string

//go:embed templates/node-cli.yaml
var nodeCLITemplate string

//go:embed templates/rust-cli.yaml
var rustCLITemplate string

// Template describes a predefined configuration template.
type Template struct {
	// Name is the unique identifier used with --template flag (e.g. "go-cli").
	Name string
	// Description is a human-readable one-line description.
	Description string
	// ProjectType is the language/ecosystem this template targets.
	ProjectType string
	// content is the raw YAML template string (Go text/template syntax).
	content string
}

// all is the registry of all built-in templates.
var all = []Template{
	{
		Name:        "go-cli",
		Description: "Go CLI tool: brew + scoop + deb + rpm + apptainer",
		ProjectType: "go",
		content:     goCLITemplate,
	},
	{
		Name:        "go-service",
		Description: "Go service/daemon: docker + deb + rpm",
		ProjectType: "go",
		content:     goServiceTemplate,
	},
	{
		Name:        "node-cli",
		Description: "Node.js CLI tool: npm + brew + scoop",
		ProjectType: "node",
		content:     nodeCLITemplate,
	},
	{
		Name:        "rust-cli",
		Description: "Rust CLI tool: cargo + brew + scoop + deb",
		ProjectType: "rust",
		content:     rustCLITemplate,
	},
}

// List returns all available templates.
func List() []Template {
	result := make([]Template, len(all))
	copy(result, all)
	return result
}

// Get returns the template with the given name, or an error if not found.
func Get(name string) (Template, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, t := range all {
		if t.Name == name {
			return t, nil
		}
	}
	return Template{}, fmt.Errorf("template %q not found; run 'bagboy init --list-templates' to see available templates", name)
}

// TemplateData holds variables available for substitution in a template.
type TemplateData struct {
	Name        string
	Version     string
	Description string
	Author      string
	Homepage    string
	License     string
	GitHubOwner string
	GitHubRepo  string
}

// Render returns the template content with all TemplateData fields substituted.
// It uses simple string replacement rather than text/template to keep the YAML
// output valid before any Go template directives are applied.
func (t Template) Render(data TemplateData) string {
	replacements := map[string]string{
		"{{.Name}}":        data.Name,
		"{{.Version}}":     data.Version,
		"{{.Description}}": data.Description,
		"{{.Author}}":      data.Author,
		"{{.Homepage}}":    data.Homepage,
		"{{.License}}":     data.License,
		"{{.GitHubOwner}}": data.GitHubOwner,
		"{{.GitHubRepo}}":  data.GitHubRepo,
	}

	result := t.content
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// Content returns the raw template YAML (before substitution).
func (t Template) Content() string {
	return t.content
}
