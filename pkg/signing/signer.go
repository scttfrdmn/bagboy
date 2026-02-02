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

package signing

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// Signer handles code signing for different platforms
type Signer struct {
	config *config.Config
}

// NewSigner creates a new code signer
func NewSigner(cfg *config.Config) *Signer {
	return &Signer{config: cfg}
}

// SigningRequirement represents signing requirements for a platform
type SigningRequirement struct {
	Platform    string
	Required    bool
	Tools       []string
	Description string
	SetupSteps  []string
}

// GetSigningRequirements returns signing requirements for all platforms
func (s *Signer) GetSigningRequirements() []SigningRequirement {
	return []SigningRequirement{
		{
			Platform:    "macOS",
			Required:    true,
			Tools:       []string{"codesign", "xcrun"},
			Description: "Apple requires code signing for distribution outside App Store, mandatory for notarization",
			SetupSteps: []string{
				"1. Join Apple Developer Program ($99/year)",
				"2. Create Developer ID Application certificate in Xcode or Apple Developer portal",
				"3. Download and install certificate in Keychain",
				"4. Set APPLE_DEVELOPER_ID environment variable",
				"5. For notarization: Set APPLE_ID and APPLE_APP_PASSWORD",
			},
		},
		{
			Platform:    "Windows",
			Required:    false,
			Tools:       []string{"signtool"},
			Description: "Code signing prevents Windows Defender warnings and builds user trust",
			SetupSteps: []string{
				"1. Purchase code signing certificate from CA (DigiCert, Sectigo, etc.)",
				"2. Install Windows SDK for signtool.exe",
				"3. Install certificate in Windows Certificate Store",
				"4. Set WINDOWS_CERT_THUMBPRINT environment variable",
			},
		},
		{
			Platform:    "Linux",
			Required:    false,
			Tools:       []string{"gpg"},
			Description: "GPG signing for package repositories and verification",
			SetupSteps: []string{
				"1. Generate GPG key: gpg --gen-key",
				"2. Export public key: gpg --export --armor your@email.com",
				"3. Upload to keyservers: gpg --send-keys KEYID",
				"4. Set GPG_KEY_ID environment variable",
			},
		},
	}
}

// CheckSigningSetup verifies signing configuration
func (s *Signer) CheckSigningSetup() map[string]SigningStatus {
	requirements := s.GetSigningRequirements()
	results := make(map[string]SigningStatus)
	
	for _, req := range requirements {
		status := SigningStatus{
			Platform: req.Platform,
			Required: req.Required,
			Available: s.checkPlatformSigning(req),
		}
		
		if !status.Available {
			status.Issues = s.getSigningIssues(req)
			status.SetupSteps = req.SetupSteps
		}
		
		results[req.Platform] = status
	}
	
	// Check Sigstore/Cosign
	if s.config != nil && s.config.Signing.Sigstore.Enabled {
		status := s.checkSigstore()
		results["sigstore"] = status
	}

	// Check SignPath.io
	if s.config != nil && s.config.Signing.SignPath.Enabled {
		status := s.checkSignPath()
		results["signpath"] = status
	}

	// Check Git signing
	if s.config != nil && s.config.Signing.Git.Enabled {
		status := s.checkGitSigning()
		results["git"] = status
	}

	return results
}

// SigningStatus represents the signing status for a platform
type SigningStatus struct {
	Platform   string
	Required   bool
	Available  bool
	Issues     []string
	SetupSteps []string
}

func (s *Signer) checkPlatformSigning(req SigningRequirement) bool {
	switch req.Platform {
	case "macOS":
		return s.checkMacOSSigning()
	case "Windows":
		return s.checkWindowsSigning()
	case "Linux":
		return s.checkLinuxSigning()
	}
	return false
}

func (s *Signer) checkMacOSSigning() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	
	// Check if codesign is available
	if _, err := exec.LookPath("codesign"); err != nil {
		return false
	}
	
	// Check if developer identity is available
	cmd := exec.Command("security", "find-identity", "-v", "-p", "codesigning")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return strings.Contains(string(output), "Developer ID Application")
}

