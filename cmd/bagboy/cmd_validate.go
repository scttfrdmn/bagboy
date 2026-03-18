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
	"github.com/scttfrdmn/bagboy/pkg/errors"
	"github.com/scttfrdmn/bagboy/pkg/ui"
)

// newValidateCmd returns the cobra Command for the 'validate' subcommand.
func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "validate",
		Aliases: []string{"v", "verify"},
		Short:   "Validate bagboy configuration",
		Long: `Validate your bagboy.yaml configuration file.

Checks for:
• Valid YAML syntax
• Required fields (name, version, binaries)
• Binary file existence
• GitHub repository access (if configured)
• Package format compatibility

Examples:
  bagboy validate               # Validate current configuration
  bagboy validate --verbose     # Show detailed validation info`,
		RunE: runValidate,
	}
}

// runValidate implements the 'validate' subcommand logic.
func runValidate(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	ui.Header("Validating Configuration")

	configPath, err := config.FindConfigFile()
	if err != nil {
		ui.Error("No bagboy configuration file found")
		ui.Info("Run 'bagboy init' to create a new configuration")
		return errors.NewConfigurationError("CONFIG_NOT_FOUND", "No bagboy configuration file found",
			"Run 'bagboy init' to create a new configuration",
			"Ensure bagboy.yaml exists in the current directory")
	}

	if verbose {
		ui.Info(fmt.Sprintf("Found config file: %s", configPath))
	}

	cfg, err := config.Load(ctx, configPath)
	if err != nil {
		ui.Error("Failed to load configuration file")
		return errors.WrapError(err, "Failed to load configuration file",
			"Check the syntax of your bagboy.yaml file",
			"Run 'bagboy init' to regenerate the configuration")
	}

	if err := cfg.Validate(); err != nil {
		ui.Error("Configuration validation failed")
		return errors.WrapError(err, "Configuration validation failed",
			"Fix the issues in your bagboy.yaml file",
			"Run 'bagboy init' to regenerate with correct structure")
	}

	ui.Success("Configuration is valid")

	if verbose {
		ui.Info(fmt.Sprintf("Project: %s v%s", cfg.Name, cfg.Version))
		ui.Info(fmt.Sprintf("Binaries: %d configured", len(cfg.Binaries)))
		if cfg.GitHub.Owner != "" {
			ui.Info(fmt.Sprintf("GitHub: %s/%s", cfg.GitHub.Owner, cfg.GitHub.Repo))
		}
	}

	return nil
}
