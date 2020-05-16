package exifSort

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
	SUFFIX_BMP = iota
	SUFFIX_CR2
	SUFFIX_DNG
	SUFFIX_GIF
	SUFFIX_JPEG
	SUFFIX_JPG
	SUFFIX_NEF
	SUFFIX_PNG
	SUFFIX_PSD
	SUFFIX_RAF
	SUFFIX_RAW
	SUFFIX_TIF
	SUFFIX_TIFF
)

// We are going to do this check a lot so let's use a map.
var mediaSuffixMap = map[string]int{
	"bmp":  SUFFIX_BMP,
	"cr2":  SUFFIX_CR2,
	"dng":  SUFFIX_DNG,
	"gif":  SUFFIX_GIF,
	"jpeg": SUFFIX_JPEG,
	"jpg":  SUFFIX_JPG,
	"nef":  SUFFIX_NEF,
	"png":  SUFFIX_PNG,
	"psd":  SUFFIX_PSD,
	"raf":  SUFFIX_RAF,
	"raw":  SUFFIX_RAW,
	"tif":  SUFFIX_TIF,
	"tiff": SUFFIX_TIFF,
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
	return !inMediaMap
}

// Methods to sort media files in nested directory structure.
const (
	METHOD_YEAR  = iota // Year : dst -> year-> media
	METHOD_MONTH        // Year : dst -> year-> month -> media
	METHOD_DAY          // Year : dst -> year-> month -> day -> media
	METHOD_NONE
)

var methodMap = map[int]string{
	METHOD_YEAR:  "Year",
	METHOD_MONTH: "Month",
	METHOD_DAY:   "Day",
}

func argChoices(argsMap map[int]string) string {
	var choices []string
	for _, str := range argsMap {
		str = fmt.Sprintf("\"%s\"", str)
		choices = append(choices, str)
	}
	return strings.Join(choices, ",")
}

func argsParse(argStr string, argsMap map[int]string) (int, error) {
	/// lower capitilazation for safe comparing
	argStr = strings.ToLower(argStr)
	for val, str := range argsMap {
		str = strings.ToLower(str)
		if argStr == str {
			return val, nil
		}
	}
	return -1, fmt.Errorf("unknown arg")
}

// ParseMethod returns the constant value of the str
// Input is case insensitive.
func ParseMethod(str string) (int, error) {
	val, err := argsParse(str, methodMap)
	if err != nil {
		retErr := fmt.Errorf("Method must be one of [%s] (case insensitive)",
			argChoices(methodMap))
		return METHOD_NONE, retErr
	}
	return val, nil
}

// When sorting media you specify action to transfer files form the src to dst
// directories.
const (
	ACTION_COPY = iota // copying
	ACTION_MOVE        // moving
	ACTION_NONE
)

var actionMap = map[int]string{
	ACTION_COPY: "Copy",
	ACTION_MOVE: "Move",
}

// ParseAction returns the constant value of the str
// Input is case insensitive.
func ParseAction(str string) (int, error) {
	val, err := argsParse(str, actionMap)
	if err != nil {
		retErr := fmt.Errorf("Action must be one of [%s] (case insensitive)",
			argChoices(actionMap))
		return ACTION_NONE, retErr
	}
	return val, nil
}
