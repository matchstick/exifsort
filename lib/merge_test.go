package exifsort

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

/*
Think of all test cases:
* Merge with bad named path to make invalid tree.
* Merge with collision files
* Merge with different files
*/
func TestMergeCheckGood(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		td := newTestDir(t, method)

		src := td.buildRoot()
		defer os.RemoveAll(src)

		dst := td.buildSortedDir(src, "dst", ActionCopy)

		err := MergeCheck(dst, method)
		if err != nil {
			t.Errorf("Err %s, method %d\n", err.Error(), method)
		}

		os.RemoveAll(dst)
	}
}

func TestMergeCheckBad(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		td := newTestDir(t, method)

		src := td.buildRoot()
		defer os.RemoveAll(src)

		dst := td.buildSortedDir(src, "dst", ActionCopy)

		badFilePath := filepath.Join(dst, "testfile.txt")
		message := []byte("Hello, Gophers!")

		_ = ioutil.WriteFile(badFilePath, message, 0600)

		err := MergeCheck(dst, method)
		if err == nil {
			t.Errorf("Unexpected Success method %d\n", method)
		}

		os.RemoveAll(badFilePath)
		os.RemoveAll(dst)
	}
}

func TestMergeDuplicate(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		td := newTestDir(t, method)

		src := td.buildRoot()

		// COpy files to two sorted directories that are identical
		fromDir := td.buildSortedDir(src, "fromDir_", ActionCopy)
		toDir := td.buildSortedDir(src, "toDir_", ActionCopy)

		// merge them
		err := Merge(fromDir, toDir, ActionMove, ioutil.Discard)
		if err != nil {
			t.Errorf("Error %s", err.Error())
		}

		// Source should have all its files removed as they are all
		// duplicates
		err = countFiles(t, fromDir, 0, "Src Dir")
		if err != nil {
			t.Errorf("Error %s", err.Error())
		}

		// Destination should have all its original contents
		err = countFiles(t, toDir, td.numTotal(), "Target Dir")
		if err != nil {
			t.Errorf("Error %s", err.Error())
		}

		os.RemoveAll(fromDir)
		os.RemoveAll(toDir)
		os.RemoveAll(src)
	}
}
