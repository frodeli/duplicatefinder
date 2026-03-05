package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func launchGUI() {
	a := app.New()
	w := a.NewWindow("Duplicate File Finder")
	w.Resize(fyne.NewSize(700, 600))

	// --- inputs ---
	dirEntry := widget.NewEntry()
	dirEntry.SetPlaceHolder("Select a directory…")

	threadsEntry := widget.NewEntry()
	threadsEntry.SetText(strconv.Itoa(runtime.NumCPU()))
	threadsEntry.Resize(fyne.NewSize(60, threadsEntry.MinSize().Height))

	browseBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			dirEntry.SetText(uri.Path())
		}, w)
	})

	// --- progress / status ---
	progress := widget.NewProgressBarInfinite()
	progress.Hide()
	statusLabel := widget.NewLabel("")
	statusLabel.Hide()

	// --- results ---
	resultsBox := container.NewVBox()
	scroll := container.NewVScroll(resultsBox)
	scroll.SetMinSize(fyne.NewSize(680, 380))

	// --- scan button ---
	var scanBtn *widget.Button
	scanBtn = widget.NewButtonWithIcon("Find Duplicates", theme.SearchIcon(), func() {
		rootDir := dirEntry.Text
		if rootDir == "" {
			dialog.ShowError(fmt.Errorf("please select a directory first"), w)
			return
		}
		threads, err := strconv.Atoi(threadsEntry.Text)
		if err != nil || threads < 1 {
			dialog.ShowError(fmt.Errorf("threads must be a positive integer"), w)
			return
		}

		// reset UI state
		resultsBox.Objects = nil
		resultsBox.Refresh()
		scanBtn.Disable()
		progress.Show()
		statusLabel.SetText("Phase 1/3: building size map…")
		statusLabel.Show()

		go func() {
			sizeMap := CreateSizeMap(rootDir)

			statusLabel.SetText("Phase 2/3: partial hash filter…")
			filteredMap := FilterByPartialHash(&sizeMap, threads)

			statusLabel.SetText("Phase 3/3: computing checksums…")
			dupMap := CreateDuplicationMap(&filteredMap, threads, false)

			// render results on the main goroutine via a channel-free approach
			// (Fyne is safe to update from goroutines for most widget ops)
			renderResults(dupMap, resultsBox, w)

			progress.Hide()
			if len(dupMap) == 0 {
				statusLabel.SetText("No duplicates found.")
			} else {
				statusLabel.SetText(fmt.Sprintf("Found %d group(s) of duplicate files.", countGroups(dupMap)))
			}
			scanBtn.Enable()
		}()
	})
	scanBtn.Importance = widget.HighImportance

	// --- layout ---
	dirRow := container.New(layout.NewBorderLayout(nil, nil, nil, browseBtn),
		browseBtn, dirEntry)

	threadsRow := container.NewHBox(
		widget.NewLabel("Threads:"),
		threadsEntry,
	)

	controls := container.NewVBox(
		widget.NewSeparator(),
		dirRow,
		threadsRow,
		container.NewCenter(scanBtn),
		container.NewHBox(progress, statusLabel),
		widget.NewSeparator(),
	)

	w.SetContent(container.NewBorder(controls, nil, nil, nil, scroll))
	w.ShowAndRun()
}

// renderResults populates resultsBox with one card per duplicate group.
func renderResults(dupMap DuplicateMap, resultsBox *fyne.Container, w fyne.Window) {
	resultsBox.Objects = nil

	for key, files := range dupMap {
		if len(files) < 2 {
			continue
		}

		// capture loop vars for closures
		groupKey := key
		groupFiles := make([]string, len(files))
		copy(groupFiles, files)

		header := widget.NewLabelWithStyle(
			fmt.Sprintf("Same files  (size: %s)", humanSize(groupKey.Size)),
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		)

		groupBox := container.NewVBox(header)
		var fileRows []*fyne.Container

		var buildRows func()
		buildRows = func() {
			groupBox.Objects = []fyne.CanvasObject{header}
			fileRows = fileRows[:0]
			for i, path := range groupFiles {
				if path == "" {
					continue
				}
				idx := i
				pathLabel := widget.NewLabel(path)
				pathLabel.Truncation = fyne.TextTruncateEllipsis

				deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
					dialog.ShowConfirm(
						"Delete file?",
						fmt.Sprintf("Permanently delete:\n%s", groupFiles[idx]),
						func(confirmed bool) {
							if !confirmed {
								return
							}
							if err := os.Remove(groupFiles[idx]); err != nil {
								dialog.ShowError(err, w)
								return
							}
							groupFiles[idx] = "" // mark deleted
							buildRows()
							resultsBox.Refresh()
						}, w)
				})
				deleteBtn.Importance = widget.DangerImportance

				row := container.New(layout.NewBorderLayout(nil, nil, nil, deleteBtn),
					deleteBtn, pathLabel)
				fileRows = append(fileRows, row)
				groupBox.Add(row)
			}

			// hide group if fewer than 2 files remain
			active := 0
			for _, p := range groupFiles {
				if p != "" {
					active++
				}
			}
			if active < 2 {
				groupBox.Hide()
			} else {
				groupBox.Show()
			}
		}

		buildRows()
		resultsBox.Add(groupBox)
		resultsBox.Add(widget.NewSeparator())
	}

	resultsBox.Refresh()
}

func countGroups(dupMap DuplicateMap) int {
	n := 0
	for _, files := range dupMap {
		if len(files) >= 2 {
			n++
		}
	}
	return n
}

func humanSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
