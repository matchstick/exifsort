package exifsort

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matchstick/exifsort/testdir"
)

func countFiles(t *testing.T, path string, correctCount int, label string) {
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

	if !winOS() && count != correctCount {
		t.Errorf("File Count error for %s on %s. Expected %d, got %d\n",
			label, path, correctCount, count)
	}
}

func testTransfer(t *testing.T, method int, action int) error {
	src := testdir.NewTestDir(t)
	defer os.RemoveAll(src)

	scanner := NewScanner()
	_ = scanner.ScanDir(src, ioutil.Discard)

	dst, _ := ioutil.TempDir("", "sort_dst_")
	defer os.RemoveAll(dst)

	sorter, err := NewSorter(scanner, method)
	if err != nil {
		return err
	}

	err = sorter.Transfer(dst, action, ioutil.Discard)
	if err != nil {
		return err
	}

	countFiles(t, dst, testdir.NumData, "Dst Data")

	switch {
	case action == ActionCopy:
		// exifErrors counted twice as they also are added to data.
		copyCount := testdir.NumTotal - testdir.NumExifError
		countFiles(t, src, copyCount, "Src Copy")
	case action == ActionMove:
		leftovers := testdir.NumScanError + testdir.NumSkipped
		countFiles(t, src, leftovers, "Src Move")
	default:
		return &sortError{"Unknown action"}
	}

	return nil
}

func TestSortDir(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		err := testTransfer(t, method, ActionCopy)
		if err != nil {
			t.Fatalf("%s\n", err.Error())
		}

		err = testTransfer(t, method, ActionMove)
		if err != nil {
			t.Fatalf("%s\n", err.Error())
		}
	}
}

func TestBadSortMethod(t *testing.T) {
	const badMethod = 888

	err := testTransfer(t, badMethod, ActionMove)
	if err == nil {
		t.Fatalf("Expected error got success\n")
	}
}

func TestBadSortAction(t *testing.T) {
	const badAction = 888

	err := testTransfer(t, MethodYear, badAction)
	if err == nil {
		t.Fatalf("Expected error got success\n")
	}
}

func TestSortNoOutputDir(t *testing.T) {
	src := testdir.NewTestDir(t)
	defer os.RemoveAll(src)

	scanner := NewScanner()
	_ = scanner.ScanDir(src, ioutil.Discard)

	sorter, _ := NewSorter(scanner, MethodYear)

	err := sorter.Transfer("dst", ActionMove, ioutil.Discard)
	if err == nil {
		t.Fatalf("Unexpected Success\n")
	}

	if !strings.Contains(err.Error(), "No Output dir") {
		t.Fatalf("Unexpected msg (%s)\n", err.Error())
	}
}
