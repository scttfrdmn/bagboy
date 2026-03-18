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

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/deploy"
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
	return cmd
}

// runDeploy implements the 'deploy' subcommand logic.
func runDeploy(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	targets, _ := cmd.Flags().GetStringSlice("targets")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

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
