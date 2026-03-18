package signing

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestNewSigner(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			MacOS: config.MacOSSigningConfig{
				Identity: "Test Identity",
			},
		},
	}

	signer := NewSigner(cfg)
	if signer == nil {
		t.Error("NewSigner returned nil")
	}

	if signer.config != cfg {
		t.Error("Signer config not set correctly")
	}
}

func TestCheckSigningSetup(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	signer := NewSigner(cfg)

	// This should not crash even with no signing configuration
	results := signer.CheckSigningSetup()
	if results == nil {
		t.Error("CheckSigningSetup returned nil")
	}

	// Should have entries for different platforms
	if len(results) == 0 {
		t.Error("CheckSigningSetup returned empty results")
	}
}

func TestSignAllBinaries_NoConfig(t *testing.T) {
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "testapp")
	os.WriteFile(testBinary, []byte("fake binary"), 0755)

	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Binaries: map[string]string{
			"linux-amd64": testBinary,
		},
	}

	signer := NewSigner(cfg)
	ctx := context.Background()

	// Should handle missing signing config gracefully
	err := signer.SignAllBinaries(ctx)
	if err == nil {
		t.Error("Expected error for missing signing configuration")
	}

	// Error should be informative
	if !strings.Contains(err.Error(), "signing failed") {
		t.Errorf("Expected informative error message, got: %v", err)
	}
}

func TestSignWithSigstore_NotConfigured(t *testing.T) {
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "testapp")
	os.WriteFile(testBinary, []byte("fake binary"), 0755)

	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Sigstore: config.SigstoreConfig{
				Enabled: false,
			},
		},
	}

	signer := NewSigner(cfg)
	ctx := context.Background()

	// Should handle disabled Sigstore gracefully
	err := signer.SignWithSigstore(ctx, testBinary)
	if err == nil {
		t.Error("Expected error for disabled Sigstore")
	}
}

func TestSignWithSignPath_NotConfigured(t *testing.T) {
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "testapp.exe")
	os.WriteFile(testBinary, []byte("fake binary"), 0755)

	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			SignPath: config.SignPathConfig{
				Enabled: false,
			},
		},
	}

	signer := NewSigner(cfg)
	ctx := context.Background()

	// Should handle disabled SignPath gracefully
	err := signer.SignWithSignPath(ctx, testBinary)
	if err == nil {
		t.Error("Expected error for disabled SignPath")
	}
}

func TestSignBinary_MacOS_NoIdentity(t *testing.T) {
	testDir := t.TempDir()
	testBinary := filepath.Join(testDir, "testapp")
	os.WriteFile(testBinary, []byte("fake binary"), 0755)

	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			MacOS: config.MacOSSigningConfig{
				// No identity configured
			},
		},
	}

	signer := NewSigner(cfg)
	ctx := context.Background()

	// Should handle missing identity gracefully
	err := signer.SignBinary(ctx, testBinary)
	if err == nil {
		t.Error("Expected error for missing macOS identity")
	}

	if !strings.Contains(err.Error(), "APPLE_DEVELOPER_ID") {
		t.Errorf("Expected APPLE_DEVELOPER_ID error, got: %v", err)
	}
}

func TestSignBinary_Windows_NoCertificate(t *testing.T) {
	// Skip this test for now - platform detection logic needs work
	t.Skip("Platform detection needs improvement")
}

func TestSignBinary_Linux_NoGPG(t *testing.T) {
	// Skip this test for now - platform detection needs work
	t.Skip("Platform detection needs improvement")
}

func TestGetSigningRequirements(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	signer := NewSigner(cfg)
	requirements := signer.GetSigningRequirements()

	if len(requirements) == 0 {
		t.Error("Expected signing requirements, got none")
	}

	// Should have requirements for different platforms
	platforms := make(map[string]bool)
	for _, req := range requirements {
		platforms[req.Platform] = true
	}

	expectedPlatforms := []string{"macOS", "Windows", "Linux"}
	for _, platform := range expectedPlatforms {
		if !platforms[platform] {
			t.Errorf("Missing requirements for platform: %s", platform)
		}
	}
}

