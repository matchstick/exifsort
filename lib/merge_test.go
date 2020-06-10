package exifsort

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeCheckGood(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		td := newTestDir(t, method)

		src := td.buildRoot()
		defer os.RemoveAll(src)

		dst := td.buildSortedDir(src, "dst", ActionCopy)

		err := MergeCheck(dst, method)
		if err != nil {
			t.Errorf("Err %s, method %d\n",
				err.Error(), method)
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
		defer os.RemoveAll(src)

		dst1 := td.buildSortedDir(src, "dst1_", ActionCopy)
		dst2 := td.buildSortedDir(src, "dst2_", ActionCopy)

		err := Merge(dst1, dst2, ActionMove, ioutil.Discard)
		if err != nil {
			t.Errorf("Error %s", err.Error())
		}

		os.RemoveAll(dst1)
		os.RemoveAll(dst2)
	}
}
