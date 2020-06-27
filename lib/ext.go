package exifsort

import (
	"path/filepath"
	"strings"
)

// Supported Extensions that we process by checking "IFD/EXIF/DateTimeOriginal"
// data then if needed modTime.
//
// Set includes: bmp, cr2, dng, gif, jpeg, jpg, nef, png, psd, raf, raw, tif,
// tiff.
func ExtensionsPhoto() []string {
	// We are going to do this check a lot so let's use a map.
	return []string{
		".bmp",
		".cr2",
		".dng",
		".gif",
		".jpeg",
		".jpg",
		".nef",
		".png",
		".psd",
		".raf",
		".raw",
		".tif",
		".tiff",
	}
}

// These extensions are for files that take a long time to extract "IFD/EXIF/DateTimeOriginal"
// so we only check modTime. Set includes: 3g2, 3gp, avi, mov, mp4, m4v.
func ExtensionsMovie() []string {
	return []string{
		".3g2",
		".3gp",
		".avi",
		".mov",
		".mp4",
		".m4v",
	}
}

// Files or directories that contain these strings we assume are metadata we
// find on Synology file servers and ignore. The check is case insensitive.
// Set includes: @eadir, @syno, synofile_thumb
func SynologySkip() []string {
	return []string{
		"@eadir",
		"@syno",
		"synofile_thumb",
	}
}

type fileCategory int

const (
	categorySkip fileCategory = iota
	categoryExif
	categoryModTime
)

func categorizeFile(path string) fileCategory {
	// All comparisons are lower case as case don't matter
	path = strings.ToLower(path)

	for _, str := range SynologySkip() {
		if strings.Contains(path, str) {
			return categorySkip
		}
	}

	extension := filepath.Ext(path)
	if extension == "" {
		// no extension found so we skip
		return categorySkip
	}

	for _, str := range ExtensionsPhoto() {
		if extension == str {
			return categoryExif
		}
	}

	for _, str := range ExtensionsMovie() {
		if extension == str {
			return categoryModTime
		}
	}

	return categorySkip
}
