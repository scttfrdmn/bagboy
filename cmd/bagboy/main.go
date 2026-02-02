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
	"github.com/scttfrdmn/bagboy/pkg/benchmark"
	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/deploy"
	"github.com/scttfrdmn/bagboy/pkg/deps"
	"github.com/scttfrdmn/bagboy/pkg/errors"
	"github.com/scttfrdmn/bagboy/pkg/requirements"
	"github.com/scttfrdmn/bagboy/pkg/signing"
	"github.com/scttfrdmn/bagboy/pkg/ui"
	"github.com/scttfrdmn/bagboy/pkg/github"
	initpkg "github.com/scttfrdmn/bagboy/pkg/init"
	"github.com/scttfrdmn/bagboy/pkg/packager"
	"github.com/scttfrdmn/bagboy/pkg/packager/appimage"
	"github.com/scttfrdmn/bagboy/pkg/packager/apptainer"
	"github.com/scttfrdmn/bagboy/pkg/packager/brew"
	"github.com/scttfrdmn/bagboy/pkg/packager/cargo"
	"github.com/scttfrdmn/bagboy/pkg/packager/chocolatey"
	"github.com/scttfrdmn/bagboy/pkg/packager/deb"
	"github.com/scttfrdmn/bagboy/pkg/packager/dmg"
	"github.com/scttfrdmn/bagboy/pkg/packager/docker"
	"github.com/scttfrdmn/bagboy/pkg/packager/flatpak"
	"github.com/scttfrdmn/bagboy/pkg/packager/installer"
	"github.com/scttfrdmn/bagboy/pkg/packager/msi"
	"github.com/scttfrdmn/bagboy/pkg/packager/msix"
	"github.com/scttfrdmn/bagboy/pkg/packager/nix"
	"github.com/scttfrdmn/bagboy/pkg/packager/npm"
	"github.com/scttfrdmn/bagboy/pkg/packager/pypi"
	"github.com/scttfrdmn/bagboy/pkg/packager/rpm"
	"github.com/scttfrdmn/bagboy/pkg/packager/scoop"
	"github.com/scttfrdmn/bagboy/pkg/packager/snap"
	"github.com/scttfrdmn/bagboy/pkg/packager/spack"
	"github.com/scttfrdmn/bagboy/pkg/packager/winget"
	"gopkg.in/yaml.v3"
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
	SilenceErrors: true,  // We handle errors ourselves
	SilenceUsage:  true,  // Don't show usage on errors
}

var initCmd = &cobra.Command{
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
	RunE: func(cmd *cobra.Command, args []string) error {
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
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Review and edit bagboy.yaml")
		fmt.Println("  2. Build your binaries")
		fmt.Println("  3. Run 'bagboy pack --all' to create packages")

		return nil
	},
}

