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

package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestProgressBar(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	pb := NewProgressBar(10, "Testing")
	pb.Update(5)
	pb.Finish()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Testing") {
		t.Errorf("Expected progress bar to contain 'Testing', got: %s", output)
	}
	if !strings.Contains(output, "50.0%") {
		t.Errorf("Expected progress bar to show 50%%, got: %s", output)
	}
}

func TestSpinner(t *testing.T) {
	spinner := NewSpinner("Loading...")
	if spinner.message != "Loading..." {
		t.Errorf("Expected spinner message 'Loading...', got: %s", spinner.message)
	}
	if len(spinner.chars) == 0 {
		t.Error("Expected spinner to have animation characters")
	}
}

func TestTable(t *testing.T) {
	table := NewTable([]string{"Name", "Status"})
	table.AddRow([]string{"test", "success"})
	table.AddRow([]string{"another", "failed"})

	if len(table.rows) != 2 {
		t.Errorf("Expected 2 rows, got: %d", len(table.rows))
	}
	if table.widths[0] < 4 { // "Name" length
		t.Errorf("Expected first column width >= 4, got: %d", table.widths[0])
	}
}

func TestIsInteractive(t *testing.T) {
	// This test is environment-dependent, just ensure it doesn't panic
	result := IsInteractive()
	_ = result // Use the result to avoid unused variable warning
}

func TestUIMessages(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Success("Test success")
	Warning("Test warning")
	Error("Test error")
	Info("Test info")
	Header("Test header")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedMessages := []string{
		"✅ Test success",
		"⚠️  Test warning",
		"❌ Test error",
		"ℹ️  Test info",
		"🎯 Test header",
	}

	for _, expected := range expectedMessages {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

func TestPrintBanner(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintBanner()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "bagboy") {
		t.Errorf("Expected banner to contain 'bagboy', got: %s", output)
	}
	if !strings.Contains(output, "Universal Software Packager") {
		t.Errorf("Expected banner to contain description, got: %s", output)
	}
}

func TestPrintVersion(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintVersion("1.0.0", "abc123", "2026-01-01")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedParts := []string{
		"bagboy version 1.0.0",
		"Git commit: abc123",
		"Built: 2026-01-01",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected version output to contain '%s', got: %s", expected, output)
		}
	}
}
