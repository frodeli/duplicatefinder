package main

import (
	"container/list"
	"io/ioutil"
	"log"
	"os"
)

// SizeMap is for finding files of same size.
type SizeMap map[int64]*list.List

// CreateSizeMap creates a SizeMap for finding groups of files of smae size.
func CreateSizeMap(dir string) SizeMap {
	sm := SizeMap{}
	sm.traverseDir(dir)
	return sm
}

// CountCandidates count how many equal file candidates is in the map.
func (sm SizeMap) CountCandidates() int {
	counter := 0
	for _, filenameList := range sm {
		if filenameList.Len() > 1 {
			counter += filenameList.Len()
		}
	}
	return counter
}

func (sm SizeMap) traverseDir(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Panic(err)
	}
	for _, f := range files {
		if f.IsDir() {
			sm.traverseDir(dir + "/" + f.Name())
		} else {
			sm.traverseFile(f, dir+"/"+f.Name())
		}
	}
}
func (sm SizeMap) traverseFile(fileinfo os.FileInfo, filename string) {
	var l = sm[fileinfo.Size()]
	if l == nil {
		l = &list.List{}
		sm[fileinfo.Size()] = l
	}
	l.PushBack(filename)
}
