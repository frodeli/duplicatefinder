package main

import (
	"log"
	"os"
	"path/filepath"
)

// SizeMap groups file paths by file size.
type SizeMap map[int64][]string

// CreateSizeMap creates a SizeMap by walking dir recursively.
func CreateSizeMap(dir string) SizeMap {
	sm := SizeMap{}
	sm.traverseDir(dir)
	return sm
}

// CountCandidates returns the number of files that share their size with at least one other file.
func (sm SizeMap) CountCandidates() int {
	counter := 0
	for _, filenames := range sm {
		if len(filenames) > 1 {
			counter += len(filenames)
		}
	}
	return counter
}

func (sm SizeMap) traverseDir(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Panic(err)
	}
	for _, f := range files {
		path := filepath.Join(dir, f.Name())
		if f.IsDir() {
			sm.traverseDir(path)
		} else {
			info, err := f.Info()
			if err != nil {
				log.Panic(err)
			}
			sm[info.Size()] = append(sm[info.Size()], path)
		}
	}
}
