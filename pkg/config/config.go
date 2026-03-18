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

// Package config provides configuration loading and validation for bagboy projects.
package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	validName    = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,62}$`)
	validVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+[a-zA-Z0-9._+-]*$`)
)

// Config is the top-level bagboy project configuration loaded from bagboy.yaml.
type Config struct {
	// Name is the project name used in generated package metadata.
	Name        string            `yaml:"name"`
	// Version is the release version (e.g. "1.2.3").
	Version     string            `yaml:"version"`
	// Description is a short human-readable description of the project.
	Description string            `yaml:"description"`
	// Homepage is the project's canonical URL.
	Homepage    string            `yaml:"homepage"`
	// License is the SPDX license identifier (e.g. "MIT", "Apache-2.0").
	License     string            `yaml:"license"`
	// Author is the primary author or maintainer.
	Author      string            `yaml:"author"`
	// Binaries maps platform identifiers (e.g. "linux-amd64") to binary paths.
	Binaries    map[string]string `yaml:"binaries"`
	// GitHub holds GitHub release and repository settings.
	GitHub      GitHubConfig      `yaml:"github"`
	// Installer holds settings for the curl|bash installer script.
	Installer   InstallerConfig   `yaml:"installer"`
	// Packages holds per-format package configuration.
	Packages     PackagesConfig     `yaml:"packages"`
	// Signing holds code-signing configuration.
	Signing      SigningConfig      `yaml:"signing"`
	// Dependencies holds system and package-manager dependency declarations.
	Dependencies DependenciesConfig `yaml:"dependencies,omitempty"`
	// VM holds VM build environment settings.
	VM           VMConfig           `yaml:"vm,omitempty"`
}

// GitHubConfig holds GitHub API and repository settings.
type GitHubConfig struct {
	// Owner is the GitHub username or organisation that owns the repository.
	Owner    string        `yaml:"owner"`
	// Repo is the GitHub repository name.
	Repo     string        `yaml:"repo"`
	// TokenEnv is the environment variable that holds the GitHub API token.
	TokenEnv string        `yaml:"token_env"`
	// Release holds settings for creating GitHub releases.
	Release  ReleaseConfig `yaml:"release"`
	// Tap holds settings for updating a Homebrew tap repository.
	Tap      TapConfig     `yaml:"tap"`
	// Bucket holds settings for updating a Scoop bucket repository.
	Bucket   BucketConfig  `yaml:"bucket"`
	// Winget holds settings for submitting Winget community manifest PRs.
	Winget   WingetConfig  `yaml:"winget"`
}

// ReleaseConfig controls GitHub release creation behaviour.
type ReleaseConfig struct {
	// Enabled enables automatic GitHub release creation.
	Enabled       bool `yaml:"enabled"`
	// Draft creates the release as a draft.
	Draft         bool `yaml:"draft"`
	// Prerelease marks the release as a pre-release.
	Prerelease    bool `yaml:"prerelease"`
	// GenerateNotes auto-generates release notes from git commits.
	GenerateNotes bool `yaml:"generate_notes"`
}

// TapConfig controls automatic Homebrew tap updates.
type TapConfig struct {
	// Enabled enables tap auto-update after a release.
	Enabled    bool   `yaml:"enabled"`
	// Repo is the tap repository in "owner/repo" format.
	Repo       string `yaml:"repo"`
	// AutoCreate creates the tap repository if it does not exist.
	AutoCreate bool   `yaml:"auto_create"`
	// AutoCommit commits the updated formula automatically.
	AutoCommit bool   `yaml:"auto_commit"`
	// AutoPush pushes the commit to the remote.
	AutoPush   bool   `yaml:"auto_push"`
}

// BucketConfig controls automatic Scoop bucket updates.
type BucketConfig struct {
	// Enabled enables bucket auto-update after a release.
	Enabled    bool   `yaml:"enabled"`
	// Repo is the bucket repository in "owner/repo" format.
	Repo       string `yaml:"repo"`
	// AutoCreate creates the bucket repository if it does not exist.
	AutoCreate bool   `yaml:"auto_create"`
	// AutoCommit commits the updated manifest automatically.
	AutoCommit bool   `yaml:"auto_commit"`
	// AutoPush pushes the commit to the remote.
	AutoPush   bool   `yaml:"auto_push"`
}

