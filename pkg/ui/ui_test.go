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
	"context"
	"os"
	"strings"
	"testing"
)

func TestProgressBar(t *testing.T) {
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

func TestProgressBarIncrement(t *testing.T) {
	pb := NewProgressBar(100, "Test")
	pb.Update(50)
	pb.Increment()
	if pb.current != 51 {
		t.Errorf("Expected current=51, got %d", pb.current)
	}
}

func TestSpinner(t *testing.T) {
	spinner := NewSpinner("Loading")
	if spinner == nil {
		t.Fatal("NewSpinner returned nil")
	}
	spinner.Start(context.Background())
	spinner.Stop()
}

func TestTable(t *testing.T) {
	table := NewTable([]string{"Name", "Value"})
	table.AddRow([]string{"test1", "value1"})
	table.AddRow([]string{"test2", "value2"})

	if len(table.rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(table.rows))
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	table.Print()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test1") {
		t.Error("Table row not found")
	}
}

func TestUIMessages(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Success("success")
	Warning("warning")
	Error("error")
	Info("info")
	Header("header")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "success") {
		t.Error("Success message not found")
	}
}

func TestPrintBanner(t *testing.T) {
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
		t.Error("Banner not found")
	}
}

func TestPrintVersion(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintVersion("1.0.0", "abc123", "2024-01-01")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "1.0.0") {
		t.Error("Version not found")
	}
}

func TestIsInteractive(t *testing.T) {
	result := IsInteractive()
	_ = result
}

func TestConfirm(t *testing.T) {
	old := os.Stdin
	defer func() { os.Stdin = old }()
	
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte("y\n"))
		w.Close()
	}()
	
	result := Confirm("Test?")
	if !result {
		t.Error("Expected true for 'y' input")
	}
}

func TestSelect(t *testing.T) {
	old := os.Stdin
	defer func() { os.Stdin = old }()
	
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte("1\n"))
		w.Close()
	}()
	
	options := []string{"Option 1", "Option 2"}
	result := Select("Choose:", options)
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}
