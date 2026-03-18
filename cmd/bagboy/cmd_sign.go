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

	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/signing"
)

// newSignCmd returns the cobra Command for the 'sign' subcommand.
func newSignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Check code signing setup and sign binaries",
		RunE:  runSignCmd,
	}
	cmd.Flags().Bool("check", false, "Check signing setup only")
	cmd.Flags().String("binary", "", "Path to binary to sign")
	return cmd
}

// runSignCmd implements the 'sign' subcommand logic.
func runSignCmd(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	checkOnly, _ := cmd.Flags().GetBool("check")
	binaryPath, _ := cmd.Flags().GetString("binary")

	configPath, err := config.FindConfigFile()
	if err != nil && !checkOnly {
		return err
	}

	var cfg *config.Config
	if configPath != "" {
		cfg, err = config.Load(ctx, configPath)
		if err != nil {
			return err
		}
	}

	signer := signing.NewSigner(cfg)

	if checkOnly || binaryPath == "" {
		results := signer.CheckSigningSetup()
		signer.PrintSigningReport(results)
		return nil
	}

	return signer.SignBinary(ctx, binaryPath)
}
