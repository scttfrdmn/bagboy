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
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "bagboy",
	Short: "Universal software packager",
	Long: `🎒 bagboy - Universal Software Packager

Pack once. Ship everywhere.

bagboy creates packages for all major platforms from a single configuration file.
Supports 20+ package formats including Homebrew, Scoop, DEB, RPM, Docker, and more.

Examples:
  bagboy init                    # Initialize new project
  bagboy pack --all              # Create all package formats
  bagboy pack --brew --deb       # Create specific formats
  bagboy publish                 # Pack and publish to registries
  bagboy sign --check            # Check code signing setup
  bagboy benchmark               # Run performance benchmarks

Learn more: https://bagboy.dev`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")

	rootCmd.AddCommand(
		newInitCmd(),
		newPackCmd(),
		newPublishCmd(),
		newCheckCmd(),
		newDeployCmd(),
		newSignCmd(),
		newValidateCmd(),
		newBenchmarkCmd(),
		newDepsCmd(),
		newVMCmd(),
		newVersionCmd(),
	)
}

// initLogger configures the default slog handler based on --verbose/--debug flags.
func initLogger() {
	level := slog.LevelWarn
	if verbose {
		level = slog.LevelInfo
	}
	if debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}
