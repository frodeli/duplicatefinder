package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"
)

func findSameFiles(rootDir string, threads int, verbose bool) {
	before := time.Now()
	sizeMap := CreateSizeMap(rootDir)
	between := time.Now()
	dupMap := CreateDuplicationMap(&sizeMap, threads, verbose)
	after := time.Now()
	fmt.Print(dupMap.dump())

	if verbose {
		fmt.Printf("Time elapsed: %s (size) %s (checksum) %s (total)\n", between.Sub(before), after.Sub(between), after.Sub(before))
	}
}

func main() {
	rootDir := flag.String("root", ".", "Root directory to search from.")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads used to compare files.")
	verbose := flag.Bool("verbose", false, "Turn on more output.")
	flag.Parse()

	findSameFiles(*rootDir, *threads, *verbose)
}
