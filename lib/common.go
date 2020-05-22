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
	case strings.Contains(path, "@eaDir"):
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

func skipFileType(path string) bool {
	// All comparisons are lower case as case don't matter
	path = strings.ToLower(path)
	if isSynologyFile(path) {
		// skip
		return true
	}

	pieces := strings.Split(path, ".")
	minSplitLen := 2 // We expect there to be at least two pieces

	numPieces := len(pieces)
	if numPieces < minSplitLen {
		// skip
		return true
	}

	suffix := pieces[numPieces-1]
	_, inMediaMap := mediaSuffixMap()[suffix]

	return !inMediaMap
}

// Methods to sort media files in nested directory structure.
const (
	MethodYear  = iota // Year : dst -> year-> media
	MethodMonth        // Year : dst -> year-> month -> media
	MethodDay          // Year : dst -> year-> month -> day -> media
	MethodNone
)

func argChoices(argsMap map[int]string) string {
	var choices = make([]string, len(argsMap))

	for _, str := range argsMap {
		str = fmt.Sprintf("\"%s\"", str)
		choices = append(choices, str)
	}

	return strings.Join(choices, ",")
}

func argsParse(argStr string, argsMap map[int]string) (int, error) {
	/// lower capitalization for safe comparing
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
	var methodMap = map[int]string{
		MethodYear:  "Year",
		MethodMonth: "Month",
		MethodDay:   "Day",
	}

	val, err := argsParse(str, methodMap)
	if err != nil {
		retErr := fmt.Errorf("Method must be one of [%s] (case insensitive)",
			argChoices(methodMap))
		return MethodNone, retErr
	}

	return val, nil
}

// When sorting media you specify action to transfer files form the src to dst
// directories.
const (
	ActionCopy = iota // copying
	ActionMove        // moving
	ActionNone
)

// ParseAction returns the constant value of the str
// Input is case insensitive.
func ParseAction(str string) (int, error) {
	var actionMap = map[int]string{
		ActionCopy: "Copy",
		ActionMove: "Move",
	}

	val, err := argsParse(str, actionMap)
	if err != nil {
		retErr := fmt.Errorf("Action must be one of [%s] (case insensitive)",
			argChoices(actionMap))
		return ActionNone, retErr
	}

	return val, nil
}
