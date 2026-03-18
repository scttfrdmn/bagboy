package checksum

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// Calculate computes the SHA256 checksum of a file
func Calculate(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CalculateBytes computes the SHA256 checksum of byte data
func CalculateBytes(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}
