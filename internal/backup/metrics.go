// metrics.go
package backup

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// metrics.go
type BackupMetrics struct {
	mu            sync.RWMutex
	totalFiles    int
	filesComplete int
	bytesComplete int64
	filesSkipped  int
	filesFailed   int
	startTime     time.Time
	quiet         bool
	updates       chan metricsUpdate // Add this
}

type metricsUpdate struct {
	operation string
	bytes     int64
}

func NewBackupMetrics(totalFiles int, quiet bool) *BackupMetrics {
	return &BackupMetrics{
		totalFiles: totalFiles,
		startTime:  time.Now(),
		quiet:      quiet,
		updates:    make(chan metricsUpdate, totalFiles), // Buffered channel
	}
}

func (m *BackupMetrics) StartTracking(ctx context.Context) {
	go func() {
		for {
			select {
			case update, ok := <-m.updates:
				if !ok {
					return // Channel was closed
				}
				m.mu.Lock()
				switch update.operation {
				case "completed":
					m.filesComplete++
					m.bytesComplete += update.bytes
				case "skipped":
					m.filesSkipped++
					m.bytesComplete += update.bytes
				case "failed":
					m.filesFailed++
				}
				m.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (m *BackupMetrics) IncrementCompleted(bytes int64) {
	select {
	case m.updates <- metricsUpdate{"completed", bytes}:
	default:
		// If channel is full, don't block
	}
}

func (m *BackupMetrics) IncrementSkipped(bytes int64) {
	select {
	case m.updates <- metricsUpdate{"skipped", bytes}:
	default:
		// If channel is full, don't block
	}
}

func (m *BackupMetrics) IncrementFailed() {
	select {
	case m.updates <- metricsUpdate{"failed", 0}:
	default:
		// If channel is full, don't block
	}
}

// Add method to get metrics for version manager
func (m *BackupMetrics) GetStats() BackupStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return BackupStats{
		TotalFiles:       m.totalFiles,
		FilesBackedUp:    m.filesComplete,
		FilesSkipped:     m.filesSkipped,
		FilesFailed:      m.filesFailed,
		TotalBytes:       m.bytesComplete,
		BytesTransferred: m.bytesComplete,
	}
}

func (m *BackupMetrics) DisplayProgress() {
	if m.quiet {
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.filesComplete + m.filesSkipped
	percentComplete := float64(total) / float64(m.totalFiles) * 100

	// Create progress bar with safety checks
	const barWidth = 30
	completed := int(percentComplete * float64(barWidth) / 100)
	if completed < 0 {
		completed = 0
	}
	if completed > barWidth {
		completed = barWidth
	}

	remaining := barWidth - completed
	if remaining < 0 {
		remaining = 0
	}

	bar := strings.Repeat("█", completed) + strings.Repeat("░", remaining)

	// Save cursor position, clear from cursor to beginning of line, write progress
	fmt.Print("\x1b[s")     // Save cursor position
	fmt.Print("\x1b[1000D") // Move cursor far left
	fmt.Print("\x1b[K")     // Clear line
	fmt.Printf("[%s] %5.1f%% | %3d copied, %3d skipped of %3d files | %6.2f MB | %6.2f MB/s",
		bar,
		percentComplete,
		m.filesComplete,
		m.filesSkipped,
		m.totalFiles,
		float64(m.bytesComplete)/1024/1024,
		float64(m.bytesComplete)/time.Since(m.startTime).Seconds()/1024/1024)
	fmt.Print("\x1b[u") // Restore cursor position
}

// metrics.go
func (m *BackupMetrics) GetStartTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.startTime
}

// metrics.go
func (m *BackupMetrics) GetDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.startTime)
}

// metrics.go
func (m *BackupMetrics) IsBackupInProgress() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.filesComplete > 0 || m.bytesComplete > 0
}

// Add this to metrics.go
func (m *BackupMetrics) DisplayFinalSummary() {
	if m.quiet {
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	duration := time.Since(m.startTime)
	fmt.Printf("\n\nBackup completed in %v\n", duration)
	fmt.Printf("Files processed: %d, Files skipped: %d, Failed: %d, Total size: %.2f MB\n",
		m.filesComplete,
		m.filesSkipped,
		m.filesFailed,
		float64(m.bytesComplete)/1024/1024)
}
