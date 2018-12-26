package main

import (
	"container/list"
	"reflect"
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
	var dupMap = CreateDuplicationMap(&sm, runtime.NumCPU(), false)

	// Verify
	var keys = reflect.ValueOf(dupMap).MapKeys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, 2, dupMap[keys[0].Interface().(EqualFile)].Len())
}

func TestTextOutput(t *testing.T) {
	// Setup
	var sm = SizeMap{}
	var l = &list.List{}
	l.PushBack("testdata/a")
	l.PushBack("testdata/subdir/c")
	sm[4] = l

	// Execute
	var dupMap = CreateDuplicationMap(&sm, runtime.NumCPU(), false)
	output := dupMap.dump()

	// Verify
	lines := strings.Split(output, "\n")
	assert.Equal(t, 5, len(lines))
	assert.Contains(t, output, "testdata/a")
	assert.Contains(t, output, "testdata/subdir/c")

}

func TestBytesToHash(t *testing.T) {
	hash := bytesToHash([]byte{1, 2, 3, 255})
	assert.Equal(t, "010203ff", hash)
}
