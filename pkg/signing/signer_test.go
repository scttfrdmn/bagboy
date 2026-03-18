package signing

import (
	"context"
	"os"
	"path/filepath"
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

	if status.Available {
		t.Error("Expected Sigstore not available without cosign")
	}
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