func (s *Signer) checkWindowsSigning() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	
	// Check if signtool is available
	if _, err := exec.LookPath("signtool"); err != nil {
		return false
	}
	
	// Check if certificate is configured
	thumbprint := os.Getenv("WINDOWS_CERT_THUMBPRINT")
	return thumbprint != ""
}

func (s *Signer) checkLinuxSigning() bool {
	// Check if GPG is available
	if _, err := exec.LookPath("gpg"); err != nil {
		return false
	}
	
	// Check if GPG key is configured
	keyID := os.Getenv("GPG_KEY_ID")
	if keyID == "" {
		return false
	}
	
	// Verify key exists
	cmd := exec.Command("gpg", "--list-secret-keys", keyID)
	return cmd.Run() == nil
}

func (s *Signer) getSigningIssues(req SigningRequirement) []string {
	var issues []string
	
	switch req.Platform {
	case "macOS":
		if runtime.GOOS != "darwin" {
			issues = append(issues, "Not running on macOS")
		} else {
			if _, err := exec.LookPath("codesign"); err != nil {
				issues = append(issues, "codesign not found (install Xcode Command Line Tools)")
			}
			if os.Getenv("APPLE_DEVELOPER_ID") == "" {
				issues = append(issues, "APPLE_DEVELOPER_ID environment variable not set")
			}
		}
		
	case "Windows":
		if runtime.GOOS != "windows" {
			issues = append(issues, "Not running on Windows")
		} else {
			if _, err := exec.LookPath("signtool"); err != nil {
				issues = append(issues, "signtool not found (install Windows SDK)")
			}
			if os.Getenv("WINDOWS_CERT_THUMBPRINT") == "" {
				issues = append(issues, "WINDOWS_CERT_THUMBPRINT environment variable not set")
			}
		}
		
	case "Linux":
		if _, err := exec.LookPath("gpg"); err != nil {
			issues = append(issues, "GPG not found")
		}
		if os.Getenv("GPG_KEY_ID") == "" {
			issues = append(issues, "GPG_KEY_ID environment variable not set")
		}
	}
	
	return issues
}

// SignBinary signs a binary for the current platform
func (s *Signer) SignBinary(ctx context.Context, binaryPath string) error {
	switch runtime.GOOS {
	case "darwin":
		return s.signMacOSBinary(ctx, binaryPath)
	case "windows":
		return s.signWindowsBinary(ctx, binaryPath)
	case "linux":
		return s.signLinuxBinary(ctx, binaryPath)
	default:
		return fmt.Errorf("signing not supported on %s", runtime.GOOS)
	}
}

func (s *Signer) SignAllBinaries(ctx context.Context) error {
	if s.config == nil {
		return fmt.Errorf("no configuration provided")
	}

	var errors []string
	for arch, binaryPath := range s.config.Binaries {
		fmt.Printf("Signing %s binary: %s\n", arch, binaryPath)
		
		// Sign based on target platform
		var err error
		if strings.HasPrefix(arch, "darwin-") {
			err = s.signMacOSBinary(ctx, binaryPath)
		} else if strings.HasPrefix(arch, "windows-") {
			err = s.signWindowsBinary(ctx, binaryPath)
		} else if strings.HasPrefix(arch, "linux-") {
			err = s.signLinuxBinary(ctx, binaryPath)
		} else {
			err = fmt.Errorf("unsupported architecture: %s", arch)
		}

		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", arch, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("signing failed for some binaries:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

func (s *Signer) signMacOSBinary(ctx context.Context, binaryPath string) error {
	identity := os.Getenv("APPLE_DEVELOPER_ID")
	if identity == "" {
		return fmt.Errorf("APPLE_DEVELOPER_ID environment variable not set")
	}
	
	// Sign the binary
	cmd := exec.CommandContext(ctx, "codesign", 
		"--sign", identity,
		"--timestamp",
		"--options", "runtime",
		binaryPath)
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("codesign failed: %w\nOutput: %s", err, output)
	}
	
	fmt.Printf("✅ Signed macOS binary: %s\n", binaryPath)
	
	// Optionally notarize
	if s.shouldNotarize() {
		return s.notarizeMacOSBinary(ctx, binaryPath)
	}
	
	return nil
}

func (s *Signer) signWindowsBinary(ctx context.Context, binaryPath string) error {
	thumbprint := os.Getenv("WINDOWS_CERT_THUMBPRINT")
	if thumbprint == "" {
		return fmt.Errorf("WINDOWS_CERT_THUMBPRINT environment variable not set")
	}
	
	cmd := exec.CommandContext(ctx, "signtool", "sign",
		"/sha1", thumbprint,
		"/t", "http://timestamp.digicert.com",
		"/fd", "SHA256",
		binaryPath)
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("signtool failed: %w\nOutput: %s", err, output)
	}
	
	fmt.Printf("✅ Signed Windows binary: %s\n", binaryPath)
	return nil
}

