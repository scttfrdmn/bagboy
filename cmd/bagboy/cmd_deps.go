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
	"github.com/scttfrdmn/bagboy/pkg/deps"
	"github.com/scttfrdmn/bagboy/pkg/ui"
)

// newDepsCmd returns the cobra Command for the 'deps' subcommand.
func newDepsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "Manage dependencies",
		Long: `Manage project dependencies across platforms and package managers.

Supports system dependencies, package manager dependencies, and runtime requirements.
Automatically detects the appropriate package manager for your platform.

Examples:
  bagboy deps check          # Check all dependencies
  bagboy deps list           # List configured dependencies
  bagboy deps install        # Install missing dependencies
  bagboy deps resolve        # Resolve dependency conflicts`,
	}
	cmd.AddCommand(
		newDepsCheckCmd(),
		newDepsListCmd(),
		newDepsInstallCmd(),
		newDepsResolveCmd(),
	)
	return cmd
}

// newDepsCheckCmd returns the cobra Command for the 'deps check' subcommand.
func newDepsCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check dependency status",
		RunE:  runDepsCheck,
	}
}

// runDepsCheck implements the 'deps check' subcommand logic.
func runDepsCheck(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	configPath, err := config.FindConfigFile()
	if err != nil {
		return err
	}

	cfg, err := config.Load(ctx, configPath)
	if err != nil {
		return err
	}

	manager := deps.NewManager(cfg)
	ui.Header("Checking Dependencies")

	results, err := manager.Check(ctx)
	if err != nil {
		return err
	}

	table := ui.NewTable([]string{"Dependency", "Status", "Version"})
	allAvailable := true

	for name, status := range results {
		statusStr := "❌ Missing"
		if status.Available {
			if status.Satisfies {
				statusStr = "✅ Available"
			} else {
				statusStr = "⚠️  Wrong Version"
				allAvailable = false
			}
		} else {
			allAvailable = false
		}
		table.AddRow([]string{name, statusStr, status.Version})
	}

	table.Print()

	if allAvailable {
		ui.Success("All dependencies are satisfied")
	} else {
		ui.Warning("Some dependencies are missing or incorrect")
		ui.Info("Run 'bagboy deps install' to install missing dependencies")
	}
	return nil
}

// newDepsListCmd returns the cobra Command for the 'deps list' subcommand.
func newDepsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured dependencies",
		RunE:  runDepsList,
	}
}

// runDepsList implements the 'deps list' subcommand logic.
func runDepsList(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	configPath, err := config.FindConfigFile()
	if err != nil {
		return err
	}

	cfg, err := config.Load(ctx, configPath)
	if err != nil {
		return err
	}

	manager := deps.NewManager(cfg)
	dependencies := manager.List()

	if len(dependencies) == 0 {
		ui.Info("No dependencies configured")
		return nil
	}

	ui.Header("Configured Dependencies")
	table := ui.NewTable([]string{"Name", "Type", "Platform/Manager", "Version"})

	for _, dep := range dependencies {
		platformOrManager := dep.Platform
		if dep.PackageManager != "" {
			platformOrManager = dep.PackageManager
		}
		table.AddRow([]string{dep.Name, dep.Type, platformOrManager, dep.Version})
	}

	table.Print()
	return nil
}

// newDepsInstallCmd returns the cobra Command for the 'deps install' subcommand.
func newDepsInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install missing dependencies",
		RunE:  runDepsInstall,
	}
}

// runDepsInstall implements the 'deps install' subcommand logic.
func runDepsInstall(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	configPath, err := config.FindConfigFile()
	if err != nil {
		return err
	}

	cfg, err := config.Load(ctx, configPath)
	if err != nil {
		return err
	}

	manager := deps.NewManager(cfg)
	ui.Header("Installing Dependencies")

	results, err := manager.Check(ctx)
	if err != nil {
		return err
	}

	var missing []string
	for name, status := range results {
		if !status.Available {
			missing = append(missing, name)
		}
	}

	if len(missing) == 0 {
		ui.Success("All dependencies are already installed")
		return nil
	}

	ui.Info(fmt.Sprintf("Installing %d missing dependencies...", len(missing)))

	if err := manager.Install(ctx, missing); err != nil {
		return err
	}

	ui.Success("Dependencies installed successfully")
	return nil
}

// newDepsResolveCmd returns the cobra Command for the 'deps resolve' subcommand.
func newDepsResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve",
		Short: "Resolve dependency conflicts",
		RunE:  runDepsResolve,
	}
}

// runDepsResolve implements the 'deps resolve' subcommand logic.
func runDepsResolve(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	configPath, err := config.FindConfigFile()
	if err != nil {
		return err
	}

	cfg, err := config.Load(ctx, configPath)
	if err != nil {
		return err
	}

	manager := deps.NewManager(cfg)
	ui.Header("Resolving Dependencies")

	result, err := manager.Resolve(ctx)
	if err != nil {
		return err
	}

	if len(result.Conflicts) > 0 {
		ui.Warning(fmt.Sprintf("Found %d dependency conflicts", len(result.Conflicts)))
		table := ui.NewTable([]string{"Dependency", "Conflicting Versions", "Reason"})
		for _, conflict := range result.Conflicts {
			versions := ""
			for i, v := range conflict.Versions {
				if i > 0 {
					versions += ", "
				}
				versions += v
			}
			table.AddRow([]string{conflict.Dependency, versions, conflict.Reason})
		}
		table.Print()
	} else {
		ui.Success("No dependency conflicts found")
	}

	ui.Info(fmt.Sprintf("Resolved %d dependencies", len(result.Resolved)))

	if err := manager.WriteLockFile(ctx, "bagboy.lock"); err != nil {
		ui.Warning(fmt.Sprintf("Failed to write lock file: %v", err))
	} else {
		ui.Success("Generated bagboy.lock file")
	}

	return nil
}