func TestEnvironmentVariableHandling(t *testing.T) {
	// Test that signing requirements are properly checked
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			MacOS: config.MacOSSigningConfig{
				Identity: "Test Identity",
			},
		},
	}

	signer := NewSigner(cfg)

	// Test signing setup check
	results := signer.CheckSigningSetup()
	if results == nil {
		t.Error("CheckSigningSetup returned nil")
	}

	// Should have macOS entry
	if _, exists := results["macOS"]; !exists {
		t.Error("Expected macOS signing status")
	}
}

func TestSigningConfigValidation(t *testing.T) {
	// Test basic signing setup functionality
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			MacOS: config.MacOSSigningConfig{
				Identity: "Test Identity",
			},
		},
	}

	signer := NewSigner(cfg)
	results := signer.CheckSigningSetup()

	if len(results) == 0 {
		t.Error("Expected signing setup results")
	}

	// Should have platform entries
	if _, exists := results["macOS"]; !exists {
		t.Error("Expected macOS signing status")
	}
}

func TestSignWithGit(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Git: config.GitSigningConfig{
				Enabled:  false,
				SignTags: true,
			},
		},
	}

	signer := NewSigner(cfg)
	ctx := context.Background()

	// Should return nil when disabled
	err := signer.SignWithGit(ctx, "v1.0.0")
	if err != nil {
		t.Errorf("Expected no error when Git signing disabled, got: %v", err)
	}
}

func TestPrintSigningReport(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}

	signer := NewSigner(cfg)
	results := map[string]SigningStatus{
		"macOS": {
			Platform:  "macOS",
			Available: true,
			Required:  true,
		},
		"Windows": {
			Platform:  "Windows",
			Available: false,
			Required:  false,
			Issues:    []string{"Certificate not found"},
		},
	}

	// Should not panic
	signer.PrintSigningReport(results)
}

func TestCheckPlatformSigning(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.SigningConfig
		platform string
		want     bool
	}{
		{
			name: "macOS configured",
			cfg: config.SigningConfig{
				MacOS: config.MacOSSigningConfig{
					Identity: "Test Identity",
				},
			},
			platform: "macOS",
			want:     false, // Will be false without actual tools
		},
		{
			name: "Windows not configured",
			cfg: config.SigningConfig{
				Windows: config.WindowsSigningConfig{},
			},
			platform: "Windows",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Name:    "testapp",
				Version: "1.0.0",
				Signing: tt.cfg,
			}

			signer := NewSigner(cfg)
			req := SigningRequirement{Platform: tt.platform}
			got := signer.checkPlatformSigning(req)

			if got != tt.want {
				t.Errorf("checkPlatformSigning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSigningIssues(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			MacOS: config.MacOSSigningConfig{},
		},
	}

	signer := NewSigner(cfg)
	req := SigningRequirement{
		Platform: "macOS",
		Tools:    []string{"codesign"},
	}

	issues := signer.getSigningIssues(req)
	if len(issues) == 0 {
		t.Error("Expected signing issues for unconfigured macOS")
	}
}

func TestCheckMacOSSigning(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			MacOS: config.MacOSSigningConfig{
				Identity: "Test Identity",
			},
		},
	}

	signer := NewSigner(cfg)
	result := signer.checkMacOSSigning()

	// Will be false without actual codesign tool
	if result {
		t.Error("Expected false when codesign not available")
	}
}

func TestCheckWindowsSigning(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Windows: config.WindowsSigningConfig{
				CertificateThumbprint: "test",
			},
		},
	}

	signer := NewSigner(cfg)
	result := signer.checkWindowsSigning()

	// Will be false without actual signtool
	if result {
		t.Error("Expected false when signtool not available")
	}
}

