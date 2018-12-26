package main

import (
	"bytes"
	"container/list"
	"crypto/md5"
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
		func(jobs chan DataMap) {
			for size, filenameList := range *sm {
				createJobs(size, filenameList, jobs)
			}
		},
		func(job DataMap) DataMap {
			return findChkSumForFile(job["filename"].(string), verbose)
		},
		threads)
	dupMap := DuplicateMap{}
	dupMap.createResult(sm, results, verbose)
	return dupMap
}

func (dupMap DuplicateMap) createResult(sm *SizeMap, results chan DataMap, verbose bool) {
	numFiles := sm.CountCandidates()
	counter := 1
	for res := range results {
		if verbose {
			fmt.Printf("\rChecking %d of %d files.", counter, numFiles)
		}
		counter++
		key := EqualFile{res["chksum"].(string), res["size"].(int64)}
		sameChkSumList := dupMap[key]

		if sameChkSumList == nil {
			sameChkSumList = &list.List{}
			dupMap[key] = sameChkSumList
		}
		sameChkSumList.PushBack(res["filename"])
	}
	if verbose {
		fmt.Println()
	}
}

func createJobs(size int64, filenameList *list.List, jobs chan DataMap) {
	if filenameList.Len() > 1 {
		for filename := filenameList.Front(); filename != nil; filename = filename.Next() {
			jobs <- DataMap{"filename": filename.Value.(string), "size": size}
		}
	}
}

func bytesToHash(bytes []byte) string {
	hashString := ""
	for _, b := range bytes {
		hashString += fmt.Sprintf("%02x", b)
	}
	return hashString
}

func findChkSumForFile(filename string, verbose bool) DataMap {
	f, err := os.Open(filename)
	if err != nil {
		if verbose {
			fmt.Printf("Got error reading %s: %s\n", filename, err)
		}
	}
	defer f.Close()
	h := md5.New()
	size, err := io.Copy(h, f)
	if err != nil {
		log.Fatal(err)
	}
	return DataMap{"chksum": bytesToHash(h.Sum(nil)), "size": size, "filename": filename}
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
