package exifsort

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func testOSPopulateFile(dir string, filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	dst := filepath.Join(dir, filepath.Base(filename))

	err = ioutil.WriteFile(dst, content, 0600)
	if err != nil {
		return err
	}

	return nil
}

func TestOSMoveFile(t *testing.T) {
	t.Parallel()

	testDir, _ := ioutil.TempDir("", "moveDir_")
	defer os.RemoveAll(testDir)

	err := testOSPopulateFile(testDir, exifPath)
	if err != nil {
		t.Fatalf("Cannot write %s %s\n", exifPath, err.Error())
	}

	err = testOSPopulateFile(testDir, noExifPath)
	if err != nil {
		t.Fatalf("Cannot write %s %s\n", exifPath, err.Error())
	}

	src := filepath.Join(testDir, filepath.Base(exifPath))
	dst := filepath.Join(testDir, filepath.Base(noExifPath))

	err = moveFile(src, dst)
	if err == nil {
		t.Fatalf("We clobbered a file.\n")
	}

	dst = filepath.Join(testDir, "nowhere.jpg")

	err = moveFile(src, dst)
	if err != nil {
		t.Fatalf("We failed to move file %s.\n", err.Error())
	}

	if !exists(dst) {
		t.Fatalf("dst does not exist.\n")
	}

	if exists(src) {
		t.Fatalf("src does exist.\n")
	}
}

func TestOSCopyFile(t *testing.T) {
	t.Parallel()

	testDir, _ := ioutil.TempDir("", "moveDir_")
	defer os.RemoveAll(testDir)

	err := testOSPopulateFile(testDir, exifPath)
	if err != nil {
		t.Fatalf("Cannot write %s %s\n", exifPath, err.Error())
	}

	err = testOSPopulateFile(testDir, noExifPath)
	if err != nil {
		t.Fatalf("Cannot write %s %s\n", exifPath, err.Error())
	}

	src := filepath.Join(testDir, filepath.Base(exifPath))
	dst := filepath.Join(testDir, filepath.Base(noExifPath))

	err = copyFile(src, dst)
	if err == nil {
		t.Fatalf("We clobbered a file.\n")
	}

	dst = filepath.Join(testDir, "nowhere.jpg")

	err = copyFile(src, dst)
	if err != nil {
		t.Fatalf("We failed to copy file %s.\n", err.Error())
	}

	if !exists(dst) {
		t.Fatalf("dst does not exist.\n")
	}

	if !exists(src) {
		t.Fatalf("src does not exist.\n")
	}
}
