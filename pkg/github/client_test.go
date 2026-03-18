package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ghlib "github.com/google/go-github/v57/github"
	"github.com/scttfrdmn/bagboy/pkg/config"
)

// newTestClient creates a Client backed by an httptest.Server so tests do not
// hit the real GitHub API.
func newTestClient(t *testing.T, mux *http.ServeMux) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}

	ghClient := ghlib.NewClient(nil)
	ghClient.BaseURL = baseURL
	ghClient.UploadURL = baseURL // avoid nil pointer in upload calls

	cfg := &config.GitHubConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		TokenEnv: "GITHUB_TOKEN",
	}

	return &Client{gh: ghClient, cfg: cfg}, server
}

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

// --- httptest-backed tests ---

// TestCreateRelease_HTTPTest uses a mock GitHub API to verify that
// CreateRelease constructs the correct request and handles the response.
func TestCreateRelease_HTTPTest(t *testing.T) {
	releaseID := int64(42)
	mux := http.NewServeMux()

	// Handle release creation.
	mux.HandleFunc("/repos/testowner/testrepo/releases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       releaseID,
			"tag_name": "v1.0.0",
			"name":     "v1.0.0",
		})
	})

	client, _ := newTestClient(t, mux)
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Repo:  "testrepo",
			Release: config.ReleaseConfig{
				Enabled:       true,
				GenerateNotes: true,
			},
		},
	}

	rel, err := client.CreateRelease(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("CreateRelease: %v", err)
	}
	if rel.GetID() != releaseID {
		t.Errorf("expected release ID %d, got %d", releaseID, rel.GetID())
	}
}

// TestCreateRelease_404 verifies that a 404 response from the API is surfaced
// as an error.
func TestCreateRelease_404(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/testowner/testrepo/releases", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	})

	client, _ := newTestClient(t, mux)
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{Owner: "testowner", Repo: "testrepo"},
	}

	_, err := client.CreateRelease(context.Background(), cfg, nil)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

// TestCreateRelease_422 verifies that a 422 (validation failed) from the API
// is surfaced as an error.
func TestCreateRelease_422(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/testowner/testrepo/releases", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message":"Validation Failed","errors":[{"code":"already_exists"}]}`))
	})

	client, _ := newTestClient(t, mux)
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{Owner: "testowner", Repo: "testrepo"},
	}

	_, err := client.CreateRelease(context.Background(), cfg, nil)
	if err == nil {
		t.Error("expected error for 422 response")
	}
}

// TestUpdateTap_Disabled verifies that UpdateTap returns nil immediately when
// the tap is not enabled.
func TestUpdateTap_Disabled(t *testing.T) {
	client, _ := newTestClient(t, http.NewServeMux())
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Tap:   config.TapConfig{Enabled: false},
		},
	}
	err := client.UpdateTap(context.Background(), cfg, "formula content")
	if err != nil {
		t.Errorf("expected nil for disabled tap, got: %v", err)
	}
}

// TestUpdateTap_InvalidRepoFormat verifies that a malformed tap repo string is
// rejected with a clear error before any API call is made.
func TestUpdateTap_InvalidRepoFormat(t *testing.T) {
	client, _ := newTestClient(t, http.NewServeMux())
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Tap: config.TapConfig{
				Enabled:    true,
				Repo:       "no-slash-here",
				AutoCreate: false,
				AutoCommit: true,
			},
		},
	}
	err := client.UpdateTap(context.Background(), cfg, "formula")
	if err == nil {
		t.Error("expected error for invalid tap repo format")
	}
	if !strings.Contains(err.Error(), "invalid tap repo format") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestUpdateTap_AutoCommitDisabled verifies that when AutoCommit is false,
// UpdateTap skips the API call and returns nil.
func TestUpdateTap_AutoCommitDisabled(t *testing.T) {
	apiCalled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		apiCalled = true
		http.Error(w, "should not be called", http.StatusInternalServerError)
	})

	client, _ := newTestClient(t, mux)
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Tap: config.TapConfig{
				Enabled:    true,
				Repo:       "testowner/homebrew-tap",
				AutoCreate: false,
				AutoCommit: false, // disabled
			},
		},
	}
	err := client.UpdateTap(context.Background(), cfg, "formula")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if apiCalled {
		t.Error("expected no API calls when AutoCommit=false")
	}
}

// TestUpdateBucket_Disabled verifies UpdateBucket returns nil when disabled.
func TestUpdateBucket_Disabled(t *testing.T) {
	client, _ := newTestClient(t, http.NewServeMux())
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner:  "testowner",
			Bucket: config.BucketConfig{Enabled: false},
		},
	}
	err := client.UpdateBucket(context.Background(), cfg, `{"json":"manifest"}`)
	if err != nil {
		t.Errorf("expected nil for disabled bucket, got: %v", err)
	}
}

// TestUpdateBucket_InvalidRepoFormat verifies malformed bucket repo is rejected.
func TestUpdateBucket_InvalidRepoFormat(t *testing.T) {
	client, _ := newTestClient(t, http.NewServeMux())
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Bucket: config.BucketConfig{
				Enabled:    true,
				Repo:       "noslash",
				AutoCreate: false,
				AutoCommit: true,
			},
		},
	}
	err := client.UpdateBucket(context.Background(), cfg, "manifest")
	if err == nil {
		t.Error("expected error for invalid bucket repo format")
	}
	if !strings.Contains(err.Error(), "invalid bucket repo format") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSubmitWingetPR_Disabled verifies that SubmitWingetPR is a no-op when
// Winget is disabled or AutoPR is false.
func TestSubmitWingetPR_Disabled(t *testing.T) {
	client, _ := newTestClient(t, http.NewServeMux())
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner:  "testowner",
			Winget: config.WingetConfig{Enabled: false, AutoPR: false},
		},
	}
	err := client.SubmitWingetPR(context.Background(), cfg, nil)
	if err != nil {
		t.Errorf("expected nil for disabled Winget, got: %v", err)
	}
}

// TestUpdateTap_HTTPTest uses a mock server to test the full UpdateTap path
// including the GetContents + CreateFile calls.
func TestUpdateTap_HTTPTest(t *testing.T) {
	mux := http.NewServeMux()

	// ensureRepository: GET to check if repo exists → 200 (exists, skip create)
	mux.HandleFunc("/repos/testowner/homebrew-tap", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "name": "homebrew-tap"})
	})

	// updateFile: GetContents → 404 (new file), then CreateFile → 201
	mux.HandleFunc("/repos/testowner/homebrew-tap/contents/Formula/testapp.rb", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
			return
		}
		// PUT / create
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": map[string]interface{}{"name": "testapp.rb"},
		})
	})

	client, _ := newTestClient(t, mux)
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Tap: config.TapConfig{
				Enabled:    true,
				Repo:       "testowner/homebrew-tap",
				AutoCreate: true,
				AutoCommit: true,
			},
		},
	}

	err := client.UpdateTap(context.Background(), cfg, "class Testapp < Formula; end")
	if err != nil {
		t.Errorf("UpdateTap: %v", err)
	}
}
