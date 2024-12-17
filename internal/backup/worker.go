// internal/backup/worker.go
package backup

import (
	"context"
	"sync"
)

type WorkerPool struct {
	workers int
	copyFn  func(CopyTask) error
}

func NewWorkerPool(workers int, copyFn func(CopyTask) error) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}
	return &WorkerPool{
		workers: workers,
		copyFn:  copyFn,
	}
}

func (p *WorkerPool) Execute(ctx context.Context, tasks []CopyTask) error {
	taskCh := make(chan CopyTask)
	errCh := make(chan error, len(tasks))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				select {
				case <-ctx.Done():
					return
				default:
					if err := p.copyFn(task); err != nil {
						errCh <- err
					}
				}
			}
		}()
	}

	// Send tasks to workers
	go func() {
		defer close(taskCh)
	TaskLoop:
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				break TaskLoop
			case taskCh <- task:
			}
		}
	}()

	// Wait for completion
	wg.Wait()
	close(errCh)

	// Check for errors
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
