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

// Package ui provides terminal output utilities including progress bars,
// spinners, tables, and formatted message helpers.
package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// ProgressBar renders a simple ASCII progress bar to stdout.
type ProgressBar struct {
	total   int
	current int
	width   int
	prefix  string
}

// NewProgressBar creates a new ProgressBar with the given total step count and
// display prefix.
func NewProgressBar(total int, prefix string) *ProgressBar {
	return &ProgressBar{
		total:  total,
		width:  40,
		prefix: prefix,
	}
}

// Update sets the progress bar to current and re-renders it.
func (pb *ProgressBar) Update(current int) {
	pb.current = current
	pb.render()
}

// Increment advances the progress bar by one step and re-renders it.
func (pb *ProgressBar) Increment() {
	pb.current++
	pb.render()
}

// Finish sets the progress bar to 100 % and prints a trailing newline.
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

// Spinner renders an animated braille spinner to stdout while work is in progress.
type Spinner struct {
	chars   []string
	current int
	message string
	active  bool
}

// NewSpinner creates a new Spinner with the given status message.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		chars:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
	}
}

// Start launches the spinner animation in a background goroutine.
// The goroutine exits when either Stop is called or ctx is cancelled,
// preventing a goroutine leak on context cancellation.
func (s *Spinner) Start(ctx context.Context) {
	s.active = true
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.active = false
				return
			case <-ticker.C:
				if !s.active {
					return
				}
				fmt.Printf("\r%s %s", s.chars[s.current], s.message)
				s.current = (s.current + 1) % len(s.chars)
			}
		}
	}()
}

// Stop halts the spinner and clears the line.
func (s *Spinner) Stop() {
	s.active = false
	fmt.Print("\r" + strings.Repeat(" ", len(s.message)+10) + "\r")
}

// Success prints a success message prefixed with a check-mark emoji.
func Success(message string) {
	fmt.Printf("✅ %s\n", message)
}

// Warning prints a warning message prefixed with a warning emoji.
func Warning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// Error prints an error message prefixed with a cross emoji.
func Error(message string) {
	fmt.Printf("❌ %s\n", message)
}

// Info prints an informational message prefixed with an info emoji.
func Info(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// Header prints a section header with a horizontal rule.
func Header(message string) {
	fmt.Printf("\n🎯 %s\n", message)
	fmt.Println(strings.Repeat("─", len(message)+4))
}

// Confirm prompts the user with a yes/no question and returns true when the
// user answers "y" or "yes" (case-insensitive).
func Confirm(message string) bool {
	fmt.Printf("❓ %s (y/N): ", message)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// Select prompts the user to choose from a numbered list of options and returns
// the zero-based index of the chosen item, or 0 on invalid input.
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

// Table renders tabular data with box-drawing borders.
type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTable creates a new Table with the given column headers.
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

// AddRow appends a row to the table, expanding column widths as necessary.
func (t *Table) AddRow(row []string) {
	for i, cell := range row {
		if i < len(t.widths) && len(cell) > t.widths[i] {
			t.widths[i] = len(cell)
		}
	}
	t.rows = append(t.rows, row)
}

// Print renders the table to stdout using Unicode box-drawing characters.
func (t *Table) Print() {
	// Top border.
	fmt.Print("┌")
	for i, width := range t.widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.widths)-1 {
			fmt.Print("┬")
		}
	}
	fmt.Println("┐")

	// Header row.
	fmt.Print("│")
	for i, header := range t.headers {
		fmt.Printf(" %-*s │", t.widths[i], header)
	}
	fmt.Println()

	// Header/body separator.
	fmt.Print("├")
	for i, width := range t.widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.widths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")

	// Data rows.
	for _, row := range t.rows {
		fmt.Print("│")
		for i, cell := range row {
			if i < len(t.widths) {
				fmt.Printf(" %-*s │", t.widths[i], cell)
			}
		}
		fmt.Println()
	}

	// Bottom border.
	fmt.Print("└")
	for i, width := range t.widths {
		fmt.Print(strings.Repeat("─", width+2))
		if i < len(t.widths)-1 {
			fmt.Print("┴")
		}
	}
	fmt.Println("┘")
}

// PrintBanner prints the bagboy welcome banner to stdout.
func PrintBanner() {
	banner := `
🎒 bagboy - Universal Software Packager
Pack once. Ship everywhere.

`
	fmt.Print(banner)
}

// PrintVersion prints version, commit, and build-date information to stdout.
func PrintVersion(version, commit, date string) {
	fmt.Printf("bagboy version %s\n", version)
	if commit != "" {
		fmt.Printf("Git commit: %s\n", commit)
	}
	if date != "" {
		fmt.Printf("Built: %s\n", date)
	}
}

// IsInteractive reports whether stdin is connected to an interactive terminal.
func IsInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
