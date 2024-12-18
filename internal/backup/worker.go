// internal/backup/worker.go
package backup

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// WorkerPool manages a pool of workers for concurrent file operations
type WorkerPool struct {
	workers       int
	copyFn        func(CopyTask) error
	retryAttempts int
	retryDelay    time.Duration
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int, copyFn func(CopyTask) error, retryAttempts int, retryDelay time.Duration) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}
	return &WorkerPool{
		workers:       workers,
		copyFn:        copyFn,
		retryAttempts: retryAttempts,
		retryDelay:    retryDelay,
	}
}

// Execute processes tasks using a pool of workers with enhanced error handling
// and progress tracking. It respects context cancellation and provides detailed
// error reporting.
func (p *WorkerPool) Execute(ctx context.Context, tasks []CopyTask) error {
	if len(tasks) == 0 {
		return nil
	}

	// Channels for task distribution and error collection
	taskCh := make(chan CopyTask)
	errCh := make(chan error, len(tasks))
	progressCh := make(chan struct{}, len(tasks))

	// Track errors and progress
	var (
		mu         sync.Mutex
		errorCount int
		errorList  []error
		completed  int
		maxErrors  = len(tasks)/10 + 1 // Allow 10% of tasks to fail, minimum 1
	)

	// Create wait group for workers
	var wg sync.WaitGroup

	// Start progress tracker
	go func() {
		for range progressCh {
			mu.Lock()
			completed++
			progress := float64(completed) / float64(len(tasks)) * 100
			mu.Unlock()

			// Log progress every 5%
			if completed%(len(tasks)/20) == 0 || completed == len(tasks) {
				log.Printf("Progress: %.1f%% (%d/%d tasks completed)",
					progress, completed, len(tasks))
			}
		}
	}()

	// Start workers
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for task := range taskCh {
				select {
				case <-ctx.Done():
					// Context was cancelled
					errCh <- fmt.Errorf("worker %d: context cancelled: %w",
						workerID, ctx.Err())
					return

				default:
					// Process the task with retry logic
					err := p.executeWithRetry(ctx, task)

					if err != nil {
						mu.Lock()
						errorCount++
						errorList = append(errorList, fmt.Errorf("worker %d: %w",
							workerID, err))

						// Check if we've exceeded error threshold
						if errorCount >= maxErrors {
							errCh <- fmt.Errorf("too many errors (%d/%d tasks failed)",
								errorCount, len(tasks))
						}
						mu.Unlock()
					}

					// Report progress
					progressCh <- struct{}{}
				}
			}
		}(i)
	}

	// Feed tasks to workers
	go func() {
		defer close(taskCh)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case taskCh <- task:
				// Task sent successfully
			}
		}
	}()

	// Wait for all workers to complete
	wg.Wait()
	close(progressCh)
	close(errCh)

	// Check for errors
	mu.Lock()
	defer mu.Unlock()

	if errorCount > 0 {
		// Create detailed error report
		var errReport strings.Builder
		errReport.WriteString(fmt.Sprintf("Backup completed with %d errors:\n", errorCount))

		for i, err := range errorList {
			errReport.WriteString(fmt.Sprintf("%d. %v\n", i+1, err))
		}

		return fmt.Errorf("%s", errReport.String())
	}

	return nil
}

// executeWithRetry attempts to execute a task with configurable retries
func (p *WorkerPool) executeWithRetry(ctx context.Context, task CopyTask) error {
	var lastErr error

	for attempt := 1; attempt <= p.retryAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := p.copyFn(task); err == nil {
				return nil
			} else {
				lastErr = err

				// Don't sleep on the last attempt
				if attempt < p.retryAttempts {
					// Exponential backoff with jitter
					backoff := p.retryDelay * time.Duration(attempt*attempt)
					jitter := time.Duration(rand.Int63n(int64(time.Second)))
					time.Sleep(backoff + jitter)
				}
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", p.retryAttempts, lastErr)
}
