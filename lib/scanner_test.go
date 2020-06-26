package exifsort

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hectane/go-acl"
)

func winOS() bool {
	return runtime.GOOS == "windows"
}

func testGetModTime(path string) (time.Time, error) {
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

func TestScanFile(t *testing.T) {
	t.Parallel()
	s := NewScanner()

	exifTime, _ := extractTimeFromStr(exifTimeStr)
	modTime, _ := testGetModTime(noExifPath)
	rootlessModTime, _ := testGetModTime(noRootExifPath)

	time, err := s.ScanFile(exifPath)
	if err != nil {
		t.Errorf("Unexpected Error with good input file\n")
	}

	if exifTime != time {
		t.Errorf("Expected Time %s but got %s\n", exifTime, time)
	}

	time, err = s.ScanFile(noExifPath)
	if err != nil {
		t.Errorf("Unexpected error with invalid Exif file.\n")
	}

	if modTime != time {
		t.Errorf("%s Should have %s not %s\n", noExifPath, "", time)
	}

	time, err = s.ScanFile(noRootExifPath)
	if err != nil {
		t.Errorf("Unexpected error with invalid Exif file.\n")
	}

	if rootlessModTime != time {
		t.Errorf("%s Should have %s not %s\n", noRootExifPath, "", time)
	}

	_, err = s.ScanFile(nonesensePath)
	if err == nil {
		t.Errorf("Expected error with nonsense path\n")
	}
}

func testCheckScanCounts(t *testing.T, td *testdir, s Scanner) {
	if td.numData != len(s.Data) {
		t.Errorf("Expected %d Data Count. Got %d\n",
			td.numData, len(s.Data))
	}

	walkData := s.Data
	if len(walkData) != td.numData {
		t.Errorf("Expected number of data to be %d got %d\n",
			td.numData, len(walkData))
	}

	if td.numExifError != len(s.ExifErrors) {
		t.Errorf("Expected %d ExifErrors got %d\n",
			td.numExifError, len(s.ExifErrors))
	}

	exifErrs := s.ExifErrors
	if len(exifErrs) != td.numExifError {
		t.Errorf("Expected number of exifErrs to be %d got %d\n",
			td.numExifError, len(exifErrs))
	}

	if td.numSkipped != s.SkippedCount {
		t.Errorf("Expected %d Skipped got %d\n",
			td.numSkipped, s.SkippedCount)
	}

	if td.numScanError != len(s.ScanErrors) {
		t.Errorf("Expected %d ScanErrors got %d\n",
			td.numScanError, len(s.ScanErrors))
	}

	if td.numTotal() != s.NumTotal() {
		t.Errorf("Expected %d Total got %d\n", td.numTotal(), s.NumTotal())
	}
}

func TestScanDir(t *testing.T) {
	t.Parallel()
	td := newTestDir(t, MethodNone, fileNoDefault)

	tmpPath := td.buildRoot()
	defer os.RemoveAll(tmpPath)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	testCheckScanCounts(t, td, s)
}

func TestScanSkipDir(t *testing.T) {
	t.Parallel()
	td := newTestDir(t, MethodNone, fileNoDefault)

	tmpPath := td.buildSkipRoot()
	defer os.RemoveAll(tmpPath)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	testCheckScanCounts(t, td, s)
}

func TestScanBadDir(t *testing.T) {
	t.Parallel()
	if winOS() {
		return
	}

	td := newTestDir(t, MethodNone, fileNoDefault)

	tmpPath := td.buildBadRoot()
	defer os.RemoveAll(tmpPath)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	testCheckScanCounts(t, td, s)
}

func TestScanSaveLoad(t *testing.T) {
	t.Parallel()
	td := newTestDir(t, MethodNone, fileNoDefault)

	tmpPath := td.buildRoot()
	defer os.RemoveAll(tmpPath)

	jsonDir, _ := ioutil.TempDir("", "jsonDir")
	defer os.RemoveAll(jsonDir)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	jsonPath := filepath.Join(jsonDir, "scanned.json")

	err := s.Save(jsonPath)
	if err != nil {
		t.Errorf("Unexpected Error %s from Save\n", err.Error())
	}

	newScanner := NewScanner()

	err = newScanner.Load(jsonPath)
	if err != nil {
		t.Errorf("Unexpected Error %s from Load\n", err.Error())
	}

	if !cmp.Equal(s, newScanner) {
		t.Errorf("Saved and Loaded Scanner do not match\n")
	}
}

func TestScanBadSave(t *testing.T) {
	t.Parallel()
	td := newTestDir(t, MethodNone, fileNoDefault)

	tmpPath := td.buildRoot()
	defer os.RemoveAll(tmpPath)

	jsonDir, _ := ioutil.TempDir("", "jsonDir")
	defer os.RemoveAll(jsonDir)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	jsonPath := filepath.Join(jsonDir, "scanned.json")

	// Windows permissions are much different than unix variants
	if runtime.GOOS == "windows" {
		_ = acl.Chmod(jsonDir, 0)
	} else {
		_ = os.Chmod(jsonDir, 0)
	}

	err := s.Save(jsonPath)
	if err == nil {
		t.Errorf("Unexpected Success from Save\n")
	}
}

func TestScanBadLoad(t *testing.T) {
	t.Parallel()
	td := newTestDir(t, MethodNone, fileNoDefault)

	tmpPath := td.buildRoot()
	defer os.RemoveAll(tmpPath)

	jsonDir, _ := ioutil.TempDir("", "jsonDir")
	defer os.RemoveAll(jsonDir)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	jsonPath := filepath.Join(jsonDir, "scanned.json")

	err := s.Save(jsonPath)
	if err != nil {
		t.Errorf("Unexpected Error %s from Save\n", err.Error())
	}

	newScanner := NewScanner()
	_ = os.Truncate(jsonPath, 2)

	err = newScanner.Load(jsonPath)
	if err == nil {
		t.Errorf("Unexpected Success from Load\n")
	}
}
