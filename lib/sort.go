package exifSort

import (
	"fmt"
	"os"
	"path/filepath"
)

var sortIndex index

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

	time, err := ExtractTime(path)
	if err != nil {
		walkState.storeInvalid(path, err.Error())
		walkState.walkPrintf("%s\n", walkErrMsg(path, err.Error()))
		return nil
	}

	walkState.storeValid()
	sortIndex.Put(path, time)
	return nil
}

func sortSummary(summarize bool) {
	if summarize == false {
		return
	}
	fmt.Printf("Sort Summary\n")
}

// SortDir examines the contents of file with acceptable suffixes in the src
// directory and transfer the file to the dst directory. The structure of
// the dst directory is specifed by 'method'. The action of transfer is
// specified by 'action'.
//
// SortDir only scans media files listed in the mediaSuffixMap, other files are
// skipped. 
//
// When 'summarize' is set to 'true' it will print a summary of stats when
// completed.
//
// If doPrint is set to false it will not print while scanning.
func SortDir(src string, dst string, method int, action int, summarize bool, doPrint bool) {
	walkState.init(doPrint)
	sortIndex = createIndex(method)
	err := filepath.Walk(src, sortFunc)
	if err != nil {
		fmt.Printf("Sort Error (%s)\n", err.Error())
	}
	sortSummary(summarize)
}
