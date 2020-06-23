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

Merge rerquirments
We want a src that forces to create a new directory
*/

const (
	exifPath    = "../data/with_exif.jpg"
	exifTimeStr = "2020:04:28 14:12:21"

	noExifPath     = "../data/no_exif.jpg"
	noRootExifPath = "../data/no_root_ifd.jpg"
	tifPath        = "../data/car.tif"
	skipPath       = "../README.md"
	nonesensePath  = "../gobofragggle"
	fileNoDefault  = 0
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

/*
// For debugging - commented out for lint

func listDir(root string) {
	_ = filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			fmt.Print(path + "\n")
			return nil
		})
}

*/

type testdir struct {
	fileNo      int
	fileNoStart int
	root        string
	t           *testing.T
	time        time.Time
	method      Method

	numExifError  int
	numData       int
	numTif        int
	numDuplicates int
	numSkipped    int
	numScanError  int
	numTimeSpread int
}

func (td *testdir) numTotal() int {
	return td.numData + td.numScanError + td.numSkipped
}

func (td *testdir) incrementTimeByMethod(delta int) {
	switch td.method {
	case MethodYear:
		td.time = td.time.AddDate(delta, 0, 0)
	case MethodMonth:
		td.time = td.time.AddDate(0, delta, 0)
	case MethodDay:
		td.time = td.time.AddDate(0, 0, delta)
	case MethodNone:
		return
	default:
		td.t.Fatalf("Invalid Method %s", td.method)
	}
	td.numTimeSpread++
}

// We need to have unique filenames often.
// Instead of worrying about if they are unique in each subdir we are using a
// testdir wide counter to ensure they are unique across the whole testdir
// instance.
func (td *testdir) buildFilename(path string) string {
	filename := filepath.Base(path)
	extension := filepath.Ext(filename)
	prefix := strings.TrimRight(filename, extension)

	filename = fmt.Sprintf("%s_%03d%s", prefix, td.fileNo, extension)
	td.fileNo++

	return filename
}

func (td *testdir) populateFiles(dir string, num int,
	contentsPath string, targetName string) {
	content, err := ioutil.ReadFile(contentsPath)
	if err != nil {
		td.t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		filename := td.buildFilename(targetName)
		targetPath := filepath.Join(dir, filename)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			td.t.Fatal(err)
		}
	}
}

// Generate 'num' files with exif date set to arg in the directory passed in
func (td *testdir) populateExifFiles(dir string, num int) {
	td.populateFiles(dir, num, exifPath, exifPath)
	td.numData += num
}

func (td *testdir) getTifRegex() string { return `.*\.tif$` }

func (td *testdir) populateTifFiles(dir string, num int) {
	td.populateFiles(dir, num, tifPath, tifPath)
	td.numData += num
	td.numTif += num
}

// The goal is to create filenames bases on zero with exif contents.
func (td *testdir) populateDuplicateFiles(dir string, path string, num int) {
	td.populateFiles(dir, num, exifPath, exifPath)
	td.numDuplicates += num
}

// This will have the same names as popualted exif files but different contents
// so should be handled as a collision.
func (td *testdir) populateCollisionFiles(dir string, num int) {
	// Note the contents will be different though. So we get files that
	// collide in name but not contents.
	td.populateFiles(dir, num, noExifPath, exifPath)
	td.numData += num
}

// Generate 'num' files that we will skip since they are not media.
func (td *testdir) populateSkipFiles(dir string, num int) {
	td.populateFiles(dir, num, skipPath, skipPath)
	td.numSkipped += num
}

func (td *testdir) populateNoExifFiles(dir string, num int) {
	td.populateFiles(dir, num, noExifPath, noExifPath)
	// These will not work in the Exif library
	td.numExifError += num
	// But with modtimes they will still be handled as data
	td.numData += num
}

// Set the modtimes that match the 'match' in 'dir' to a spread based on testdir
func (td *testdir) setModTimes(dir string, match string) {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		td.t.Fatal(err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		err = os.Chtimes(path, td.time, td.time)
		if err != nil {
			td.t.Fatal(err)
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

	mixedDir, _ := ioutil.TempDir(td.root, "mixed_exif")
	td.populateExifFiles(mixedDir, 25)
	td.populateNoExifFiles(mixedDir, 25)

	return td.root
}

func (td *testdir) buildCollisionWithRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "with_exif")
	td.populateCollisionFiles(exifDir, 50)

	nestedDir, _ := ioutil.TempDir(exifDir, "nested_exif")
	td.populateCollisionFiles(nestedDir, 25)

	noExifDir, _ := ioutil.TempDir(td.root, "no_exif")
	td.populateCollisionFiles(noExifDir, 25)

	mixedDir, _ := ioutil.TempDir(td.root, "mixed_exif")
	td.populateCollisionFiles(mixedDir, 25)

	return td.root
}

func (td *testdir) buildTifRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "dir_")
	td.populateNoExifFiles(exifDir, 25)
	td.populateTifFiles(exifDir, 125)

	return td.root
}

// We are building a root with no exif data to keep times straight.
// All modtimes.
func (td *testdir) buildTimeSpreadRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "with_exif")
	td.populateNoExifFiles(exifDir, 25)
	td.setModTimes(exifDir, noExifPath)

	td.incrementTimeByMethod(10)

	nestedDir, _ := ioutil.TempDir(exifDir, "nested_exif")
	td.populateNoExifFiles(nestedDir, 25)
	td.setModTimes(nestedDir, noExifPath)

	td.incrementTimeByMethod(10)

	noExifDir, _ := ioutil.TempDir(td.root, "no_exif")
	td.populateNoExifFiles(noExifDir, 25)
	td.setModTimes(noExifDir, noExifPath)

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

// this root has collisions within itself when sorted
func (td *testdir) buildCollisionWithinThisRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	collisionDir, _ := ioutil.TempDir(td.root, "collision")

	// First add exif files
	td.populateExifFiles(exifDir, 50)

	// What we are doing is setting our fileno counter back to the start
	// This will construct filenames that equal the ones created alredy.
	td.fileNo = td.fileNoStart

	// Add collision files to first exif set using noExifPath
	td.populateCollisionFiles(collisionDir, 25)

	return td.root
}

// This root has duplicates in this exact root
func (td *testdir) buildDuplicateWithinThisRoot() string {
	exifDir, _ := ioutil.TempDir(td.root, "exif")
	duplicateDir, _ := ioutil.TempDir(td.root, "duplicate")

	// First add exif files
	td.populateExifFiles(exifDir, 50)

	// What we are doing is setting our fileno counter back to the start
	// This will construct filenames that equal the ones created already.
	td.fileNo = td.fileNoStart

	// Add collision files to first exif set using noExifPath
	td.populateDuplicateFiles(duplicateDir, exifPath, 25)

	return td.root
}

func (td *testdir) buildSortedDir(src string, dst string, action Action) string {
	scanner := NewScanner()
	_ = scanner.ScanDir(src, ioutil.Discard)

	dst, _ = ioutil.TempDir("", dst)

	sorter, _ := NewSorter(scanner, td.method)
	_ = sorter.Transfer(dst, action, ioutil.Discard)

	return dst
}

func newTestDir(t *testing.T, method Method, fileNo int) *testdir {
	var td testdir

	td.fileNoStart = fileNo
	td.fileNo = fileNo
	td.root, _ = ioutil.TempDir("", "root")
	td.t = t
	td.time = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.Local)
	td.method = method

	return &td
}
