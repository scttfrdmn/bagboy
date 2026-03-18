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
	"strings"
	"testing"
)

func TestVMCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"vm", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("vm --help: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "vm") {
		t.Errorf("help output missing 'vm': %q", out)
	}
}

func TestVMCheckCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"vm", "check", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("vm check --help: %v", err)
	}
}

func TestVMCheckCmd_Runs(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"vm", "check"})
	// Should succeed regardless of Docker availability.
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("vm check: %v", err)
	}
}
