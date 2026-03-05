package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// EqualFile holds the checksum and size of the equal files.
type EqualFile struct {
	Chksum string
	Size   int64
}

// DuplicateMap holds groups of duplicate files.
type DuplicateMap map[EqualFile][]string

const partialReadSize = 4096

// bufPool reuses 1 MB read buffers across goroutines for the full checksum pass.
var bufPool = sync.Pool{
	New: func() interface{} { return make([]byte, 1<<20) },
}

// FilterByPartialHash removes candidates whose first 4 KB differ from all other same-sized files,
// avoiding a full read of files that cannot possibly be duplicates.
func FilterByPartialHash(sm *SizeMap, threads int) SizeMap {
	results := RunWorkers(
		func(jobs chan ChecksumItem) {
			for size, filenames := range *sm {
				createJobs(size, filenames, jobs)
			}
		},
		func(job ChecksumItem) ChecksumItem {
			return partialChkSumForFile(job.Filename, job.Size)
		},
		threads)

	type partialKey struct {
		size   int64
		chksum string
	}
	grouped := make(map[partialKey][]string)
	for res := range results {
		if res.Chksum == "" {
			continue
		}
		key := partialKey{res.Size, res.Chksum}
		grouped[key] = append(grouped[key], res.Filename)
	}

	filtered := SizeMap{}
	for key, filenames := range grouped {
		if len(filenames) > 1 {
			filtered[key.size] = append(filtered[key.size], filenames...)
		}
	}
	return filtered
}

// CreateDuplicationMap creates a duplicatin map objekt.
func CreateDuplicationMap(sm *SizeMap, threads int, verbose bool) DuplicateMap {
	results := RunWorkers(
		func(jobs chan ChecksumItem) {
			for size, filenames := range *sm {
				createJobs(size, filenames, jobs)
			}
		},
		func(job ChecksumItem) ChecksumItem {
			return findChkSumForFile(job.Filename, verbose)
		},
		threads)
	dupMap := DuplicateMap{}
	dupMap.createResult(sm, results, verbose)
	return dupMap
}

func (dupMap DuplicateMap) createResult(sm *SizeMap, results chan ChecksumItem, verbose bool) {
	numFiles := sm.CountCandidates()
	counter := 1
	for res := range results {
		if verbose {
			fmt.Printf("\rChecking %d of %d files.", counter, numFiles)
		}
		counter++
		key := EqualFile{res.Chksum, res.Size}
		dupMap[key] = append(dupMap[key], res.Filename)
	}
	if verbose {
		fmt.Println()
	}
}

func createJobs(size int64, filenames []string, jobs chan ChecksumItem) {
	if len(filenames) > 1 {
		for _, filename := range filenames {
			jobs <- ChecksumItem{Filename: filename, Size: size}
		}
	}
}

func partialChkSumForFile(filename string, size int64) ChecksumItem {
	f, err := os.Open(filename)
	if err != nil {
		return ChecksumItem{Filename: filename, Size: size}
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.CopyN(h, f, partialReadSize); err != nil && err != io.EOF {
		return ChecksumItem{Filename: filename, Size: size}
	}
	return ChecksumItem{Filename: filename, Size: size, Chksum: hex.EncodeToString(h.Sum(nil))}
}

func findChkSumForFile(filename string, verbose bool) ChecksumItem {
	f, err := os.Open(filename)
	if err != nil {
		if verbose {
			fmt.Printf("Got error reading %s: %s\n", filename, err)
		}
		return ChecksumItem{Filename: filename}
	}
	defer f.Close()
	h := md5.New()
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)
	size, err := io.CopyBuffer(h, f, buf)
	if err != nil {
		log.Fatal(err)
	}
	return ChecksumItem{Filename: filename, Size: size, Chksum: hex.EncodeToString(h.Sum(nil))}
}

func (dupMap DuplicateMap) dump() string {
	var buffer bytes.Buffer

	for key, files := range dupMap {
		if len(files) > 1 {
			fmt.Fprintf(&buffer, "Same files: (size: %d)\n", key.Size)
			for _, f := range files {
				fmt.Fprintf(&buffer, "  ->  %s\n", f)
			}
			fmt.Fprintln(&buffer)
		}
	}
	return buffer.String()
}
