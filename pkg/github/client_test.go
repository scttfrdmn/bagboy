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

package github

import (
	"testing"

	"github.com/scttfrdmn/bagboy/pkg/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.GitHubConfig{
		TokenEnv: "NONEXISTENT_TOKEN",
	}

	// Should fail without token
	_, err := NewClient(cfg)
	if err == nil {
		t.Error("Expected error when token not found")
	}

	// Test with valid token env var name but no actual token
	cfg.TokenEnv = "PATH" // Use existing env var for test
	client, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Error("Expected client to be created")
	}
}
