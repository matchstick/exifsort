package exifsort

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matchstick/exifsort/testdir"
)

func countFiles(t *testing.T, path string, correctCount int) {
	var count = 0

	_ = filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				count++
				return nil
			}

			if info.IsDir() {
				return nil
			}

			count++

			return nil
		})

	if count != correctCount {
		t.Errorf("File Count error for %s. Expected %d, got %d\n",
			path, correctCount, count)
	}
}

func testTransfer(t *testing.T, method int, action int) {
	src := testdir.NewTestDir(t)
	defer os.RemoveAll(src)

	scanner := NewScanner()
	_ = scanner.ScanDir(src, ioutil.Discard)

	const dst = "/tmp/dst"
	defer os.RemoveAll(dst)

	sorter, _ := NewSorter(scanner, method)

	err := sorter.Transfer(dst, action, ioutil.Discard)
	if err != nil {
		t.Errorf("Sort failed. Action: %d, Method: %d, Err: %s\n",
			action, method, err.Error())
	}

	countFiles(t, dst, testdir.CorrectNumValid)

	switch {
	case action == ActionCopy:
		countFiles(t, src, testdir.CorrectNumTotal)
	case action == ActionMove:
		leftovers := testdir.CorrectNumInvalid + testdir.CorrectNumSkipped
		countFiles(t, src, leftovers)
	default:
		t.Fatal("Unknown action")
	}
}

func TestSortDir(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		testTransfer(t, method, ActionCopy)
		testTransfer(t, method, ActionMove)
	}
}

func TestSortLoad(t *testing.T) {
	tmpPath := testdir.NewTestDir(t)
	defer os.RemoveAll(tmpPath)

	jsonDir, _ := ioutil.TempDir("", "jsonDir")
	defer os.RemoveAll(jsonDir)

	s := NewScanner()
	_ = s.ScanDir(tmpPath, ioutil.Discard)

	jsonPath := fmt.Sprintf("%s/%s", jsonDir, "scanned.json")

	err := s.Save(jsonPath)
	if err != nil {
		t.Errorf("Unexpected Error %s from Save\n", err.Error())
	}

	newScanner := NewScanner()

	err = newScanner.Load(jsonPath)
	if err != nil {
		t.Errorf("Unexpected Error %s from Load\n", err.Error())
	}

	if !cmp.Equal(s, newScanner) {
		t.Errorf("Saved and Loaded Scanner do not match\n")
	}
}
