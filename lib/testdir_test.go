package exifsort

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

/*

This file began as a quick hack but it is becoming apparrent it is going to
need work and focus if we are going to put exifsort into real production use
for the family.

Requirements are accumaltive as the stages of the pipeline refine the needs of
the input directories.

Scanner Requirements:
* A directory with two levels
* Some files (variable) with valid exifdata
* Some files (variable) with invalid exifdata so they use mod times
* Some examples of filepath errors
* Some files to skip

Sort Requirements
* All the same ones scanner has
* Create input files to allow for media sets of one and multiple files.
* Have media sets be spread across all methods
* Have transfer be all actions.
* Exercise collision paths
* Exercise duplicate paths

Merge rerquirments
* All the ones sorter has
* Take as input a sorted directory with correct structure
* A directory with broken structure
* A directory with a disjoint set of media files to create new leaves
* A directory with a same set of media files to not add anything to the dst
* A directory with the same media names and sort times but different contents to create collisions.

*/

const (
	exifPath    = "../data/with_exif.jpg"
	exifTimeStr = "2020:04:28 14:12:21"

	noExifPath     = "../data/no_exif.jpg"
	noRootExifPath = "../data/no_root_ifd.jpg"

	skipPath      = "../README.md"
	nonesensePath = "../gobofragggle"

	numExifError = 50
	numData      = 150
	numSkipped   = 25
	numScanError = 1
	numTotal     = 176
)

type testdir struct {
	fileNo    int
	root      string
	t         *testing.T
	startTime time.Time
}

func (td *testdir) addTimeByMethod(t time.Time, method int, delta int) time.Time {
	var time time.Time

	switch method {
	case MethodYear:
		return t.AddDate(delta, 0, 0)
	case MethodMonth:
		return t.AddDate(0, delta, 0)
	case MethodDay:
		return t.AddDate(0, 0, delta)
	default:
		td.t.Fatalf("Invalid Method %d", method)
	}

	return time
}

// We need to have unique filenames often.
// Instead of worrying about if they are unique in each subdir we are using a
// testdir wide counter to ensure they are unique across the whole testdir
// instance.
func (td *testdir) uniqueFilename(path string) string {
	filename := filepath.Base(path)
	extension := filepath.Ext(filename)
	prefix := strings.TrimRight(filename, extension)

	// increment our global counter
	td.fileNo++

	return fmt.Sprintf("%s_%d.%s", prefix, td.fileNo, extension)
}

// Generate 'num' files with exif date set to arg in the directory passed in
func (td *testdir) populateExifFiles(dir string, num int) {
	content, err := ioutil.ReadFile(exifPath)
	if err != nil {
		td.t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		filename := td.uniqueFilename(exifPath)
		targetPath := filepath.Join(dir, filename)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			td.t.Fatal(err)
		}
	}
}

// Generate 'num' files with exif date set to arg in the directory passed in
func (td *testdir) populateModFiles(dir string, num int, method int, delta int) {
	content, err := ioutil.ReadFile(noExifPath)
	if err != nil {
		td.t.Fatal(err)
	}

	time := td.startTime

	for i := 0; i < num; i++ {
		filename := td.uniqueFilename(noExifPath)
		targetPath := filepath.Join(dir, filename)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			td.t.Fatal(err)
		}

		// set time for file
		err = os.Chtimes(targetPath, time, time)
		if err != nil {
			td.t.Fatal(err)
		}

		time = td.addTimeByMethod(time, method, delta)
	}
}

// Generate 'num' files with exif date set to arg in the directory passed in
func (td *testdir) populateSkipFiles(dir string, num int) {
	content, err := ioutil.ReadFile(skipPath)
	if err != nil {
		td.t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		filename := td.uniqueFilename(skipPath)
		targetPath := filepath.Join(dir, filename)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			td.t.Fatal(err)
		}
	}
}

func newTestDir(t *testing.T) string {
	var td testdir
	td.fileNo = 0
	td.root, _ = ioutil.TempDir("", "root")
	td.t = t
	td.startTime = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.Local)

	exifDir, _ := ioutil.TempDir(td.root, "with_exif")
	badDir, _ := ioutil.TempDir(td.root, "badPerms")
	skipDir, _ := ioutil.TempDir(td.root, "skip")
	nestedDir, _ := ioutil.TempDir(exifDir, "nested_exif")
	noExifDir, _ := ioutil.TempDir(td.root, "no_exif")
	mixedDir, _ := ioutil.TempDir(td.root, "mixed_exif")

	td.populateExifFiles(exifDir, 50)
	td.populateModFiles(noExifDir, 25, MethodYear, 1)
	td.populateExifFiles(mixedDir, 25)
	td.populateModFiles(mixedDir, 25, MethodYear, 1)
	td.populateExifFiles(nestedDir, 25)
	td.populateSkipFiles(skipDir, 25)

	err := os.Chmod(badDir, 0)
	if err != nil {
		td.t.Errorf("Chmod failed on %s with %s\n", badDir, err.Error())
	}

	return td.root
}

// Returns the path to the root of a directory full of files and nested
// structures. This is only intended for test code. Some of the media has
// exifdata some does not, some are not even media files. All of the files and
// directories were created as golang tmp files or directories.
//
// The bult directory even has a subdir with perms 0 for testing errorn
// handling.
func newMethodTestDir(t *testing.T, method int) string {
	var td testdir
	td.fileNo = 0
	td.root, _ = ioutil.TempDir("", "root")
	td.t = t
	td.startTime = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.Local)

	exifDir, _ := ioutil.TempDir(td.root, "with_exif")
	badDir, _ := ioutil.TempDir(td.root, "badPerms")
	skipDir, _ := ioutil.TempDir(td.root, "skip")
	nestedDir, _ := ioutil.TempDir(exifDir, "nested_exif")
	noExifDir, _ := ioutil.TempDir(td.root, "no_exif")
	mixedDir, _ := ioutil.TempDir(td.root, "mixed_exif")

	td.populateExifFiles(exifDir, 50)
	td.populateModFiles(noExifDir, 25, method, 1)
	td.populateExifFiles(mixedDir, 25)
	td.populateModFiles(mixedDir, 25, method, 1)
	td.populateExifFiles(nestedDir, 25)
	td.populateSkipFiles(skipDir, 25)

	err := os.Chmod(badDir, 0)
	if err != nil {
		td.t.Errorf("Chmod failed on %s with %s\n", badDir, err.Error())
	}

	return td.root
}
