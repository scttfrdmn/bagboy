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

package deps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Cache handles dependency check result caching
type Cache struct {
	cacheDir string
}

// NewCache creates a new dependency cache
func NewCache() *Cache {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".bagboy", "cache", "deps")
	os.MkdirAll(cacheDir, 0755)
	
	return &Cache{cacheDir: cacheDir}
}

// CacheEntry represents a cached dependency check result
type CacheEntry struct {
	Status    DependencyStatus `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	TTL       time.Duration    `json:"ttl"`
}

// Get retrieves a cached dependency status
func (c *Cache) Get(key string) (*DependencyStatus, bool) {
	cachePath := filepath.Join(c.cacheDir, key+".json")
	
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}
	
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	
	// Check if cache entry is still valid
	if time.Since(entry.Timestamp) > entry.TTL {
		os.Remove(cachePath) // Clean up expired entry
		return nil, false
	}
	
	return &entry.Status, true
}

// Set stores a dependency status in cache
func (c *Cache) Set(key string, status DependencyStatus, ttl time.Duration) error {
	entry := CacheEntry{
		Status:    status,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
	
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	
	cachePath := filepath.Join(c.cacheDir, key+".json")
	return os.WriteFile(cachePath, data, 0644)
}

// Clear removes all cached entries
func (c *Cache) Clear() error {
	return os.RemoveAll(c.cacheDir)
}

// CleanExpired removes expired cache entries
func (c *Cache) CleanExpired() error {
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			cachePath := filepath.Join(c.cacheDir, entry.Name())
			
			data, err := os.ReadFile(cachePath)
			if err != nil {
				continue
			}
			
			var cacheEntry CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err != nil {
				continue
			}
			
			if time.Since(cacheEntry.Timestamp) > cacheEntry.TTL {
				os.Remove(cachePath)
			}
		}
	}
	
	return nil
}
