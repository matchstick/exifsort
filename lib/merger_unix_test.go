// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package exifsort

import (
	"testing"
)

func TestMergePathGood(t *testing.T) {
	var goodInput = map[string]Method{
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

	for input, goodMethod := range goodInput {
		method := mergePathValid(root, input)
		if method != goodMethod {
			t.Errorf("Input %s, %s != %s\n", input, goodMethod, method)
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
		method := mergePathValid(root, input)
		if method != MethodYear {
			t.Errorf("Expected year but got %s on %s\n",
				method, input)
		}
	}
}

func TestMergePathBad(t *testing.T) {
	var badInput = []string{
		"gobo",
		"gobo/",
		"gobo/0/m.jpg",
		"gobo/2010_/m.jpg",
		"gobo/00000/m.jpg",
		"gobo/10000/m.jpg",
		"gobo/bad/m.jpg",
		"gobo/2010/2010_00/m.jpg",
		"gobo/2010/2010_13/m.jpg",
		"gobo/2010/2010_32/m.jpg",
		"gobo/2010/2010_gobo/m.jpg",
		"gobo/2010/2010_02/2010__01/m.jpg",
		"gobo/2010/2010_02/2010_gobo_11/m.jpg",
		"gobo/2010/2010_02/2010_02_gobo/m.jpg",
		"gobo/2010/2010_02/2010_02_00/m.jpg",
		"gobo/2010/2010_02/2010_02_32/m.jpg",
	}

	for _, input := range badInput {
		method := mergePathValid("gobo", input)
		if method != MethodNone {
			t.Errorf("Input of %s yields method %s\n", input, method)
		}
	}
}
