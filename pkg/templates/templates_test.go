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

package templates

import (
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	templates := List()
	if len(templates) == 0 {
		t.Fatal("expected at least one template")
	}
	names := map[string]bool{}
	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("template has empty name")
		}
		if tmpl.Description == "" {
			t.Errorf("template %q has empty description", tmpl.Name)
		}
		if tmpl.Content() == "" {
			t.Errorf("template %q has empty content", tmpl.Name)
		}
		if names[tmpl.Name] {
			t.Errorf("duplicate template name: %q", tmpl.Name)
		}
		names[tmpl.Name] = true
	}
}

func TestGet_KnownTemplates(t *testing.T) {
	known := []string{"go-cli", "go-service", "node-cli", "rust-cli"}
	for _, name := range known {
		t.Run(name, func(t *testing.T) {
			tmpl, err := Get(name)
			if err != nil {
				t.Fatalf("Get(%q): %v", name, err)
			}
			if tmpl.Name != name {
				t.Errorf("expected name %q, got %q", name, tmpl.Name)
			}
			if tmpl.Content() == "" {
				t.Errorf("template %q has empty content", name)
			}
		})
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("nonexistent-template")
	if err == nil {
		t.Error("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "nonexistent-template") {
		t.Errorf("error should mention the template name: %v", err)
	}
}

func TestGet_CaseInsensitive(t *testing.T) {
	_, err := Get("GO-CLI")
	if err != nil {
		t.Fatalf("Get should be case-insensitive: %v", err)
	}
}

func TestTemplate_Render(t *testing.T) {
	tmpl, err := Get("go-cli")
	if err != nil {
		t.Fatal(err)
	}

	data := TemplateData{
		Name:        "mytool",
		Version:     "1.2.3",
		Description: "My awesome tool",
		Author:      "Jane Doe",
		Homepage:    "https://example.com",
		License:     "MIT",
		GitHubOwner: "janedoe",
		GitHubRepo:  "mytool",
	}

	rendered := tmpl.Render(data)

	checks := []string{
		`name: "mytool"`,
		`version: "1.2.3"`,
		`description: "My awesome tool"`,
		`author: "Jane Doe"`,
		`owner: "janedoe"`,
		`repo: "mytool"`,
	}
	for _, want := range checks {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered output missing %q\nfull output:\n%s", want, rendered)
		}
	}

	// Ensure no unreplaced placeholders remain.
	if strings.Contains(rendered, "{{.Name}}") {
		t.Error("rendered output still contains {{.Name}} placeholder")
	}
}

func TestTemplate_ContentUnchanged(t *testing.T) {
	tmpl, err := Get("go-cli")
	if err != nil {
		t.Fatal(err)
	}
	// Content() should return the raw template.
	if !strings.Contains(tmpl.Content(), "{{.Name}}") {
		t.Error("Content() should return raw template with placeholders")
	}
}

func TestAllTemplatesHaveValidYAMLStructure(t *testing.T) {
	for _, tmpl := range List() {
		t.Run(tmpl.Name, func(t *testing.T) {
			content := tmpl.Content()
			// Basic structural checks.
			if !strings.Contains(content, "name:") {
				t.Errorf("template %q missing 'name:' field", tmpl.Name)
			}
			if !strings.Contains(content, "version:") {
				t.Errorf("template %q missing 'version:' field", tmpl.Name)
			}
			if !strings.Contains(content, "binaries:") {
				t.Errorf("template %q missing 'binaries:' field", tmpl.Name)
			}
		})
	}
}