var packCmd = &cobra.Command{
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
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		sign, _ := cmd.Flags().GetBool("sign")
		brewFlag, _ := cmd.Flags().GetBool("brew")
		scoopFlag, _ := cmd.Flags().GetBool("scoop")
		debFlag, _ := cmd.Flags().GetBool("deb")
		rpmFlag, _ := cmd.Flags().GetBool("rpm")
		chocolateyFlag, _ := cmd.Flags().GetBool("chocolatey")
		wingetFlag, _ := cmd.Flags().GetBool("winget")
		snapFlag, _ := cmd.Flags().GetBool("snap")
		appimageFlag, _ := cmd.Flags().GetBool("appimage")
		flatpakFlag, _ := cmd.Flags().GetBool("flatpak")
		npmFlag, _ := cmd.Flags().GetBool("npm")
		pypiFlag, _ := cmd.Flags().GetBool("pypi")
		dockerFlag, _ := cmd.Flags().GetBool("docker")
		apptainerFlag, _ := cmd.Flags().GetBool("apptainer")
		dmgFlag, _ := cmd.Flags().GetBool("dmg")
		msiFlag, _ := cmd.Flags().GetBool("msi")
		msixFlag, _ := cmd.Flags().GetBool("msix")
		cargoFlag, _ := cmd.Flags().GetBool("cargo")
		nixFlag, _ := cmd.Flags().GetBool("nix")
		spackFlag, _ := cmd.Flags().GetBool("spack")
		installerFlag, _ := cmd.Flags().GetBool("installer")

		configPath, err := config.FindConfigFile()
		if err != nil {
			return err
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

		registry := packager.NewRegistry()
		registry.Register(brew.New())
		registry.Register(scoop.New())
		registry.Register(deb.New())
		registry.Register(rpm.New())
		registry.Register(chocolatey.New())
		registry.Register(winget.New())
		registry.Register(snap.New())
		registry.Register(appimage.New())
		registry.Register(flatpak.New())
		registry.Register(npm.New())
		registry.Register(pypi.New())
		registry.Register(docker.New())
		registry.Register(apptainer.New())
		registry.Register(dmg.New())
		registry.Register(msi.New())
		registry.Register(msix.New())
		registry.Register(cargo.New())
		registry.Register(nix.New())
		registry.Register(spack.New())
		registry.Register(installer.New())

		ctx := context.Background()

		// Sign binaries first if requested
		if sign {
			fmt.Println("🔐 Signing binaries...")
			signer := signing.NewSigner(cfg)
			if err := signer.SignAllBinaries(ctx); err != nil {
				fmt.Printf("⚠️  Signing failed: %v\n", err)
				// Continue with packaging even if signing fails
			}

			// Sigstore signing if enabled
			if cfg.Signing.Sigstore.Enabled {
				for arch, binaryPath := range cfg.Binaries {
					fmt.Printf("Signing %s with Sigstore...\n", arch)
					if err := signer.SignWithSigstore(ctx, binaryPath); err != nil {
						fmt.Printf("⚠️  Sigstore signing failed for %s: %v\n", arch, err)
					}
				}
			}

			// SignPath.io signing if enabled
			if cfg.Signing.SignPath.Enabled {
				for arch, binaryPath := range cfg.Binaries {
					// Only sign Windows binaries with SignPath.io (typical use case)
					if strings.HasPrefix(arch, "windows-") {
						fmt.Printf("Signing %s with SignPath.io...\n", arch)
						if err := signer.SignWithSignPath(ctx, binaryPath); err != nil {
							fmt.Printf("⚠️  SignPath.io signing failed for %s: %v\n", arch, err)
						}
					}
				}
			}
		}

		if all {
			ui.Header("Creating All Package Formats")
			
			// Get total count for progress
			totalPackagers := registry.Count()
			progress := ui.NewProgressBar(totalPackagers, "📦 Packaging")
			
			results, err := registry.PackAll(ctx, cfg)
			progress.Finish()
			
			if err != nil {
				return err
			}

			ui.Success(fmt.Sprintf("Created %d packages", len(results)))
			
			// Display results in a table
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

		// Individual packagers
		if brewFlag {
			if p, ok := registry.Get("brew"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created brew formula: %s\n", output)
			}
		}

		if scoopFlag {
			if p, ok := registry.Get("scoop"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created scoop manifest: %s\n", output)
			}
		}

		if debFlag {
			if p, ok := registry.Get("deb"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created deb package: %s\n", output)
			}
		}

		if rpmFlag {
			if p, ok := registry.Get("rpm"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created rpm package: %s\n", output)
			}
		}

		if chocolateyFlag {
			if p, ok := registry.Get("chocolatey"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created chocolatey package: %s\n", output)
			}
		}

		if wingetFlag {
			if p, ok := registry.Get("winget"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created winget manifests: %s\n", output)
			}
		}

		if snapFlag {
			if p, ok := registry.Get("snap"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created snap package: %s\n", output)
			}
		}

		if appimageFlag {
			if p, ok := registry.Get("appimage"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created appimage: %s\n", output)
			}
		}

		if flatpakFlag {
			if p, ok := registry.Get("flatpak"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created flatpak manifest: %s\n", output)
			}
		}

		if npmFlag {
			if p, ok := registry.Get("npm"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created npm package: %s\n", output)
			}
		}

		if pypiFlag {
			if p, ok := registry.Get("pypi"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created pypi package: %s\n", output)
			}
		}

		if dockerFlag {
			if p, ok := registry.Get("docker"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created docker files: %s\n", output)
			}
		}

		if apptainerFlag {
			if p, ok := registry.Get("apptainer"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created apptainer container: %s\n", output)
			}
		}

		if dmgFlag {
			if p, ok := registry.Get("dmg"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created dmg installer: %s\n", output)
			}
		}

		if msiFlag {
			if p, ok := registry.Get("msi"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created msi installer: %s\n", output)
			}
		}

		if msixFlag {
			if p, ok := registry.Get("msix"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created msix package: %s\n", output)
			}
		}

		if cargoFlag {
			if p, ok := registry.Get("cargo"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created cargo package: %s\n", output)
			}
		}

		if nixFlag {
			if p, ok := registry.Get("nix"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created nix package: %s\n", output)
			}
		}

		if spackFlag {
			if p, ok := registry.Get("spack"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created spack package: %s\n", output)
			}
		}

		if installerFlag {
			if p, ok := registry.Get("installer"); ok {
				output, err := p.Pack(ctx, cfg)
				if err != nil {
					return err
				}
				fmt.Printf("✅ Created installer script: %s\n", output)
			}
		}

		return nil
	},
}

var publishCmd = &cobra.Command{
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
	RunE: func(cmd *cobra.Command, args []string) error {
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

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

		if dryRun {
			ui.Info("Would create packages for:")
			for _, format := range []string{"brew", "scoop", "deb", "rpm", "docker"} {
				ui.Info(fmt.Sprintf("  • %s", format))
			}
			if !skipGitHub && cfg.GitHub.Owner != "" {
				ui.Info("Would create GitHub release and update repositories")
			}
			return nil
		}

		fmt.Println("🚀 Publishing", cfg.Name, cfg.Version)

		// Create packages
		registry := packager.NewRegistry()
		registry.Register(brew.New())
		registry.Register(scoop.New())
		registry.Register(deb.New())
		registry.Register(rpm.New())
		registry.Register(chocolatey.New())
		registry.Register(winget.New())
		registry.Register(snap.New())
		registry.Register(appimage.New())
		registry.Register(flatpak.New())
		registry.Register(npm.New())
		registry.Register(pypi.New())
		registry.Register(docker.New())
		registry.Register(apptainer.New())
		registry.Register(dmg.New())
		registry.Register(msi.New())
		registry.Register(msix.New())
		registry.Register(cargo.New())
		registry.Register(nix.New())
		registry.Register(installer.New())
		registry.Register(spack.New())
		ctx := context.Background()
		results, err := registry.PackAll(ctx, cfg)
		if err != nil {
			return err
		}

		fmt.Println("✅ Created packages:")
		var assets []string
		for name, path := range results {
			fmt.Printf("  %s: %s\n", name, path)
			assets = append(assets, path)
		}

		if dryRun {
			fmt.Println("🔍 Dry run - would create GitHub release with assets:", assets)
			return nil
		}

		// Create GitHub release
		if cfg.GitHub.Release.Enabled {
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

			// Update tap and bucket
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

			// Submit Winget PR
			if cfg.GitHub.Winget.Enabled && cfg.GitHub.Winget.AutoPR {
				fmt.Println("Submitting Winget PR...")
				wingetResult, exists := results["winget"]
				if exists && wingetResult != "" {
					// Read all manifest files from the winget output directory
					manifests := make(map[string]string)
					manifestFiles := []string{
						fmt.Sprintf("%s.yaml", cfg.Packages.Winget.PackageIdentifier),
						fmt.Sprintf("%s.installer.yaml", cfg.Packages.Winget.PackageIdentifier),
						fmt.Sprintf("%s.locale.en-US.yaml", cfg.Packages.Winget.PackageIdentifier),
					}
					
					for _, filename := range manifestFiles {
						manifestPath := filepath.Join(wingetResult, filename)
						if content, err := os.ReadFile(manifestPath); err == nil {
							manifests[filename] = string(content)
						}
					}
					
					if len(manifests) > 0 {
						if err := client.SubmitWingetPR(ctx, cfg, manifests); err != nil {
							fmt.Printf("⚠️  Failed to submit Winget PR: %v\n", err)
						}
					}
				}
			}
		}

		fmt.Println("\n🎉 Publish complete!")
		return nil
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check system requirements for package formats",
	RunE: func(cmd *cobra.Command, args []string) error {
		formats, _ := cmd.Flags().GetStringSlice("formats")
		if len(formats) == 0 {
			formats = []string{"brew", "scoop", "deb", "rpm", "dmg", "msi", "docker", "snap", "appimage"}
		}
		
		checker := requirements.NewRequirementChecker()
		results := checker.CheckRequirements(formats)
		checker.PrintRequirementReport(results)
		
		return nil
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy packages to repositories and registries",
	RunE: func(cmd *cobra.Command, args []string) error {
		targets, _ := cmd.Flags().GetStringSlice("targets")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		
		if len(targets) == 0 {
			// Show available targets
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
		
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		
		deployer := deploy.NewDeployer(cfg)
		ctx := context.Background()
		
		return deployer.Deploy(ctx, targets, dryRun)
	},
}

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Check code signing setup and sign binaries",
	RunE: func(cmd *cobra.Command, args []string) error {
		checkOnly, _ := cmd.Flags().GetBool("check")
		binaryPath, _ := cmd.Flags().GetString("binary")
		
		configPath, err := config.FindConfigFile()
		if err != nil && !checkOnly {
			return err
		}
		
		var cfg *config.Config
		if configPath != "" {
			cfg, err = config.Load(configPath)
			if err != nil {
				return err
			}
		}
		
		signer := signing.NewSigner(cfg)
		
		if checkOnly || binaryPath == "" {
			// Check signing setup
			results := signer.CheckSigningSetup()
			signer.PrintSigningReport(results)
			return nil
		}
		
		// Sign specific binary
		ctx := context.Background()
		return signer.SignBinary(ctx, binaryPath)
	},
}

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"v", "check", "verify"},
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
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		
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

		cfg, err := config.Load(configPath)
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
	},
}

