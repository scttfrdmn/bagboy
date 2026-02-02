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

package deploy

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

// Deployer handles deployment of packages to various repositories
type Deployer struct {
	cfg *config.Config
}

// NewDeployer creates a new deployer
func NewDeployer(cfg *config.Config) *Deployer {
	return &Deployer{cfg: cfg}
}

// DeploymentTarget represents a deployment destination
type DeploymentTarget struct {
	Name        string
	Format      string
	Command     string
	Description string
	Instructions []string
}

// GetDeploymentTargets returns available deployment targets
func (d *Deployer) GetDeploymentTargets() []DeploymentTarget {
	return []DeploymentTarget{
		{
			Name:        "Homebrew Tap",
			Format:      "brew",
			Description: "Deploy to your Homebrew tap repository",
			Instructions: []string{
				"1. Create tap repository: gh repo create homebrew-tap",
				"2. Add Formula/ directory to your tap repo",
				"3. Copy generated .rb file to Formula/",
				"4. Commit and push to GitHub",
				"5. Users install with: brew install yourname/tap/appname",
			},
		},
		{
			Name:        "Scoop Bucket",
			Format:      "scoop",
			Description: "Deploy to your Scoop bucket repository",
			Instructions: []string{
				"1. Create bucket repository: gh repo create scoop-bucket",
				"2. Add bucket/ directory to your repo",
				"3. Copy generated .json file to bucket/",
				"4. Commit and push to GitHub",
				"5. Users add bucket: scoop bucket add yourname https://github.com/yourname/scoop-bucket",
				"6. Users install with: scoop install appname",
			},
		},
		{
			Name:        "npm Registry",
			Format:      "npm",
			Description: "Deploy to npm registry",
			Instructions: []string{
				"1. Login to npm: npm login",
				"2. Navigate to generated npm package directory",
				"3. Publish: npm publish",
				"4. Users install with: npm install -g appname",
			},
		},
		{
			Name:        "PyPI",
			Format:      "pypi",
			Description: "Deploy to Python Package Index",
			Instructions: []string{
				"1. Install twine: pip install twine",
				"2. Build package: python setup.py sdist bdist_wheel",
				"3. Upload: twine upload dist/*",
				"4. Users install with: pip install appname",
			},
		},
		{
			Name:        "Crates.io",
			Format:      "cargo",
			Description: "Deploy to Rust package registry",
			Instructions: []string{
				"1. Login to crates.io: cargo login",
				"2. Navigate to generated cargo package directory",
				"3. Publish: cargo publish",
				"4. Users install with: cargo install appname",
			},
		},
		{
			Name:        "Docker Hub",
			Format:      "docker",
			Description: "Deploy Docker image to Docker Hub",
			Instructions: []string{
				"1. Login to Docker Hub: docker login",
				"2. Navigate to generated docker directory",
				"3. Build image: docker build -t yourname/appname:version .",
				"4. Push image: docker push yourname/appname:version",
				"5. Users run with: docker run yourname/appname",
			},
		},
		{
			Name:        "GitHub Releases",
			Format:      "github",
			Description: "Deploy packages as GitHub release assets",
			Instructions: []string{
				"1. Create GitHub release: gh release create v1.0.0",
				"2. Upload packages: gh release upload v1.0.0 dist/*",
				"3. Users download from GitHub releases page",
			},
		},
		{
			Name:        "Snap Store",
			Format:      "snap",
			Description: "Deploy to Ubuntu Snap Store",
			Instructions: []string{
				"1. Register app name: snapcraft register appname",
				"2. Build snap: snapcraft",
				"3. Upload: snapcraft upload appname.snap",
				"4. Release: snapcraft release appname revision stable",
				"5. Users install with: snap install appname",
			},
		},
	}
}

// Deploy executes deployment for specified targets
func (d *Deployer) Deploy(ctx context.Context, targets []string, dryRun bool) error {
	deploymentTargets := d.GetDeploymentTargets()
	
	for _, target := range targets {
		found := false
		for _, dt := range deploymentTargets {
			if dt.Format == target || dt.Name == target {
				found = true
				if dryRun {
					fmt.Printf("🔍 Would deploy %s (%s)\n", dt.Name, dt.Format)
					d.printInstructions(dt)
				} else {
					fmt.Printf("🚀 Deploying %s...\n", dt.Name)
					if err := d.executeDeploy(ctx, dt); err != nil {
						return fmt.Errorf("deployment failed for %s: %w", dt.Name, err)
					}
				}
				break
			}
		}
		
		if !found {
			return fmt.Errorf("unknown deployment target: %s", target)
		}
	}
	
	return nil
}

func (d *Deployer) printInstructions(target DeploymentTarget) {
	fmt.Printf("📋 %s Deployment Instructions:\n", target.Name)
	for _, instruction := range target.Instructions {
		fmt.Printf("   %s\n", instruction)
	}
	fmt.Println()
}

func (d *Deployer) executeDeploy(ctx context.Context, target DeploymentTarget) error {
	switch target.Format {
	case "npm":
		return d.deployNpm(ctx)
	case "docker":
		return d.deployDocker(ctx)
	case "github":
		return d.deployGitHub(ctx)
	default:
		// For most targets, we provide instructions rather than automated deployment
		fmt.Printf("📋 Manual deployment required for %s:\n", target.Name)
		d.printInstructions(target)
		return nil
	}
}

func (d *Deployer) deployNpm(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "npm", "publish", "dist/npm")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm publish failed: %w\nOutput: %s", err, output)
	}
	fmt.Printf("✅ Published to npm: %s\n", strings.TrimSpace(string(output)))
	return nil
}

func (d *Deployer) deployDocker(ctx context.Context) error {
	// Build Docker image
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", 
		fmt.Sprintf("%s:%s", d.cfg.Name, d.cfg.Version), "dist/docker")
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}
	
	// Push Docker image (requires docker login)
	pushCmd := exec.CommandContext(ctx, "docker", "push", 
		fmt.Sprintf("%s:%s", d.cfg.Name, d.cfg.Version))
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}
	
	fmt.Printf("✅ Pushed Docker image: %s:%s\n", d.cfg.Name, d.cfg.Version)
	return nil
}

func (d *Deployer) deployGitHub(ctx context.Context) error {
	// Create GitHub release using gh CLI
	releaseCmd := exec.CommandContext(ctx, "gh", "release", "create", 
		"v"+d.cfg.Version, "dist/*", "--title", "v"+d.cfg.Version)
	output, err := releaseCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("github release failed: %w\nOutput: %s", err, output)
	}
	
	fmt.Printf("✅ Created GitHub release: %s\n", strings.TrimSpace(string(output)))
	return nil
}
