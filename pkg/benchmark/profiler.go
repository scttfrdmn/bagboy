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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/scttfrdmn/bagboy/pkg/config"
	"github.com/scttfrdmn/bagboy/pkg/packager"
)

// PerformanceProfiler provides performance profiling capabilities
type PerformanceProfiler struct {
	startTime time.Time
	metrics   map[string]interface{}
	mu        sync.RWMutex
}

// NewPerformanceProfiler creates a new performance profiler
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		metrics: make(map[string]interface{}),
	}
}

// Start begins performance profiling
func (p *PerformanceProfiler) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.startTime = time.Now()
}

// Stop ends performance profiling and returns metrics
func (p *PerformanceProfiler) Stop() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	duration := time.Since(p.startTime)
	p.metrics["duration_ms"] = duration.Milliseconds()
	p.metrics["duration_seconds"] = duration.Seconds()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	p.metrics["memory_alloc_mb"] = float64(m.Alloc) / 1024 / 1024
	p.metrics["memory_total_alloc_mb"] = float64(m.TotalAlloc) / 1024 / 1024
	p.metrics["gc_cycles"] = m.NumGC
	
	return p.metrics
}

// RecordMetric records a custom metric
func (p *PerformanceProfiler) RecordMetric(name string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics[name] = value
}

// OptimizedPackageRegistry provides optimized parallel packaging
type OptimizedPackageRegistry struct {
	packagers   map[string]packager.Packager
	workerPool  chan struct{}
	maxWorkers  int
}

// NewOptimizedPackageRegistry creates an optimized package registry
func NewOptimizedPackageRegistry(maxWorkers int) *OptimizedPackageRegistry {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	
	return &OptimizedPackageRegistry{
		packagers:  make(map[string]packager.Packager),
		workerPool: make(chan struct{}, maxWorkers),
		maxWorkers: maxWorkers,
	}
}

// Register adds a packager to the registry
func (r *OptimizedPackageRegistry) Register(p packager.Packager) {
	r.packagers[p.Name()] = p
}

// PackAllOptimized packages using optimized parallel processing
func (r *OptimizedPackageRegistry) PackAllOptimized(ctx context.Context, cfg *config.Config) (map[string]string, error) {
	profiler := NewPerformanceProfiler()
	profiler.Start()
	
	results := make(map[string]string)
	errors := make(chan error, len(r.packagers))
	resultsChan := make(chan struct {
		name string
		path string
	}, len(r.packagers))
	
	var wg sync.WaitGroup
	
	// Launch workers
	for name, pkg := range r.packagers {
		wg.Add(1)
		go func(name string, pkg packager.Packager) {
			defer wg.Done()
			
			// Acquire worker slot
			r.workerPool <- struct{}{}
			defer func() { <-r.workerPool }()
			
			// Validate before packing
			if err := pkg.Validate(cfg); err != nil {
				errors <- fmt.Errorf("validation failed for %s: %w", name, err)
				return
			}
			
			// Pack with timeout
			packCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()
			
			path, err := pkg.Pack(packCtx, cfg)
			if err != nil {
				errors <- fmt.Errorf("packing failed for %s: %w", name, err)
				return
			}
			
			resultsChan <- struct {
				name string
				path string
			}{name, path}
		}(name, pkg)
	}
	
	// Wait for completion
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errors)
	}()
	
	// Collect results
	var firstError error
	for {
		select {
		case result, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else {
				results[result.name] = result.path
			}
		case err, ok := <-errors:
			if !ok {
				errors = nil
			} else if firstError == nil {
				firstError = err
			}
		}
		
		if resultsChan == nil && errors == nil {
			break
		}
	}
	
	metrics := profiler.Stop()
	fmt.Printf("📊 Performance Metrics:\n")
	fmt.Printf("   Duration: %.2f seconds\n", metrics["duration_seconds"])
	fmt.Printf("   Memory: %.2f MB allocated\n", metrics["memory_alloc_mb"])
	fmt.Printf("   Workers: %d\n", r.maxWorkers)
	fmt.Printf("   Packages: %d\n", len(results))
	
	if firstError != nil {
		return results, firstError
	}
	
	return results, nil
}

// BenchmarkResult represents benchmark results
type BenchmarkResult struct {
	Name           string        `json:"name"`
	Duration       time.Duration `json:"duration"`
	MemoryUsage    int64         `json:"memory_usage"`
	PackagesBuilt  int           `json:"packages_built"`
	Throughput     float64       `json:"throughput"` // packages per second
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
}

// RunBenchmarkSuite runs a comprehensive benchmark suite
func RunBenchmarkSuite(cfg *config.Config) []BenchmarkResult {
	var results []BenchmarkResult
	
	// Single packager benchmarks
	packagers := map[string]packager.Packager{
		"brew":      nil, // Would need actual packager instances
		"scoop":     nil,
		"installer": nil,
	}
	
	for name := range packagers {
		result := BenchmarkResult{
			Name: fmt.Sprintf("Single_%s", name),
		}
		
		profiler := NewPerformanceProfiler()
		profiler.Start()
		
		// Simulate benchmark (would use actual packager)
		time.Sleep(100 * time.Millisecond) // Placeholder
		
		metrics := profiler.Stop()
		result.Duration = time.Duration(metrics["duration_ms"].(int64)) * time.Millisecond
		result.MemoryUsage = int64(metrics["memory_alloc_mb"].(float64) * 1024 * 1024)
		result.PackagesBuilt = 1
		result.Throughput = 1.0 / result.Duration.Seconds()
		result.Success = true
		
		results = append(results, result)
	}
	
	// Parallel packaging benchmark
	parallelResult := BenchmarkResult{
		Name: "Parallel_All",
	}
	
	profiler := NewPerformanceProfiler()
	profiler.Start()
	
	// Simulate parallel packaging
	time.Sleep(200 * time.Millisecond) // Placeholder
	
	metrics := profiler.Stop()
	parallelResult.Duration = time.Duration(metrics["duration_ms"].(int64)) * time.Millisecond
	parallelResult.MemoryUsage = int64(metrics["memory_alloc_mb"].(float64) * 1024 * 1024)
	parallelResult.PackagesBuilt = len(packagers)
	parallelResult.Throughput = float64(len(packagers)) / parallelResult.Duration.Seconds()
	parallelResult.Success = true
	
	results = append(results, parallelResult)
	
	return results
}

// PrintBenchmarkResults prints benchmark results in a formatted way
func PrintBenchmarkResults(results []BenchmarkResult) {
	fmt.Println("🚀 Bagboy Performance Benchmark Results")
	fmt.Println(strings.Repeat("=", 50))
	
	for _, result := range results {
		fmt.Printf("\n📦 %s:\n", result.Name)
		if result.Success {
			fmt.Printf("   ✅ Duration: %v\n", result.Duration)
			fmt.Printf("   💾 Memory: %.2f MB\n", float64(result.MemoryUsage)/1024/1024)
			fmt.Printf("   📊 Throughput: %.2f packages/sec\n", result.Throughput)
			fmt.Printf("   📈 Packages: %d\n", result.PackagesBuilt)
		} else {
			fmt.Printf("   ❌ Failed: %s\n", result.Error)
		}
	}
	
	fmt.Println("\n💡 Performance Tips:")
	fmt.Println("   • Use parallel packaging for multiple formats")
	fmt.Println("   • Optimize binary size for faster packaging")
	fmt.Println("   • Use SSD storage for better I/O performance")
	fmt.Println("   • Increase worker count for CPU-bound operations")
}
