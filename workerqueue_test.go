package main

import (
	"testing"

	"github.com/test-go/testify/assert"
)

func TestWorkerQueue(t *testing.T) {

	results := RunWorkers(
		func(jobs chan ChecksumItem) {
			jobs <- ChecksumItem{Filename: "bar"}
		},
		func(job ChecksumItem) ChecksumItem {
			if job.Filename == "bar" {
				return ChecksumItem{Chksum: "ok"}
			}
			return ChecksumItem{}
		},
		4)

	for result := range results {
		assert.Equal(t, "ok", result.Chksum)
	}
}
