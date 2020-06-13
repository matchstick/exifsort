package exifsort

import (
	"errors"
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
* A directory with the same media names and sort times but different contents
  to create collisions.

*/

const (
	exifPath    = "../data/with_exif.jpg"
	exifTimeStr = "2020:04:28 14:12:21"

	noExifPath     = "../data/no_exif.jpg"
	noRootExifPath = "../data/no_root_ifd.jpg"
	skipPath       = "../README.md"
	nonesensePath  = "../gobofragggle"
)

func countFiles(t *testing.T, path string, correctCount int, label string) error {
	var count = 0

	_ = filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				count++
				return nil
			}

			if info.IsDir() {
				return nil
			}

			count++

			return nil
		})

	if count != correctCount {
		errStr := fmt.Sprintf("count error for %s on %s. "+
			"Expected %d got %d",
			label, path, correctCount, count)

		return errors.New(errStr)
	}

	return nil
}

type testdir struct {
	fileNo    int // We want files that are unique across an entire testdir instance
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
func (td *testdir) buildFilename(path string, counter *int) string {
	filename := filepath.Base(path)
	extension := filepath.Ext(filename)
	prefix := strings.TrimRight(filename, extension)

	filename = fmt.Sprintf("%s_%d%s", prefix, *counter, extension)
	*counter = (*counter) + 1

	return filename
}

func (td *testdir) populateFiles(dir string, num int, counter *int,
	contentsPath string, targetName string) {
	content, err := ioutil.ReadFile(contentsPath)
	if err != nil {
		td.t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		filename := td.buildFilename(targetName, counter)
		targetPath := filepath.Join(dir, filename)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			td.t.Fatal(err)
		}
	}
}

// Generate 'num' files with exif date set to arg in the directory passed in
func (td *testdir) populateExifFiles(dir string, num int) {
	td.populateFiles(dir, num, &td.fileNo, exifPath, exifPath)
	td.numData += num
}

// The goal is to create filenames bases on zero with exif contents.
func (td *testdir) populateDuplicateFilenames(dir string, path string, num int) {
	// What we are doing is setting our fileno counter back to 0.
	// THis will construct filenames that equal the ones created alredy.
	var count = 0

	td.populateFiles(dir, num, &count, exifPath, exifPath)
	td.numDuplicates += num
}

// This will have the same names as popualted exif files but different contents
// so should be handled as a collision.
func (td *testdir) populateCollisionFilenames(dir string, num int) {
	// What we are doing is setting our fileno counter back to 0.
	// This will construct filenames that equal the ones created alredy.
	var count = 0

	// Note the contents will be different though. So we get files that
	// collide in name but not contents.
	td.populateFiles(dir, num, &count, noExifPath, exifPath)
	td.numData += num
}

// Generate 'num' files that we will skip since they are not media.
func (td *testdir) populateSkipFiles(dir string, num int) {
	td.populateFiles(dir, num, &td.fileNo, skipPath, skipPath)
	td.numSkipped += num
}

func (td *testdir) populateNoExifFiles(dir string, num int) {
	td.populateFiles(dir, num, &td.fileNo, noExifPath, noExifPath)
	// THese will not work in the Exif library
	td.numExifError += num
	// But with modtimes they will still be handled as data
	td.numData += num
}

// Set the modtimes that match the 'match' in 'dir' to a spread based on testdir
func (td *testdir) setModTimes(dir string, match string,
	startTime time.Time, delta int) {
	time := startTime

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		td.t.Fatal(err)
	}

	for _, entry := range entries {
		filename := entry.Name()
		extension := filepath.Ext(filename)
		prefix := strings.TrimRight(filename, extension)

		// If not a matchjust skip it
		if !strings.Contains(prefix, match) {
			continue
		}

		// set time for file
		path := filepath.Join(dir, filename)

		err = os.Chtimes(path, time, time)
		if err != nil {
			td.t.Fatal(err)
		}

		if td.method != MethodNone {
			time = td.addTimeByMethod(time, delta)
		}
	}
}

func (td *testdir) buildRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "with_exif")
	td.populateExifFiles(exifDir, 50)

	nestedDir, _ := ioutil.TempDir(exifDir, "nested_exif")
	td.populateExifFiles(nestedDir, 25)

	noExifDir, _ := ioutil.TempDir(td.root, "no_exif")
	td.populateNoExifFiles(noExifDir, 25)
	td.setModTimes(noExifDir, noExifPath, td.startTime, 1)

	mixedDir, _ := ioutil.TempDir(td.root, "mixed_exif")
	td.populateExifFiles(mixedDir, 25)
	td.populateNoExifFiles(mixedDir, 25)
	td.setModTimes(mixedDir, noExifPath, td.startTime, 1)

	return td.root
}

func (td *testdir) buildSkipRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	skipDir, _ := ioutil.TempDir(td.root, "skip")

	// First add exif files
	td.populateExifFiles(exifDir, 50)
	// Add files we want to skip.
	td.populateSkipFiles(skipDir, 25)

	return td.root
}

func (td *testdir) buildBadRoot() string {
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

func (td *testdir) buildCollisionRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	collisionDir, _ := ioutil.TempDir(td.root, "collision")

	// First add exif files
	td.populateExifFiles(exifDir, 50)
	// Add collision files to first exif set using noExifPath
	td.populateCollisionFilenames(collisionDir, 25)

	return td.root
}

func (td *testdir) buildDuplicateRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	duplicateDir, _ := ioutil.TempDir(td.root, "duplicate")

	// First add exif files
	td.populateExifFiles(exifDir, 50)
	// Add collision files to first exif set using noExifPath
	td.populateDuplicateFilenames(duplicateDir, exifPath, 25)

	return td.root
}

func (td *testdir) buildSortedDir(src string, dst string, action int) string {
	scanner := NewScanner()
	_ = scanner.ScanDir(src, ioutil.Discard)

	dst, _ = ioutil.TempDir("", dst)

	sorter, _ := NewSorter(scanner, td.method)
	_ = sorter.Transfer(dst, action, ioutil.Discard)

	return dst
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
