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
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/bagboy/pkg/requirements"
)

// newCheckCmd returns the cobra Command for the 'check' subcommand.
func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check system requirements for package formats",
		RunE:  runCheck,
	}
	cmd.Flags().StringSlice("formats", []string{}, "Package formats to check (default: all)")
	return cmd
}

// runCheck implements the 'check' subcommand logic.
func runCheck(cmd *cobra.Command, _ []string) error {
	formats, _ := cmd.Flags().GetStringSlice("formats")
	if len(formats) == 0 {
		formats = []string{"brew", "scoop", "deb", "rpm", "dmg", "msi", "docker", "snap", "appimage"}
	}

	checker := requirements.NewRequirementChecker()
	results := checker.CheckRequirements(formats)
	checker.PrintRequirementReport(results)
	return nil
}
