package exifSort

import (
	"fmt"
	"strings"
	"sync/atomic"
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
	if strings.Contains(path, "@eaDir") {
		return true
	}
	if strings.Contains(path, "@syno") {
		return true
	}
	if strings.Contains(path, "synofile_thumb") {
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

type walkStateType struct {
	skippedCount uint64
	validCount   uint64
	invalidCount uint64
	walkDoPrint  bool
	errMsgs      map[string]string
}

func (w *walkStateType) skipped() uint64 {
	return w.skippedCount
}

func (w *walkStateType) valid() uint64 {
	return w.validCount
}

func (w *walkStateType) invalid() uint64 {
	return w.invalidCount
}

func (w *walkStateType) errs() map[string]string {
	return w.errMsgs
}

func (w *walkStateType) total() uint64 {
	return w.skippedCount + w.validCount + w.invalidCount
}

func (w *walkStateType) storeValid() {
	atomic.AddUint64(&w.validCount, 1)
}

// We don't check if you have a path duplicate
func (w *walkStateType) storeInvalid(path string, errStr string) {
	atomic.AddUint64(&w.invalidCount, 1)
	w.errMsgs[path] = errStr
}

func (w *walkStateType) storeSkipped() {
	atomic.AddUint64(&w.skippedCount, 1)
}

// has to be a global so it can be accessed via walk routines
var walkState walkStateType

func (w *walkStateType) resetWalkState(walkDoPrint bool) {
	w.skippedCount = 0
	w.validCount = 0
	w.invalidCount = 0
	w.walkDoPrint = walkDoPrint
	walkState.errMsgs = make(map[string]string)
}

func (w *walkStateType) walkPrintf(s string, params ...interface{}) {
	if w.walkDoPrint == false {
		return
	}

	if len(params) == 0 {
		fmt.Printf(s)
	}

	fmt.Printf(s, params...)
}

func walkErrMsg(path string, errMsg string) string {
	return fmt.Sprintf("%s with (%s)", path, errMsg)
}
