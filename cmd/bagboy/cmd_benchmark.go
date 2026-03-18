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

	"github.com/scttfrdmn/bagboy/pkg/benchmark"
	"github.com/scttfrdmn/bagboy/pkg/config"
)

// newBenchmarkCmd returns the cobra Command for the 'benchmark' subcommand.
func newBenchmarkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "benchmark",
		Short: "Run performance benchmarks",
		RunE:  runBenchmark,
	}
}

// runBenchmark implements the 'benchmark' subcommand logic.
func runBenchmark(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	configPath, err := config.FindConfigFile()
	if err != nil {
		return err
	}

	cfg, err := config.Load(ctx, configPath)
	if err != nil {
		return err
	}

	fmt.Println("🚀 Running bagboy performance benchmarks...")
	results := benchmark.RunBenchmarkSuite(cfg)
	benchmark.PrintBenchmarkResults(results)
	return nil
}
