package exifSort

import (
	"fmt"
	"os"
	"path/filepath"
)

var sortedState SortedType

func sortFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Printf("Error accessing path %s\n", path)
		return err
	}

	// Don't need to scan directories
	if info.IsDir() {
		return nil
	}
	// Only looking for media files that may have exif.
	if skipFileType(path) {
		walkState.storeSkipped()
		return nil
	}

	time, err := ExtractExifTime(path)
	if err != nil {
		walkState.storeInvalid(path, err.Error())
		walkState.walkPrintf("%s\n", walkErrMsg(path, err.Error()))
		return nil
	}

	walkState.storeValid()
	sortedState.Add(path, time)
	return nil
}

func sortSummary(summarize bool) {
	if summarize == false {
		return
	}
	fmt.Printf("Sort Summary\n")
}

func SortDir(root string, method int, summarize bool, doPrint bool) {
	walkState.Init(doPrint)
	sortedState.Init(root, method)
	err := filepath.Walk(root, scanFunc)

	if err != nil {
		fmt.Printf("Sort Error (%s)\n", err.Error())
	}

	sortSummary(summarize)
}
