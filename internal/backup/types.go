// types.go
package backup

import (
	"sync"
	"time"
)

// Service represents the backup service with all required dependencies
type Service struct {
	config    *Config
	logger    *Logger
	metrics   *Metrics
	pool      *WorkerPool
	versioner *VersionManager
}

// Metrics tracks backup operation statistics
type Metrics struct {
	mu           sync.Mutex
	BytesCopied  int64
	FilesCopied  int
	FilesSkipped int
	Errors       int
	StartTime    time.Time
	EndTime      time.Time
}

// FileMetadata holds file comparison information
type FileMetadata struct {
	Path     string
	Size     int64
	ModTime  time.Time
	Checksum string
}

// CopyTask represents a single file copy operation
type CopyTask struct {
	Source      string
	Destination string
	Size        int64
	ModTime     time.Time
}
