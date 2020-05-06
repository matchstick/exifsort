package exifSort

import (
	"fmt"
	"github.com/stretchr/powerwalk"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
)

// We are going to do this check a lot so let's use a map.
var mediaSuffixMap = map[string]bool{
	"bmp":  true,
	"cr2":  true,
	"dng":  true,
	"gif":  true,
	"jpeg": true,
	"jpg":  true,
	"nef":  true,
	"png":  true,
	"psd":  true,
	"raf":  true,
	"raw":  true,
	"tif":  true,
	"tiff": true,
}

// Running this on a synology results in the file server creating all these
// useless media files. We want to skip them.
func isSynologyFile(path string) bool {
	if strings.Contains(path, "@eaDir") {
		return true
	}
	if strings.Contains(path, "@syno") {
		return true
	}
	if strings.Contains(path, "synofile_thumb") {
		return true
	}

	return false
}

func skipFileType(path string) bool {
	// All comparisons are lower case as case don't matter
	path = strings.ToLower(path)
	if isSynologyFile(path) {
		// skip
		return true
	}
	pieces := strings.Split(path, ".")
	numPieces := len(pieces)
	if numPieces < 2 {
		// skip
		return true
	}
	suffix := pieces[numPieces-1]
	_, inMediaMap := mediaSuffixMap[suffix]
	if inMediaMap == false {
		// skip
		return true
	}
	return false
}

type scanStateType struct {
	skipped     uint64
	validDate   uint64
	invalidDate uint64
	printScan   bool
	errFiles    []string
}

var scanState scanStateType

func resetScanState(printScan bool) {
	scanState = scanStateType{0, 0, 0, printScan, nil}
}

func scanPrintf(s string, params ...interface{}) {
	if scanState.printScan == false {
		return
	}

	if len(params) == 0 {
		fmt.Printf(s)
	}

	fmt.Printf(s, params...)
}

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
		atomic.AddUint64(&scanState.skipped, 1)
		return nil
	}

	entry, err := ExtractExifDate(path)
	if err != nil {
		returnErr := fmt.Errorf("ERROR with File %s with (%s)",
			path, err.Error())
		fmt.Println(returnErr)
		scanState.errFiles = append(scanState.errFiles, path)
		return nil
	}

	if entry.Valid == false {
		atomic.AddUint64(&scanState.invalidDate, 1)
		scanPrintf("%s, %s\n", entry.Path, "None")
		return nil
	}

	scanPrintf("%s, %s\n", entry.Path, ExifTime(entry.Time))
	atomic.AddUint64(&scanState.validDate, 1)
	return nil
}

func scanSummary(summarize bool) {
	if summarize == false {
		return
	}
	total := scanState.skipped + scanState.invalidDate +
		scanState.validDate
	fmt.Printf("Scanned Valid: %d\n", scanState.validDate)
	fmt.Printf("Scanned Invalid: %d\n", scanState.invalidDate)
	fmt.Printf("Scanned Skipped: %d\n", scanState.skipped)
	fmt.Printf("Scanned Total: %d\n", total)
	if len(scanState.errFiles) == 0 {
		fmt.Println("No Files caused Errors")
		return
	}
	fmt.Println("Error Files were:")
	for _, path := range scanState.errFiles {
		fmt.Printf("\t%s\n", path)
	}
}

func ScanDir(root string, summarize bool, printScan bool, cpus int) {
	resetScanState(printScan)

	if cpus == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else if cpus > runtime.NumCPU() {
		fmt.Printf("Specified %d cpu but only have %d\n",
			cpus, runtime.NumCPU())
		return
	} else {
		fmt.Printf("Using %d CPUs\n", cpus)
		runtime.GOMAXPROCS(cpus)
	}

	err := powerwalk.Walk(root, scanFunc)
	runtime.GOMAXPROCS(1)

	if err != nil {
		fmt.Printf("Scan Error (%s)\n", err.Error())
	}

	scanSummary(summarize)
}