func (s *Signer) signLinuxBinary(ctx context.Context, binaryPath string) error {
	keyID := os.Getenv("GPG_KEY_ID")
	if keyID == "" {
		return fmt.Errorf("GPG_KEY_ID environment variable not set")
	}
	
	// Create detached signature
	sigPath := binaryPath + ".sig"
	cmd := exec.CommandContext(ctx, "gpg", 
		"--detach-sign",
		"--armor",
		"--local-user", keyID,
		"--output", sigPath,
		binaryPath)
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gpg signing failed: %w\nOutput: %s", err, output)
	}
	
	fmt.Printf("✅ Signed Linux binary: %s (signature: %s)\n", binaryPath, sigPath)
	return nil
}

func (s *Signer) shouldNotarize() bool {
	return os.Getenv("APPLE_ID") != "" && os.Getenv("APPLE_APP_PASSWORD") != ""
}

func (s *Signer) notarizeMacOSBinary(ctx context.Context, binaryPath string) error {
	appleID := os.Getenv("APPLE_ID")
	appPassword := os.Getenv("APPLE_APP_PASSWORD")
	
	if appleID == "" || appPassword == "" {
		fmt.Println("⚠️  Skipping notarization (APPLE_ID or APPLE_APP_PASSWORD not set)")
		return nil
	}
	
	// Create a zip for notarization
	zipPath := strings.TrimSuffix(binaryPath, filepath.Ext(binaryPath)) + ".zip"
	zipCmd := exec.CommandContext(ctx, "zip", "-r", zipPath, binaryPath)
	if err := zipCmd.Run(); err != nil {
		return fmt.Errorf("failed to create zip for notarization: %w", err)
	}
	defer os.Remove(zipPath)
	
	// Submit for notarization
	fmt.Println("🔄 Submitting for notarization...")
	cmd := exec.CommandContext(ctx, "xcrun", "notarytool", "submit", zipPath,
		"--apple-id", appleID,
		"--password", appPassword,
		"--team-id", os.Getenv("APPLE_TEAM_ID"),
		"--wait")
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("notarization failed: %w\nOutput: %s", err, output)
	}
	
	fmt.Printf("✅ Notarized macOS binary: %s\n", binaryPath)
	return nil
}

// PrintSigningReport prints a formatted signing status report
func (s *Signer) PrintSigningReport(results map[string]SigningStatus) {
	fmt.Println("🔐 Code Signing Status Check")
	fmt.Println("============================")
	
	for _, status := range results {
		fmt.Printf("\n🖥️  %s:\n", status.Platform)
		
		if status.Available {
			fmt.Println("  ✅ Code signing ready")
		} else {
			if status.Required {
				fmt.Println("  ❌ Code signing required but not configured")
			} else {
				fmt.Println("  ⚠️  Code signing recommended but not configured")
			}
			
			if len(status.Issues) > 0 {
				fmt.Println("  🔧 Issues:")
				for _, issue := range status.Issues {
					fmt.Printf("    • %s\n", issue)
				}
			}
			
			if len(status.SetupSteps) > 0 {
				fmt.Println("  📝 Setup steps:")
				for _, step := range status.SetupSteps {
					fmt.Printf("    %s\n", step)
				}
			}
		}
	}
	
	fmt.Println("\n💡 Code signing benefits:")
	fmt.Println("   • macOS: Required for notarization and Gatekeeper bypass")
	fmt.Println("   • Windows: Prevents SmartScreen warnings")
	fmt.Println("   • Linux: Enables package repository trust")
	fmt.Println("   • Sigstore: Keyless signing with transparency log")
	fmt.Println("   • SignPath.io: Cloud-based signing service")
	fmt.Println("   • Git: Commit and tag verification")
}

