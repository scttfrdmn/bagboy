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
	"strings"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/packager"
	"github.com/scttfrdmn/bagboy/pkg/signing"
	"github.com/scttfrdmn/bagboy/pkg/ui"
)

// newPackCmd returns the cobra Command for the 'pack' subcommand.
func newPackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pack",
		Aliases: []string{"p", "package", "build"},
		Short:   "Create packages for distribution",
		Long: `Create packages for various platforms and package managers.

Supports 20+ package formats including:
• Package Managers: Homebrew, Scoop, Chocolatey, Winget
• Linux Packages: DEB, RPM, AppImage, Snap, Flatpak
• Containers: Docker, Apptainer
• Language Packages: npm, PyPI, Cargo, Nix, Spack
• Platform Installers: DMG, MSI, MSIX, curl|bash

Examples:
  bagboy pack --all              # Create all supported formats
  bagboy pack --brew --scoop     # Create Homebrew and Scoop packages
  bagboy pack --deb --rpm        # Create Linux packages
  bagboy pack --docker --sign    # Create Docker image with signing`,
		RunE: runPack,
	}

	cmd.Flags().Bool("all", false, "Create all package types")
	cmd.Flags().Bool("sign", false, "Sign binaries before packaging")
	cmd.Flags().Bool("brew", false, "Create Homebrew formula")
	cmd.Flags().Bool("scoop", false, "Create Scoop manifest")
	cmd.Flags().Bool("deb", false, "Create DEB package")
	cmd.Flags().Bool("rpm", false, "Create RPM package")
	cmd.Flags().Bool("chocolatey", false, "Create Chocolatey package")
	cmd.Flags().Bool("winget", false, "Create Winget manifests")
	cmd.Flags().Bool("snap", false, "Create Snap package")
	cmd.Flags().Bool("appimage", false, "Create AppImage")
	cmd.Flags().Bool("flatpak", false, "Create Flatpak manifest")
	cmd.Flags().Bool("npm", false, "Create npm package")
	cmd.Flags().Bool("pypi", false, "Create PyPI package")
	cmd.Flags().Bool("docker", false, "Create Docker files")
	cmd.Flags().Bool("apptainer", false, "Create Apptainer container")
	cmd.Flags().Bool("dmg", false, "Create macOS DMG installer")
	cmd.Flags().Bool("msi", false, "Create Windows MSI installer")
	cmd.Flags().Bool("msix", false, "Create Windows MSIX package")
	cmd.Flags().Bool("cargo", false, "Create Rust Cargo package")
	cmd.Flags().Bool("nix", false, "Create Nix package")
	cmd.Flags().Bool("spack", false, "Create Spack package")
	cmd.Flags().Bool("installer", false, "Create curl|bash installer")

	return cmd
}

// runPack implements the 'pack' subcommand logic.
func runPack(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	all, _ := cmd.Flags().GetBool("all")
	sign, _ := cmd.Flags().GetBool("sign")

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

	// Sign binaries first if requested.
	if sign {
		if err := runSign(ctx, cfg); err != nil {
			fmt.Printf("⚠️  Some signing operations failed: %v\n", err)
		}
	}

	if all {
		return runPackAll(ctx, cfg, reg)
	}
	return runPackFlags(ctx, cmd, cfg, reg)
}

// runSign signs binaries using all configured signing providers.
func runSign(ctx context.Context, cfg *config.Config) error {
	fmt.Println("🔐 Signing binaries...")
	signer := signing.NewSigner(cfg)
	if err := signer.SignAllBinaries(ctx); err != nil {
		fmt.Printf("⚠️  Signing failed: %v\n", err)
	}

	if cfg.Signing.Sigstore.Enabled {
		for arch, binaryPath := range cfg.Binaries {
			fmt.Printf("Signing %s with Sigstore...\n", arch)
			if err := signer.SignWithSigstore(ctx, binaryPath); err != nil {
				fmt.Printf("⚠️  Sigstore signing failed for %s: %v\n", arch, err)
			}
		}
	}

	if cfg.Signing.SignPath.Enabled {
		for arch, binaryPath := range cfg.Binaries {
			if strings.HasPrefix(arch, "windows-") {
				fmt.Printf("Signing %s with SignPath.io...\n", arch)
				if err := signer.SignWithSignPath(ctx, binaryPath); err != nil {
					fmt.Printf("⚠️  SignPath.io signing failed for %s: %v\n", arch, err)
				}
			}
		}
	}
	return nil
}

// runPackAll packs all registered formats and prints a results table.
func runPackAll(ctx context.Context, cfg *config.Config, reg *packager.Registry) error {
	ui.Header("Creating All Package Formats")
	progress := ui.NewProgressBar(reg.Count(), "📦 Packaging")
	results, err := reg.PackAll(ctx, cfg)
	progress.Finish()
	if err != nil {
		return err
	}
	ui.Success(fmt.Sprintf("Created %d packages", len(results)))
	table := ui.NewTable([]string{"Format", "Output Path", "Status"})
	for name, path := range results {
		status := "✅ Success"
		if path == "" {
			status = "⚠️  Skipped"
		}
		table.AddRow([]string{name, path, status})
	}
	table.Print()
	return nil
}

// runPackFlags iterates the boolean flags and packs only the requested formats.
// Using a map dispatch reduces cyclomatic complexity from ~21 to ~3.
func runPackFlags(ctx context.Context, cmd *cobra.Command, cfg *config.Config, reg *packager.Registry) error {
	flagToName := map[string]string{
		"brew":       "brew",
		"scoop":      "scoop",
		"deb":        "deb",
		"rpm":        "rpm",
		"chocolatey": "chocolatey",
		"winget":     "winget",
		"snap":       "snap",
		"appimage":   "appimage",
		"flatpak":    "flatpak",
		"npm":        "npm",
		"pypi":       "pypi",
		"docker":     "docker",
		"apptainer":  "apptainer",
		"dmg":        "dmg",
		"msi":        "msi",
		"msix":       "msix",
		"cargo":      "cargo",
		"nix":        "nix",
		"spack":      "spack",
		"installer":  "installer",
	}
	for flag, name := range flagToName {
		enabled, _ := cmd.Flags().GetBool(flag)
		if !enabled {
			continue
		}
		p, ok := reg.Get(name)
		if !ok {
			continue
		}
		output, err := p.Pack(ctx, cfg)
		if err != nil {
			return err
		}
		fmt.Printf("✅ Created %s: %s\n", name, output)
	}
	return nil
}
