package exifsort

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// ScannerInput specifies how scanner receives data
type ScannerInput int

const (
	// ScannerInputJSON : Input via JSON file
	ScannerInputJSON ScannerInput = iota
	// ScannerInputDir : Input via dir
	ScannerInputDir
	// ScannerInputNone : Error valueI
	ScannerInputNone
)

type scanError struct {
	path string
	err  error
}

type scanData struct {
	path string
	time time.Time
}

type scanDirState struct {
	scanErrorChannel chan scanError
	exifErrorChannel chan scanError
	dataChannel      chan scanData
	done             chan bool
	walkComplete     sync.WaitGroup
}

func (s *scanDirState) gatherResults(scanner *Scanner) {
	go func() {
		for {
			select {
			case scanErr := <-s.scanErrorChannel:
				scanner.storeScanError(scanErr.path, scanErr.err)
			case scanErr := <-s.exifErrorChannel:
				scanner.storeExifError(scanErr.path, scanErr.err)
			case scanData := <-s.dataChannel:
				scanner.storeData(scanData.path, scanData.time)
			case <-s.done:
				s.walkComplete.Done()
				return
			}
		}
	}()
}

func (s *scanDirState) sendScanError(logger io.Writer, path string, err error) {
	var payload scanError
	payload.path = path
	payload.err = err
	fmt.Fprintf(logger, "Error: %s: (%s)\n", path, err.Error())
	s.scanErrorChannel <- payload
}

func (s *scanDirState) sendExifError(path string, err error) {
	var payload scanError
	payload.path = path
	payload.err = err
	s.exifErrorChannel <- payload
}

func (s *scanDirState) sendScanData(logger io.Writer, path string, time time.Time) {
	var payload scanData
	payload.path = path
	payload.time = time

	str := fmt.Sprintf("%d:%02d:%02d %02d:%02d:%02d",
		time.Year(), time.Month(), time.Day(),
		time.Hour(), time.Minute(), time.Second())
	fmt.Fprintf(logger, "%s, %s\n", path, str)
	s.dataChannel <- payload
}

func (s *scanDirState) Wait() {
	close(s.done)

	s.walkComplete.Wait()

	close(s.scanErrorChannel)
	close(s.exifErrorChannel)
	close(s.dataChannel)
}

func newRunState() *scanDirState {
	var s scanDirState
	s.exifErrorChannel = make(chan scanError)
	s.scanErrorChannel = make(chan scanError)
	s.dataChannel = make(chan scanData)
	s.done = make(chan bool)

	s.walkComplete.Add(1)

	return &s
}

// Scanner is your API to scan directory of media.
//
// It holds errors and data results of the scan after scanning.
type Scanner struct {
	Input             ScannerInput
	SkippedCount      int32
	Data              map[string]time.Time
	NumDataTypes      map[string]int
	ExifErrors        map[string]string
	NumExifErrorTypes map[string]int
	ScanErrors        map[string]string
	DirState          *scanDirState
}

// NumTotal returns the total number of files skipped, scanned and errors.
func (s *Scanner) NumTotal() int {
	return s.NumSkipped() + len(s.Data) + len(s.ScanErrors)
}

// We don't check if you have a path duplicate.
func (s *Scanner) storeData(path string, time time.Time) {
	s.Data[path] = time

	extension := filepath.Ext(path)

	_, present := s.NumDataTypes[extension]
	if present {
		s.NumDataTypes[extension]++
	} else {
		s.NumDataTypes[extension] = 1
	}
}

// We don't check if you have a path duplicate.
func (s *Scanner) storeExifError(path string, err error) {
	s.ExifErrors[path] = err.Error()

	extension := filepath.Ext(path)

	_, present := s.NumExifErrorTypes[extension]
	if present {
		s.NumExifErrorTypes[extension]++
	} else {
		s.NumExifErrorTypes[extension] = 1
	}
}

func (s *Scanner) storeScanError(path string, err error) {
	s.ScanErrors[path] = err.Error()
}

func (s *Scanner) storeSkipped() {
	atomic.AddInt32(&s.SkippedCount, 1)
}

func (s *Scanner) NumSkipped() int {
	return int(atomic.LoadInt32(&s.SkippedCount))
}

func (s *Scanner) modTime(path string) (time.Time, error) {
	var t time.Time

	info, err := os.Stat(path)
	if err != nil {
		return t, err
	}

	t = info.ModTime()

	// We are clearing the nanoseconds for consistency
	t = time.Date(t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), 0, time.Local)

	return t, nil
}

// ScanFile accepts a filepath, reads the exifdata stored inside and
// returns the 'Exif/DateTimeOriginal' value as a golang time.Time format. If
// the exifData is not valid it will return the time based on FileInfo's
// ModTime.
//
// It returns an error if the file has no exif data and cannot be statted.
func (s *Scanner) ScanFile(path string) (time.Time, error) {
	time, err := ExifTimeGet(path)
	// Note we are checking == not !=
	if err == nil {
		return time, nil
	}

	// If we are running a ScanDir while calling ScanFile we need to use the
	// state tooling to gather stats concurrently.
	if s.DirState != nil {
		s.DirState.sendExifError(path, err)
	} else {
		s.storeExifError(path, err)
	}

	return s.modTime(path)
}

func (s *Scanner) scanFunc(logger io.Writer) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		var time time.Time

		if err != nil {
			s.DirState.sendScanError(logger, path, err)
			return nil
		}

		// Don't need to scan directories
		if info.IsDir() {
			return nil
		}

		fileCategory := categorizeFile(path)
		switch fileCategory {
		case categorySkip:
			s.storeSkipped()
			return nil
		case categoryExif:
			time, err = s.ScanFile(path)
		case categoryModTime:
			time, err = s.modTime(path)
		}

		if err != nil {
			s.DirState.sendScanError(logger, path, err)
		} else {
			s.DirState.sendScanData(logger, path, time)
		}

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
	s.Input = ScannerInputDir

	info, err := os.Stat(src)
	if err != nil || !info.IsDir() {
		return err
	}

	s.DirState = newRunState()

	s.DirState.gatherResults(s)

	// scanFunc never returns an error
	// We don't want to walk for an hour and then fail on one error.
	// Consult the walkstate for errors.
	_ = filepath.Walk(src, s.scanFunc(logger))
	s.DirState.Wait()

	s.DirState = nil

	return nil
}

// Save Scanner to a json file.
func (s *Scanner) Save(jsonPath string) error {
	json, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(jsonPath, json, 0600)
	if err != nil {
		return err
	}

	return nil
}

// Load Scanner from a json file.
func (s *Scanner) Load(jsonPath string) error {
	s.Input = ScannerInputJSON

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

// Reset clears data so Scanner can be reused.
func (s *Scanner) Reset() {
	s.Input = ScannerInputNone
	s.SkippedCount = 0
	s.Data = make(map[string]time.Time)
	s.NumDataTypes = make(map[string]int)
	s.ExifErrors = make(map[string]string)
	s.NumExifErrorTypes = make(map[string]int)
	s.ScanErrors = make(map[string]string)
	s.DirState = nil
}

// NewScanner allocates a new Scanner.
func NewScanner() Scanner {
	var s Scanner

	s.Reset()

	return s
}
