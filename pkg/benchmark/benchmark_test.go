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

package benchmark

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/packager"
	"github.com/scttfrdmn/bagboy/pkg/packager/brew"
	"github.com/scttfrdmn/bagboy/pkg/packager/deb"
	"github.com/scttfrdmn/bagboy/pkg/packager/docker"
	"github.com/scttfrdmn/bagboy/pkg/packager/installer"
	"github.com/scttfrdmn/bagboy/pkg/packager/scoop"
)

// BenchmarkConfig represents benchmark configuration
type BenchmarkConfig struct {
	BinarySize    int64  // Size of test binary in bytes
	NumPackagers  int    // Number of packagers to test
	Iterations    int    // Number of iterations per test
	OutputDir     string // Directory for benchmark outputs
}

// setupBenchmarkEnvironment creates test environment for benchmarks
func setupBenchmarkEnvironment(b *testing.B, cfg BenchmarkConfig) (*config.Config, string) {
	b.Helper()
	
	tmpDir := b.TempDir()
	
	// Create test binary
	binaryPath := filepath.Join(tmpDir, "testapp")
	testData := make([]byte, cfg.BinarySize)
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	
	if err := os.WriteFile(binaryPath, testData, 0755); err != nil {
		b.Fatalf("Failed to create test binary: %v", err)
	}
	
	// Create bagboy config
	config := &config.Config{
		Name:        "benchmarkapp",
		Version:     "1.0.0",
		Description: "Benchmark test application",
		License:     "MIT",
		Homepage:    "https://example.com",
		Author:      "Benchmark Test <test@example.com>",
		Binaries: map[string]string{
			"linux-amd64":   binaryPath,
			"darwin-amd64":  binaryPath,
			"windows-amd64": binaryPath + ".exe",
		},
		Packages: config.PackagesConfig{
			Deb: config.DebConfig{
				Maintainer: "test@example.com",
				Section:    "utils",
			},
		},
	}
	
	return config, tmpDir
}

// BenchmarkSinglePackager benchmarks individual packager performance
func BenchmarkSinglePackager(b *testing.B) {
	benchmarks := []struct {
		name     string
		packager packager.Packager
	}{
		{"Brew", brew.New()},
		{"Scoop", scoop.New()},
		{"DEB", deb.New()},
		{"Docker", docker.New()},
		{"Installer", installer.New()},
	}
	
	cfg, tmpDir := setupBenchmarkEnvironment(b, BenchmarkConfig{
		BinarySize: 1024 * 1024, // 1MB binary
	})
	
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)
	
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			ctx := context.Background()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := bm.packager.Pack(ctx, cfg)
				if err != nil {
					// Allow errors for missing external tools
					if !isExpectedError(err) {
						b.Errorf("Unexpected error in %s: %v", bm.name, err)
					}
				}
			}
		})
	}
}

// BenchmarkParallelPackaging benchmarks parallel packaging performance
func BenchmarkParallelPackaging(b *testing.B) {
	cfg, tmpDir := setupBenchmarkEnvironment(b, BenchmarkConfig{
		BinarySize: 1024 * 1024, // 1MB binary
	})
	
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)
	
	registry := packager.NewRegistry()
	registry.Register(brew.New())
	registry.Register(scoop.New())
	registry.Register(installer.New())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := registry.PackAll(ctx, cfg)
		if err != nil && !isExpectedError(err) {
			b.Errorf("Unexpected error in parallel packaging: %v", err)
		}
	}
}

// BenchmarkMemoryUsage benchmarks memory usage during packaging
func BenchmarkMemoryUsage(b *testing.B) {
	cfg, tmpDir := setupBenchmarkEnvironment(b, BenchmarkConfig{
		BinarySize: 10 * 1024 * 1024, // 10MB binary
	})
	
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)
	
	packager := brew.New()
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		_, err := packager.Pack(ctx, cfg)
		if err != nil && !isExpectedError(err) {
			b.Errorf("Unexpected error: %v", err)
		}
		
		runtime.ReadMemStats(&m2)
		
		// Report memory allocation
		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc), "bytes/op")
	}
}

// BenchmarkLargeBinary benchmarks performance with large binaries
func BenchmarkLargeBinary(b *testing.B) {
	sizes := []int64{
		1024 * 1024,      // 1MB
		10 * 1024 * 1024, // 10MB
		50 * 1024 * 1024, // 50MB
	}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%dMB", size/(1024*1024)), func(b *testing.B) {
			cfg, tmpDir := setupBenchmarkEnvironment(b, BenchmarkConfig{
				BinarySize: size,
			})
			
			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(tmpDir)
			
			packager := installer.New() // Fast packager for large binary test
			ctx := context.Background()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				start := time.Now()
				_, err := packager.Pack(ctx, cfg)
				duration := time.Since(start)
				
				if err != nil && !isExpectedError(err) {
					b.Errorf("Unexpected error: %v", err)
				}
				
				b.ReportMetric(duration.Seconds(), "sec/op")
			}
		})
	}
}

// BenchmarkConcurrentPackaging benchmarks concurrent packaging operations
func BenchmarkConcurrentPackaging(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency%d", concurrency), func(b *testing.B) {
			cfg, tmpDir := setupBenchmarkEnvironment(b, BenchmarkConfig{
				BinarySize: 1024 * 1024, // 1MB binary
			})
			
			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(tmpDir)
			
			b.SetParallelism(concurrency)
			b.ResetTimer()
			
			b.RunParallel(func(pb *testing.PB) {
				packager := installer.New()
				ctx := context.Background()
				
				for pb.Next() {
					_, err := packager.Pack(ctx, cfg)
					if err != nil && !isExpectedError(err) {
						b.Errorf("Unexpected error: %v", err)
					}
				}
			})
		})
	}
}

// isExpectedError checks if an error is expected (e.g., missing external tools)
func isExpectedError(err error) bool {
	if err == nil {
		return false
	}
	
	errorStr := err.Error()
	expectedErrors := []string{
		"not found",
		"no such file",
		"command not found",
		"executable file not found",
		"rpmbuild not found",
		"docker not found",
		"zip failed",
	}
	
	for _, expected := range expectedErrors {
		if contains(errorStr, expected) {
			return true
		}
	}
	
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
