// types.go
package backup

import (
	"time"
)

// Service represents the backup service with all required dependencies
type Service struct {
	config    *Config
	logger    *Logger
	metrics   *BackupMetrics
	pool      *WorkerPool
	versioner *VersionManager
}

// CopyTask represents a single file copy operation
type CopyTask struct {
	Source      string
	Destination string
	Size        int64
	ModTime     time.Time
}

// FileMetadata holds file comparison information
type FileMetadata struct {
	Path     string
	Size     int64
	ModTime  time.Time
	Checksum string
}

// BackupStats holds statistical information about the backup
type BackupStats struct {
	TotalFiles       int   // Total number of files processed
	FilesBackedUp    int   // Number of files actually copied
	FilesSkipped     int   // Number of unchanged files
	FilesFailed      int   // Number of files that failed to backup
	TotalBytes       int64 // Total bytes processed
	BytesTransferred int64 // Actual bytes copied
}

// WorkerPool manages a pool of workers for concurrent file operations
type WorkerPool struct {
	workers       int
	copyFn        func(CopyTask) error
	retryAttempts int
	retryDelay    time.Duration
}
