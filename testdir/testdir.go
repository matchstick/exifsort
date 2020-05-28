package testdir

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	ExifPath          = "../data/with_exif.jpg"
	NoExifPath        = "../data/no_exif.jpg"
	SkipPath          = "../README.md"
	NoRootExifPath    = "../data/no_root_ifd.jpg"
	NonesensePath     = "../gobofragggle"
	CorrectNumInvalid = 51
	CorrectNumValid   = 100
	CorrectNumSkipped = 25
	CorrectNumTotal   = 176
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

func (td *testdir) setDirPerms(dirPath string, perms os.FileMode) {
	infos, _ := ioutil.ReadDir(dirPath)
	for _, info := range infos {
		targetPath := fmt.Sprintf("%s/%s", dirPath, info.Name())

		err := os.Chmod(targetPath, perms)
		if err != nil {
			td.t.Errorf("Chmod failed on %s with %s\n", info.Name(), err.Error())
		}
	}

	err := os.Chmod(dirPath, perms)
	if err != nil {
		td.t.Errorf("Chmod failed on %s with %s\n", dirPath, err.Error())
	}
}

func TmpDir(parent string, name string) string {
	newDir, err := ioutil.TempDir(parent, name)
	if err != nil {
		panic("cannot create tmp dir")
	}

	return newDir
}

/*
	Root
	-with_exif // valid exif
	  -nested_exif // nested dir with valid exit
	-no_exif // no exif
	-mixed_exif // mix of both
*/
func NewTestDir(t *testing.T) string {
	var td testdir
	td.fileNo = 0
	td.root = TmpDir("", "root")
	td.t = t

	exifDir := TmpDir(td.root, "with_exif")
	badDir := TmpDir(td.root, "badPerms")
	skipDir := TmpDir(td.root, "skip")
	nestedDir := TmpDir(exifDir, "nested_exif")
	noExifDir := TmpDir(td.root, "no_exif")
	mixedDir := TmpDir(td.root, "mixed_exif")

	td.populateExifDir(exifDir, ExifPath, 50)
	td.populateExifDir(noExifDir, NoExifPath, 25)
	td.populateExifDir(mixedDir, ExifPath, 25)
	td.populateExifDir(mixedDir, NoExifPath, 25)
	td.populateExifDir(nestedDir, ExifPath, 25)
	td.populateExifDir(skipDir, SkipPath, 25)

	td.populateExifDir(badDir, ExifPath, 25)
	td.setDirPerms(badDir, 0)

	return td.root
}
