package exifsort

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const (
	scanExifPath      = "../data/with_exif.jpg"
	scanNoExifPath    = "../data/no_exif.jpg"
	scanSkipPath      = "../README.md"
	correctNumInvalid = 51
	correctNumValid   = 100
	correctNumSkipped = 25
	correctNumTotal   = 176
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

const validExifPath = "../data/with_exif.jpg"
const invalidExifPath = "../data/no_exif.jpg"
const noRootExifPath = "../data/no_root_ifd.jpg"
const goodDateExifStr = "2020:04:28 14:12:21"

func TestExtractTime(t *testing.T) {
	s := NewScanner()
	goodTime, _ := s.extractTimeFromStr(goodDateExifStr)

	time, err := s.ScanFile(validExifPath)
	if err != nil {
		t.Errorf("Unexpected Error with good input file\n")
	}

	if goodTime != time {
		t.Errorf("Expected Time %s but got %s\n", goodTime, time)
	}

	_, err = s.ScanFile(invalidExifPath)
	if err == nil {
		t.Errorf("Unexpected success with invalid Exif file.\n")
	}

	_, err = s.ScanFile(noRootExifPath)
	if err == nil {
		t.Errorf("Unexpected success with invalid Exif file.\n")
	}

	nonePath := "../gobofragggle"

	_, err = s.ScanFile(nonePath)
	if err == nil {
		t.Errorf("Unexpected success with nonsense path\n")
	}
}

func stampFileNo(path string, fileno *int) string {
	basename := filepath.Base(path)
	pieces := strings.Split(basename, ".")
	newPath := fmt.Sprintf("%s_%d.%s", pieces[0], *fileno, pieces[1])
	*fileno++

	return newPath
}

func populateExifDir(t *testing.T, dir string, readPath string, num int, fileno *int) {
	content, err := ioutil.ReadFile(readPath)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		newBase := stampFileNo(readPath, fileno)
		targetPath := fmt.Sprintf("%s/%s", dir, newBase)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testTmpDir(t *testing.T, parent string, name string) string {
	newDir, err := ioutil.TempDir(parent, name)
	if err != nil {
		t.Fatal(err)
	}

	return newDir
}

func setDirPerms(t *testing.T, dirPath string, perms os.FileMode) {
	infos, _ := ioutil.ReadDir(dirPath)
	for _, info := range infos {
		targetPath := fmt.Sprintf("%s/%s", dirPath, info.Name())

		err := os.Chmod(targetPath, perms)
		if err != nil {
			t.Errorf("Chmod failed on %s with %s\n", info.Name(), err.Error())
		}
	}

	err := os.Chmod(dirPath, perms)
	if err != nil {
		t.Errorf("Chmod failed on %s with %s\n", dirPath, err.Error())
	}
}

/*
	Root
	-with_exif // valid exif
	  -nested_exif // nested dir with valid exit
	-no_exif // no exif
	-mixed_exif // mix of both
*/
func buildTestDir(t *testing.T) string {
	fileNo := 0
	rootDir := testTmpDir(t, "", "root")
	exifDir := testTmpDir(t, rootDir, "with_exif")
	badDir := testTmpDir(t, rootDir, "badPerms")
	skipDir := testTmpDir(t, rootDir, "skip")
	nestedDir := testTmpDir(t, exifDir, "nested_exif")
	noExifDir := testTmpDir(t, rootDir, "no_exif")
	mixedDir := testTmpDir(t, rootDir, "mixed_exif")

	populateExifDir(t, exifDir, scanExifPath, 50, &fileNo)
	populateExifDir(t, badDir, scanExifPath, 25, &fileNo)
	populateExifDir(t, noExifDir, scanNoExifPath, 25, &fileNo)
	populateExifDir(t, mixedDir, scanExifPath, 25, &fileNo)
	populateExifDir(t, mixedDir, scanNoExifPath, 25, &fileNo)
	populateExifDir(t, nestedDir, scanExifPath, 25, &fileNo)
	populateExifDir(t, skipDir, scanSkipPath, 25, &fileNo)

	setDirPerms(t, badDir, 0)

	return rootDir
}

func TestScanDir(t *testing.T) {
	tmpPath := buildTestDir(t)
	defer os.RemoveAll(tmpPath)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	if correctNumSkipped != s.NumSkipped() {
		t.Errorf("Expected %d Skipped Count. Got %d\n",
			correctNumSkipped, s.NumSkipped())
	}

	if correctNumInvalid != s.NumInvalid() {
		t.Errorf("Expected %d Invalid Count. Got %d\n",
			correctNumInvalid, s.NumInvalid())
	}

	walkData := s.Data
	if len(walkData) != correctNumValid {
		t.Errorf("Expected number of data to be %d. Got %d\n",
			correctNumValid, len(walkData))
	}

	walkErrs := s.Errors
	if len(walkErrs) != correctNumInvalid {
		t.Errorf("Expected number of walkErrs to be %d. Got %d\n",
			correctNumInvalid, len(walkErrs))
	}

	if correctNumValid != s.NumValid() {
		t.Errorf("Expected %d Valid Count. Got %d\n",
			correctNumValid, s.NumValid())
	}

	if correctNumTotal != s.NumTotal() {
		t.Errorf("Expected %d Total Count. Got %d\n",
			correctNumTotal, s.NumTotal())
	}
}

func TestScanSaveLoad(t *testing.T) {
	tmpPath := buildTestDir(t)
	defer os.RemoveAll(tmpPath)

	jsonDir := testTmpDir(t, "", "jsonDir")
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
	tmpPath := buildTestDir(t)
	defer os.RemoveAll(tmpPath)

	jsonDir := testTmpDir(t, "", "jsonDir")
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
	tmpPath := buildTestDir(t)
	defer os.RemoveAll(tmpPath)

	jsonDir := testTmpDir(t, "", "jsonDir")
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
