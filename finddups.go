package main

import (
	"bytes"
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

// EqualFile holds the checksum and size of the equal files.
type EqualFile struct {
	Chksum string
	Size   int64
}

// DuplicateMap holds duplicate files.
type DuplicateMap map[EqualFile]*list.List

// CreateDuplicationMap creates a duplicatin map objekt.
func CreateDuplicationMap(sm *SizeMap, threads int, verbose bool) DuplicateMap {
	results := RunWorkers(
		func(jobs chan ChecksumItem) {
			for size, filenameList := range *sm {
				createJobs(size, filenameList, jobs)
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
		sameChkSumList := dupMap[key]

		if sameChkSumList == nil {
			sameChkSumList = &list.List{}
			dupMap[key] = sameChkSumList
		}
		sameChkSumList.PushBack(res.Filename)
	}
	if verbose {
		fmt.Println()
	}
}

func createJobs(size int64, filenameList *list.List, jobs chan ChecksumItem) {
	if filenameList.Len() > 1 {
		for filename := filenameList.Front(); filename != nil; filename = filename.Next() {
			jobs <- ChecksumItem{Filename: filename.Value.(string), Size: size}
		}
	}
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
	size, err := io.Copy(h, f)
	if err != nil {
		log.Fatal(err)
	}
	return ChecksumItem{Filename: filename, Size: size, Chksum: hex.EncodeToString(h.Sum(nil))}
}

func (dupMap DuplicateMap) dump() string {
	var buffer bytes.Buffer

	for key, filelist := range dupMap {
		if filelist.Len() > 1 {
			fmt.Fprintf(&buffer, "Same files: (size: %d)\n", key.Size)
			for f := filelist.Front(); f != nil; f = f.Next() {
				fmt.Fprintf(&buffer, "  ->  %s\n", f.Value.(string))
			}
			fmt.Fprintln(&buffer)
		}
	}
	return buffer.String()
}
