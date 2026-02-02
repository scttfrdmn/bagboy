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
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestNewDeployer(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	
	deployer := NewDeployer(cfg)
	
	if deployer == nil {
		t.Fatal("NewDeployer returned nil")
	}
	
	if deployer.cfg != cfg {
		t.Error("Deployer config not set correctly")
	}
}

func TestGetDeploymentTargets(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	
	targets := deployer.GetDeploymentTargets()
	
	if len(targets) == 0 {
		t.Fatal("No deployment targets returned")
	}
	
	// Verify expected targets are present
	expectedTargets := map[string]string{
		"Homebrew Tap":    "brew",
		"Scoop Bucket":    "scoop",
		"npm Registry":    "npm",
		"PyPI":           "pypi",
		"Crates.io":      "cargo",
		"Docker Hub":     "docker",
		"GitHub Releases": "github",
		"Snap Store":     "snap",
	}
	
	targetMap := make(map[string]DeploymentTarget)
	for _, target := range targets {
		targetMap[target.Name] = target
	}
	
	for expectedName, expectedFormat := range expectedTargets {
		target, exists := targetMap[expectedName]
		if !exists {
			t.Errorf("Expected deployment target '%s' not found", expectedName)
			continue
		}
		
		if target.Format != expectedFormat {
			t.Errorf("Target '%s' has format '%s', expected '%s'", expectedName, target.Format, expectedFormat)
		}
		
		if target.Description == "" {
			t.Errorf("Target '%s' has empty description", expectedName)
		}
		
		if len(target.Instructions) == 0 {
			t.Errorf("Target '%s' has no instructions", expectedName)
		}
	}
}

func TestDeploymentTarget_Structure(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	
	targets := deployer.GetDeploymentTargets()
	
	for _, target := range targets {
		t.Run(target.Name, func(t *testing.T) {
			if target.Name == "" {
				t.Error("Target name is empty")
			}
			
			if target.Format == "" {
				t.Error("Target format is empty")
			}
			
			if target.Description == "" {
				t.Error("Target description is empty")
			}
			
			if len(target.Instructions) == 0 {
				t.Error("Target has no instructions")
			}
			
			// Verify instructions are meaningful
			for i, instruction := range target.Instructions {
				if strings.TrimSpace(instruction) == "" {
					t.Errorf("Instruction %d is empty or whitespace", i)
				}
			}
		})
	}
}

func TestDeploy_DryRun(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	// Test dry run with known targets
	targets := []string{"brew", "npm", "docker"}
	
	err := deployer.Deploy(ctx, targets, true)
	if err != nil {
		t.Errorf("Dry run deployment failed: %v", err)
	}
}

func TestDeploy_UnknownTarget(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	// Test with unknown target
	targets := []string{"unknown-target"}
	
	err := deployer.Deploy(ctx, targets, true)
	if err == nil {
		t.Error("Expected error for unknown deployment target")
	}
	
	if !strings.Contains(err.Error(), "unknown deployment target") {
		t.Errorf("Expected 'unknown deployment target' error, got: %v", err)
	}
}

func TestDeploy_EmptyTargets(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	// Test with empty targets
	err := deployer.Deploy(ctx, []string{}, true)
	if err != nil {
		t.Errorf("Deploy with empty targets should not fail: %v", err)
	}
}

func TestDeploy_MultipleTargets(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	// Test with multiple valid targets
	targets := []string{"brew", "scoop", "pypi"}
	
	err := deployer.Deploy(ctx, targets, true)
	if err != nil {
		t.Errorf("Deploy with multiple targets failed: %v", err)
	}
}

func TestDeploy_TargetByName(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	// Test targeting by name instead of format
	targets := []string{"Homebrew Tap", "Docker Hub"}
	
	err := deployer.Deploy(ctx, targets, true)
	if err != nil {
		t.Errorf("Deploy with target names failed: %v", err)
	}
}

func TestExecuteDeploy_ManualTargets(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	// Test manual deployment targets (should not fail, just print instructions)
	manualTargets := []DeploymentTarget{
		{Name: "Test Target", Format: "manual", Instructions: []string{"Step 1", "Step 2"}},
	}
	
	for _, target := range manualTargets {
		err := deployer.executeDeploy(ctx, target)
		if err != nil {
			t.Errorf("Manual deployment should not fail: %v", err)
		}
	}
}

func TestExecuteDeploy_NPM(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	target := DeploymentTarget{
		Name:   "npm Registry",
		Format: "npm",
	}
	
	// Create mock npm directory
	os.MkdirAll("dist/npm", 0755)
	defer os.RemoveAll("dist")
	
	err := deployer.executeDeploy(ctx, target)
	
	// This will likely fail because npm publish requires authentication
	// But we're testing that the command is attempted correctly
	if err != nil {
		// Check if it's the expected npm error (not a code error)
		if !strings.Contains(err.Error(), "npm publish failed") {
			t.Errorf("Unexpected error type: %v", err)
		}
	}
}

func TestExecuteDeploy_Docker(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	target := DeploymentTarget{
		Name:   "Docker Hub",
		Format: "docker",
	}
	
	// Create mock docker directory with Dockerfile
	os.MkdirAll("dist/docker", 0755)
	os.WriteFile("dist/docker/Dockerfile", []byte("FROM alpine\nCOPY testapp /usr/local/bin/\n"), 0644)
	defer os.RemoveAll("dist")
	
	err := deployer.executeDeploy(ctx, target)
	
	// This will likely fail if Docker is not available or not logged in
	// But we're testing that the command structure is correct
	if err != nil {
		// Check if it's a Docker-related error (not a code error)
		if !strings.Contains(err.Error(), "docker") {
			t.Errorf("Expected Docker-related error, got: %v", err)
		}
	}
}

