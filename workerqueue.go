package main

import (
	"sync"
)

// ChecksumItem holds file information passed between worker queue stages.
type ChecksumItem struct {
	Filename string
	Size     int64
	Chksum   string
}

// Provider is a function that sends jobs into a worker queue.
type Provider func(jobs chan ChecksumItem)

// Consumer is a function that processes a job and returns a result.
type Consumer func(job ChecksumItem) ChecksumItem

// RunWorkers starts workers with given provider and consumer functions. The work in the consumers are
// spread across a given number of threads.
func RunWorkers(provider Provider, consumer Consumer, threads int) chan ChecksumItem {

	jobs := make(chan ChecksumItem)
	results := make(chan ChecksumItem)

	// create jobs
	go func() {
		provider(jobs)
		close(jobs)
	}()

	var wg sync.WaitGroup
	wg.Add(threads)
	go func() {
		wg.Wait()
		close(results)
	}()

	for i := 0; i < threads; i++ {
		go func() {

			defer wg.Done()
			for job := range jobs {
				result := consumer(job)
				results <- result
			}

		}()
	}
	return results
}
