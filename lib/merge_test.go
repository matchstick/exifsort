package exifsort

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/matchstick/exifsort/testdir"
)

func TestMergeCheckGood(t *testing.T) {
	src := testdir.NewTestDir(t)
	defer os.RemoveAll(src)

	scanner := NewScanner()
	_ = scanner.ScanDir(src, ioutil.Discard)

	for method := MethodYear; method < MethodNone; method++ {
		dst, _ := ioutil.TempDir("", "sort_dst_")

		sorter, _ := NewSorter(scanner, method)
		_ = sorter.Transfer(dst, ActionCopy, ioutil.Discard)

		err := MergeCheck(dst, method)
		if err != nil {
			t.Errorf("MergeCheck err of Sorted Dir %s, method %d\n",
				err.Error(), method)
		}

		os.RemoveAll(dst)
	}
}
