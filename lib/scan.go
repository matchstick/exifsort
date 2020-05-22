package exifsort

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

	time, err := ExtractTime(path)
	if err != nil {
		walkState.storeInvalid(path, err.Error())
		walkState.walkPrintf("%s\n", walkErrMsg(path, err.Error()))
		return nil
	}

	walkState.walkPrintf("%s, %s\n", path, exifTimeToStr(time))
	walkState.storeValid()
	return nil
}

func scanSummary(summarize bool) {
	if !summarize {
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
	for path, msg := range walkState.walkErrs() {
		fmt.Printf("\t%s\n", walkErrMsg(path, msg))
	}
}

// ScanDir will examine the contents of every file in the src directory and
// print it's time of creation as stored by exifdata as it scans.
//
// ScanDir only scans media files listed in the mediaSuffixMap, other files are
// skipped.
//
// When 'summarize' is set to 'true' it will print a summary of stats when
// completed.
//
// If doPrint is set to false it will not print while scanning.
func ScanDir(src string, summarize bool, doPrint bool) {
	walkState.init(doPrint)

	err := filepath.Walk(src, scanFunc)

	if err != nil {
		fmt.Printf("Scan Error (%s)\n", err.Error())
	}

	scanSummary(summarize)
}
