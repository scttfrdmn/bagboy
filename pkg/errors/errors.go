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
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeDependency    ErrorType = "dependency"
	ErrorTypeFileSystem    ErrorType = "filesystem"
	ErrorTypeNetwork       ErrorType = "network"
	ErrorTypeExternal      ErrorType = "external"
	ErrorTypeInternal      ErrorType = "internal"
)

// BagboyError represents a structured error with context and suggestions
type BagboyError struct {
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
	Cause       error     `json:"-"`
}

// Error implements the error interface
func (e *BagboyError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error
func (e *BagboyError) Unwrap() error {
	return e.Cause
}

// String returns a formatted error message with suggestions
func (e *BagboyError) String() string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("❌ %s\n", e.Message))
	
	if e.Details != "" {
		sb.WriteString(fmt.Sprintf("   Details: %s\n", e.Details))
	}
	
	if len(e.Suggestions) > 0 {
		sb.WriteString("   💡 Suggestions:\n")
		for _, suggestion := range e.Suggestions {
			sb.WriteString(fmt.Sprintf("      • %s\n", suggestion))
		}
	}
	
	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf("   Caused by: %v\n", e.Cause))
	}
	
	return sb.String()
}

// NewValidationError creates a validation error
func NewValidationError(code, message string, suggestions ...string) *BagboyError {
	return &BagboyError{
		Type:        ErrorTypeValidation,
		Code:        code,
		Message:     message,
		Suggestions: suggestions,
	}
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(code, message string, suggestions ...string) *BagboyError {
	return &BagboyError{
		Type:        ErrorTypeConfiguration,
		Code:        code,
		Message:     message,
		Suggestions: suggestions,
	}
}

// NewDependencyError creates a dependency error
func NewDependencyError(code, message string, suggestions ...string) *BagboyError {
	return &BagboyError{
		Type:        ErrorTypeDependency,
		Code:        code,
		Message:     message,
		Suggestions: suggestions,
	}
}

// NewFileSystemError creates a filesystem error
func NewFileSystemError(code, message string, cause error, suggestions ...string) *BagboyError {
	return &BagboyError{
		Type:        ErrorTypeFileSystem,
		Code:        code,
		Message:     message,
		Cause:       cause,
		Suggestions: suggestions,
	}
}

// NewExternalError creates an external tool error
func NewExternalError(code, message string, cause error, suggestions ...string) *BagboyError {
	return &BagboyError{
		Type:        ErrorTypeExternal,
		Code:        code,
		Message:     message,
		Cause:       cause,
		Suggestions: suggestions,
	}
}

// Common error codes and constructors
const (
	CodeMissingBinary      = "MISSING_BINARY"
	CodeMissingDependency  = "MISSING_DEPENDENCY"
	CodeInvalidConfig      = "INVALID_CONFIG"
	CodeExternalToolFailed = "EXTERNAL_TOOL_FAILED"
	CodeFileNotFound       = "FILE_NOT_FOUND"
	CodePermissionDenied   = "PERMISSION_DENIED"
)

// MissingBinaryError creates a standardized missing binary error
func MissingBinaryError(platform string) *BagboyError {
	return NewValidationError(
		CodeMissingBinary,
		fmt.Sprintf("No %s binary specified", platform),
		fmt.Sprintf("Add a %s binary to the 'binaries' section in bagboy.yaml", platform),
		"Example: binaries:\n  "+platform+": dist/myapp-"+platform,
		"Run 'bagboy init' to regenerate configuration with detected binaries",
	)
}

// MissingDependencyError creates a standardized missing dependency error
func MissingDependencyError(tool, installCmd string) *BagboyError {
	suggestions := []string{
		fmt.Sprintf("Install %s: %s", tool, installCmd),
		"Check if the tool is in your PATH",
		"Use bagboy's built-in implementation (if available)",
	}
	
	return NewDependencyError(
		CodeMissingDependency,
		fmt.Sprintf("Required tool '%s' not found", tool),
		suggestions...,
	)
}

// ExternalToolError creates a standardized external tool error
func ExternalToolError(tool string, cause error) *BagboyError {
	return NewExternalError(
		CodeExternalToolFailed,
		fmt.Sprintf("External tool '%s' failed", tool),
		cause,
		fmt.Sprintf("Check if %s is properly installed and configured", tool),
		"Verify the tool works independently outside of bagboy",
		"Check the tool's documentation for troubleshooting",
	)
}

// FileNotFoundError creates a standardized file not found error
func FileNotFoundError(path string) *BagboyError {
	return NewFileSystemError(
		CodeFileNotFound,
		fmt.Sprintf("File not found: %s", path),
		nil,
		"Check if the file path is correct",
		"Ensure the file exists and is readable",
		"Use absolute paths to avoid confusion",
	)
}

// InvalidConfigError creates a standardized configuration error
func InvalidConfigError(field, reason string) *BagboyError {
	return NewConfigurationError(
		CodeInvalidConfig,
		fmt.Sprintf("Invalid configuration for '%s': %s", field, reason),
		fmt.Sprintf("Fix the '%s' field in bagboy.yaml", field),
		"Run 'bagboy validate' to check your configuration",
		"See documentation for valid configuration options",
	)
}

// WrapError wraps an existing error with bagboy context
func WrapError(err error, message string, suggestions ...string) *BagboyError {
	return &BagboyError{
		Type:        ErrorTypeInternal,
		Code:        "WRAPPED_ERROR",
		Message:     message,
		Cause:       err,
		Suggestions: suggestions,
	}
}

// IsType checks if an error is of a specific type
func IsType(err error, errorType ErrorType) bool {
	if bagboyErr, ok := err.(*BagboyError); ok {
		return bagboyErr.Type == errorType
	}
	return false
}

// HasCode checks if an error has a specific code
func HasCode(err error, code string) bool {
	if bagboyErr, ok := err.(*BagboyError); ok {
		return bagboyErr.Code == code
	}
	return false
}

// FormatError formats any error for user-friendly display
func FormatError(err error) string {
	if bagboyErr, ok := err.(*BagboyError); ok {
		return bagboyErr.String()
	}
	return fmt.Sprintf("❌ %v", err)
}
