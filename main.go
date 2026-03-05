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
	between1 := time.Now()
	filteredMap := FilterByPartialHash(&sizeMap, threads)
	between2 := time.Now()
	dupMap := CreateDuplicationMap(&filteredMap, threads, verbose)
	after := time.Now()
	fmt.Print(dupMap.dump())

	if verbose {
		fmt.Printf("Time elapsed: %s (size) %s (partial) %s (checksum) %s (total)\n",
			between1.Sub(before), between2.Sub(between1), after.Sub(between2), after.Sub(before))
	}
}

func main() {
	rootDir := flag.String("root", ".", "Root directory to search from.")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads used to compare files.")
	verbose := flag.Bool("verbose", false, "Turn on more output.")
	gui := flag.Bool("gui", false, "Launch graphical user interface.")
	flag.Parse()

	if *gui {
		launchGUI()
		return
	}

	findSameFiles(*rootDir, *threads, *verbose)
}