// WingetConfig controls Winget community manifest pull-request submission.
type WingetConfig struct {
	// Enabled enables Winget PR submission.
	Enabled  bool   `yaml:"enabled"`
	// AutoPR automatically opens a PR to the winget-pkgs community repository.
	AutoPR   bool   `yaml:"auto_pr"`
	// ForkRepo is the "owner/repo" of the user's winget-pkgs fork.
	ForkRepo string `yaml:"fork_repo"`
}

// InstallerConfig controls the generated curl|bash install script.
type InstallerConfig struct {
	// BaseURL is the download base URL template (supports {{.Version}}).
	BaseURL        string `yaml:"base_url"`
	// InstallPath is the directory where the binary is installed.
	InstallPath    string `yaml:"install_path"`
	// DetectOS enables runtime OS/arch detection in the script.
	DetectOS       bool   `yaml:"detect_os"`
	// VerifyChecksum enables checksum verification in the script.
	VerifyChecksum bool   `yaml:"verify_checksum"`
}

// PackagesConfig holds per-format package generation settings.
type PackagesConfig struct {
	// Brew holds Homebrew-specific settings.
	Brew       BrewConfig       `yaml:"brew"`
	// Scoop holds Scoop-specific settings.
	Scoop      ScoopConfig      `yaml:"scoop"`
	// Chocolatey holds Chocolatey-specific settings.
	Chocolatey ChocolateyConfig `yaml:"chocolatey"`
	// Winget holds Winget package-identifier settings.
	Winget     WingetPkgConfig  `yaml:"winget"`
	// Deb holds Debian package settings.
	Deb        DebConfig        `yaml:"deb"`
	// RPM holds RPM package settings.
	RPM        RPMConfig        `yaml:"rpm"`
	// AppImage holds AppImage settings.
	AppImage   AppImageConfig   `yaml:"appimage"`

	// configured tracks which package-format keys were explicitly present in the YAML.
	configured map[string]bool
}

// UnmarshalYAML implements yaml.Unmarshaler so we can record which format keys
// were explicitly listed under packages: — even when their value is an empty map.
func (p *PackagesConfig) UnmarshalYAML(value *yaml.Node) error {
	// Decode into the struct fields normally.
	type plain PackagesConfig
	if err := value.Decode((*plain)(p)); err != nil {
		return err
	}
	// Walk the mapping node to record every present key.
	p.configured = make(map[string]bool)
	for i := 0; i+1 < len(value.Content); i += 2 {
		p.configured[value.Content[i].Value] = true
	}
	return nil
}

