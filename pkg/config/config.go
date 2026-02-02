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

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Homepage    string            `yaml:"homepage"`
	License     string            `yaml:"license"`
	Author      string            `yaml:"author"`
	Binaries    map[string]string `yaml:"binaries"`
	GitHub      GitHubConfig      `yaml:"github"`
	Installer   InstallerConfig   `yaml:"installer"`
	Packages     PackagesConfig     `yaml:"packages"`
	Signing      SigningConfig      `yaml:"signing"`
	Dependencies DependenciesConfig `yaml:"dependencies,omitempty"`
}

type GitHubConfig struct {
	Owner    string        `yaml:"owner"`
	Repo     string        `yaml:"repo"`
	TokenEnv string        `yaml:"token_env"`
	Release  ReleaseConfig `yaml:"release"`
	Tap      TapConfig     `yaml:"tap"`
	Bucket   BucketConfig  `yaml:"bucket"`
	Winget   WingetConfig  `yaml:"winget"`
}

type ReleaseConfig struct {
	Enabled       bool `yaml:"enabled"`
	Draft         bool `yaml:"draft"`
	Prerelease    bool `yaml:"prerelease"`
	GenerateNotes bool `yaml:"generate_notes"`
}

type TapConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Repo       string `yaml:"repo"`
	AutoCreate bool   `yaml:"auto_create"`
	AutoCommit bool   `yaml:"auto_commit"`
	AutoPush   bool   `yaml:"auto_push"`
}

type BucketConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Repo       string `yaml:"repo"`
	AutoCreate bool   `yaml:"auto_create"`
	AutoCommit bool   `yaml:"auto_commit"`
	AutoPush   bool   `yaml:"auto_push"`
}

type WingetConfig struct {
	Enabled  bool   `yaml:"enabled"`
	AutoPR   bool   `yaml:"auto_pr"`
	ForkRepo string `yaml:"fork_repo"`
}

type InstallerConfig struct {
	BaseURL        string `yaml:"base_url"`
	InstallPath    string `yaml:"install_path"`
	DetectOS       bool   `yaml:"detect_os"`
	VerifyChecksum bool   `yaml:"verify_checksum"`
}

type PackagesConfig struct {
	Brew       BrewConfig       `yaml:"brew"`
	Scoop      ScoopConfig      `yaml:"scoop"`
	Chocolatey ChocolateyConfig `yaml:"chocolatey"`
	Winget     WingetPkgConfig  `yaml:"winget"`
	Deb        DebConfig        `yaml:"deb"`
	RPM        RPMConfig        `yaml:"rpm"`
	AppImage   AppImageConfig   `yaml:"appimage"`
}

type BrewConfig struct {
	Test string `yaml:"test"`
}

type ScoopConfig struct {
	Bin       string     `yaml:"bin"`
	Shortcuts [][]string `yaml:"shortcuts"`
}

type ChocolateyConfig struct {
	PackageSourceURL string `yaml:"package_source_url"`
	DocsURL          string `yaml:"docs_url"`
}

type WingetPkgConfig struct {
	PackageIdentifier string `yaml:"package_identifier"`
	Publisher         string `yaml:"publisher"`
	MinimumOSVersion  string `yaml:"minimum_os_version"`
}

type DebConfig struct {
	Maintainer string `yaml:"maintainer"`
	Section    string `yaml:"section"`
	Priority   string `yaml:"priority"`
}

type RPMConfig struct {
	Group  string `yaml:"group"`
	Vendor string `yaml:"vendor"`
}

type AppImageConfig struct {
	Categories   []string              `yaml:"categories"`
	Icon         string                `yaml:"icon"`
	DesktopEntry AppImageDesktopConfig `yaml:"desktop_entry"`
}

type AppImageDesktopConfig struct {
	Terminal bool   `yaml:"terminal"`
	Type     string `yaml:"type"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if len(c.Binaries) == 0 {
		return fmt.Errorf("at least one binary is required")
	}
	return nil
}

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

type SigningConfig struct {
	MacOS    MacOSSigningConfig    `yaml:"macos"`
	Windows  WindowsSigningConfig  `yaml:"windows"`
	Linux    LinuxSigningConfig    `yaml:"linux"`
	Sigstore SigstoreConfig       `yaml:"sigstore"`
	SignPath SignPathConfig       `yaml:"signpath"`
	Git      GitSigningConfig     `yaml:"git"`
}

// DependenciesConfig represents dependency configuration
type DependenciesConfig struct {
	System          map[string][]string `yaml:"system,omitempty"`
	PackageManagers map[string][]string `yaml:"package_managers,omitempty"`
	Runtime         map[string]string   `yaml:"runtime,omitempty"`
}

type MacOSSigningConfig struct {
	Identity     string `yaml:"identity"`
	Notarize     bool   `yaml:"notarize"`
	AppleID      string `yaml:"apple_id"`
	TeamID       string `yaml:"team_id"`
	AppPassword  string `yaml:"app_password"`
}

type WindowsSigningConfig struct {
	CertificateThumbprint string `yaml:"certificate_thumbprint"`
	TimestampURL          string `yaml:"timestamp_url"`
}

type LinuxSigningConfig struct {
	GPGKeyID string `yaml:"gpg_key_id"`
}

type SigstoreConfig struct {
	Enabled    bool   `yaml:"enabled"`
	OIDCIssuer string `yaml:"oidc_issuer"`
	Keyless    bool   `yaml:"keyless"`
}

type SignPathConfig struct {
	Enabled       bool   `yaml:"enabled"`
	OrganizationID string `yaml:"organization_id"`
	ProjectSlug   string `yaml:"project_slug"`
	APIToken      string `yaml:"api_token"`
}

type GitSigningConfig struct {
	Enabled   bool   `yaml:"enabled"`
	GPGKeyID  string `yaml:"gpg_key_id"`
	SignTags  bool   `yaml:"sign_tags"`
	SignCommits bool `yaml:"sign_commits"`
}
