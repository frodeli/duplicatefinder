package main

import (
	"container/list"
	"encoding/hex"
	"runtime"
	"strings"
	"testing"

	"github.com/test-go/testify/assert"
)

func TestEqualsFilesShouldGiveDups(t *testing.T) {
	// Setup
	sm := SizeMap{}
	l := list.List{}
	l.PushBack("testdata/a")
	l.PushBack("testdata/subdir/c")
	sm[4] = &l

	// Execute
	dupMap := CreateDuplicationMap(&sm, runtime.NumCPU(), false)

	// Verify
	assert.Equal(t, 1, len(dupMap))
	for _, filelist := range dupMap {
		assert.Equal(t, 2, filelist.Len())
	}
}

func TestTextOutput(t *testing.T) {
	// Setup
	sm := SizeMap{}
	l := &list.List{}
	l.PushBack("testdata/a")
	l.PushBack("testdata/subdir/c")
	sm[4] = l

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