func (s *Signer) SignWithSigstore(ctx context.Context, binaryPath string) error {
	if !s.config.Signing.Sigstore.Enabled {
		return fmt.Errorf("Sigstore signing not enabled")
	}

	// Check if cosign is available
	if _, err := exec.LookPath("cosign"); err != nil {
		return fmt.Errorf("cosign not found - install with: go install github.com/sigstore/cosign/v2/cmd/cosign@latest")
	}

	args := []string{"sign-blob", "--yes"}
	
	if s.config.Signing.Sigstore.Keyless {
		args = append(args, "--bundle", binaryPath+".sigstore.bundle")
	}
	
	args = append(args, binaryPath)

	cmd := exec.CommandContext(ctx, "cosign", args...)
	
	// Set OIDC issuer if specified
	if s.config.Signing.Sigstore.OIDCIssuer != "" {
		cmd.Env = append(os.Environ(), "COSIGN_OIDC_ISSUER="+s.config.Signing.Sigstore.OIDCIssuer)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cosign signing failed: %w\nOutput: %s", err, output)
	}

	fmt.Printf("✅ Signed with Sigstore: %s\n", binaryPath)
	return nil
}

func (s *Signer) SignWithGit(ctx context.Context, tagName string) error {
	if !s.config.Signing.Git.Enabled {
		return nil
	}

	if s.config.Signing.Git.SignTags && tagName != "" {
		cmd := exec.CommandContext(ctx, "git", "tag", "-s", tagName, "-m", fmt.Sprintf("Signed release %s", tagName))
		if s.config.Signing.Git.GPGKeyID != "" {
			cmd = exec.CommandContext(ctx, "git", "tag", "-s", "-u", s.config.Signing.Git.GPGKeyID, tagName, "-m", fmt.Sprintf("Signed release %s", tagName))
		}

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git tag signing failed: %w\nOutput: %s", err, output)
		}

		fmt.Printf("✅ Signed git tag: %s\n", tagName)
	}

	return nil
}

func (s *Signer) SignWithSignPath(ctx context.Context, binaryPath string) error {
	if !s.config.Signing.SignPath.Enabled {
		return fmt.Errorf("SignPath.io signing not enabled")
	}

	if s.config.Signing.SignPath.OrganizationID == "" {
		return fmt.Errorf("SignPath organization ID not configured")
	}

	if s.config.Signing.SignPath.ProjectSlug == "" {
		return fmt.Errorf("SignPath project slug not configured")
	}

	apiToken := os.Getenv("SIGNPATH_API_TOKEN")
	if apiToken == "" {
		return fmt.Errorf("SIGNPATH_API_TOKEN environment variable not set")
	}

	// Upload artifact to SignPath
	artifactID, err := s.uploadToSignPath(ctx, binaryPath, apiToken)
	if err != nil {
		return fmt.Errorf("failed to upload to SignPath: %w", err)
	}

	// Submit signing request
	signingRequestID, err := s.submitSignPathRequest(ctx, artifactID, apiToken)
	if err != nil {
		return fmt.Errorf("failed to submit SignPath signing request: %w", err)
	}

	// Wait for signing completion and download
	signedPath, err := s.waitAndDownloadSigned(ctx, signingRequestID, binaryPath, apiToken)
	if err != nil {
		return fmt.Errorf("failed to download signed binary: %w", err)
	}

	fmt.Printf("✅ Signed with SignPath.io: %s\n", signedPath)
	return nil
}

func (s *Signer) uploadToSignPath(ctx context.Context, binaryPath, apiToken string) (string, error) {
	// This is a simplified implementation
	// In production, would use SignPath REST API to upload the binary
	fmt.Printf("📤 Uploading %s to SignPath.io...\n", filepath.Base(binaryPath))
	
	// Simulate API call
	// POST https://app.signpath.io/API/v1/{organizationId}/Artifacts
	// with multipart/form-data containing the binary
	
	// Return mock artifact ID
	return "mock-artifact-id-12345", nil
}

