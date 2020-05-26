package exifsort

import (
	"fmt"
	"time"
)

// WalkState holds all the data gathered during a scan of the src directories
type WalkState struct {
	skippedCount int
	data         map[string]time.Time
	errors       map[string]error
}

// Returns how many files were skipped.
func (w *WalkState) Skipped() int {
	return w.skippedCount
}

// Returns how many files had valid exif DateTimeOriginal data.
func (w *WalkState) Valid() int {
	return len(w.data)
}

// Returns how many files had invalid exif DateTimeOriginal data.
func (w *WalkState) Invalid() int {
	return len(w.errors)
}

// Returns a map from path to error scanning.
func (w *WalkState) Errors() map[string]error {
	return w.errors
}

// Returns a map from path to time of valid media
func (w *WalkState) Data() map[string]time.Time {
	return w.data
}

// Returns the total number of files skipped and scanned.
func (w *WalkState) Total() int {
	return w.skippedCount + w.Valid() + w.Invalid()
}

// We don't check if you have a path duplicate.
func (w *WalkState) storeValid(path string, time time.Time) {
	w.data[path] = time
}

// We don't check if you have a path duplicate.
func (w *WalkState) storeInvalid(path string, err error) {
	w.errors[path] = err
}

func (w *WalkState) storeSkipped() {
	w.skippedCount++
}

func newWalkState() WalkState {
	var w WalkState
	w.skippedCount = 0
	w.data = make(map[string]time.Time)
	w.errors = make(map[string]error)

	return w
}

func (w *WalkState) ErrStr(path string, err error) string {
	return fmt.Sprintf("%s with (%s)", path, err.Error())
}
