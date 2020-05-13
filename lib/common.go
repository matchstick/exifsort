package exifSort

import (
	"fmt"
	"strings"
	"time"
)

func ExifTime(t time.Time) string {
	return fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

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

	switch {
	case strings.Contains(path, "@eaDir"):
		return true
	case strings.Contains(path, "@syno"):
		return true
	case strings.Contains(path, "synofile_thumb"):
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

const (
	METHOD_NONE = iota
	METHOD_YEAR
	METHOD_MONTH
	METHOD_DAY
)

var methodMap = map[int]string{
	METHOD_NONE:  "None",
	METHOD_YEAR:  "Year",
	METHOD_MONTH: "Month",
	METHOD_DAY:   "Day",
}

func methodStr(method int) string {
	str, present := methodMap[method]
	if present == false {
		return "unknown"
	}
	return str
}
