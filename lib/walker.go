package exifSort

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// We are going to do this check a lot so let's use a map.
var exifTypes = map[string]bool{
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

func skipFileType(path string) bool {
	pieces := strings.Split(path, ".")
	numPieces := len(pieces)
	if numPieces < 2 {
		return false
	}
	suffix := strings.ToLower(pieces[numPieces-1])
	return exifTypes[suffix]
}

var entries = make([]ExifDateEntry, 0)

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
	entries = append(entries, entry)
	return nil
}

func ScanDir(root string) error {

	err := filepath.Walk(root, scanFunc)
	if err != nil {
		return err
	}
	return nil
}