func init() {
	initCmd.Flags().BoolP("interactive", "i", false, "Interactive mode")

	validateCmd.Flags().BoolP("verbose", "v", false, "Show detailed validation information")

	packCmd.Flags().Bool("all", false, "Create all package types")
	packCmd.Flags().Bool("sign", false, "Sign binaries before packaging")
	packCmd.Flags().Bool("brew", false, "Create Homebrew formula")
	packCmd.Flags().Bool("scoop", false, "Create Scoop manifest")
	packCmd.Flags().Bool("deb", false, "Create DEB package")
	packCmd.Flags().Bool("rpm", false, "Create RPM package")
	packCmd.Flags().Bool("chocolatey", false, "Create Chocolatey package")
	packCmd.Flags().Bool("winget", false, "Create Winget manifests")
	packCmd.Flags().Bool("snap", false, "Create Snap package")
	packCmd.Flags().Bool("appimage", false, "Create AppImage")
	packCmd.Flags().Bool("flatpak", false, "Create Flatpak manifest")
	packCmd.Flags().Bool("npm", false, "Create npm package")
	packCmd.Flags().Bool("pypi", false, "Create PyPI package")
	packCmd.Flags().Bool("docker", false, "Create Docker files")
	packCmd.Flags().Bool("apptainer", false, "Create Apptainer container")
	packCmd.Flags().Bool("dmg", false, "Create macOS DMG installer")
	packCmd.Flags().Bool("msi", false, "Create Windows MSI installer")
	packCmd.Flags().Bool("msix", false, "Create Windows MSIX package")
	packCmd.Flags().Bool("cargo", false, "Create Rust Cargo package")
	packCmd.Flags().Bool("nix", false, "Create Nix package")
	packCmd.Flags().Bool("spack", false, "Create Spack package")
	packCmd.Flags().Bool("installer", false, "Create curl|bash installer")

	publishCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	publishCmd.Flags().Bool("skip-github", false, "Skip GitHub operations (release, tap, bucket)")
	
	checkCmd.Flags().StringSlice("formats", []string{}, "Package formats to check (default: all)")
	
	deployCmd.Flags().StringSlice("targets", []string{}, "Deployment targets (brew,npm,docker,etc)")
	deployCmd.Flags().Bool("dry-run", false, "Show deployment instructions without executing")
	
	signCmd.Flags().Bool("check", false, "Check signing setup only")
	signCmd.Flags().String("binary", "", "Path to binary to sign")

	var benchmarkCmd = &cobra.Command{
		Use:   "benchmark",
		Short: "Run performance benchmarks",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.FindConfigFile()
			if err != nil {
				return err
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			fmt.Println("🚀 Running bagboy performance benchmarks...")
			
			// Run basic benchmark suite
			results := benchmark.RunBenchmarkSuite(cfg)
			benchmark.PrintBenchmarkResults(results)
			
			return nil
		},
	}

	var depsCmd = &cobra.Command{
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

	var depsCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Check dependency status",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.FindConfigFile()
			if err != nil {
				return err
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			manager := deps.NewManager(cfg)
			ctx := context.Background()
			
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
		},
	}

	var depsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List configured dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.FindConfigFile()
			if err != nil {
				return err
			}

			cfg, err := config.Load(configPath)
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
				
				table.AddRow([]string{
					dep.Name,
					dep.Type,
					platformOrManager,
					dep.Version,
				})
			}
			
			table.Print()
			
			return nil
		},
	}

	var depsInstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Install missing dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.FindConfigFile()
			if err != nil {
				return err
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			manager := deps.NewManager(cfg)
			ctx := context.Background()
			
			ui.Header("Installing Dependencies")
			
			// Check which dependencies are missing
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
		},
	}

	depsCmd.AddCommand(depsCheckCmd)
	depsCmd.AddCommand(depsListCmd)
	depsCmd.AddCommand(depsInstallCmd)

	var versionCmd = &cobra.Command{
		Use:     "version",
		Aliases: []string{"v", "--version"},
		Short:   "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintVersion("0.7.0-dev", "", "")
			return nil
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(benchmarkCmd)
	rootCmd.AddCommand(depsCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Enhanced error handling with recovery suggestions
		if bagboyErr, ok := err.(*errors.BagboyError); ok {
			ui.Error(bagboyErr.Message)
			if len(bagboyErr.Suggestions) > 0 {
				ui.Info("💡 " + bagboyErr.Suggestions[0])
			}
			if bagboyErr.Details != "" {
				ui.Info("📋 " + bagboyErr.Details)
			}
		} else {
			ui.Error(err.Error())
			ui.Info("💡 Run 'bagboy --help' for usage information")
		}
		os.Exit(1)
	}
}
