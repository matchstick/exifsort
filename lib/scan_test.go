package exifsort

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	s.ScanDir(tmpPath, ioutil.Discard)

	if correctNumSkipped != s.Skipped() {
		t.Errorf("Expected %d Skipped Count. Got %d\n",
			correctNumSkipped, s.Skipped())
	}

	if correctNumInvalid != s.Invalid() {
		t.Errorf("Expected %d Invalid Count. Got %d\n",
			correctNumInvalid, s.Invalid())
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

	if correctNumValid != s.Valid() {
		t.Errorf("Expected %d Valid Count. Got %d\n",
			correctNumValid, s.Valid())
	}

	if correctNumTotal != s.Total() {
		t.Errorf("Expected %d Total Count. Got %d\n",
			correctNumTotal, s.Total())
	}
}

func TestScanSaveLoad(t *testing.T) {
	tmpPath := buildTestDir(t)
	defer os.RemoveAll(tmpPath)

	jsonDir := testTmpDir(t, "", "jsonDir")
	defer os.RemoveAll(jsonDir)

	s := NewScanner()
	s.ScanDir(tmpPath, ioutil.Discard)

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