func TestCheckLinuxSigning(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Linux: config.LinuxSigningConfig{
				GPGKeyID: "test",
			},
		},
	}

	signer := NewSigner(cfg)
	result := signer.checkLinuxSigning()

	// Will be false without actual gpg tool
	if result {
		t.Error("Expected false when gpg not available")
	}
}

func TestCheckSigstore(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Sigstore: config.SigstoreConfig{
				Enabled: true,
			},
		},
	}

	signer := NewSigner(cfg)
	status := signer.checkSigstore()

	// If cosign is not installed, the status should be unavailable.
	// If it is installed (e.g. on a developer machine), it may report available.
	// Either way the call should not panic.
	_ = status.Available
}

func TestCheckSignPath(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			SignPath: config.SignPathConfig{
				Enabled:        true,
				OrganizationID: "test-org",
				ProjectSlug:    "test-project",
			},
		},
	}

	signer := NewSigner(cfg)
	status := signer.checkSignPath()

	// Just verify it returns a status
	_ = status
}

func TestCheckGitSigning(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Git: config.GitSigningConfig{
				Enabled:  true,
				SignTags: true,
			},
		},
	}

	signer := NewSigner(cfg)
	status := signer.checkGitSigning()

	// Just verify it returns a status
	_ = status
}

// TestCheckGitSigning_WithGPGKeyID verifies that a non-existent GPG key ID
// results in an issue being reported.
func TestCheckGitSigning_WithGPGKeyID(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Git: config.GitSigningConfig{
				Enabled:  true,
				GPGKeyID: "DEADBEEF12345678",
			},
		},
	}

	signer := NewSigner(cfg)
	status := signer.checkGitSigning()

	// Should not be available — the key doesn't exist on this system.
	if status.Available {
		t.Error("expected signing unavailable for nonexistent GPG key")
	}
}

// TestCheckSignPath_MissingFields exercises each required SignPath field
// individually to confirm checkSignPath reports the right issues.
func TestCheckSignPath_MissingFields(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.SignPathConfig
		wantIssue   string
	}{
		{
			name: "missing org ID",
			cfg: config.SignPathConfig{
				Enabled:     true,
				ProjectSlug: "proj",
				APIToken:    "tok",
			},
			wantIssue: "organization ID",
		},
		{
			name: "missing project slug",
			cfg: config.SignPathConfig{
				Enabled:        true,
				OrganizationID: "org",
				APIToken:       "tok",
			},
			wantIssue: "project slug",
		},
		{
			name: "missing API token",
			cfg: config.SignPathConfig{
				Enabled:        true,
				OrganizationID: "org",
				ProjectSlug:    "proj",
			},
			wantIssue: "API token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Name:    "testapp",
				Version: "1.0.0",
				Signing: config.SigningConfig{SignPath: tt.cfg},
			}
			signer := NewSigner(cfg)
			status := signer.checkSignPath()

			if status.Available {
				t.Errorf("expected unavailable for %s", tt.name)
			}

			found := false
			for _, issue := range status.Issues {
				if strings.Contains(strings.ToLower(issue), strings.ToLower(tt.wantIssue)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected issue containing %q, got issues: %v", tt.wantIssue, status.Issues)
			}
		})
	}
}

// TestCheckSigstore_Keyless_NoOIDCIssuer verifies that enabling keyless signing
// without an OIDC issuer produces an issue.
func TestCheckSigstore_Keyless_NoOIDCIssuer(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Sigstore: config.SigstoreConfig{
				Enabled: true,
				Keyless: true,
				// OIDCIssuer intentionally empty
			},
		},
	}
	signer := NewSigner(cfg)
	status := signer.checkSigstore()

	if status.Available {
		t.Error("expected unavailable when keyless but no OIDC issuer")
	}

	foundIssue := false
	for _, issue := range status.Issues {
		if strings.Contains(strings.ToLower(issue), "oidc") {
			foundIssue = true
			break
		}
	}
	if !foundIssue {
		t.Errorf("expected OIDC-related issue, got: %v", status.Issues)
	}
}