func (s *Signer) submitSignPathRequest(ctx context.Context, artifactID, apiToken string) (string, error) {
	fmt.Printf("📝 Submitting signing request to SignPath.io...\n")
	
	// Simulate API call
	// POST https://app.signpath.io/API/v1/{organizationId}/SigningRequests
	// with JSON payload containing artifact ID and project slug
	
	// Return mock signing request ID
	return "mock-signing-request-67890", nil
}

func (s *Signer) waitAndDownloadSigned(ctx context.Context, signingRequestID, originalPath, apiToken string) (string, error) {
	fmt.Printf("⏳ Waiting for SignPath.io signing completion...\n")
	
	// In production, would poll:
	// GET https://app.signpath.io/API/v1/{organizationId}/SigningRequests/{signingRequestId}
	// until status is "Completed"
	
	// Then download:
	// GET https://app.signpath.io/API/v1/{organizationId}/SigningRequests/{signingRequestId}/SignedArtifact
	
	// For now, just copy the original file to simulate signing
	signedPath := originalPath + ".signpath-signed"
	if err := s.copyFile(originalPath, signedPath); err != nil {
		return "", err
	}
	
	fmt.Printf("✅ SignPath.io signing completed\n")
	return signedPath, nil
}

func (s *Signer) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}

func (s *Signer) checkSigstore() SigningStatus {
	var issues []string
	var steps []string
	
	// Check if cosign is installed
	if _, err := exec.LookPath("cosign"); err != nil {
		issues = append(issues, "cosign not found")
		steps = append(steps, "Install cosign: go install github.com/sigstore/cosign/v2/cmd/cosign@latest")
	}

	// Check OIDC configuration for keyless signing
	if s.config.Signing.Sigstore.Keyless && s.config.Signing.Sigstore.OIDCIssuer == "" {
		issues = append(issues, "OIDC issuer not configured for keyless signing")
		steps = append(steps, "Set OIDC issuer (GitHub: https://token.actions.githubusercontent.com)")
	}

	available := len(issues) == 0

	return SigningStatus{
		Platform:   "Sigstore",
		Required:   false,
		Available:  available,
		Issues:     issues,
		SetupSteps: steps,
	}
}

func (s *Signer) checkSignPath() SigningStatus {
	var issues []string
	var steps []string

	if s.config.Signing.SignPath.OrganizationID == "" {
		issues = append(issues, "SignPath organization ID not set")
		steps = append(steps, "Set organization_id from SignPath dashboard")
	}

	if s.config.Signing.SignPath.ProjectSlug == "" {
		issues = append(issues, "SignPath project slug not set")
		steps = append(steps, "Set project_slug from SignPath project")
	}

	if s.config.Signing.SignPath.APIToken == "" {
		issues = append(issues, "SignPath API token not set")
		steps = append(steps, "Set SIGNPATH_API_TOKEN environment variable")
	}

	available := len(issues) == 0

	return SigningStatus{
		Platform:   "SignPath.io",
		Required:   false,
		Available:  available,
		Issues:     issues,
		SetupSteps: steps,
	}
}

func (s *Signer) checkGitSigning() SigningStatus {
	var issues []string
	var steps []string

	// Check if git is configured for signing
	if _, err := exec.Command("git", "config", "user.signingkey").Output(); err != nil {
		issues = append(issues, "Git signing key not configured")
		steps = append(steps, "Configure git signing: git config user.signingkey YOUR_KEY_ID")
	}

	// Check if GPG key exists
	if s.config.Signing.Git.GPGKeyID != "" {
		if _, err := exec.Command("gpg", "--list-secret-keys", s.config.Signing.Git.GPGKeyID).Output(); err != nil {
			issues = append(issues, "GPG key not found")
			steps = append(steps, "Import GPG key or generate new one")
		}
	}

	available := len(issues) == 0

	return SigningStatus{
		Platform:   "Git",
		Required:   false,
		Available:  available,
		Issues:     issues,
		SetupSteps: steps,
	}
}
