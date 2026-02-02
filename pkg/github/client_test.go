package github

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
	}

	// Test with missing token
	os.Unsetenv("GITHUB_TOKEN")
	_, err := NewClient(cfg)
	if err == nil {
		t.Error("Expected error for missing GitHub token")
	}

	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("Expected GITHUB_TOKEN error, got: %v", err)
	}

	// Test with valid token
	os.Setenv("GITHUB_TOKEN", "mock-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Error("NewClient returned nil")
	}

	if client.cfg != cfg {
		t.Error("Client config not set correctly")
	}
}

func TestCreateRelease_MissingToken(t *testing.T) {
	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
	}

	// Test without token - should fail at client creation
	os.Unsetenv("GITHUB_TOKEN")
	_, err := NewClient(cfg)
	if err == nil {
		t.Error("Expected error for missing GitHub token")
	}
}

func TestUploadAsset_InvalidFile(t *testing.T) {
	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
	}

	os.Setenv("GITHUB_TOKEN", "mock-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	// Test with non-existent file - this tests the private method indirectly
	// by testing the public CreateRelease method that uses uploadAsset
	appCfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	ctx := context.Background()
	_, err = client.CreateRelease(ctx, appCfg, []string{"/non/existent/file.txt"})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestUploadAsset_ValidFile(t *testing.T) {
	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
	}

	os.Setenv("GITHUB_TOKEN", "mock-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	// Create test file
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test-asset.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Test CreateRelease with valid file - this will fail with API call but should pass file validation
	appCfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	ctx := context.Background()
	_, err = client.CreateRelease(ctx, appCfg, []string{testFile})
	// We expect this to fail with API error, not file error
	if err != nil && strings.Contains(err.Error(), "no such file") {
		t.Error("File validation failed unexpectedly")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *config.GitHubConfig
		token  string
		valid  bool
	}{
		{
			name: "Valid config with token",
			config: &config.GitHubConfig{
				Owner:    "testowner",
				Repo:     "testrepo",
				TokenEnv: "GITHUB_TOKEN",
			},
			token: "mock-token",
			valid: true,
		},
		{
			name: "Valid config without token",
			config: &config.GitHubConfig{
				Owner:    "testowner",
				Repo:     "testrepo",
				TokenEnv: "GITHUB_TOKEN",
			},
			token: "",
			valid: false,
		},
		{
			name: "Config with missing fields but valid token",
			config: &config.GitHubConfig{
				TokenEnv: "GITHUB_TOKEN",
			},
			token: "mock-token",
			valid: true, // NewClient only validates token
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.token != "" {
				os.Setenv("GITHUB_TOKEN", test.token)
			} else {
				os.Unsetenv("GITHUB_TOKEN")
			}
			defer os.Unsetenv("GITHUB_TOKEN")

			_, err := NewClient(test.config)
			
			if test.valid && err != nil {
				t.Errorf("Expected valid config to succeed, got error: %v", err)
			}
			
			if !test.valid && err == nil {
				t.Error("Expected invalid config to fail")
			}
		})
	}
}

func TestTapConfiguration(t *testing.T) {
	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
		Tap: config.TapConfig{
			Enabled:    true,
			Repo:       "testowner/homebrew-tap",
			AutoCreate: true,
			AutoCommit: true,
			AutoPush:   true,
		},
	}

	os.Setenv("GITHUB_TOKEN", "mock-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	// Verify tap configuration is accessible
	if !client.cfg.Tap.Enabled {
		t.Error("Tap should be enabled")
	}

	if client.cfg.Tap.Repo != "testowner/homebrew-tap" {
		t.Errorf("Expected tap repo 'testowner/homebrew-tap', got %s", client.cfg.Tap.Repo)
	}
}

func TestBucketConfiguration(t *testing.T) {
	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
		Bucket: config.BucketConfig{
			Enabled:    true,
			Repo:       "testowner/scoop-bucket",
			AutoCreate: true,
			AutoCommit: true,
			AutoPush:   true,
		},
	}

	os.Setenv("GITHUB_TOKEN", "mock-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	// Verify bucket configuration is accessible
	if !client.cfg.Bucket.Enabled {
		t.Error("Bucket should be enabled")
	}

	if client.cfg.Bucket.Repo != "testowner/scoop-bucket" {
		t.Errorf("Expected bucket repo 'testowner/scoop-bucket', got %s", client.cfg.Bucket.Repo)
	}
}

func TestWingetConfiguration(t *testing.T) {
	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
		Winget: config.WingetConfig{
			Enabled:  true,
			AutoPR:   true,
			ForkRepo: "testowner/winget-pkgs",
		},
	}

	os.Setenv("GITHUB_TOKEN", "mock-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	// Verify winget configuration is accessible
	if !client.cfg.Winget.Enabled {
		t.Error("Winget should be enabled")
	}

	if !client.cfg.Winget.AutoPR {
		t.Error("Winget AutoPR should be enabled")
	}

	if client.cfg.Winget.ForkRepo != "testowner/winget-pkgs" {
		t.Errorf("Expected fork repo 'testowner/winget-pkgs', got %s", client.cfg.Winget.ForkRepo)
	}
}

func TestReleaseConfiguration(t *testing.T) {
	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
		Release: config.ReleaseConfig{
			Enabled:       true,
			Draft:         false,
			Prerelease:    false,
			GenerateNotes: true,
		},
	}

	os.Setenv("GITHUB_TOKEN", "mock-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	// Verify release configuration is accessible
	if !client.cfg.Release.Enabled {
		t.Error("Release should be enabled")
	}

	if client.cfg.Release.Draft {
		t.Error("Release should not be draft")
	}

	if !client.cfg.Release.GenerateNotes {
		t.Error("Release should generate notes")
	}
}