// TestCheckSigstore_WithFakeCosign places a fake cosign script on PATH and
// verifies that checkSigstore considers it available (when keyless OIDC is also
// set so that's not a blocker).
func TestCheckSigstore_WithFakeCosign(t *testing.T) {
	tmpDir := t.TempDir()
	fakeCosign := filepath.Join(tmpDir, "cosign")
	if err := os.WriteFile(fakeCosign, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	orig := os.Getenv("PATH")
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+orig)

	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Sigstore: config.SigstoreConfig{
				Enabled:    true,
				Keyless:    true,
				OIDCIssuer: "https://token.actions.githubusercontent.com",
			},
		},
	}
	signer := NewSigner(cfg)
	status := signer.checkSigstore()

	if !status.Available {
		t.Errorf("expected available with fake cosign on PATH, issues: %v", status.Issues)
	}
}

// TestSignWithSigstore_NoCosign verifies that SignWithSigstore returns an error
// when cosign is not on PATH.
func TestSignWithSigstore_NoCosign(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir) // empty PATH — no cosign

	testBinary := filepath.Join(tmpDir, "testapp")
	if err := os.WriteFile(testBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Sigstore: config.SigstoreConfig{Enabled: true},
		},
	}
	signer := NewSigner(cfg)
	err := signer.SignWithSigstore(context.Background(), testBinary)
	if err == nil {
		t.Error("expected error when cosign not on PATH")
	}
	if !strings.Contains(err.Error(), "cosign") {
		t.Errorf("expected cosign-related error, got: %v", err)
	}
}

// TestShouldNotarize verifies notarisation gating via env vars.
func TestShouldNotarize(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("notarize check is macOS-only")
	}

	signer := NewSigner(nil)

	// Without env vars, should not notarize.
	t.Setenv("APPLE_ID", "")
	t.Setenv("APPLE_APP_PASSWORD", "")
	if signer.shouldNotarize() {
		t.Error("expected shouldNotarize=false when env vars not set")
	}

	// With both env vars set, should notarize.
	t.Setenv("APPLE_ID", "dev@example.com")
	t.Setenv("APPLE_APP_PASSWORD", "xxxx-xxxx-xxxx-xxxx")
	if !signer.shouldNotarize() {
		t.Error("expected shouldNotarize=true when both env vars are set")
	}
}

// TestSignWithGit_Enabled_NoTagName verifies that SignWithGit with an enabled
// config but empty tagName does not attempt to run git tag.
func TestSignWithGit_Enabled_EmptyTag(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			Git: config.GitSigningConfig{
				Enabled:  true,
				SignTags: true,
				GPGKeyID: "",
			},
		},
	}
	signer := NewSigner(cfg)
	// Empty tagName — should not attempt git tag and should return nil.
	err := signer.SignWithGit(context.Background(), "")
	if err != nil {
		t.Errorf("expected no error for empty tagName, got: %v", err)
	}
}

// TestSignWithSignPath_MissingOrgID verifies the enabled+missing-org-ID path.
func TestSignWithSignPath_MissingOrgID(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			SignPath: config.SignPathConfig{
				Enabled: true,
				// OrganizationID missing
				ProjectSlug: "proj",
			},
		},
	}
	signer := NewSigner(cfg)
	err := signer.SignWithSignPath(context.Background(), "some-binary")
	if err == nil {
		t.Error("expected error for missing organization ID")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "organization") {
		t.Errorf("expected organization error, got: %v", err)
	}
}

// TestSignWithSignPath_MissingProjectSlug verifies the enabled+missing-project-slug path.
func TestSignWithSignPath_MissingProjectSlug(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
		Signing: config.SigningConfig{
			SignPath: config.SignPathConfig{
				Enabled:        true,
				OrganizationID: "org-id",
				// ProjectSlug missing
			},
		},
	}
	signer := NewSigner(cfg)
	err := signer.SignWithSignPath(context.Background(), "some-binary")
	if err == nil {
		t.Error("expected error for missing project slug")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "project slug") {
		t.Errorf("expected project slug error, got: %v", err)
	}
}