func TestExecuteDeploy_GitHub(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	target := DeploymentTarget{
		Name:   "GitHub Releases",
		Format: "github",
	}
	
	// Create mock dist directory
	os.MkdirAll("dist", 0755)
	os.WriteFile("dist/testfile.txt", []byte("test content"), 0644)
	defer os.RemoveAll("dist")
	
	err := deployer.executeDeploy(ctx, target)
	
	// This will likely fail if gh CLI is not available or not authenticated
	// But we're testing that the command structure is correct
	if err != nil {
		// Check if it's a GitHub-related error (not a code error)
		if !strings.Contains(err.Error(), "github release failed") {
			t.Errorf("Expected GitHub release error, got: %v", err)
		}
	}
}

func TestPrintInstructions(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	
	target := DeploymentTarget{
		Name: "Test Target",
		Instructions: []string{
			"Step 1: Do something",
			"Step 2: Do something else",
		},
	}
	
	// This test just ensures printInstructions doesn't panic
	// In a real scenario, you might capture stdout to verify output
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printInstructions panicked: %v", r)
		}
	}()
	
	deployer.printInstructions(target)
}

func TestDeploymentWorkflow_Integration(t *testing.T) {
	cfg := &config.Config{
		Name:        "testapp",
		Version:     "1.0.0",
		Description: "Test application",
		Author:      "Test Author",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	// Test complete workflow: get targets -> deploy (dry run)
	targets := deployer.GetDeploymentTargets()
	if len(targets) == 0 {
		t.Fatal("No deployment targets available")
	}
	
	// Test deploying first few targets in dry run mode
	targetFormats := []string{targets[0].Format, targets[1].Format}
	
	err := deployer.Deploy(ctx, targetFormats, true)
	if err != nil {
		t.Errorf("Integration workflow failed: %v", err)
	}
}

func TestDeploymentTargets_Coverage(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	
	targets := deployer.GetDeploymentTargets()
	
	// Verify we have comprehensive coverage of deployment targets
	formats := make(map[string]bool)
	for _, target := range targets {
		formats[target.Format] = true
	}
	
	expectedFormats := []string{
		"brew", "scoop", "npm", "pypi", "cargo", "docker", "github", "snap",
	}
	
	for _, format := range expectedFormats {
		if !formats[format] {
			t.Errorf("Missing deployment target for format: %s", format)
		}
	}
	
	// Verify each target has meaningful instructions
	for _, target := range targets {
		if len(target.Instructions) < 2 {
			t.Errorf("Target %s has insufficient instructions (%d)", target.Name, len(target.Instructions))
		}
		
		// Check that instructions contain actionable steps
		hasActionableStep := false
		for _, instruction := range target.Instructions {
			if strings.Contains(instruction, ":") || strings.Contains(instruction, "install") || 
			   strings.Contains(instruction, "create") || strings.Contains(instruction, "upload") {
				hasActionableStep = true
				break
			}
		}
		
		if !hasActionableStep {
			t.Errorf("Target %s instructions don't contain actionable steps", target.Name)
		}
	}
}

func TestDeploymentError_Handling(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	ctx := context.Background()
	
	// Test error handling for mixed valid/invalid targets
	targets := []string{"brew", "invalid-target", "npm"}
	
	err := deployer.Deploy(ctx, targets, true)
	if err == nil {
		t.Error("Expected error for invalid target in list")
	}
	
	if !strings.Contains(err.Error(), "invalid-target") {
		t.Errorf("Error should mention the invalid target: %v", err)
	}
}

func TestCommandAvailability(t *testing.T) {
	// Test helper function to check if commands are available
	commands := map[string]string{
		"npm":    "npm",
		"docker": "docker", 
		"gh":     "gh",
	}
	
	for name, cmd := range commands {
		t.Run(name, func(t *testing.T) {
			_, err := exec.LookPath(cmd)
			if err != nil {
				t.Logf("Command %s not available: %v", cmd, err)
			} else {
				t.Logf("Command %s is available", cmd)
			}
		})
	}
}

func TestDeploymentTarget_Validation(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	
	targets := deployer.GetDeploymentTargets()
	
	for _, target := range targets {
		t.Run(target.Format, func(t *testing.T) {
			// Validate target structure
			if target.Name == "" {
				t.Error("Target name cannot be empty")
			}
			
			if target.Format == "" {
				t.Error("Target format cannot be empty")
			}
			
			if target.Description == "" {
				t.Error("Target description cannot be empty")
			}
			
			// Validate instructions quality
			if len(target.Instructions) == 0 {
				t.Error("Target must have instructions")
			}
			
			for i, instruction := range target.Instructions {
				if len(strings.TrimSpace(instruction)) < 10 {
					t.Errorf("Instruction %d is too short: %s", i, instruction)
				}
				
				// Instructions should be numbered or have clear structure
				if !strings.Contains(instruction, ".") && !strings.Contains(instruction, ":") {
					t.Errorf("Instruction %d lacks clear structure: %s", i, instruction)
				}
			}
		})
	}
}

func TestDeployment_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		Name:    "testapp",
		Version: "1.0.0",
	}
	deployer := NewDeployer(cfg)
	
	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	err := deployer.Deploy(ctx, []string{"brew"}, true)
	// Dry run should complete even with cancelled context
	// since it doesn't execute external commands
	if err != nil {
		t.Errorf("Dry run should not be affected by context cancellation: %v", err)
	}
}
