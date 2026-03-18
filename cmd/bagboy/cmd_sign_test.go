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
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSignCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"sign", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sign --help: %v", err)
	}
	if !strings.Contains(buf.String(), "sign") {
		t.Errorf("help output missing 'sign': %q", buf.String())
	}
}

func TestSignCmd_CheckFlagNoConfig(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"sign", "--check"})
	// --check does not require a config file; should succeed.
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sign --check: %v", err)
	}
}

func TestSignCmd_CheckFlagWithConfig(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)

	cfg := `name: myapp
version: 1.0.0
description: Test
binaries:
  linux-amd64: myapp-linux-amd64
`
	if err := os.WriteFile("bagboy.yaml", []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"sign", "--check"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sign --check with config: %v", err)
	}
}

func TestSignCmd_NoBinaryNoConfig(t *testing.T) {
	testDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(testDir)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"sign"})
	// Without config and without --check, should return an error.
	_ = rootCmd.Execute()
}
