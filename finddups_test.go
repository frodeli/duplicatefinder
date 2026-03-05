package main

import (
	"encoding/hex"
	"runtime"
	"strings"
	"testing"

	"github.com/test-go/testify/assert"
)

func TestEqualsFilesShouldGiveDups(t *testing.T) {
	// Setup
	sm := SizeMap{4: []string{"testdata/a", "testdata/subdir/c"}}

	// Execute
	dupMap := CreateDuplicationMap(&sm, runtime.NumCPU(), false)

	// Verify
	assert.Equal(t, 1, len(dupMap))
	for _, files := range dupMap {
		assert.Equal(t, 2, len(files))
	}
}

func TestTextOutput(t *testing.T) {
	// Setup
	sm := SizeMap{4: []string{"testdata/a", "testdata/subdir/c"}}

	// Execute
	dupMap := CreateDuplicationMap(&sm, runtime.NumCPU(), false)
	output := dupMap.dump()

	// Verify
	lines := strings.Split(output, "\n")
	assert.Equal(t, 5, len(lines))
	assert.Contains(t, output, "testdata/a")
	assert.Contains(t, output, "testdata/subdir/c")
}

func TestBytesToHash(t *testing.T) {
	hash := hex.EncodeToString([]byte{1, 2, 3, 255})
	assert.Equal(t, "010203ff", hash)
}
