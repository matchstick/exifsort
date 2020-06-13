package exifsort

import (
	"errors"
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
		td := newTestDir(t, method, fileNoDefault)

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
		td := newTestDir(t, method, fileNoDefault)

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

func testMerge(t *testing.T, method int, action int, dstFileNo int, dup bool) error {
	tdSrc := newTestDir(t, method, fileNoDefault)
	tdDst := newTestDir(t, method, fileNoDefault)

	// If we want to not have duplicate testdirs we need to modify the fileNo
	if dstFileNo != fileNoDefault {
		tdDst.setFileNo(dstFileNo)
	}

	src := tdSrc.buildRoot()
	dst := tdDst.buildRoot()

	// Copy files to two sorted directories that are identical
	fromDir := tdSrc.buildSortedDir(src, "fromDir_", ActionCopy)
	toDir := tdDst.buildSortedDir(dst, "toDir_", ActionCopy)

	// merge them
	err := Merge(fromDir, toDir, action, ioutil.Discard)
	if err != nil {
		return err
	}

	var leftOvers int

	switch action {
	case ActionCopy:
		// src dir should have all its media untouched
		leftOvers = tdSrc.numTotal()
	case ActionMove:
		// src dir should have all media merged and be empty
		leftOvers = 0
	default:
		return errors.New("unknown action")
	}

	err = countFiles(t, fromDir, leftOvers, "Src Dir")
	if err != nil {
		return err
	}

	var total int
	if dup {
		// Destination should have data only from dst as src is all duplicates
		total = tdDst.numData
	} else {
		// Destination should have all data from both sources
		total = tdDst.numData + tdSrc.numData
	}

	err = countFiles(t, toDir, total, "Target Dir")
	if err != nil {
		return err
	}

	defer os.RemoveAll(fromDir)
	defer os.RemoveAll(toDir)
	defer os.RemoveAll(src)

	return nil
}

func TestMergeGood(t *testing.T) {
	// By setting the fileNo so high the files will have different names
	// between tesdirs We are hoping that this number is just high enough
	// but as of this writing testdir has 150 files, and our countfiles
	// checking should protect us.
	fileNo := 10000

	for method := MethodYear; method < MethodNone; method++ {
		err := testMerge(t, method, ActionCopy, fileNo, false)
		if err != nil {
			t.Fatalf("Method %d, Action Copy Error: %s\n",
				method, err.Error())
		}
	}

	for method := MethodYear; method < MethodNone; method++ {
		err := testMerge(t, method, ActionMove, fileNo, false)
		if err != nil {
			t.Fatalf("Method %d, Action Move Error: %s\n",
				method, err.Error())
		}
	}
}

func TestMergeDuplicate(t *testing.T) {
	// By setting the fileNo to the default we ensure the dst directory will have
	// files with the same names as src and then get duplicates.
	fileNo := fileNoDefault

	for method := MethodYear; method < MethodNone; method++ {
		err := testMerge(t, method, ActionCopy, fileNo, true)
		if err != nil {
			t.Fatalf("Method %d, Action Copy Error: %s\n",
				method, err.Error())
		}
	}

	for method := MethodYear; method < MethodNone; method++ {
		err := testMerge(t, method, ActionMove, fileNo, true)
		if err != nil {
			t.Fatalf("Method %d, Action Move Error: %s\n",
				method, err.Error())
		}
	}
}
