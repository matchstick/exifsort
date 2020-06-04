package exifsort

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/matchstick/exifsort/testdir"
)

func TestMergePathGood(t *testing.T) {
	var goodInput = map[string]int{
		"gobo/1920/m.jpg":                    MethodYear,
		"gobo/2010/m.jpg":                    MethodYear,
		"gobo/2020/m.jpg":                    MethodYear,
		"gobo/9999/m.jpg":                    MethodYear,
		"gobo/2010/2010_02/m.jpg":            MethodMonth,
		"gobo/2010/2010_09/m.jpg":            MethodMonth,
		"gobo/2010/2010_10/m.jpg":            MethodMonth,
		"gobo/2010/2010_12/m.jpg":            MethodMonth,
		"gobo/2020/2020_04/m.jpg":            MethodMonth,
		"gobo/2010/2010_02/2010_02_01/m.jpg": MethodDay,
		"gobo/2010/2010_02/2010_02_11/m.jpg": MethodDay,
		"gobo/2010/2010_02/2010_02_19/m.jpg": MethodDay,
		"gobo/2010/2010_02/2010_02_20/m.jpg": MethodDay,
		"gobo/2010/2010_02/2010_02_29/m.jpg": MethodDay,
		"gobo/2010/2010_02/2010_02_30/m.jpg": MethodDay,
		"gobo/2010/2010_02/2010_02_31/m.jpg": MethodDay,
		"gobo/2020/2020_04/2020_04_28/m.jpg": MethodDay,
	}

	root := "gobo"

	for input, method := range goodInput {
		if !mergePathValid(root, input, method) {
			t.Errorf("Expected Match with %s on method %d\n",
				input, method)
		}
	}
}

func TestMergePathRoots(t *testing.T) {
	var goodRootInput = map[string]string{
		"gobo/gobo":      "gobo/gobo/1920/m.jpg",
		"gobo/dogo/gobo": "gobo/dogo/gobo/1920/m.jpg",
		"gobo/":          "gobo/1920/m.jpg",
		"../gobo/":       "../gobo/1920/m.jpg",
	}

	for root, input := range goodRootInput {
		if !mergePathValid(root, input, MethodYear) {
			t.Errorf("Expected Match with %s on method %d\n",
				input, MethodYear)
		}
	}
}

func TestMergePathBad(t *testing.T) {
	var goodInput = map[string]int{
		"gobo":                                 MethodYear,
		"gobo/":                                MethodYear,
		"gobo/0/m.jpg":                         MethodYear,
		"gobo/2010_/m.jpg":                     MethodYear,
		"gobo/00000/m.jpg":                     MethodYear,
		"gobo/10000/m.jpg":                     MethodYear,
		"gobo/bad/m.jpg":                       MethodYear,
		"gobo/2010/2010_00/m.jpg":              MethodMonth,
		"gobo/2010/2010_13/m.jpg":              MethodMonth,
		"gobo/2010/2010_32/m.jpg":              MethodMonth,
		"gobo/2010/2010_gobo/m.jpg":            MethodMonth,
		"gobo/2010/2010_02/2010__01/m.jpg":     MethodDay,
		"gobo/2010/2010_02/2010_gobo_11/m.jpg": MethodDay,
		"gobo/2010/2010_02/2010_02_gobo/m.jpg": MethodDay,
		"gobo/2010/2010_02/2010_02_00/m.jpg":   MethodDay,
		"gobo/2010/2010_02/2010_02_32/m.jpg":   MethodDay,
	}

	for input, method := range goodInput {
		if mergePathValid("gobo", input, method) {
			t.Errorf("Unexpected Match with %s on method %d\n",
				input, method)
		}
	}
}

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
