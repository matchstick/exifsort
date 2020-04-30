package exifSort

import (
	"fmt"
	"github.com/stretchr/powerwalk"
	"os"
	"runtime"
	"strings"
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
		return nil
	}

	entry, err := ExtractExifDate(path)
	if err != nil {
		return err
	}
	if entry.Valid == false {
		fmt.Printf("%s,%s\n", entry.Path, "None")
		return nil
	}

	fmt.Printf("%s,%s\n", entry.Path, ExifTime(entry.Time))
	return nil
}

func ScanDir(root string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	err := powerwalk.Walk(root, scanFunc)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}
