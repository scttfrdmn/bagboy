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

func TestCheckCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"check", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("check --help: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "check") {
		t.Errorf("help output missing 'check': %q", out)
	}
}

func TestCheckCmd_DefaultFormats(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"check"})
	// check inspects installed tools; should always succeed.
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("check: %v", err)
	}
}

func TestCheckCmd_SpecificFormats(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"check", "--formats", "brew,deb"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("check --formats brew,deb: %v", err)
	}
}

func TestCheckCmd_SingleFormat(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"check", "--formats", "docker"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("check --formats docker: %v", err)
	}
}
