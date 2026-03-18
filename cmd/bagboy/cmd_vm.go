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

	"github.com/scttfrdmn/bagboy/pkg/ui"
	"github.com/scttfrdmn/bagboy/pkg/vm"
)

// newVMCmd returns the cobra Command for the 'vm' subcommand.
func newVMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Manage VM build environments",
	}
	cmd.AddCommand(newVMCheckCmd())
	return cmd
}

// newVMCheckCmd returns the cobra Command for the 'vm check' subcommand.
func newVMCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check VM availability",
		RunE:  runVMCheck,
	}
}

// runVMCheck implements the 'vm check' subcommand logic.
func runVMCheck(_ *cobra.Command, _ []string) error {
	ui.Header("VM Status")
	vmMgr := vm.NewManager(&vm.Config{Enabled: true, Provider: "docker"})
	if !vmMgr.IsAvailable() {
		ui.Warning("Docker not available")
		return nil
	}
	ui.Success("✅ Docker available")
	return nil
}
