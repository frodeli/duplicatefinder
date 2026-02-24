package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

const chunkSize = 4096

// EqualFile holds the checksum and size of the equal files.
type EqualFile struct {
	Chksum string
	Size   int64
}

// DuplicateMap holds duplicate files.
type DuplicateMap map[EqualFile][]string

// CreateDuplicationMap creates a duplicatin map objekt.
func CreateDuplicationMap(sm *SizeMap, threads int, verbose bool) DuplicateMap {
	// Partial hash calculation
	partialResults := RunWorkers(
		func(jobs chan DataMap) {
			for size, filenameList := range *sm {
				if len(filenameList) > 1 {
					for _, filename := range filenameList {
						jobs <- DataMap{"filename": filename, "size": size}
					}
				}
			}
		},
		func(job DataMap) DataMap {
			filename := job["filename"].(string)
			partialHash, err := partialHash(filename)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: Could not calculate partial hash for %s: %s\n", filename, err)
				}
				return nil
			}
			job["partialHash"] = partialHash
			return job
		},
		threads)

	// Group by partial hash
	partialMap := make(map[string][]DataMap)
	for res := range partialResults {
		if res != nil {
			partialHash := res["partialHash"].(string)
			partialMap[partialHash] = append(partialMap[partialHash], res)
		}
	}

	// Full hash calculation for files with same partial hash
	fullResults := RunWorkers(
		func(jobs chan DataMap) {
			for _, jobList := range partialMap {
				if len(jobList) > 1 {
					for _, job := range jobList {
						jobs <- job
					}
				}
			}
		},
		func(job DataMap) DataMap {
			filename := job["filename"].(string)
			fullHash, err := fullHash(filename)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: Could not calculate full hash for %s: %s\n", filename, err)
				}
				return nil
			}
			job["fullHash"] = fullHash
			return job
		},
		threads)

	dupMap := DuplicateMap{}
	dupMap.createResult(fullResults, verbose)
	return dupMap
}

func (dupMap DuplicateMap) createResult(results chan DataMap, verbose bool) {
	for res := range results {
		if res == nil {
			continue
		}

		filename := res["filename"].(string)
		size := res["size"].(int64)
		chksum := res["fullHash"].(string)

		key := EqualFile{chksum, size}
		dupMap[key] = append(dupMap[key], filename)
	}
}

func createJobs(size int64, filenameList []string, jobs chan DataMap) {
	if len(filenameList) > 1 {
		for _, filename := range filenameList {
			jobs <- DataMap{"filename": filename, "size": size}
		}
	}
}

func bytesToHash(bytes []byte) string {
	return fmt.Sprintf("%x", bytes)
}

func partialHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	chunk := make([]byte, chunkSize)
	bytesRead, err := f.Read(chunk)
	if err != nil && err != io.EOF {
		return "", err
	}

	_, err = h.Write(chunk[:bytesRead])
	if err != nil {
		return "", err
	}

	return bytesToHash(h.Sum(nil)), nil
}

func fullHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return bytesToHash(h.Sum(nil)), nil
}

func (dupMap DuplicateMap) dump() string {
	var buffer bytes.Buffer

	for key, filelist := range dupMap {
		if len(filelist) > 1 {
			fmt.Fprintf(&buffer, "Same files: (size: %d)\n", key.Size)
			for _, f := range filelist {
				fmt.Fprintf(&buffer, "  ->  %s\n", f)
			}
			fmt.Fprintln(&buffer)
		}
	}
	return buffer.String()
}