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
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/github"
	"github.com/scttfrdmn/bagboy/pkg/ui"
)

// newPublishCmd returns the cobra Command for the 'publish' subcommand.
func newPublishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "publish",
		Aliases: []string{"pub", "release", "deploy"},
		Short:   "Pack all formats and create GitHub release",
		Long: `Complete publishing workflow: pack, release, and distribute.

This command will:
• Create all package formats
• Create GitHub release with assets
• Update Homebrew tap (if configured)
• Update Scoop bucket (if configured)
• Submit Winget PR (if configured)

Examples:
  bagboy publish                # Full publish workflow
  bagboy publish --dry-run      # Preview what would happen
  bagboy publish --skip-github  # Skip GitHub operations`,
		RunE: runPublish,
	}
	cmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	cmd.Flags().Bool("skip-github", false, "Skip GitHub operations (release, tap, bucket)")
	return cmd
}

// runPublish implements the 'publish' subcommand logic.
func runPublish(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	skipGitHub, _ := cmd.Flags().GetBool("skip-github")

	if dryRun {
		ui.Warning("DRY RUN MODE - No changes will be made")
	}

	ui.PrintBanner()
	ui.Header("Publishing Workflow")

	configPath, err := config.FindConfigFile()
	if err != nil {
		return err
	}

	cfg, err := config.Load(ctx, configPath)
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	reg := newPackagerRegistry()
	configuredNames := cfg.Packages.ConfiguredNames()

	if dryRun {
		ui.Info("Would create packages for:")
		if len(configuredNames) > 0 {
			for _, format := range configuredNames {
				ui.Info(fmt.Sprintf("  • %s", format))
			}
		} else {
			for _, format := range reg.List() {
				ui.Info(fmt.Sprintf("  • %s", format))
			}
		}
		if !skipGitHub && cfg.GitHub.Owner != "" {
			ui.Info("Would create GitHub release and update repositories")
		}
		return nil
	}

	fmt.Println("🚀 Publishing", cfg.Name, cfg.Version)

	// Determine which packager names to run.
	namesToPack := configuredNames
	if len(namesToPack) == 0 {
		namesToPack = reg.List()
	}

	// Pack each format; warn and continue on error so one missing build tool
	// (e.g. WiX on macOS) does not abort the entire publish workflow.
	results := make(map[string]string)
	for _, name := range namesToPack {
		p, ok := reg.Get(name)
		if !ok {
			continue
		}
		if err := p.Validate(cfg); err != nil {
			fmt.Printf("⚠️  Skipping %s: %v\n", name, err)
			continue
		}
		output, err := p.Pack(ctx, cfg)
		if err != nil {
			fmt.Printf("⚠️  Skipping %s: %v\n", name, err)
			continue
		}
		results[name] = output
	}

	fmt.Println("✅ Created packages:")
	var assets []string
	for name, path := range results {
		fmt.Printf("  %s: %s\n", name, path)
		assets = append(assets, path)
	}

	if !cfg.GitHub.Release.Enabled {
		fmt.Println("\n🎉 Publish complete!")
		return nil
	}

	client, err := github.NewClient(&cfg.GitHub)
	if err != nil {
		fmt.Printf("⚠️  GitHub integration disabled: %v\n", err)
		return nil
	}

	release, err := client.CreateRelease(ctx, cfg, assets)
	if err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}
	fmt.Printf("✅ Created GitHub release: %s\n", release.GetHTMLURL())

	if cfg.GitHub.Tap.Enabled {
		if err := client.UpdateTap(ctx, cfg, results["brew"]); err != nil {
			fmt.Printf("⚠️  Failed to update tap: %v\n", err)
		} else {
			fmt.Printf("✅ Updated Homebrew tap: %s\n", cfg.GitHub.Tap.Repo)
		}
	}

	if cfg.GitHub.Bucket.Enabled {
		if err := client.UpdateBucket(ctx, cfg, results["scoop"]); err != nil {
			fmt.Printf("⚠️  Failed to update bucket: %v\n", err)
		} else {
			fmt.Printf("✅ Updated Scoop bucket: %s\n", cfg.GitHub.Bucket.Repo)
		}
	}

	if cfg.GitHub.Winget.Enabled && cfg.GitHub.Winget.AutoPR {
		if err := submitWingetPR(ctx, client, cfg, results); err != nil {
			fmt.Printf("⚠️  Failed to submit Winget PR: %v\n", err)
		}
	}

	fmt.Println("\n🎉 Publish complete!")
	return nil
}

// submitWingetPR reads winget manifest files and submits a pull request.
func submitWingetPR(ctx context.Context, client *github.Client, cfg *config.Config, results map[string]string) error {
	wingetResult, exists := results["winget"]
	if !exists || wingetResult == "" {
		return nil
	}
	manifests := make(map[string]string)
	for _, suffix := range []string{"", ".installer", ".locale.en-US"} {
		filename := fmt.Sprintf("%s%s.yaml", cfg.Packages.Winget.PackageIdentifier, suffix)
		content, err := os.ReadFile(filepath.Join(wingetResult, filename))
		if err == nil {
			manifests[filename] = string(content)
		}
	}
	if len(manifests) == 0 {
		return nil
	}
	return client.SubmitWingetPR(ctx, cfg, manifests)
}
