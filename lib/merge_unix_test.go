// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package exifsort

import (
	"testing"
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

func TestMergeFilter(t *testing.T) {
	for method := range MethodMap() {
		err := testMergeFilter(t, method, ActionCopy)
		if err != nil {
			t.Fatalf("Method %d, Action Copy Error: %s\n",
				method, err.Error())
		}
	}

	for method := range MethodMap() {
		err := testMergeFilter(t, method, ActionMove)
		if err != nil {
			t.Fatalf("Method %d, Action Move Error: %s\n",
				method, err.Error())
		}
	}
}
