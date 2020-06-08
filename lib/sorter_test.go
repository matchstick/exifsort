package exifsort

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func countFiles(t *testing.T, path string, correctCount int, label string) error {
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
		errStr := fmt.Sprintf("count error for %s on %s. Expected %d got %d",
			label, path, correctCount, count)
		return &sortError{errStr}
	}

	return nil
}

func testTransfer(t *testing.T, td *testdir, method int, action int) error {
	scanner := NewScanner()
	_ = scanner.ScanDir(td.root, ioutil.Discard)

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

	err = countFiles(t, dst, td.numData, "Dst Data")
	if err != nil {
		return err
	}

	switch {
	case action == ActionCopy:
		err := countFiles(t, td.root, td.numTotal(), "Src Copy")
		if err != nil {
			return err
		}

	case action == ActionMove:
		leftovers := td.numScanError + td.numSkipped

		err := countFiles(t, td.root, leftovers, "Src Move")
		if err != nil {
			fmt.Printf("hey 4\n")
			return err
		}
	default:
		return &sortError{"Unknown action"}
	}

	return nil
}

func TestSortDir(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		td := newTestDir(t, method)

		src := td.getRoot()
		defer os.RemoveAll(src)

		err := testTransfer(t, td, method, ActionCopy)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}

		err = testTransfer(t, td, method, ActionMove)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}
	}
}

func TestSortDuplicates(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		td := newTestDir(t, method)

		src := td.getDuplicateRoot()
		defer os.RemoveAll(src)

		err := testTransfer(t, td, method, ActionCopy)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}

		err = testTransfer(t, td, method, ActionMove)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}
	}
}

func TestSortCollisions(t *testing.T) {
	for method := MethodYear; method < MethodNone; method++ {
		td := newTestDir(t, method)

		src := td.getCollisionRoot()
		defer os.RemoveAll(src)

		err := testTransfer(t, td, method, ActionCopy)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}

		err = testTransfer(t, td, method, ActionMove)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}
	}
}

func TestBadSortMethod(t *testing.T) {
	const badMethod = 888

	td := newTestDir(t, badMethod)

	err := testTransfer(t, td, badMethod, ActionMove)
	if err == nil {
		t.Fatalf("Expected error got success\n")
	}
}

func TestBadSortAction(t *testing.T) {
	const badAction = 888

	td := newTestDir(t, MethodNone)

	err := testTransfer(t, td, MethodYear, badAction)
	if err == nil {
		t.Fatalf("Expected error got success\n")
	}
}

func TestSortNoOutputDir(t *testing.T) {
	td := newTestDir(t, MethodNone)

	src := td.getRoot()
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
