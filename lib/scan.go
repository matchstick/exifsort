package exifsort

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	exifknife "github.com/dsoprea/go-exif-knife"
	"github.com/dsoprea/go-exif/v2"
)

type scanError struct {
	prob string
}

func (e scanError) Error() string {
	return e.prob
}

func newScanError(label string, dateString string) error {
	var e scanError
	e.prob = fmt.Sprintf("bad format for %s: %s Problem", dateString, label)

	return e
}

// Scanner is your API to scan directory of media.
//
// It holds errors and data results of the scan after scanning.
type Scanner struct {
	SkippedCount int
	Data         map[string]time.Time
	Errors       map[string]string
}

// Returns how many files were skipped.
func (s *Scanner) NumSkipped() int {
	return s.SkippedCount
}

// Returns how many files had valid exif DateTimeOriginal data.
func (s *Scanner) NumValid() int {
	return len(s.Data)
}

// Returns how many files had invalid exif DateTimeOriginal data.
func (s *Scanner) NumInvalid() int {
	return len(s.Errors)
}

// Returns the total number of files skipped and scanned.
func (s *Scanner) NumTotal() int {
	return s.SkippedCount + s.NumValid() + s.NumInvalid()
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

func ErrStr(path string, errStr string) string {
	return fmt.Sprintf("%s with (%s)", path, errStr)
}

const numSecsSplit = 2 // we expect two pieces

// Seconds are funny. The format may be "<sec> <milli>"
// or it may be with an extra decmial place such as <sec>.<hundredths>.
func (s *Scanner) secsFractionFromStr(secsStr string) (int, error) {
	splitSecs := strings.Split(secsStr, ".")
	if len(splitSecs) != numSecsSplit {
		return 0, &scanError{"Not a fraction second"}
	}

	// We only care about what is in front of the "."
	secs, err := strconv.Atoi(splitSecs[0])
	if err != nil {
		return 0, &scanError{"Not a convertable second"}
	}

	return secs, nil
}

const numDateSplit = 3 // We expect the date to be X:X:X

func (s *Scanner) dateFromStr(str string, exifDateTime string) (int, int, int, error) {
	splitDate := strings.Split(str, ":")
	if len(splitDate) != numDateSplit {
		return 0, 0, 0, newScanError("Date Split", exifDateTime)
	}

	year, err := strconv.Atoi(splitDate[0])
	if err != nil {
		return 0, 0, 0, newScanError("Year", exifDateTime)
	}

	month, err := strconv.Atoi(splitDate[1])
	if err != nil {
		return 0, 0, 0, newScanError("Month", exifDateTime)
	}

	day, err := strconv.Atoi(splitDate[2])
	if err != nil {
		return 0, 0, 0, newScanError("Day", exifDateTime)
	}

	return year, month, day, nil
}

const numTimeSplit = 3 // We expect time to be X:X:X

func (s *Scanner) timeFromStr(str string, exifDateTime string) (int, int, int, error) {
	splitTime := strings.Split(str, ":")
	if len(splitTime) != numTimeSplit {
		return 0, 0, 0, newScanError("Time Split", exifDateTime)
	}

	hour, err := strconv.Atoi(splitTime[0])
	if err != nil {
		return 0, 0, 0, newScanError("Hour", exifDateTime)
	}

	minute, err := strconv.Atoi(splitTime[1])
	if err != nil {
		return 0, 0, 0, newScanError("Minute", exifDateTime)
	}

	second, err := strconv.Atoi(splitTime[2])
	if err != nil {
		second, err = s.secsFractionFromStr(splitTime[2])
		if err != nil {
			return 0, 0, 0, newScanError("Sec", exifDateTime)
		}
	}

	return hour, minute, second, nil
}

const numDateTimeSplit = 2 // We expect DateTime to be "Date Time"

func (s *Scanner) extractTimeFromStr(exifDateTime string) (time.Time, error) {
	var t time.Time

	splitDateTime := strings.Split(exifDateTime, " ")
	if len(splitDateTime) != numDateTimeSplit {
		return t, newScanError("Space Problem", exifDateTime)
	}

	date := splitDateTime[0]
	timeOfDay := splitDateTime[1]

	year, month, day, err := s.dateFromStr(date, exifDateTime)
	if err != nil {
		return t, err
	}

	hour, minute, second, err := s.timeFromStr(timeOfDay, exifDateTime)
	if err != nil {
		return t, err
	}

	t = time.Date(year, time.Month(month), day,
		hour, minute, second, 0, time.Local)

	return t, nil
}

const validDateTimeOrigninalTagNum = 1

// ScanFile accepts a filepath, reads the exifdata stored inside and
// returns the 'Exif/DateTimeOriginal' value as a golang time.Time format.
//
// It returns an error if the file has no exif data, mangled exif data, or the
// contents are unexpected.
func (s *Scanner) ScanFile(filepath string) (time.Time, error) {
	var time time.Time
	// Get the Exif Data and Ifd root
	mc, err := exifknife.GetExif(filepath)
	if err != nil {
		return time, err
	}
	// If the root is not there there is no exif data
	if mc.RootIfd == nil {
		return time, &scanError{"root ifd not found"}
	}

	// See if the EXIF info path is there. We want DateTimeOriginal
	exifIfd, err := exif.FindIfdFromRootIfd(mc.RootIfd, "IFD/Exif")
	if err != nil {
		return time, &scanError{"media IFD/Exif not found"}
	}

	// Query for DateTimeOriginal
	results, err := exifIfd.FindTagWithName("DateTimeOriginal")
	if err != nil {
		return time, &scanError{"the DateTimeOriginal Tag was not found"}
	}

	if len(results) != validDateTimeOrigninalTagNum {
		return time, &scanError{"too many DateTimeOriginal Tags found"}
	}

	// Found it, so extract value
	value, _ := results[0].Value()

	// Parse string into Time
	time, err = s.extractTimeFromStr(value.(string))
	if err != nil {
		return time, err
	}

	return time, nil
}

func (s *Scanner) scanFunc(logger io.Writer) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.storeInvalid(path, err)
			fmt.Fprintf(logger, "%s\n", ErrStr(path, err.Error()))

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

		time, err := s.ScanFile(path)
		if err != nil {
			s.storeInvalid(path, err)
			fmt.Fprintf(logger, "%s\n", ErrStr(path, err.Error()))

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
func (s *Scanner) ScanDir(src string, logger io.Writer) error {
	info, err := os.Stat(src)
	if err != nil || !info.IsDir() {
		return err
	}

	// scanFunc never returns an error
	// We don't want to walk for an hour and then fail on one error.
	// Consult the walkstate for errors.
	_ = filepath.Walk(src, s.scanFunc(logger))

	return nil
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
