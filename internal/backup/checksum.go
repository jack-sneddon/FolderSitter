// checksum.go
package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// calculateChecksum computes SHA-256 hash of file
func (s *Service) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
