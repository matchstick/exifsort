package exifsort

import (
	"fmt"
	"sync/atomic"
)

type WalkState struct {
	skippedCount     uint64
	validCount       uint64
	invalidCount     uint64
	transferErrCount uint64
	doPrint          bool
	walkErrMsgs      map[string]string
	transferErrMsgs  map[string]string
}

func (w *WalkState) Skipped() uint64 {
	return w.skippedCount
}

func (w *WalkState) Valid() uint64 {
	return w.validCount
}

func (w *WalkState) Invalid() uint64 {
	return w.invalidCount
}

func (w *WalkState) WalkErrs() map[string]string {
	return w.walkErrMsgs
}

func (w *WalkState) TransferErrs() map[string]string {
	return w.transferErrMsgs
}

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

func newWalkState(doPrint bool) WalkState {
	var w WalkState
	w.skippedCount = 0
	w.validCount = 0
	w.invalidCount = 0
	w.doPrint = doPrint
	w.walkErrMsgs = make(map[string]string)
	w.transferErrMsgs = make(map[string]string)

	return w
}

func (w *WalkState) walkPrintf(s string, params ...interface{}) {
	if !w.doPrint {
		return
	}

	if len(params) == 0 {
		fmt.Print(s)
	}

	fmt.Printf(s, params...)
}

func (w *WalkState) ErrStr(path string, errMsg string) string {
	return fmt.Sprintf("%s with (%s)", path, errMsg)
}
