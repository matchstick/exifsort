package exifSort

import (
	"fmt"
	"os"
	"path/filepath"
)

func scanFunc(path string, info os.FileInfo, err error) error {
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

	walkState.walkPrintf("%s, %s\n", path, ExifTime(time))
	walkState.storeValid()
	return nil
}

func scanSummary(summarize bool) {
	if summarize == false {
		return
	}
	fmt.Printf("Scanned Valid: %d\n", walkState.valid())
	fmt.Printf("Scanned Invalid: %d\n", walkState.invalid())
	fmt.Printf("Scanned Skipped: %d\n", walkState.skipped())
	fmt.Printf("Scanned Total: %d\n", walkState.total())
	if walkState.invalid() == 0 {
		fmt.Println("No Files caused Errors")
		return
	}

	fmt.Println("Error Files were:")
	for path, msg := range walkState.errs() {
		fmt.Printf("\t%s\n", walkErrMsg(path, msg))
	}
}

func ScanDir(root string, summarize bool, doPrint bool) {
	walkState.resetWalkState(doPrint)

	err := filepath.Walk(root, scanFunc)

	if err != nil {
		fmt.Printf("Scan Error (%s)\n", err.Error())
	}

	scanSummary(summarize)
}