// ConfiguredNames returns the package-format names that were explicitly listed
// under packages: in the config file, in sorted order. Returns nil when the
// packages: section was absent or empty.
func (p *PackagesConfig) ConfiguredNames() []string {
	if len(p.configured) == 0 {
		return nil
	}
	names := make([]string, 0, len(p.configured))
	for name := range p.configured {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// BrewConfig holds Homebrew formula settings.
type BrewConfig struct {
	// Test is the shell command used in the formula's test block.
	Test string `yaml:"test"`
}

// ScoopConfig holds Scoop manifest settings.
type ScoopConfig struct {
	// Bin is the relative path to the executable within the extracted archive.
	Bin       string     `yaml:"bin"`
	// Shortcuts lists Start-Menu shortcut definitions.
	Shortcuts [][]string `yaml:"shortcuts"`
}

// ChocolateyConfig holds Chocolatey package settings.
type ChocolateyConfig struct {
	// PackageSourceURL is the URL to the package source code.
	PackageSourceURL string `yaml:"package_source_url"`
	// DocsURL is the URL to the package documentation.
	DocsURL          string `yaml:"docs_url"`
}

// WingetPkgConfig holds Winget package identifier settings.
type WingetPkgConfig struct {
	// PackageIdentifier is the fully qualified Winget package ID (e.g. "Publisher.AppName").
	PackageIdentifier string `yaml:"package_identifier"`
	// Publisher is the publisher display name.
	Publisher         string `yaml:"publisher"`
	// MinimumOSVersion is the minimum Windows version required.
	MinimumOSVersion  string `yaml:"minimum_os_version"`
}

// DebConfig holds Debian package settings.
type DebConfig struct {
	// Maintainer is the Debian package maintainer in "Name <email>" format.
	Maintainer string `yaml:"maintainer"`
	// Section is the Debian package section (e.g. "utils").
	Section    string `yaml:"section"`
	// Priority is the Debian package priority (e.g. "optional").
	Priority   string `yaml:"priority"`
}

// RPMConfig holds RPM package settings.
type RPMConfig struct {
	// Group is the RPM package group (e.g. "Applications/System").
	Group  string `yaml:"group"`
	// Vendor is the RPM vendor name.
	Vendor string `yaml:"vendor"`
}

// AppImageConfig holds AppImage settings.
type AppImageConfig struct {
	// Categories lists the XDG application categories.
	Categories   []string              `yaml:"categories"`
	// Icon is the path to the application icon.
	Icon         string                `yaml:"icon"`
	// DesktopEntry holds .desktop file settings.
	DesktopEntry AppImageDesktopConfig `yaml:"desktop_entry"`
}

// AppImageDesktopConfig holds AppImage .desktop entry settings.
type AppImageDesktopConfig struct {
	// Terminal indicates whether the application runs in a terminal.
	Terminal bool   `yaml:"terminal"`
	// Type is the XDG application type (typically "Application").
	Type     string `yaml:"type"`
}

// SigningConfig holds code-signing configuration for all platforms.
type SigningConfig struct {
	// MacOS holds macOS notarisation and codesign settings.
	MacOS    MacOSSigningConfig    `yaml:"macos"`
	// Windows holds Windows Authenticode signing settings.
	Windows  WindowsSigningConfig  `yaml:"windows"`
	// Linux holds Linux GPG signing settings.
	Linux    LinuxSigningConfig    `yaml:"linux"`
	// Sigstore holds keyless Sigstore/Cosign signing settings.
	Sigstore SigstoreConfig       `yaml:"sigstore"`
	// SignPath holds SignPath.io signing settings.
	SignPath SignPathConfig       `yaml:"signpath"`
	// Git holds git-commit and tag signing settings.
	Git      GitSigningConfig     `yaml:"git"`
}

// DependenciesConfig represents dependency configuration.
type DependenciesConfig struct {
	// System maps OS names to lists of system package dependencies.
	System          map[string][]string `yaml:"system,omitempty"`
	// PackageManagers maps package manager names to lists of package dependencies.
	PackageManagers map[string][]string `yaml:"package_managers,omitempty"`
	// Runtime maps runtime names to minimum version requirements.
	Runtime         map[string]string   `yaml:"runtime,omitempty"`
}

// MacOSSigningConfig holds macOS codesign and notarisation settings.
type MacOSSigningConfig struct {
	// Identity is the signing certificate identity (e.g. "Developer ID Application: ...").
	Identity     string `yaml:"identity"`
	// Notarize enables Apple notarisation.
	Notarize     bool   `yaml:"notarize"`
	// AppleID is the Apple ID email used for notarisation.
	AppleID      string `yaml:"apple_id"`
	// TeamID is the Apple Developer Team ID.
	TeamID       string `yaml:"team_id"`
	// AppPassword is the app-specific password for notarisation.
	AppPassword  string `yaml:"app_password"`
}

// WindowsSigningConfig holds Windows Authenticode signing settings.
type WindowsSigningConfig struct {
	// CertificateThumbprint is the SHA1 thumbprint of the signing certificate.
	CertificateThumbprint string `yaml:"certificate_thumbprint"`
	// TimestampURL is the RFC 3161 timestamp authority URL.
	TimestampURL          string `yaml:"timestamp_url"`
}

// LinuxSigningConfig holds Linux GPG signing settings.
type LinuxSigningConfig struct {
	// GPGKeyID is the GPG key ID used to sign release artefacts.
	GPGKeyID string `yaml:"gpg_key_id"`
}

// SigstoreConfig holds keyless Sigstore/Cosign signing settings.
type SigstoreConfig struct {
	// Enabled enables Sigstore signing.
	Enabled    bool   `yaml:"enabled"`
	// OIDCIssuer is the OIDC token issuer URL.
	OIDCIssuer string `yaml:"oidc_issuer"`
	// Keyless uses keyless Sigstore signing (no private key required).
	Keyless    bool   `yaml:"keyless"`
}

// SignPathConfig holds SignPath.io signing settings.
type SignPathConfig struct {
	// Enabled enables SignPath.io signing.
	Enabled       bool   `yaml:"enabled"`
	// OrganizationID is the SignPath.io organisation ID.
	OrganizationID string `yaml:"organization_id"`
	// ProjectSlug is the SignPath.io project slug.
	ProjectSlug   string `yaml:"project_slug"`
	// APIToken is the SignPath.io API token.
	APIToken      string `yaml:"api_token"`
}

// GitSigningConfig holds git commit and tag signing settings.
type GitSigningConfig struct {
	// Enabled enables git signing.
	Enabled   bool   `yaml:"enabled"`
	// GPGKeyID is the GPG key ID used to sign commits and tags.
	GPGKeyID  string `yaml:"gpg_key_id"`
	// SignTags enables signing of git tags.
	SignTags  bool   `yaml:"sign_tags"`
	// SignCommits enables signing of git commits.
	SignCommits bool `yaml:"sign_commits"`
}

// VMConfig holds VM build environment settings.
type VMConfig struct {
	// Enabled enables VM-based builds.
	Enabled  bool         `yaml:"enabled"`
	// Provider is the VM provider to use ("docker", "multipass", or "vagrant").
	Provider string       `yaml:"provider"`
	// Docker holds Docker-specific VM settings.
	Docker   DockerVMConfig `yaml:"docker"`
}

// DockerVMConfig holds Docker-based VM build settings.
type DockerVMConfig struct {
	// Image is the Docker image used for builds.
	Image string `yaml:"image"`
}

// Load reads and parses the bagboy configuration file at path.
// ctx is reserved for future cancellation support.
func Load(_ context.Context, path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the configuration contains all required fields.
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !validName.MatchString(c.Name) {
		return fmt.Errorf("validation failed: name %q contains invalid characters (must match ^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,62}$)", c.Name)
	}
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if !validVersion.MatchString(c.Version) {
		return fmt.Errorf("validation failed: version %q is not a valid semver-like string", c.Version)
	}
	if len(c.Binaries) == 0 {
		return fmt.Errorf("at least one binary is required")
	}
	c.validateSecrets()
	return nil
}

func (c *Config) validateSecrets() {
	if c.Signing.MacOS.AppPassword != "" && !strings.HasPrefix(c.Signing.MacOS.AppPassword, "$") {
		slog.Warn("signing.macos.app_password appears to be a plaintext secret; consider using an env-var reference like $APPLE_APP_PASSWORD")
	}
	if c.Signing.SignPath.APIToken != "" && !strings.HasPrefix(c.Signing.SignPath.APIToken, "$") {
		slog.Warn("signing.signpath.api_token appears to be a plaintext secret; consider using an env-var reference like $SIGNPATH_API_TOKEN")
	}
}

// FindConfigFile searches the current directory for a bagboy configuration file
// and returns its absolute path. It checks bagboy.yaml, bagboy.yml, .bagboy.yaml,
// and .bagboy.yml in that order.
func FindConfigFile() (string, error) {
	candidates := []string{"bagboy.yaml", "bagboy.yml", ".bagboy.yaml", ".bagboy.yml"}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			abs, err := filepath.Abs(candidate)
			if err != nil {
				return candidate, nil
			}
			return abs, nil
		}
	}

	return "", fmt.Errorf("no bagboy config file found")
}
