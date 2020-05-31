package exifsort

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/matchstick/exifsort/testdir"
)

func TestSkipFileType(t *testing.T) {
	s := NewScanner()
	// Try just gobo.<suffix>
	for suffix := range s.mediaSuffixMap() {
		goodInput := fmt.Sprintf("gobo.%s", suffix)

		_, skip := s.skipFileType(goodInput)
		if skip {
			t.Errorf("Expected False for %s\n", goodInput)
		}
	}
	// Try a simple upper case just gobo.<suffix>
	for suffix := range s.mediaSuffixMap() {
		goodInput := strings.ToUpper(fmt.Sprintf("gobo.%s", suffix))

		_, skip := s.skipFileType(goodInput)
		if skip {
			t.Errorf("Expected False for %s\n", goodInput)
		}
	}

	// Try with many "." hey.gobo.<suffix>
	for suffix := range s.mediaSuffixMap() {
		goodInput := fmt.Sprintf("hey.gobo.%s", suffix)

		_, skip := s.skipFileType(goodInput)
		if skip {
			t.Errorf("Expected False for %s\n", goodInput)
		}
	}

	badInput := "gobobob.."

	_, skip := s.skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}

	badInput = "gobo"

	_, skip = s.skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}

	// Try ".." at the end.<suffix>
	for suffix := range s.mediaSuffixMap() {
		badInput := fmt.Sprintf("gobo.%s..", suffix)

		_, skip := s.skipFileType(badInput)
		if !skip {
			t.Errorf("Expected True for %s\n", badInput)
		}
	}
}

func TestSkipSynologyTypes(t *testing.T) {
	s := NewScanner()

	badInput := "@eaDir"

	_, skip := s.skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}

	badInput = "@syno"

	_, skip = s.skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}

	badInput = "synofile_thumb"

	_, skip = s.skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}
}

func TestFormatError(t *testing.T) {
	testErrStr := "bad format for dingle: dangle Problem"
	err := newScanError("dangle", "dingle")

	if err.Error() != testErrStr {
		t.Errorf("Errors do not match: %s %s", err, testErrStr)
	}
}

func TestGoodTimes(t *testing.T) {
	good1String := "2008:03:01 12:36:01"
	good2String := "2008:03:01 12:36:01.34"
	testMonth := 3
	goodTime := time.Date(2008, time.Month(testMonth), 1, 12, 36, 1, 0, time.Local)

	s := NewScanner()

	testTime, err := s.extractTimeFromStr(good1String)
	if testTime != goodTime {
		t.Errorf("Return Time is incorrect %q\n", testTime)
	}

	if err != nil {
		t.Errorf("Error is incorrectly not nil %q\n", err)
	}

	testTime, err = s.extractTimeFromStr(good2String)
	if testTime != goodTime {
		t.Errorf("Return Time is incorrect %q\n", testTime)
	}

	if err != nil {
		t.Errorf("Error is incorrectly not nil %q\n", err)
	}
}

func TestExtractBadTimeFromStr(t *testing.T) {
	var formBadInput = map[string]string{
		"Gobo":                    "Space Problem",
		"Gobo a a a a":            "Space Problem",
		"Gobo Hey":                "Date Split",
		"Gobo:03:01 12:36:11":     "Year",
		"2008:Gobo:01 12:36:11":   "Month",
		"2008:03:Gobo 12:36:11":   "Day",
		"2008:03:01 Gobo":         "Time Split",
		"2008:03:01 Gobo:36:11":   "Hour",
		"2008:03:01 12:Gobo:11":   "Minute",
		"2008:03:01 12:36:Gobo":   "Sec",
		"2008:03:01 12:36:Gobo.2": "Sec",
	}

	for input, errLabel := range formBadInput {
		s := NewScanner()

		_, err := s.extractTimeFromStr(input)
		if err == nil {
			t.Fatalf("Expected error on input: %s\n", input)
		}

		if strings.Contains(err.Error(), errLabel) == false {
			t.Errorf("Improper error reporting on input %s: %s\n",
				input, err.Error())
		}
	}
}

const goodDateExifStr = "2020:04:28 14:12:21"

func TestScanFile(t *testing.T) {
	s := NewScanner()
	goodTime, _ := s.extractTimeFromStr(goodDateExifStr)

	time, err := s.ScanFile(testdir.ExifPath)
	if err != nil {
		t.Errorf("Unexpected Error with good input file\n")
	}

	if goodTime != time {
		t.Errorf("Expected Time %s but got %s\n", goodTime, time)
	}

	_, err = s.ScanFile(testdir.NoExifPath)
	if err == nil {
		t.Errorf("Unexpected success with invalid Exif file.\n")
	}

	_, err = s.ScanFile(testdir.NoRootExifPath)
	if err == nil {
		t.Errorf("Unexpected success with invalid Exif file.\n")
	}

	_, err = s.ScanFile(testdir.NonesensePath)
	if err == nil {
		t.Errorf("Unexpected success with nonsense path\n")
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

	if testdir.NumScanError != s.NumScanErrors() {
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
	if len(scanErrs) != testdir.NumScanError {
		t.Errorf("Expected number of walkErrs to be %d. Got %d\n",
			testdir.NumScanError, len(scanErrs))
	}

	if testdir.NumTotal != s.NumTotal() {
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

	jsonPath := fmt.Sprintf("%s/%s", jsonDir, "scanned.json")

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

	jsonPath := fmt.Sprintf("%s/%s", jsonDir, "scanned.json")

	_ = os.Chmod(jsonDir, 0)

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

	jsonPath := fmt.Sprintf("%s/%s", jsonDir, "scanned.json")

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
