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

package cicd_test

import (
	"strings"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/cicd"
	_ "github.com/scttfrdmn/bagboy/pkg/cicd/github"
	_ "github.com/scttfrdmn/bagboy/pkg/cicd/gitlab"
	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestGet_GitHub(t *testing.T) {
	p, err := cicd.Get("github")
	if err != nil {
		t.Fatalf("Get(github): %v", err)
	}
	if p.Name() != "github" {
		t.Errorf("expected 'github', got %q", p.Name())
	}
}

func TestGet_GitLab(t *testing.T) {
	p, err := cicd.Get("gitlab")
	if err != nil {
		t.Fatalf("Get(gitlab): %v", err)
	}
	if p.Name() != "gitlab" {
		t.Errorf("expected 'gitlab', got %q", p.Name())
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := cicd.Get("circleci")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestList(t *testing.T) {
	names := cicd.List()
	if len(names) < 2 {
		t.Fatalf("expected at least 2 providers, got %d", len(names))
	}
}

func TestGitHubProvider_GeneratePipeline(t *testing.T) {
	p, err := cicd.Get("github")
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:    "myapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner: "acme",
			Repo:  "myapp",
		},
	}

	files, err := p.GeneratePipeline(cfg)
	if err != nil {
		t.Fatalf("GeneratePipeline: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one file")
	}

	f := files[0]
	if !strings.HasSuffix(f.Path, ".yml") && !strings.HasSuffix(f.Path, ".yaml") {
		t.Errorf("expected YAML file path, got %q", f.Path)
	}
	if !strings.Contains(f.Content, "bagboy") {
		t.Errorf("workflow should mention bagboy; got:\n%s", f.Content)
	}
	if !strings.Contains(f.Content, "push") {
		t.Errorf("workflow should trigger on push; got:\n%s", f.Content)
	}
}

func TestGitLabProvider_GeneratePipeline(t *testing.T) {
	p, err := cicd.Get("gitlab")
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:    "myapp",
		Version: "1.0.0",
	}

	files, err := p.GeneratePipeline(cfg)
	if err != nil {
		t.Fatalf("GeneratePipeline: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one file")
	}

	f := files[0]
	if f.Path != ".gitlab-ci.yml" {
		t.Errorf("expected .gitlab-ci.yml, got %q", f.Path)
	}
	if !strings.Contains(f.Content, "stages") {
		t.Errorf("GitLab CI file should define stages; got:\n%s", f.Content)
	}
	if !strings.Contains(f.Content, "bagboy") {
		t.Errorf("pipeline should mention bagboy; got:\n%s", f.Content)
	}
}

func TestGitHubProvider_NilConfig(t *testing.T) {
	p, err := cicd.Get("github")
	if err != nil {
		t.Fatal(err)
	}
	// nil config should not panic.
	files, err := p.GeneratePipeline(nil)
	if err != nil {
		t.Fatalf("GeneratePipeline(nil): %v", err)
	}
	if len(files) == 0 {
		t.Error("expected files even with nil config")
	}
}
