package main

import (
	"sync"
)

// DataMap is a map containing data sent in and out of worker queue.
type DataMap map[string]interface{}

// Provider is a function that send data into a worker queue.
type Provider func(jobs chan DataMap)

// Consumer is a function that consums data sent into a worker queue.
type Consumer func(job DataMap) DataMap

// RunWorkers starts workers with given provider and consumer fucntions. The work in the consumers are
// spread across a given number of threads.
func RunWorkers(provider Provider, consumer Consumer, threads int) chan DataMap {

	jobs := make(chan DataMap)
	results := make(chan DataMap)

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
