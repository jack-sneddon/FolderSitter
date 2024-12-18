// internal/backup/worker.go
package backup

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

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
// worker.go - updated Execute function
func (p *WorkerPool) Execute(ctx context.Context, tasks []CopyTask) error {
	if len(tasks) == 0 {
		return nil
	}

	taskCh := make(chan CopyTask, len(tasks))
	var wg sync.WaitGroup

	// Feed tasks to channel first
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	// Start workers
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskCh {
				select {
				case <-ctx.Done():
					return
				default:
					if err := p.executeWithRetry(ctx, task); err != nil {
						log.Printf("Worker %d: Error processing task: %v", workerID, err)
					}
				}
			}
		}(i)
	}

	wg.Wait()
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
