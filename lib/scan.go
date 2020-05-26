package exifsort

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Scanner is your API to scan directory of media.
//
// It holds errors and data results of the scan after scanning.
type Scanner struct {
	SkippedCount int
	Data         map[string]time.Time
	Errors       map[string]string
}

// Returns how many files were skipped.
func (s *Scanner) Skipped() int {
	return s.SkippedCount
}

// Returns how many files had valid exif DateTimeOriginal data.
func (s *Scanner) Valid() int {
	return len(s.Data)
}

// Returns how many files had invalid exif DateTimeOriginal data.
func (s *Scanner) Invalid() int {
	return len(s.Errors)
}

// Returns the total number of files skipped and scanned.
func (s *Scanner) Total() int {
	return s.SkippedCount + s.Valid() + s.Invalid()
}

// We don't check if you have a path duplicate.
func (s *Scanner) storeValid(path string, time time.Time) {
	s.Data[path] = time
}

// We don't check if you have a path duplicate.
func (s *Scanner) storeInvalid(path string, err error) {
	s.Errors[path] = err.Error()
}

func (s *Scanner) storeSkipped() {
	s.SkippedCount++
}

func (s *Scanner) ErrStr(path string, errStr string) string {
	return fmt.Sprintf("%s with (%s)", path, errStr)
}

func (s *Scanner) scanFunc(logger io.Writer) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.storeInvalid(path, err)
			fmt.Fprintf(logger, "%s\n", s.ErrStr(path, err.Error()))

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
			fmt.Fprintf(logger, "%s\n", s.ErrStr(path, err.Error()))

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

// Save Scanner to a json file.
func (s *Scanner) Save(jsonPath string) error {
	json, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(jsonPath, json, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Load Scanner from a json file.
func (s *Scanner) Load(jsonPath string) error {
	content, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &s)
	if err != nil {
		return err
	}

	return nil
}

// Clears data so scanner can be reused.
func (s *Scanner) Reset() {
	s.SkippedCount = 0
	s.Data = make(map[string]time.Time)
	s.Errors = make(map[string]string)
}

// Allocates a new Scanner.
func NewScanner() Scanner {
	var s Scanner

	s.Reset()

	return s
}
