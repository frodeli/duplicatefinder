package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// SizeMap is for finding files of same size.
type SizeMap map[int64][]string

// CreateSizeMap creates a SizeMap for finding groups of files of same size.
func CreateSizeMap(dir string) SizeMap {
	sm := SizeMap{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Warning: Could not access %s: %s\n", path, err)
			return filepath.SkipDir
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				fmt.Printf("Warning: Could not get info for %s: %s\n", path, err)
				return nil
			}
			sm.add(info.Size(), path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("An unexpected error occurred: %s\n", err)
	}
	return sm
}

// CountCandidates count how many equal file candidates is in the map.
func (sm SizeMap) CountCandidates() int {
	counter := 0
	for _, filenameList := range sm {
		if len(filenameList) > 1 {
			counter += len(filenameList)
		}
	}
	return counter
}

func (sm SizeMap) add(size int64, path string) {
	sm[size] = append(sm[size], path)
}
