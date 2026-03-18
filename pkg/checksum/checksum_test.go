package checksum

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculate(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello world")
	
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate checksum
	checksum, err := Calculate(testFile)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// Verify it's a valid SHA256 (64 hex characters)
	if len(checksum) != 64 {
		t.Errorf("Expected 64 character checksum, got %d", len(checksum))
	}

	// Known SHA256 of "hello world"
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if checksum != expected {
		t.Errorf("Expected checksum %s, got %s", expected, checksum)
	}
}

func TestCalculateBytes(t *testing.T) {
	data := []byte("hello world")
	checksum := CalculateBytes(data)

	// Known SHA256 of "hello world"
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if checksum != expected {
		t.Errorf("Expected checksum %s, got %s", expected, checksum)
	}
}

func TestCalculateNonExistentFile(t *testing.T) {
	_, err := Calculate("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCalculateEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")
	
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checksum, err := Calculate(testFile)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// SHA256 of empty string
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if checksum != expected {
		t.Errorf("Expected checksum %s, got %s", expected, checksum)
	}
}
