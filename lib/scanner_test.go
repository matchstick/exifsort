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
	"github.com/matchstick/exifsort/testdir"
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
	s := NewScanner()

	exifTime, _ := extractTimeFromStr(testdir.ExifTimeStr)
	modTime, _ := testGetModTime(testdir.NoExifPath)
	rootlessModTime, _ := testGetModTime(testdir.NoRootExifPath)

	time, err := s.ScanFile(testdir.ExifPath)
	if err != nil {
		t.Errorf("Unexpected Error with good input file\n")
	}

	if exifTime != time {
		t.Errorf("Expected Time %s but got %s\n", exifTime, time)
	}

	time, err = s.ScanFile(testdir.NoExifPath)
	if err != nil {
		t.Errorf("Unexpected error with invalid Exif file.\n")
	}

	if modTime != time {
		t.Errorf("%s Should have %s not %s\n",
			testdir.NoExifPath, "", time)
	}

	time, err = s.ScanFile(testdir.NoRootExifPath)
	if err != nil {
		t.Errorf("Unexpected error with invalid Exif file.\n")
	}

	if rootlessModTime != time {
		t.Errorf("%s Should have %s not %s\n",
			testdir.NoRootExifPath, "", time)
	}

	_, err = s.ScanFile(testdir.NonesensePath)
	if err == nil {
		t.Errorf("Expected error with nonsense path\n")
	}
}

func TestScanDir(t *testing.T) {
	tmpPath := testdir.NewTestDir(t)
	defer os.RemoveAll(tmpPath)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	if testdir.NumData != s.NumData() {
		t.Errorf("Expected %d Valid Count. Got %d\n",
			testdir.NumData, s.NumData())
	}

	if testdir.NumSkipped != s.NumSkipped() {
		t.Errorf("Expected %d Skipped Count. Got %d\n",
			testdir.NumSkipped, s.NumSkipped())
	}

	if testdir.NumExifError != s.NumExifErrors() {
		t.Errorf("Expected %d ExifErrors Count. Got %d\n",
			testdir.NumExifError, s.NumExifErrors())
	}

	if !winOS() && testdir.NumScanError != s.NumScanErrors() {
		t.Errorf("Expected %d ExifErrors Count. Got %d\n",
			testdir.NumScanError, s.NumScanErrors())
	}

	walkData := s.Data
	if len(walkData) != testdir.NumData {
		t.Errorf("Expected number of data to be %d. Got %d\n",
			testdir.NumData, len(walkData))
	}

	exifErrs := s.ExifErrors
	if len(exifErrs) != testdir.NumExifError {
		t.Errorf("Expected number of walkErrs to be %d. Got %d\n",
			testdir.NumExifError, len(exifErrs))
	}

	scanErrs := s.ScanErrors
	if !winOS() && len(scanErrs) != testdir.NumScanError {
		t.Errorf("Expected number of walkErrs to be %d. Got %d\n",
			testdir.NumScanError, len(scanErrs))
	}

	if !winOS() && testdir.NumTotal != s.NumTotal() {
		t.Errorf("Expected %d Total Count. Got %d\n",
			testdir.NumTotal, s.NumTotal())
	}
}

func TestScanSaveLoad(t *testing.T) {
	tmpPath := testdir.NewTestDir(t)
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
	tmpPath := testdir.NewTestDir(t)
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
	tmpPath := testdir.NewTestDir(t)
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
