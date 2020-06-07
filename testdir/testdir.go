package testdir

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	ExifPath    = "../data/with_exif.jpg"
	ExifTimeStr = "2020:04:28 14:12:21"

	NoExifPath     = "../data/no_exif.jpg"
	NoRootExifPath = "../data/no_root_ifd.jpg"

	SkipPath      = "../README.md"
	NonesensePath = "../gobofragggle"

	NumExifError = 50
	NumData      = 150
	NumSkipped   = 25
	NumScanError = 1
	NumTotal     = 176
)

type testdir struct {
	fileNo int
	root   string
	t      *testing.T
}

func (td *testdir) stampFileNo(path string) string {
	basename := filepath.Base(path)
	pieces := strings.Split(basename, ".")
	newPath := fmt.Sprintf("%s_%d.%s", pieces[0], td.fileNo, pieces[1])
	td.fileNo++

	return newPath
}

func (td *testdir) populateExifDir(dir string, readPath string, num int) {
	content, err := ioutil.ReadFile(readPath)
	if err != nil {
		td.t.Fatal(err)
	}

	for i := 0; i < num; i++ {
		newBase := td.stampFileNo(readPath)
		targetPath := fmt.Sprintf("%s/%s", dir, newBase)

		err := ioutil.WriteFile(targetPath, content, 0600)
		if err != nil {
			td.t.Fatal(err)
		}
	}
}

// Returns the path to the root of a directory full of files and nested
// structures. This is only intended for test code. Some of the media has
// exifdata some does not, some are not even media files. All of the files and
// directories were created as golang tmp files or directories.
//
// The bult directory even has a subdir with perms 0 for testing errorn
// handling.
func NewTestDir(t *testing.T) string {
	var td testdir
	td.fileNo = 0
	td.root, _ = ioutil.TempDir("", "root")
	td.t = t

	exifDir, _ := ioutil.TempDir(td.root, "with_exif")
	badDir, _ := ioutil.TempDir(td.root, "badPerms")
	skipDir, _ := ioutil.TempDir(td.root, "skip")
	nestedDir, _ := ioutil.TempDir(exifDir, "nested_exif")
	noExifDir, _ := ioutil.TempDir(td.root, "no_exif")
	mixedDir, _ := ioutil.TempDir(td.root, "mixed_exif")

	td.populateExifDir(exifDir, ExifPath, 50)
	td.populateExifDir(noExifDir, NoExifPath, 25)
	td.populateExifDir(mixedDir, ExifPath, 25)
	td.populateExifDir(mixedDir, NoExifPath, 25)
	td.populateExifDir(nestedDir, ExifPath, 25)
	td.populateExifDir(skipDir, SkipPath, 25)

	err := os.Chmod(badDir, 0)
	if err != nil {
		td.t.Errorf("Chmod failed on %s with %s\n",
			badDir, err.Error())
	}

	return td.root
}
