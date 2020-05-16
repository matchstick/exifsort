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
	METHOD_NONE
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

func ParseMethod(argStr string) (int, error) {
	val, err := argsParse(argStr, methodMap)
	if err != nil {
		retErr := fmt.Errorf("Method must be one of [%s] (case insensitive)",
			argChoices(methodMap))
		return METHOD_NONE, retErr
	}
	return val, nil
}

const (
	ACTION_COPY = iota
	ACTION_MOVE
	ACTION_NONE
)

var actionMap = map[int]string{
	ACTION_COPY: "Copy",
	ACTION_MOVE: "Move",
}

func ParseAction(argStr string) (int, error) {
	val, err := argsParse(argStr, actionMap)
	if err != nil {
		retErr := fmt.Errorf("Action must be one of [%s] (case insensitive)",
			argChoices(actionMap))
		return ACTION_NONE, retErr
	}
	return val, nil
}
