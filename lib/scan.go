package exifsort

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Scanner is your API to scan directory of media.
//
// It holds errors and data results of the scan after scanning.
type Scanner struct {
	skippedCount int
	data         map[string]time.Time
	errors       map[string]error
}

// Returns how many files were skipped.
func (s *Scanner) Skipped() int {
	return s.skippedCount
}

// Returns how many files had valid exif DateTimeOriginal data.
func (s *Scanner) Valid() int {
	return len(s.data)
}

// Returns how many files had invalid exif DateTimeOriginal data.
func (s *Scanner) Invalid() int {
	return len(s.errors)
}

// Returns a map from path to error scanning.
func (s *Scanner) Errors() map[string]error {
	return s.errors
}

// Returns a map from path to time of valid media
func (s *Scanner) Data() map[string]time.Time {
	return s.data
}

// Returns the total number of files skipped and scanned.
func (s *Scanner) Total() int {
	return s.skippedCount + s.Valid() + s.Invalid()
}

// We don't check if you have a path duplicate.
func (s *Scanner) storeValid(path string, time time.Time) {
	s.data[path] = time
}

// We don't check if you have a path duplicate.
func (s *Scanner) storeInvalid(path string, err error) {
	s.errors[path] = err
}

func (s *Scanner) storeSkipped() {
	s.skippedCount++
}

func (s *Scanner) ErrStr(path string, err error) string {
	return fmt.Sprintf("%s with (%s)", path, err.Error())
}

func (s *Scanner) scanFunc(logger io.Writer) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.storeInvalid(path, err)
			fmt.Fprintf(logger, "%s\n", s.ErrStr(path, err))

			return nil
		}

		// Don't need to scan directories
		if info.IsDir() {
			return nil
		}
		// Only looking for media files that may have exif.
		if skipFileType(path) {
			s.storeSkipped()
			return nil
		}

		time, err := ExtractTime(path)
		if err != nil {
			s.storeInvalid(path, err)
			fmt.Fprintf(logger, "%s\n", s.ErrStr(path, err))

			return nil
		}

		fmt.Fprintf(logger, "%s, %s\n", path, exifTimeToStr(time))
		s.storeValid(path, time)

		return nil
	}
}

// ScanDir will examine the contents of every file in the src directory and
// print it's time of creation as stored by exifdata as it scans.
//
// ScanDir only scans media files listed as constants as documented, other
// files are skipped.
//
// logger specifies where to send output while scanning.
func (s *Scanner) ScanDir(src string, logger io.Writer) {
	// scanFunc never returns an error
	// We don't want to walk for an hour and then fail on one error.
	// Consult the walkstate for errors.
	_ = filepath.Walk(src, s.scanFunc(logger))
}

// Clears data so scanner can be reused.
func (s *Scanner) Reset() {
	s.skippedCount = 0
	s.data = make(map[string]time.Time)
	s.errors = make(map[string]error)
}

// Allocates a new Scanner.
func NewScanner() Scanner {
	var s Scanner

	s.Reset()

	return s
}
