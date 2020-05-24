package exifsort

import (
	"fmt"
	"strings"
	"time"
)

func exifTimeToStr(t time.Time) string {
	return fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// Files suffixes that are processed and not skipped.
const (
	SuffixBMP = iota
	SuffixCR2
	SuffixDNG
	SuffixGIF
	SuffixJPEG
	SuffixJPG
	SuffixNEF
	SuffixPNG
	SuffixPSD
	SuffixRAF
	SuffixRAW
	SuffixTIF
	SuffixTIFF
)

// Running this on a synology results in the file server creating all these
// useless media files. We want to skip them.
func isSynologyFile(path string) bool {
	switch {
	case strings.Contains(path, "@eadir"):
		return true
	case strings.Contains(path, "@syno"):
		return true
	case strings.Contains(path, "synofile_thumb"):
		return true
	}

	return false
}

func mediaSuffixMap() map[string]int {
	// We are going to do this check a lot so let's use a map.
	return map[string]int{
		"bmp":  SuffixBMP,
		"cr2":  SuffixCR2,
		"dng":  SuffixDNG,
		"gif":  SuffixGIF,
		"jpeg": SuffixJPEG,
		"jpg":  SuffixJPG,
		"nef":  SuffixNEF,
		"png":  SuffixPNG,
		"psd":  SuffixPSD,
		"raf":  SuffixRAF,
		"raw":  SuffixRAW,
		"tif":  SuffixTIF,
		"tiff": SuffixTIFF,
	}
}

const minSplitLen = 2 // We expect there to be at least two pieces

func skipFileType(path string) bool {
	// All comparisons are lower case as case don't matter
	path = strings.ToLower(path)
	if isSynologyFile(path) {
		// skip
		return true
	}

	pieces := strings.Split(path, ".")

	numPieces := len(pieces)
	if numPieces < minSplitLen {
		// skip
		return true
	}

	suffix := pieces[numPieces-1]
	_, inMediaMap := mediaSuffixMap()[suffix]

	return !inMediaMap
}
