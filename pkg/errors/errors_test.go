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

package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestBagboyError_Error(t *testing.T) {
	err := &BagboyError{
		Type:    ErrorTypeValidation,
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}
	
	if err.Error() != "Test error message" {
		t.Errorf("Expected 'Test error message', got '%s'", err.Error())
	}
}

func TestBagboyError_String(t *testing.T) {
	err := &BagboyError{
		Type:    ErrorTypeValidation,
		Code:    "TEST_ERROR",
		Message: "Test error message",
		Details: "Additional details",
		Suggestions: []string{
			"First suggestion",
			"Second suggestion",
		},
		Cause: fmt.Errorf("underlying error"),
	}
	
	result := err.String()
	
	// Check that all components are present
	if !strings.Contains(result, "❌ Test error message") {
		t.Error("Missing error message")
	}
	if !strings.Contains(result, "Details: Additional details") {
		t.Error("Missing details")
	}
	if !strings.Contains(result, "💡 Suggestions:") {
		t.Error("Missing suggestions header")
	}
	if !strings.Contains(result, "• First suggestion") {
		t.Error("Missing first suggestion")
	}
	if !strings.Contains(result, "• Second suggestion") {
		t.Error("Missing second suggestion")
	}
	if !strings.Contains(result, "Caused by: underlying error") {
		t.Error("Missing cause")
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("TEST_CODE", "Test message", "suggestion1", "suggestion2")
	
	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, err.Type)
	}
	if err.Code != "TEST_CODE" {
		t.Errorf("Expected code 'TEST_CODE', got '%s'", err.Code)
	}
	if err.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", err.Message)
	}
	if len(err.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(err.Suggestions))
	}
}

func TestMissingBinaryError(t *testing.T) {
	err := MissingBinaryError("linux-amd64")
	
	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected validation error type")
	}
	if err.Code != CodeMissingBinary {
		t.Errorf("Expected code %s, got %s", CodeMissingBinary, err.Code)
	}
	if !strings.Contains(err.Message, "linux-amd64") {
		t.Error("Error message should contain platform")
	}
	if len(err.Suggestions) == 0 {
		t.Error("Should have suggestions")
	}
}

func TestMissingDependencyError(t *testing.T) {
	err := MissingDependencyError("docker", "brew install docker")
	
	if err.Type != ErrorTypeDependency {
		t.Errorf("Expected dependency error type")
	}
	if err.Code != CodeMissingDependency {
		t.Errorf("Expected code %s, got %s", CodeMissingDependency, err.Code)
	}
	if !strings.Contains(err.Message, "docker") {
		t.Error("Error message should contain tool name")
	}
	
	// Check that install command is in suggestions
	found := false
	for _, suggestion := range err.Suggestions {
		if strings.Contains(suggestion, "brew install docker") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Install command should be in suggestions")
	}
}

func TestExternalToolError(t *testing.T) {
	cause := fmt.Errorf("command failed")
	err := ExternalToolError("rpmbuild", cause)
	
	if err.Type != ErrorTypeExternal {
		t.Errorf("Expected external error type")
	}
	if err.Code != CodeExternalToolFailed {
		t.Errorf("Expected code %s, got %s", CodeExternalToolFailed, err.Code)
	}
	if err.Cause != cause {
		t.Error("Cause should be preserved")
	}
	if !strings.Contains(err.Message, "rpmbuild") {
		t.Error("Error message should contain tool name")
	}
}

func TestFileNotFoundError(t *testing.T) {
	err := FileNotFoundError("/path/to/missing/file")
	
	if err.Type != ErrorTypeFileSystem {
		t.Errorf("Expected filesystem error type")
	}
	if err.Code != CodeFileNotFound {
		t.Errorf("Expected code %s, got %s", CodeFileNotFound, err.Code)
	}
	if !strings.Contains(err.Message, "/path/to/missing/file") {
		t.Error("Error message should contain file path")
	}
}

func TestInvalidConfigError(t *testing.T) {
	err := InvalidConfigError("name", "cannot be empty")
	
	if err.Type != ErrorTypeConfiguration {
		t.Errorf("Expected configuration error type")
	}
	if err.Code != CodeInvalidConfig {
		t.Errorf("Expected code %s, got %s", CodeInvalidConfig, err.Code)
	}
	if !strings.Contains(err.Message, "name") {
		t.Error("Error message should contain field name")
	}
	if !strings.Contains(err.Message, "cannot be empty") {
		t.Error("Error message should contain reason")
	}
}

func TestWrapError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := WrapError(originalErr, "wrapped message", "suggestion1")
	
	if err.Type != ErrorTypeInternal {
		t.Errorf("Expected internal error type")
	}
	if err.Cause != originalErr {
		t.Error("Original error should be preserved as cause")
	}
	if err.Message != "wrapped message" {
		t.Errorf("Expected 'wrapped message', got '%s'", err.Message)
	}
	if len(err.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(err.Suggestions))
	}
}

func TestIsType(t *testing.T) {
	validationErr := NewValidationError("TEST", "test")
	configErr := NewConfigurationError("TEST", "test")
	regularErr := fmt.Errorf("regular error")
	
	if !IsType(validationErr, ErrorTypeValidation) {
		t.Error("Should identify validation error")
	}
	if IsType(validationErr, ErrorTypeConfiguration) {
		t.Error("Should not identify as configuration error")
	}
	if IsType(configErr, ErrorTypeValidation) {
		t.Error("Should not identify as validation error")
	}
	if IsType(regularErr, ErrorTypeValidation) {
		t.Error("Regular error should not be identified as bagboy error")
	}
}

func TestHasCode(t *testing.T) {
	err := NewValidationError("TEST_CODE", "test")
	regularErr := fmt.Errorf("regular error")
	
	if !HasCode(err, "TEST_CODE") {
		t.Error("Should identify correct code")
	}
	if HasCode(err, "WRONG_CODE") {
		t.Error("Should not identify wrong code")
	}
	if HasCode(regularErr, "TEST_CODE") {
		t.Error("Regular error should not have bagboy code")
	}
}

func TestFormatError(t *testing.T) {
	bagboyErr := NewValidationError("TEST", "bagboy error", "suggestion")
	regularErr := fmt.Errorf("regular error")
	
	bagboyResult := FormatError(bagboyErr)
	regularResult := FormatError(regularErr)
	
	if !strings.Contains(bagboyResult, "❌ bagboy error") {
		t.Error("Bagboy error should be formatted with suggestions")
	}
	if !strings.Contains(bagboyResult, "💡 Suggestions:") {
		t.Error("Bagboy error should include suggestions")
	}
	if !strings.Contains(regularResult, "❌ regular error") {
		t.Error("Regular error should be formatted simply")
	}
	if strings.Contains(regularResult, "💡 Suggestions:") {
		t.Error("Regular error should not include suggestions")
	}
}

func TestBagboyError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := &BagboyError{
		Type:    ErrorTypeInternal,
		Code:    "TEST",
		Message: "test",
		Cause:   cause,
	}
	
	if err.Unwrap() != cause {
		t.Error("Unwrap should return the cause")
	}
	
	errNoCause := &BagboyError{
		Type:    ErrorTypeInternal,
		Code:    "TEST",
		Message: "test",
	}
	
	if errNoCause.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause")
	}
}
