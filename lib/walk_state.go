package exifsort

import (
	"fmt"
	"io"
	"sync/atomic"
)

// WalkState holds all the data gathered as exifstort scans the src directories
// and if sorting transfers media.
type WalkState struct {
	skippedCount     uint64
	validCount       uint64
	invalidCount     uint64
	transferErrCount uint64
	printer          io.Writer
	walkErrMsgs      map[string]string
	transferErrMsgs  map[string]string
}

// Returns how many files were skipped.
func (w *WalkState) Skipped() uint64 {
	return w.skippedCount
}

// Returns how many files had valid exif DateTimeOriginal data.
func (w *WalkState) Valid() uint64 {
	return w.validCount
}

// Returns how many files had invalid exif DateTimeOriginal data.
func (w *WalkState) Invalid() uint64 {
	return w.invalidCount
}

// Returns a map from path to error scanning.
func (w *WalkState) WalkErrs() map[string]string {
	return w.walkErrMsgs
}

// Returns a map from path to error transferring.
func (w *WalkState) TransferErrs() map[string]string {
	return w.transferErrMsgs
}

// Returns the total number of files skipped and scanned.
func (w *WalkState) Total() uint64 {
	return w.skippedCount + w.validCount + w.invalidCount
}

func (w *WalkState) storeValid() {
	atomic.AddUint64(&w.validCount, 1)
}

// We don't check if you have a path duplicate.
func (w *WalkState) storeInvalid(path string, errStr string) {
	atomic.AddUint64(&w.invalidCount, 1)
	w.walkErrMsgs[path] = errStr
}

// We don't check if you have a path duplicate.
func (w *WalkState) storeTransferErr(path string, errStr string) {
	atomic.AddUint64(&w.transferErrCount, 1)
	w.transferErrMsgs[path] = errStr
}

func (w *WalkState) storeSkipped() {
	atomic.AddUint64(&w.skippedCount, 1)
}

func newWalkState(printer io.Writer) WalkState {
	var w WalkState
	w.skippedCount = 0
	w.validCount = 0
	w.invalidCount = 0
	w.walkErrMsgs = make(map[string]string)
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
