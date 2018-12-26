package main

import (
	"testing"

	"github.com/test-go/testify/assert"
)

func TestWorkerQueue(t *testing.T) {

	results := RunWorkers(
		func(jobs chan DataMap) {
			job := DataMap{}
			job["foo"] = "bar"
			jobs <- job
		},
		func(job DataMap) DataMap {
			result := DataMap{}
			if job["foo"] == "bar" {
				result["ok"] = "true"
				return result
			}
			result["ok"] = "false"
			return result
		},
		4)

	for result := range results {
		assert.Equal(t, "true", result["ok"])
	}
}
