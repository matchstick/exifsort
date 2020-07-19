package exifsort

import (
	"path/filepath"
	"strings"
)

// ExtensionsPhoto returns set of supported Extensions that we process.
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

// ExtensionsMovie returns the set of extensions for files that take a long time
// to extract exif data so we only check modTime for these extensions. Set
// includes: 3g2, 3gp, avi, m4v, mov, mp4, mpg, wmv.
func ExtensionsMovie() []string {
	return []string{
		".3g2",
		".3gp",
		".avi",
		".m4v",
		".mpg",
		".mov",
		".mp4",
		".wmv",
	}
}

// SynologySkip returns the set of files or directories that contain strings we
// find on Synology file servers and ignore. The check is case insensitive.
//
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
