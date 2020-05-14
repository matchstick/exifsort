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
	METHOD_YEAR = iota
	METHOD_MONTH
	METHOD_DAY
	METHOD_NONE // for testing
)

var methodMap = map[int]string{
	METHOD_YEAR:  "Year",
	METHOD_MONTH: "Month",
	METHOD_DAY:   "Day",
}

func methodLookup(method int) string {
	str, present := methodMap[method]
	if present == false {
		return "unknown"
	}
	return str
}

func methodChoices() string {
	var methods []string
	for _, str := range methodMap {
		str = fmt.Sprintf("\"%s\"", str)
		methods = append(methods, str)
	}
	return strings.Join(methods, ",")
}

func MethodArgCheck(argStr string) (int, error) {
	/// lower capitilazation for safe comparing
	argStr = strings.ToLower(argStr)
	for method, methodStr := range methodMap {
		methodStr = strings.ToLower(methodStr)
		if argStr == methodStr {
			return method, nil
		}
	}
	return METHOD_NONE, fmt.Errorf("Method must be one of [%s] (case insensitive)", methodChoices())
}
