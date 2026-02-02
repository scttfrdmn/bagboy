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
	"fmt"
	"os"
	"strings"
	"time"
)

// ProgressBar represents a simple progress bar
type ProgressBar struct {
	total   int
	current int
	width   int
	prefix  string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, prefix string) *ProgressBar {
	return &ProgressBar{
		total:  total,
		width:  40,
		prefix: prefix,
	}
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int) {
	pb.current = current
	pb.render()
}

// Increment increments the progress bar by 1
func (pb *ProgressBar) Increment() {
	pb.current++
	pb.render()
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish() {
	pb.current = pb.total
	pb.render()
	fmt.Println()
}

func (pb *ProgressBar) render() {
	percent := float64(pb.current) / float64(pb.total)
	filled := int(percent * float64(pb.width))
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)
	
	fmt.Printf("\r%s [%s] %d/%d (%.1f%%)", 
		pb.prefix, bar, pb.current, pb.total, percent*100)
}

// Spinner represents a simple spinner
type Spinner struct {
	chars   []string
	current int
	message string
	active  bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		chars:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
	}
}

// Start starts the spinner
func (s *Spinner) Start() {
	s.active = true
	go func() {
		for s.active {
			fmt.Printf("\r%s %s", s.chars[s.current], s.message)
			s.current = (s.current + 1) % len(s.chars)
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.active = false
	fmt.Print("\r" + strings.Repeat(" ", len(s.message)+10) + "\r")
}

// Success displays a success message
func Success(message string) {
	fmt.Printf("✅ %s\n", message)
}

// Warning displays a warning message
func Warning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// Error displays an error message
func Error(message string) {
	fmt.Printf("❌ %s\n", message)
}

// Info displays an info message
func Info(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// Header displays a section header
func Header(message string) {
	fmt.Printf("\n🎯 %s\n", message)
	fmt.Println(strings.Repeat("─", len(message)+4))
}

// Confirm prompts for user confirmation
func Confirm(message string) bool {
	fmt.Printf("❓ %s (y/N): ", message)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// Select prompts user to select from options
func Select(message string, options []string) int {
	fmt.Printf("❓ %s\n", message)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}
	fmt.Print("Enter choice (1-", len(options), "): ")
	
	var choice int
	fmt.Scanln(&choice)
	
	if choice < 1 || choice > len(options) {
		return 0
	}
	return choice - 1
}

// Table displays data in a table format
type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTable creates a new table
func NewTable(headers []string) *Table {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	return &Table{
		headers: headers,
		widths:  widths,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	for i, cell := range row {
		if i < len(t.widths) && len(cell) > t.widths[i] {
			t.widths[i] = len(cell)
		}
	}
	t.rows = append(t.rows, row)
}

// Print prints the table
func (t *Table) Print() {
	// Print header
	fmt.Print("┌")
	for i, width := range t.widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.widths)-1 {
			fmt.Print("┬")
		}
	}
	fmt.Println("┐")
	
	// Print header row
	fmt.Print("│")
	for i, header := range t.headers {
		fmt.Printf(" %-*s │", t.widths[i], header)
	}
	fmt.Println()
	
	// Print separator
	fmt.Print("├")
	for i, width := range t.widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.widths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")
	
	// Print rows
	for _, row := range t.rows {
		fmt.Print("│")
		for i, cell := range row {
			if i < len(t.widths) {
				fmt.Printf(" %-*s │", t.widths[i], cell)
			}
		}
		fmt.Println()
	}
	
	// Print bottom border
	fmt.Print("└")
	for i, width := range t.widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.widths)-1 {
			fmt.Print("┴")
		}
	}
	fmt.Println("┘")
}

// PrintBanner prints a welcome banner
func PrintBanner() {
	banner := `
🎒 bagboy - Universal Software Packager
Pack once. Ship everywhere.

`
	fmt.Print(banner)
}

// PrintVersion prints version information
func PrintVersion(version, commit, date string) {
	fmt.Printf("bagboy version %s\n", version)
	if commit != "" {
		fmt.Printf("Git commit: %s\n", commit)
	}
	if date != "" {
		fmt.Printf("Built: %s\n", date)
	}
}

// IsInteractive checks if we're running in an interactive terminal
func IsInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
