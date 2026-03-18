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

package vm

import (
	"context"
	"errors"
	"testing"
)

// mockProvider is a test double for the Provider interface.
type mockProvider struct {
	available   bool
	setupErr    error
	buildOutput string
	buildErr    error
	cleanupErr  error
	setupCalled bool
	buildCalled bool
	cleanupCalled bool
}

func (m *mockProvider) IsAvailable() bool { return m.available }

func (m *mockProvider) Setup(_ context.Context) error {
	m.setupCalled = true
	return m.setupErr
}

func (m *mockProvider) Build(_ context.Context, _ string) (string, error) {
	m.buildCalled = true
	return m.buildOutput, m.buildErr
}

func (m *mockProvider) CopyArtifact(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockProvider) Cleanup(_ context.Context) error {
	m.cleanupCalled = true
	return m.cleanupErr
}

// --- Manager tests ---

func TestManagerIsAvailable_WithProvider(t *testing.T) {
	mgr := &Manager{
		config:   &Config{Enabled: true, Provider: "mock"},
		provider: &mockProvider{available: true},
	}
	if !mgr.IsAvailable() {
		t.Error("expected IsAvailable() == true")
	}
}

func TestManagerIsAvailable_NilProvider(t *testing.T) {
	mgr := &Manager{config: &Config{}}
	if mgr.IsAvailable() {
		t.Error("expected IsAvailable() == false with nil provider")
	}
}

func TestManagerIsAvailable_UnavailableProvider(t *testing.T) {
	mgr := &Manager{
		config:   &Config{},
		provider: &mockProvider{available: false},
	}
	if mgr.IsAvailable() {
		t.Error("expected IsAvailable() == false when provider is not available")
	}
}

func TestManagerBuildInVM_NotEnabled(t *testing.T) {
	mgr := &Manager{
		config:   &Config{Enabled: false},
		provider: &mockProvider{available: true},
	}
	_, err := mgr.BuildInVM(context.Background(), "ubuntu:22.04", ".", "uname -s")
	if err == nil {
		t.Fatal("expected error when VM support not enabled")
	}
}

func TestManagerBuildInVM_NotAvailable(t *testing.T) {
	mgr := &Manager{
		config:   &Config{Enabled: true},
		provider: &mockProvider{available: false},
	}
	_, err := mgr.BuildInVM(context.Background(), "ubuntu:22.04", ".", "uname -s")
	if err == nil {
		t.Fatal("expected error when provider not available")
	}
}

func TestManagerBuildInVM_SetupError(t *testing.T) {
	mock := &mockProvider{available: true, setupErr: errors.New("setup failed")}
	mgr := &Manager{
		config:   &Config{Enabled: true},
		provider: mock,
	}
	_, err := mgr.BuildInVM(context.Background(), "", ".", "echo hi")
	if err == nil {
		t.Fatal("expected error from setup failure")
	}
	if !mock.setupCalled {
		t.Error("expected Setup() to have been called")
	}
}

func TestManagerBuildInVM_BuildError(t *testing.T) {
	mock := &mockProvider{available: true, buildErr: errors.New("build failed")}
	mgr := &Manager{
		config:   &Config{Enabled: true},
		provider: mock,
	}
	_, err := mgr.BuildInVM(context.Background(), "", ".", "false")
	if err == nil {
		t.Fatal("expected error from build failure")
	}
	if !mock.buildCalled {
		t.Error("expected Build() to have been called")
	}
	if !mock.cleanupCalled {
		t.Error("expected Cleanup() to have been called via defer")
	}
}

func TestManagerBuildInVM_Success(t *testing.T) {
	mock := &mockProvider{available: true, buildOutput: "Linux"}
	mgr := &Manager{
		config:   &Config{Enabled: true},
		provider: mock,
	}
	out, err := mgr.BuildInVM(context.Background(), "", ".", "uname -s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Linux" {
		t.Errorf("expected 'Linux', got %q", out)
	}
	if !mock.setupCalled {
		t.Error("expected Setup() to be called")
	}
	if !mock.buildCalled {
		t.Error("expected Build() to be called")
	}
	if !mock.cleanupCalled {
		t.Error("expected Cleanup() to be called")
	}
}

// --- NewManager tests ---

func TestNewManager_DefaultsToDocker(t *testing.T) {
	mgr := NewManager(nil)
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
	if mgr.config.Provider != "docker" {
		t.Errorf("expected default provider 'docker', got %q", mgr.config.Provider)
	}
}

func TestNewManager_DockerProvider(t *testing.T) {
	mgr := NewManager(&Config{
		Enabled:  true,
		Provider: "docker",
		Docker:   DockerConfig{Image: "ubuntu:22.04", WorkDir: "."},
	})
	if mgr.provider == nil {
		t.Error("expected docker provider to be set")
	}
}

func TestNewManager_UnknownProvider(t *testing.T) {
	mgr := NewManager(&Config{
		Enabled:  true,
		Provider: "nonexistent",
	})
	if mgr.provider != nil {
		t.Error("expected nil provider for unknown provider name")
	}
}

// --- DockerConfig validation ---

func TestDockerConfig_DefaultImage(t *testing.T) {
	cfg := &DockerConfig{WorkDir: "."}
	provider := NewDockerProvider(cfg)
	if provider.config.Image != "ubuntu:22.04" {
		t.Errorf("expected default image 'ubuntu:22.04', got %q", provider.config.Image)
	}
}

func TestDockerConfig_CustomImage(t *testing.T) {
	cfg := &DockerConfig{Image: "debian:bookworm", WorkDir: "/build"}
	provider := NewDockerProvider(cfg)
	if provider.config.Image != "debian:bookworm" {
		t.Errorf("expected 'debian:bookworm', got %q", provider.config.Image)
	}
}

// --- Docker integration tests (require Docker) ---

func TestDockerProvider_Integration(t *testing.T) {
	provider := NewDockerProvider(&DockerConfig{
		Image:   "ubuntu:22.04",
		WorkDir: ".",
	})

	if !provider.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	output, err := provider.Build(ctx, "echo 'hello from docker'")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if output != "hello from docker" {
		t.Errorf("expected 'hello from docker', got %q", output)
	}
}

func TestManager_Integration(t *testing.T) {
	mgr := NewManager(&Config{
		Enabled:  true,
		Provider: "docker",
		Docker: DockerConfig{
			Image:   "ubuntu:22.04",
			WorkDir: ".",
		},
	})

	if !mgr.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	output, err := mgr.BuildInVM(ctx, "ubuntu:22.04", ".", "uname -s")
	if err != nil {
		t.Fatalf("BuildInVM failed: %v", err)
	}
	if output != "Linux" {
		t.Errorf("expected 'Linux', got %q", output)
	}
}

func TestDockerProvider_IsAvailableNoDocker(t *testing.T) {
	provider := NewDockerProvider(&DockerConfig{})
	// Ensure it doesn't panic regardless of docker availability.
	_ = provider.IsAvailable()
}
