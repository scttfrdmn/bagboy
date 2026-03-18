package vm

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Provider represents a VM/container provider
type Provider interface {
	Setup(ctx context.Context) error
	Build(ctx context.Context, cmd string) (string, error)
	CopyArtifact(ctx context.Context, src, dst string) error
	Cleanup(ctx context.Context) error
	IsAvailable() bool
}

// Config holds VM configuration
type Config struct {
	Enabled  bool
	Provider string
	Docker   DockerConfig
}

type DockerConfig struct {
	Image   string
	WorkDir string
}

// Manager manages VM providers
type Manager struct {
	config   *Config
	provider Provider
}

// NewManager creates a new VM manager
func NewManager(cfg *Config) *Manager {
	if cfg == nil {
		cfg = &Config{Provider: "docker"}
	}
	
	m := &Manager{config: cfg}
	
	if cfg.Provider == "docker" {
		m.provider = NewDockerProvider(&cfg.Docker)
	}
	
	return m
}

// IsAvailable checks if VM support is available
func (m *Manager) IsAvailable() bool {
	return m.provider != nil && m.provider.IsAvailable()
}

// BuildInVM builds a package using a VM
func (m *Manager) BuildInVM(ctx context.Context, image, workDir, cmd string) (string, error) {
	if !m.config.Enabled {
		return "", fmt.Errorf("VM support not enabled")
	}
	
	if !m.IsAvailable() {
		return "", fmt.Errorf("VM provider not available")
	}
	
	// Setup
	if err := m.provider.Setup(ctx); err != nil {
		return "", fmt.Errorf("setup failed: %w", err)
	}
	defer m.provider.Cleanup(ctx)
	
	// Build
	output, err := m.provider.Build(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}
	
	return output, nil
}

// DockerProvider implements Provider for Docker
type DockerProvider struct {
	config      *DockerConfig
	containerID string
	workDir     string
}

func NewDockerProvider(cfg *DockerConfig) *DockerProvider {
	if cfg.Image == "" {
		cfg.Image = "ubuntu:22.04"
	}
	return &DockerProvider{config: cfg}
}

func (p *DockerProvider) IsAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func (p *DockerProvider) Setup(ctx context.Context) error {
	return nil // Lazy setup on first build
}

func (p *DockerProvider) Build(ctx context.Context, cmd string) (string, error) {
	// Use docker run with volume mount
	absWorkDir, err := filepath.Abs(p.config.WorkDir)
	if err != nil {
		absWorkDir = p.config.WorkDir
	}
	
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/work", absWorkDir),
		"-w", "/work",
		p.config.Image,
		"sh", "-c", cmd,
	}
	
	execCmd := exec.CommandContext(ctx, "docker", args...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(output))
	}
	
	// Get last line (actual output, not pull messages)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines[len(lines)-1], nil
}

func (p *DockerProvider) CopyArtifact(ctx context.Context, src, dst string) error {
	return nil // Not needed with volume mounts
}

func (p *DockerProvider) Cleanup(ctx context.Context) error {
	return nil // Using --rm flag
}


