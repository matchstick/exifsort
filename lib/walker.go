package exifSort

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func skipFileType(path string) (bool, error) {

	// We are going to do this check a lot so let's use a map.
	exifTypes := map[string]int{
		"bmp":  1,
		"CR2":  1,
		"dng":  1,
		"gif":  1,
		"jpeg": 1,
		"jpg":  1,
		"nef":  1,
		"png":  1,
		"psd":  1,
		"RAF":  1,
		"raw":  1,
		"tif":  1,
		"tiff": 1,
	}

	pieces := strings.Split(path, ".")
	numPieces := len(pieces)
	if numPieces < 2 {
		return false, fmt.Errorf("No suffix to split for %s\n", path)
	}
	suffix := pieces[numPieces-1]
	fmt.Printf("Suffix %s\n", suffix)
	skip := exifTypes[suffix] == 0
	return skip, nil
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
	skip, err := skipFileType(path)
	if err != nil {
		fmt.Printf("%q\n", err)
		return nil
	}

	if skip {
		return nil
	}

	entry, err := ExtractExifDate(path)
	if err != nil {
		return err
	}
	if entry.Valid == false {
		fmt.Printf("No Exif Data\n")
		return nil
	}
	fmt.Printf("Retrieved %+v\n", entry)
	return nil
}

func ScanDir(root string) error {
	err := filepath.Walk(root, scanFunc)
	if err != nil {
		return err
	}
	return nil
}
