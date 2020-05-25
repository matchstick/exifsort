package exifsort

import (
	"fmt"
	"io"
	"time"
)

// WalkState holds all the data gathered as exifstort scans the src directories
// and if sorting transfers media.
type WalkState struct {
	skippedCount     int
	printer          io.Writer
	walkData	 map[string]time.Time
	walkErrors       map[string]string
	transferErrMsgs  map[string]string
}

// Returns how many files were skipped.
func (w *WalkState) Skipped() int {
	return w.skippedCount
}

// Returns how many files had valid exif DateTimeOriginal data.
func (w *WalkState) Valid() int {
	return len(w.walkData)
}

// Returns how many files had invalid exif DateTimeOriginal data.
func (w *WalkState) Invalid() int {
	return  len(w.walkErrors)
}

// Returns a map from path to error scanning.
func (w *WalkState) Errors() map[string]string {
	return w.walkErrors
}

// Returns a map from path to error transferring.
func (w *WalkState) TransferErrs() map[string]string {
	return w.transferErrMsgs
}

// Returns the total number of files skipped and scanned.
func (w *WalkState) Total() int {
	return w.skippedCount + w.Valid() + w.Invalid()
}

// We don't check if you have a path duplicate.
func (w *WalkState) storeValid(path string, time time.Time) {
	w.walkData[path] = time
}

// We don't check if you have a path duplicate.
func (w *WalkState) storeInvalid(path string, errStr string) {
	w.walkErrors[path] = errStr
}

// We don't check if you have a path duplicate.
func (w *WalkState) storeTransferErr(path string, errStr string) {
	w.transferErrMsgs[path] = errStr
}

func (w *WalkState) storeSkipped() {
	w.skippedCount++
}

func newWalkState(printer io.Writer) WalkState {
	var w WalkState
	w.skippedCount = 0
	w.walkData = make(map[string]time.Time)
	w.walkErrors = make(map[string]string)
	w.transferErrMsgs = make(map[string]string)
	w.printer = printer

	return w
}

func (w *WalkState) Printf(s string, params ...interface{}) {
	if w.printer == nil {
		return
	}

	if len(params) == 0 {
		fmt.Fprintf(w.printer, "%s", s)
	}

	fmt.Fprintf(w.printer, s, params...)
}

func (w *WalkState) ErrStr(path string, errMsg string) string {
	return fmt.Sprintf("%s with (%s)", path, errMsg)
}
