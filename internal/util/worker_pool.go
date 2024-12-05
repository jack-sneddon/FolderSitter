package util

import (
	"sync"
)

// WorkerPool executes tasks concurrently using a fixed number of workers.
func WorkerPool(numWorkers int, tasks []string, workerFunc func(string)) {
	// Create a channel to manage tasks
	taskChannel := make(chan string, len(tasks))

	// Add tasks to the channel
	for _, task := range tasks {
		taskChannel <- task
	}
	close(taskChannel)

	// Create a WaitGroup to wait for all workers to complete
	var wg sync.WaitGroup

	// Launch workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChannel {
				workerFunc(task)
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
}
