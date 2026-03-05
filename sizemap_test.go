package main

import (
	"testing"

	"github.com/test-go/testify/assert"
)

func TestSizeMap(t *testing.T) {
	sm := CreateSizeMap("testdata")

	assert.Equal(t, 1, len(sm))
	assert.Equal(t, 3, sm.CountCandidates())
}
