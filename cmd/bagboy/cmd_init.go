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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/scttfrdmn/bagboy/pkg/config"
	initpkg "github.com/scttfrdmn/bagboy/pkg/init"
	"github.com/scttfrdmn/bagboy/pkg/ui"
)

// newInitCmd returns the cobra Command for the 'init' subcommand.
func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"i", "new", "create"},
		Short:   "Initialize a new bagboy project",
		Long: `Initialize a new bagboy project with smart detection.

Automatically detects:
• Project type (Go, Node.js, Rust, Python)
• Project metadata (name, version, description)
• GitHub repository information
• Existing binary locations

Examples:
  bagboy init                    # Auto-detect project settings
  bagboy init --interactive      # Interactive configuration
  bagboy init --name myapp       # Override detected name`,
		RunE: runInit,
	}
	cmd.Flags().BoolP("interactive", "i", false, "Interactive mode")
	return cmd
}

// runInit implements the 'init' subcommand logic.
func runInit(cmd *cobra.Command, _ []string) error {
	interactive, _ := cmd.Flags().GetBool("interactive")

	ui.PrintBanner()
	ui.Info("Initializing bagboy project...")

	info, err := initpkg.DetectProject()
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	if interactive {
		fmt.Println("\nDetected project information:")
		if err := initpkg.PromptUser(info); err != nil {
			return err
		}
	}

	cfg := &config.Config{
		Name:        info.Name,
		Version:     info.Version,
		Description: info.Description,
		Author:      info.Author,
		Homepage:    info.Homepage,
		License:     info.License,
		Binaries:    info.Binaries,
		GitHub: config.GitHubConfig{
			Owner:    info.GitHubOwner,
			Repo:     info.GitHubRepo,
			TokenEnv: "GITHUB_TOKEN",
			Release: config.ReleaseConfig{
				Enabled:       true,
				GenerateNotes: true,
			},
			Tap: config.TapConfig{
				Enabled:    true,
				Repo:       fmt.Sprintf("%s/homebrew-tap", info.GitHubOwner),
				AutoCreate: true,
				AutoCommit: true,
				AutoPush:   true,
			},
			Bucket: config.BucketConfig{
				Enabled:    true,
				Repo:       fmt.Sprintf("%s/scoop-bucket", info.GitHubOwner),
				AutoCreate: true,
				AutoCommit: true,
				AutoPush:   true,
			},
		},
		Installer: config.InstallerConfig{
			BaseURL:        fmt.Sprintf("https://github.com/%s/%s/releases/download/v{{.Version}}", info.GitHubOwner, info.GitHubRepo),
			InstallPath:    "/usr/local/bin",
			DetectOS:       true,
			VerifyChecksum: true,
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("bagboy.yaml", data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("✅ Created bagboy.yaml")

	ui.Header("Next Steps")
	fmt.Println("1. Review and customize bagboy.yaml")
	fmt.Println("2. Build your binaries for target platforms")
	fmt.Println("3. Run 'bagboy pack --all' to create packages")
	fmt.Println("4. Run 'bagboy publish' to distribute everywhere")
	fmt.Println()
	ui.Info("Learn more at https://bagboy.dev")

	return nil
}
