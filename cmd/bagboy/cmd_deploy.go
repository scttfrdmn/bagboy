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
	"strings"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/bagboy/pkg/cicd"
	_ "github.com/scttfrdmn/bagboy/pkg/cicd/github"
	_ "github.com/scttfrdmn/bagboy/pkg/cicd/gitlab"
	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/deploy"
	"github.com/scttfrdmn/bagboy/pkg/ui"
)

// newDeployCmd returns the cobra Command for the 'deploy' subcommand.
func newDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy packages to repositories and registries",
		RunE:  runDeploy,
	}
	cmd.Flags().StringSlice("targets", []string{}, "Deployment targets (brew,npm,docker,etc)")
	cmd.Flags().Bool("dry-run", false, "Show deployment instructions without executing")
	cmd.Flags().String("generate-ci", "", "Generate a CI/CD pipeline file (github or gitlab)")
	return cmd
}

// runDeploy implements the 'deploy' subcommand logic.
func runDeploy(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	targets, _ := cmd.Flags().GetStringSlice("targets")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	generateCI, _ := cmd.Flags().GetString("generate-ci")

	// CI/CD pipeline generation takes priority.
	if generateCI != "" {
		return runGenerateCI(generateCI)
	}

	if len(targets) == 0 {
		deployer := deploy.NewDeployer(nil)
		deploymentTargets := deployer.GetDeploymentTargets()
		fmt.Println("📦 Available Deployment Targets:")
		fmt.Println("================================")
		for _, target := range deploymentTargets {
			fmt.Printf("\n🎯 %s (%s)\n", target.Name, target.Format)
			fmt.Printf("   %s\n", target.Description)
		}
		fmt.Println("\nUsage: bagboy deploy --targets brew,npm,docker")
		fmt.Println("\nTo generate a CI/CD pipeline:")
		fmt.Println("  bagboy deploy --generate-ci github")
		fmt.Println("  bagboy deploy --generate-ci gitlab")
		return nil
	}

	configPath, err := config.FindConfigFile()
	if err != nil {
		return err
	}

	cfg, err := config.Load(ctx, configPath)
	if err != nil {
		return err
	}

	deployer := deploy.NewDeployer(cfg)
	return deployer.Deploy(ctx, targets, dryRun)
}

// runGenerateCI generates CI/CD pipeline files using the named provider.
func runGenerateCI(providerName string) error {
	providerName = strings.ToLower(strings.TrimSpace(providerName))

	provider, err := cicd.Get(providerName)
	if err != nil {
		return err
	}

	// Try to load config; it's optional for pipeline generation.
	var cfg *config.Config
	if configPath, lerr := config.FindConfigFile(); lerr == nil {
		if loaded, lerr2 := config.Load(context.Background(), configPath); lerr2 == nil {
			cfg = loaded
		}
	}

	files, err := provider.GeneratePipeline(cfg)
	if err != nil {
		return fmt.Errorf("generating %s pipeline: %w", providerName, err)
	}

	ui.Header(fmt.Sprintf("Generating %s CI/CD Pipeline", strings.ToUpper(providerName)))

	for _, f := range files {
		// Ensure parent directory exists.
		dir := filepath.Dir(f.Path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dir, err)
			}
		}

		// Check if file already exists and confirm overwrite.
		if _, statErr := os.Stat(f.Path); statErr == nil {
			if !ui.Confirm(fmt.Sprintf("%s already exists. Overwrite?", f.Path)) {
				ui.Info(fmt.Sprintf("Skipped %s", f.Path))
				continue
			}
		}

		if err := os.WriteFile(f.Path, []byte(f.Content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", f.Path, err)
		}
		ui.Success(fmt.Sprintf("Created %s", f.Path))
	}

	fmt.Println()
	ui.Info("Review the generated file(s) and commit them to your repository.")
	return nil
}
