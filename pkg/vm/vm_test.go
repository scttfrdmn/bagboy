package vm

import (
	"context"
	"testing"
)

func TestDockerProvider(t *testing.T) {
	provider := NewDockerProvider(&DockerConfig{
		Image:   "ubuntu:22.04",
		WorkDir: ".",
	})

	if !provider.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()

	// Test simple command
	output, err := provider.Build(ctx, "echo 'hello from docker'")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if output != "hello from docker" {
		t.Errorf("Expected 'hello from docker', got %q", output)
	}
}

func TestManager(t *testing.T) {
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
		t.Errorf("Expected 'Linux', got %q", output)
	}
}

func TestDockerProviderNotAvailable(t *testing.T) {
	provider := NewDockerProvider(&DockerConfig{})
	
	// Just test that IsAvailable doesn't panic
	_ = provider.IsAvailable()
}
