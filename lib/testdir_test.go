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
	skipPath       = "../README.md"
	nonesensePath  = "../gobofragggle"
)

type testdir struct {
	fileNo    int
	root      string
	t         *testing.T
	startTime time.Time
	method    int

	numExifError  int
	numData       int
	numDuplicates int
	numSkipped    int
	numScanError  int
}

func (td *testdir) numTotal() int {
	return td.numData + td.numScanError + td.numSkipped
}

func (td *testdir) addTimeByMethod(t time.Time, delta int) time.Time {
	var time time.Time

	switch td.method {
	case MethodYear:
		return t.AddDate(delta, 0, 0)
	case MethodMonth:
		return t.AddDate(0, delta, 0)
	case MethodDay:
		return t.AddDate(0, 0, delta)
	default:
		td.t.Fatalf("Invalid Method %d", td.method)
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

	filename = fmt.Sprintf("%s_%d%s", prefix, td.fileNo, extension)
	td.fileNo++

	return filename
}

func (td *testdir) testFilename(path string, index int) string {
	filename := filepath.Base(path)
	extension := filepath.Ext(filename)
	prefix := strings.TrimRight(filename, extension)

	return fmt.Sprintf("%s_%d%s", prefix, index, extension)
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

		td.numData++
	}
}

// What we are doing is setting our fileno counter back to 0.
// The goal is to create filenames bases on zero with exif contents.
func (td *testdir) populateDuplicateFilenames(dir string, path string, num int) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		td.t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		filename := td.testFilename(path, i)
		targetPath := filepath.Join(dir, filename)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			td.t.Fatal(err)
		}

		td.numDuplicates++
	}
}

// What we are doing is setting our fileno counter back to 0 and using exif filenames.
// The goal is to create filenames bases on zero with nonexif contents but with exif names.
// This will have the same names as popualted exif files but different contents
// so should be handled as a collision.

func (td *testdir) populateCollisionFilenames(dir string, num int) {
	content, err := ioutil.ReadFile(noExifPath)
	if err != nil {
		td.t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		filename := td.testFilename(exifPath, i)
		targetPath := filepath.Join(dir, filename)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			td.t.Fatal(err)
		}

		td.numData++
	}
}

// Generate 'num' files with exif date set to arg in the directory passed in
func (td *testdir) populateModFiles(dir string, num int, delta int) {
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

		if td.method != MethodNone {
			time = td.addTimeByMethod(time, delta)
		}

		td.numExifError++

		// We end up adding these as data if they have exif problems
		// since we parse the mod time as input.
		td.numData++
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

		td.numSkipped++
	}
}

func (td *testdir) getRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "with_exif")
	nestedDir, _ := ioutil.TempDir(exifDir, "nested_exif")
	noExifDir, _ := ioutil.TempDir(td.root, "no_exif")
	mixedDir, _ := ioutil.TempDir(td.root, "mixed_exif")

	// First add exif files
	td.populateExifFiles(exifDir, 50)
	// Then add mod file
	td.populateModFiles(noExifDir, 25, 1)
	// Add more exif files with different names
	td.populateExifFiles(mixedDir, 25)
	// Add more mod files with different names
	td.populateModFiles(mixedDir, 25, 1)
	// To exercise the walk algo we are creating a nested directory.
	td.populateExifFiles(nestedDir, 25)

	return td.root
}

func (td *testdir) getSkipRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	skipDir, _ := ioutil.TempDir(td.root, "skip")

	// First add exif files
	td.populateExifFiles(exifDir, 50)
	// Add files we want to skip.
	td.populateSkipFiles(skipDir, 25)

	return td.root
}

func (td *testdir) getBadRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	badDir, _ := ioutil.TempDir(td.root, "bad")

	// First add exif files
	td.populateExifFiles(exifDir, 50)

	// Taint the badDir so we can exercise the error paths
	err := os.Chmod(badDir, 0)
	if err != nil {
		td.t.Errorf("Chmod failed on %s with %s\n", badDir, err.Error())
	}

	td.numScanError++

	return td.root
}

func (td *testdir) getCollisionRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	collisionDir, _ := ioutil.TempDir(td.root, "collision")

	// First add exif files
	td.populateExifFiles(exifDir, 50)
	// Add collision files to first exif set using noExifPath
	td.populateCollisionFilenames(collisionDir, 25)

	return td.root
}

func (td *testdir) getDuplicateRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	duplicateDir, _ := ioutil.TempDir(td.root, "duplicate")

	// First add exif files
	td.populateExifFiles(exifDir, 50)
	// Add collision files to first exif set using noExifPath
	td.populateDuplicateFilenames(duplicateDir, exifPath, 25)

	return td.root
}

func newTestDir(t *testing.T, method int) *testdir {
	var td testdir

	td.fileNo = 0
	td.root, _ = ioutil.TempDir("", "root")
	td.t = t
	td.startTime = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.Local)
	td.method = method

	return &td
}
